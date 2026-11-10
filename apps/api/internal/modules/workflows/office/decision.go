package office

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	runcommands "riders.agency/pulse/api/internal/modules/runs/commands"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (s *Service) Decide(ctx context.Context, pid, uid, rid uuid.UUID, approved bool) (map[string]any, error) {
	if err := s.authorize(ctx, pid, uid); err != nil {
		return nil, err
	}
	err := uow.Run(ctx, s.P, func(ctx context.Context, tx pgx.Tx) error {
		var linked uuid.UUID
		err := tx.QueryRow(ctx, `SELECT workflow_run_id FROM workflow_office_messages WHERE workflow_run_id=$1 AND project_id=$2`, rid, pid).Scan(&linked)
		if err != nil {
			return err
		}
		if err := runcommands.LockWaiting(ctx, tx, pid, rid, "office_approval_required"); err != nil {
			return ErrConflict
		}
		if _, err = tx.Exec(ctx, `INSERT INTO workflow_office_decisions(workflow_run_id,approved,decided_by_user_id) VALUES($1,$2,$3)`, rid, approved, uid); err != nil {
			return err
		}
		return runcommands.ResumeAfterDecision(ctx, tx, rid, uid, approved)
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "approved": approved}, nil
}
