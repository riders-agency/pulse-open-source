package service

import (
	"context"

	"github.com/google/uuid"
	workrepo "riders.agency/pulse/api/internal/modules/worktracker/repository"
)

// SetTaskArchived preserves comments, attachments and accounting history.
func (c *Commands) SetTaskArchived(ctx context.Context, taskID, userID uuid.UUID, archived bool) error {
	access, _, err := c.WorkTracker.AccessForStoredTask(ctx, taskID, userID)
	if err != nil {
		return err
	}
	tx, err := c.WorkTracker.P.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var task workrepo.Task
	var current bool
	if err = tx.QueryRow(ctx, `SELECT task_workflow_id::text, COALESCE(assignee_user_id::text,''), COALESCE(reporter_user_id::text,''), archived_at IS NOT NULL FROM tasks WHERE task_id=$1 FOR UPDATE`, taskID).Scan(&task.WorkflowID, &task.AssigneeUserID, &task.ReporterUserID, &current); err != nil {
		return err
	}
	if !access.CanEditTaskInWorkflow(task.WorkflowID) || access.IsClientGuest || !access.CanSeeTask(task) {
		return workrepo.ErrForbidden
	}
	if current == archived {
		return nil
	}
	if archived {
		var running bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM timer_sessions WHERE task_id=$1 AND state IN ('running','paused'))`, taskID).Scan(&running); err != nil {
			return err
		}
		if running {
			return workrepo.ErrTaskTimerActive
		}
	}
	if _, err = tx.Exec(ctx, `UPDATE tasks SET archived_at=CASE WHEN $2 THEN now() ELSE NULL END, revision=revision+1 WHERE task_id=$1`, taskID, archived); err != nil {
		return err
	}
	event := "task.restored"
	if archived {
		event = "task.archived"
	}
	if err = insertActivity(ctx, tx, access.CompanyID, taskID, userID, event, map[string]any{}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
