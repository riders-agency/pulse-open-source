package office

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/errs"
	"riders.agency/pulse/api/internal/platform/uow"
)

var ErrConflict = errors.New("office_conflict")

type Service struct{ P *pgxpool.Pool }

func (s *Service) authorize(ctx context.Context, pid, uid uuid.UUID) error {
	if err := authz.RequireProjectProcess(ctx, s.P, pid); err != nil {
		return err
	}
	var allowed bool
	err := s.P.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pulse_read.projects_v1 p JOIN pulse_read.companies_v1 c USING(company_id)
 JOIN pulse_read.company_memberships_v1 m USING(company_id) JOIN pulse_read.users_v1 u ON u.user_id=m.user_id
 WHERE p.project_id=$1 AND m.user_id=$2 AND u.email::text LIKE 'guest-%@demo.example.test'
 AND c.slug::text LIKE 'demo-%' AND m.role='owner')`, pid, uid).Scan(&allowed)
	if err != nil {
		return err
	}
	if !allowed {
		return errs.ErrForbidden
	}
	return nil
}

func (s *Service) Ensure(ctx context.Context, pid, uid uuid.UUID) (map[string]any, error) {
	if err := s.authorize(ctx, pid, uid); err != nil {
		return nil, err
	}
	err := uow.Run(ctx, s.P, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, pid.String()); err != nil {
			return err
		}
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM workflow_office_workspaces WHERE project_id=$1)`, pid).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return nil
		}
		wid, vid := uuid.New(), uuid.New()
		def, graph := Definition()
		_, err := tx.Exec(ctx, `INSERT INTO workflows(workflow_id,project_id,name,slug,status,draft_definition_json,draft_graph_json,created_by_user_id)
  VALUES($1,$2,'Входящая почта → ИИ-отделы',$3,'draft',$4,$5,$6)`, wid, pid, WorkflowSlug, def, graph, uid)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO workflow_versions(workflow_version_id,project_id,workflow_id,version_no,status,definition_json,graph_json,created_by_user_id,description)
  VALUES($1,$2,$3,1,'published',$4,$5,$6,'Почта каждые 30 секунд, ИИ-отделы и подтверждение человеком')`, vid, pid, wid, def, graph, uid)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `UPDATE workflows SET status='active',current_workflow_version_id=$2 WHERE workflow_id=$1`, wid, vid); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `INSERT INTO workflow_office_workspaces(project_id,workflow_id,owner_user_id) VALUES($1,$2,$3)`, pid, wid, uid); err != nil {
			return err
		}
		return insertLetter(ctx, tx, pid, uuid.New())
	})
	if err != nil {
		return nil, err
	}
	return s.State(ctx, pid, uid)
}

func (s *Service) Deliver(ctx context.Context, pid, uid, key uuid.UUID) (map[string]any, error) {
	if err := s.authorize(ctx, pid, uid); err != nil {
		return nil, err
	}
	if key == uuid.Nil {
		return nil, ErrConflict
	}
	err := uow.Run(ctx, s.P, func(ctx context.Context, tx pgx.Tx) error {
		var id uuid.UUID
		if err := tx.QueryRow(ctx, `SELECT project_id FROM workflow_office_workspaces WHERE project_id=$1 FOR UPDATE`, pid).Scan(&id); err != nil {
			return err
		}
		var count int
		if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM workflow_office_messages WHERE project_id=$1 AND workflow_run_id IS NULL`, pid).Scan(&count); err != nil {
			return err
		}
		if count >= 30 {
			return ErrConflict
		}
		return insertLetter(ctx, tx, pid, key)
	})
	if err != nil {
		return nil, err
	}
	return s.State(ctx, pid, uid)
}

func insertLetter(ctx context.Context, tx pgx.Tx, pid, key uuid.UUID) error {
	_, err := tx.Exec(ctx, `INSERT INTO workflow_office_messages(project_id,delivery_key,sender,subject,body,amount_minor)
 VALUES($1,$2,'Клиент демо <client@example.test>','Добавьте услугу на сайт и подготовьте счёт',
 'Здравствуйте! Добавьте на сайт услугу аудита интерфейса: изучение пути клиента, поиск неудобных мест и рекомендации. Подготовьте счёт для ООО «Демо» по согласованному заказу. Спасибо!',4500000)
 ON CONFLICT(project_id,delivery_key) DO NOTHING`, pid, key)
	return err
}

func (s *Service) State(ctx context.Context, pid, uid uuid.UUID) (map[string]any, error) {
	if err := s.authorize(ctx, pid, uid); err != nil {
		return nil, err
	}
	var raw []byte
	err := s.P.QueryRow(ctx, `SELECT jsonb_build_object('workflow_id',o.workflow_id,'enabled',o.enabled,'interval_seconds',30,'retained_runs',30,
 'last_poll_at',o.last_poll_at,'next_poll_at',o.next_poll_at,'poll_count',o.poll_count,
 'pending_mail',(SELECT count(*) FROM workflow_office_messages m WHERE m.project_id=o.project_id AND m.workflow_run_id IS NULL),
 'waiting',(SELECT count(*) FROM pulse_read.workflow_runs_v1 r WHERE r.workflow_id=o.workflow_id AND r.status='waiting'),
 'artifacts',COALESCE((SELECT jsonb_agg(to_jsonb(a) ORDER BY a.created_at DESC) FROM workflow_office_artifacts a WHERE a.project_id=o.project_id),'[]'))
 FROM workflow_office_workspaces o WHERE o.project_id=$1`, pid).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return map[string]any{"configured": false}, nil
	}
	if err != nil {
		return nil, err
	}
	var value map[string]any
	err = json.Unmarshal(raw, &value)
	value["configured"] = true
	return value, err
}

func (s *Service) SetEnabled(ctx context.Context, pid, uid uuid.UUID, enabled bool) (map[string]any, error) {
	if err := s.authorize(ctx, pid, uid); err != nil {
		return nil, err
	}
	_, err := s.P.Exec(ctx, `UPDATE workflow_office_workspaces SET enabled=$2,next_poll_at=now()+interval '30 seconds' WHERE project_id=$1`, pid, enabled)
	if err != nil {
		return nil, err
	}
	return s.State(ctx, pid, uid)
}
