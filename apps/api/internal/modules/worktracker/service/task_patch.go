package service

import (
	"context"

	"github.com/google/uuid"

	workrepo "riders.agency/pulse/api/internal/modules/worktracker/repository"
)

func (c *Commands) PatchTask(ctx context.Context, userID uuid.UUID, in workrepo.PatchTaskInput) (*workrepo.Task, error) {
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
	if !access.CanEditTaskInWorkflow(current.WorkflowID) {
		return nil, workrepo.ErrForbidden
	}
	if in.ClientIDSet && in.ClientID != nil {
		var ok bool
		if err := c.WorkTracker.P.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1 FROM pulse_read.clients_v1
   WHERE client_id = $1
     AND company_id = $2
)`, in.ClientID.String(), access.CompanyID.String()).Scan(&ok); err != nil {
			return nil, err
		}
		if !ok {
			return nil, workrepo.ErrValidation
		}
	}

	tx, err := c.WorkTracker.P.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `
UPDATE tasks SET
  title = COALESCE($2, title),
  description = COALESCE($3, description),
  type = COALESCE($4, type),
  priority = COALESCE($5, priority),
  severity = COALESCE($6, severity),
  assignee_user_id = COALESCE($7::uuid, assignee_user_id),
  reviewer_user_id = COALESCE($8::uuid, reviewer_user_id),
  start_date = COALESCE($9::date, start_date),
  due_date = COALESCE($10::date, due_date),
  estimate_minutes = COALESCE($11, estimate_minutes),
  remaining_minutes = COALESCE($11, remaining_minutes),
  story_points = COALESCE($12, story_points),
  labels = CASE WHEN $13 THEN $14 ELSE labels END,
  client_visible = COALESCE($15, client_visible),
  billable = COALESCE($16, billable),
  is_blocked = COALESCE($17, is_blocked),
  blocked_reason = COALESCE($18, blocked_reason),
  client_id = CASE WHEN $19 THEN $20::uuid ELSE client_id END,
  revision = revision + 1
 WHERE task_id = $1
   AND revision = $21`, in.TaskID.String(), textPtr(in.Title), textPtr(in.Description), textPtr(in.Type),
		textPtr(in.Priority), textPtr(in.Severity), nullableUUID(in.AssigneeUserID), nullableUUID(in.ReviewerUserID),
		dateArg(in.StartDate), dateArg(in.DueDate), in.EstimateMinutes, in.StoryPoints, in.LabelsSet, in.Labels,
		in.ClientVisible, in.Billable, in.IsBlocked, textPtr(in.BlockedReason), in.ClientIDSet, nullableUUID(in.ClientID),
		expectedRevision(current.Revision, in.ExpectedRevision))
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, workrepo.ErrConflict
	}
	if err := insertActivity(ctx, tx, access.CompanyID, in.TaskID, userID, "task.updated", map[string]any{}); err != nil {
		return nil, err
	}
	if c.WorkTracker.TaskEvents != nil {
		if err := c.WorkTracker.TaskEvents.EnqueueTaskModuleTriggers(ctx, tx, access.CompanyID, access.ProjectID, uuid.MustParse(current.WorkflowID), in.TaskID, "task.updated", map[string]any{}); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return c.WorkTracker.TaskRow(ctx, in.TaskID)
}
