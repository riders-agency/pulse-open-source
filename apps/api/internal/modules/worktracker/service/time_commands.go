package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	workrepo "riders.agency/pulse/api/internal/modules/worktracker/repository"
)

func (c *Commands) StartTimer(ctx context.Context, taskID, userID uuid.UUID) (*workrepo.TimerSession, error) {
	if c == nil || c.WorkTracker == nil || c.WorkTracker.P == nil {
		return nil, workrepo.ErrNotFound
	}
	access, _, err := c.WorkTracker.AccessForTask(ctx, taskID, userID)
	if err != nil {
		return nil, err
	}
	task, err := c.WorkTracker.TaskRow(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if !access.CanLogTimeInWorkflow(task.WorkflowID) || !access.CanSeeTask(*task) {
		return nil, workrepo.ErrForbidden
	}
	tx, err := c.WorkTracker.P.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var active bool
	if err := tx.QueryRow(ctx, `SELECT archived_at IS NULL FROM tasks WHERE task_id=$1 FOR UPDATE`, taskID).Scan(&active); err != nil {
		return nil, err
	}
	if !active {
		return nil, workrepo.ErrNotFound
	}
	timerID := uuid.New()
	if _, err := tx.Exec(ctx, `
INSERT INTO timer_sessions (timer_session_id, company_id, task_id, user_id, state)
VALUES ($1,$2,$3,$4,'running')`, timerID.String(), access.CompanyID.String(), taskID.String(), userID.String()); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return c.WorkTracker.TimerSession(ctx, timerID)
}

func (c *Commands) StopTimer(ctx context.Context, userID uuid.UUID, taskIDFilter *uuid.UUID) (*workrepo.TimeEntry, error) {
	if c == nil || c.WorkTracker == nil || c.WorkTracker.P == nil {
		return nil, workrepo.ErrNotFound
	}
	tx, err := c.WorkTracker.P.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var timerID, companyID, taskID, projectID, workflowID uuid.UUID
	var clientID *uuid.UUID
	var statusID uuid.UUID
	var startedAt time.Time
	err = tx.QueryRow(ctx, `
SELECT ts.timer_session_id, ts.company_id, ts.task_id, t.project_id, t.client_id,
       t.task_workflow_id, t.task_workflow_status_id, ts.started_at
  FROM timer_sessions ts
  JOIN tasks t ON t.task_id = ts.task_id
 WHERE ts.user_id = $1
   AND ts.state IN ('running','paused')
   AND ($2::uuid IS NULL OR ts.task_id = $2)
 ORDER BY ts.started_at DESC
 LIMIT 1
 FOR UPDATE`, userID.String(), nullableUUID(taskIDFilter)).Scan(&timerID, &companyID, &taskID, &projectID, &clientID, &workflowID, &statusID, &startedAt)
	if err == pgx.ErrNoRows {
		return nil, workrepo.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	duration := int(time.Since(startedAt).Minutes())
	if duration < 1 {
		duration = 1
	}
	entryID := uuid.New()
	if _, err := tx.Exec(ctx, `
INSERT INTO time_entries (
  time_entry_id, company_id, task_id, project_id, client_id, user_id, entry_date,
  started_at, ended_at, duration_minutes, description, billable, source, task_workflow_status_id
) VALUES ($1,$2,$3,$4,$5,$6,CURRENT_DATE,$7,now(),$8,'Timer session',true,'timer',$9)`,
		entryID.String(), companyID.String(), taskID.String(), projectID.String(), nullableUUID(clientID), userID.String(), startedAt, duration, statusID.String()); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
UPDATE timer_sessions
   SET state = 'stopped',
       stopped_at = now(),
       duration_minutes_cached = $2,
       updated_at = now()
 WHERE timer_session_id = $1`, timerID.String(), duration); err != nil {
		return nil, err
	}
	if err := refreshSpent(ctx, tx, taskID); err != nil {
		return nil, err
	}
	if c.WorkTracker.TaskEvents != nil {
		if err := c.WorkTracker.TaskEvents.EnqueueTaskModuleTriggers(ctx, tx, companyID, projectID, workflowID, taskID, "time.logged", map[string]any{"duration_minutes": duration, "source": "timer"}); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return c.WorkTracker.TimeEntry(ctx, entryID)
}

func (c *Commands) AddManualTime(ctx context.Context, userID uuid.UUID, in workrepo.ManualTimeInput) (*workrepo.TimeEntry, error) {
	if c == nil || c.WorkTracker == nil || c.WorkTracker.P == nil {
		return nil, workrepo.ErrNotFound
	}
	access, _, err := c.WorkTracker.AccessForTask(ctx, in.TaskID, userID)
	if err != nil {
		return nil, err
	}
	task, err := c.WorkTracker.TaskRow(ctx, in.TaskID)
	if err != nil {
		return nil, err
	}
	if (!access.CanLogTimeInWorkflow(task.WorkflowID) && access.Role != "finance") || !access.CanSeeTask(*task) {
		return nil, workrepo.ErrForbidden
	}
	tx, err := c.WorkTracker.P.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	entryID := uuid.New()
	if _, err := tx.Exec(ctx, `
INSERT INTO time_entries (
  time_entry_id, company_id, task_id, project_id, client_id, user_id, entry_date,
  duration_minutes, description, billable, source, task_workflow_status_id
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'manual',$11)`,
		entryID.String(), access.CompanyID.String(), in.TaskID.String(), task.ProjectID,
		nullableTextUUID(task.ClientID), userID.String(), in.EntryDate.Format("2006-01-02"),
		in.DurationMinutes, in.Description, in.Billable, task.StatusID); err != nil {
		return nil, err
	}
	if err := refreshSpent(ctx, tx, in.TaskID); err != nil {
		return nil, err
	}
	if c.WorkTracker.TaskEvents != nil {
		if err := c.WorkTracker.TaskEvents.EnqueueTaskModuleTriggers(ctx, tx, access.CompanyID, access.ProjectID, uuid.MustParse(task.WorkflowID), in.TaskID, "time.logged", map[string]any{"duration_minutes": in.DurationMinutes, "source": "manual"}); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return c.WorkTracker.TimeEntry(ctx, entryID)
}
