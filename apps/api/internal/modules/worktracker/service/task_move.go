package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	workrepo "riders.agency/pulse/api/internal/modules/worktracker/repository"
)

func (c *Commands) MoveTask(ctx context.Context, userID uuid.UUID, in workrepo.MoveTaskInput) (*workrepo.Task, error) {
	if c == nil || c.WorkTracker == nil || c.WorkTracker.P == nil {
		return nil, workrepo.ErrNotFound
	}
	access, _, err := c.WorkTracker.AccessForTask(ctx, in.TaskID, userID)
	if err != nil {
		return nil, err
	}
	current, err := c.WorkTracker.TaskRow(ctx, in.TaskID)
	if err != nil {
		return nil, err
	}
	if current.StatusID == in.StatusID.String() {
		return current, nil
	}
	rule, err := c.transitionRule(ctx, uuid.MustParse(current.StatusID), in.StatusID)
	if err != nil {
		if err != workrepo.ErrNotFound || !access.CanManageProcesses {
			return nil, err
		}
		rule = TransitionRule{}
	}
	if err := ValidateMovePolicy(*access, *current, rule, in.Comment); err != nil {
		return nil, err
	}

	tx, err := c.WorkTracker.P.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	clientVisible := current.ClientVisible || rule.Automation.SetClientVisible
	tag, err := tx.Exec(ctx, `
UPDATE tasks
   SET task_workflow_status_id = $2,
       rank = $3,
       position = $4,
       client_visible = $5,
       revision = revision + 1,
       completed_at = CASE
         WHEN (SELECT is_done FROM task_workflow_statuses WHERE task_workflow_status_id = $2) THEN COALESCE(completed_at, now())
         ELSE NULL
       END
 WHERE task_id = $1
   AND revision = $6`, in.TaskID.String(), in.StatusID.String(), in.Rank, int(in.Rank), clientVisible, expectedRevision(current.Revision, in.ExpectedRevision))
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, workrepo.ErrConflict
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO task_status_history (task_id, from_status_id, to_status_id, changed_by_user_id)
VALUES ($1,$2,$3,$4)`, in.TaskID.String(), current.StatusID, in.StatusID.String(), userID.String()); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Comment) != "" {
		if _, err := tx.Exec(ctx, `
INSERT INTO task_comments (task_id, company_id, author_user_id, body, visibility)
VALUES ($1,$2,$3,$4,'internal')`, in.TaskID.String(), access.CompanyID.String(), userID.String(), strings.TrimSpace(in.Comment)); err != nil {
			return nil, err
		}
	}
	if err := insertActivity(ctx, tx, access.CompanyID, in.TaskID, userID, "task.status_changed", map[string]any{
		"from_status_id": current.StatusID,
		"to_status_id":   in.StatusID.String(),
	}); err != nil {
		return nil, err
	}
	if c.WorkTracker.TaskEvents != nil {
		if err := c.WorkTracker.TaskEvents.EnqueueTaskModuleTriggers(ctx, tx, access.CompanyID, access.ProjectID, uuid.MustParse(current.WorkflowID), in.TaskID, "task.status_changed", map[string]any{
			"from_status_id": current.StatusID,
			"to_status_id":   in.StatusID.String(),
		}); err != nil {
			return nil, err
		}
	}
	if rule.Automation.StopTimer {
		if err := stopTaskTimersFromTransition(ctx, tx, in.TaskID); err != nil {
			return nil, err
		}
	}
	if rule.Automation.StartTimer {
		if _, err := tx.Exec(ctx, `
INSERT INTO timer_sessions (timer_session_id, company_id, task_id, user_id, state)
VALUES ($1,$2,$3,$4,'running')`, uuid.NewString(), access.CompanyID.String(), in.TaskID.String(), userID.String()); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return c.WorkTracker.TaskRow(ctx, in.TaskID)
}

func (c *Commands) transitionRule(ctx context.Context, fromStatusID, toStatusID uuid.UUID) (TransitionRule, error) {
	row := c.WorkTracker.P.QueryRow(ctx, `
SELECT allowed_roles, required_fields, automation_rules
  FROM task_workflow_transitions
 WHERE from_status_id = $1
   AND to_status_id = $2
 ORDER BY position
 LIMIT 1`, fromStatusID.String(), toStatusID.String())
	var rule TransitionRule
	var raw []byte
	if err := row.Scan(&rule.AllowedRoles, &rule.RequiredFields, &raw); err != nil {
		if err == pgx.ErrNoRows {
			return rule, workrepo.ErrNotFound
		}
		return rule, err
	}
	_ = json.Unmarshal(raw, &rule.Automation)
	return rule, nil
}

func insertActivity(ctx context.Context, tx pgx.Tx, companyID, taskID, userID uuid.UUID, eventType string, payload map[string]any) error {
	raw, _ := json.Marshal(payload)
	_, err := tx.Exec(ctx, `
INSERT INTO task_activity_events (company_id, task_id, actor_user_id, event_type, payload_json)
VALUES ($1,$2,$3,$4,$5::jsonb)`, companyID.String(), taskID.String(), userID.String(), eventType, string(raw))
	return err
}

func stopTaskTimersFromTransition(ctx context.Context, tx pgx.Tx, taskID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
WITH active AS (
  SELECT ts.timer_session_id,
         ts.company_id,
         ts.task_id,
         t.project_id,
         t.client_id,
         ts.user_id,
         ts.started_at,
         t.task_workflow_status_id,
         GREATEST(1, FLOOR(EXTRACT(EPOCH FROM (now() - ts.started_at)) / 60)::int) AS duration_minutes
    FROM timer_sessions ts
    JOIN tasks t ON t.task_id = ts.task_id
   WHERE ts.task_id = $1
     AND ts.state IN ('running','paused')
   FOR UPDATE
),
stopped AS (
  UPDATE timer_sessions ts
     SET state = 'stopped',
         stopped_at = now(),
         duration_minutes_cached = active.duration_minutes,
         updated_at = now()
    FROM active
   WHERE ts.timer_session_id = active.timer_session_id
   RETURNING active.company_id,
             active.task_id,
             active.project_id,
             active.client_id,
             active.user_id,
             active.started_at,
             active.task_workflow_status_id,
             active.duration_minutes
)
INSERT INTO time_entries (
  time_entry_id, company_id, task_id, project_id, client_id, user_id, entry_date,
  started_at, ended_at, duration_minutes, description, billable, source, task_workflow_status_id
)
SELECT gen_random_uuid(), company_id, task_id, project_id, client_id, user_id, CURRENT_DATE,
       started_at, now(), duration_minutes, 'Auto-stopped by workflow transition',
       true, 'status_auto', task_workflow_status_id
  FROM stopped`, taskID.String())
	if err != nil {
		return err
	}
	return refreshSpent(ctx, tx, taskID)
}

func refreshSpent(ctx context.Context, tx pgx.Tx, taskID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
UPDATE tasks t
   SET spent_minutes_cached = COALESCE((
         SELECT SUM(duration_minutes)::int FROM time_entries te
          WHERE te.task_id = t.task_id
       ), 0),
       remaining_minutes = CASE
         WHEN t.estimate_minutes IS NULL THEN NULL
         ELSE GREATEST(t.estimate_minutes - COALESCE((
           SELECT SUM(duration_minutes)::int FROM time_entries te
            WHERE te.task_id = t.task_id
         ), 0), 0)
       END,
       revision = revision + 1
 WHERE t.task_id = $1`, taskID.String())
	return err
}
