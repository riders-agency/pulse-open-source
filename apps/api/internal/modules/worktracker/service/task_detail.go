package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"

	workrepo "riders.agency/pulse/api/internal/modules/worktracker/repository"
)

func (c *Commands) AddComment(ctx context.Context, userID uuid.UUID, in workrepo.AddCommentInput) (*workrepo.TaskDetail, error) {
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
	if !access.CanSeeTask(*task) || !access.CanCommentTaskInWorkflow(task.WorkflowID) {
		return nil, workrepo.ErrForbidden
	}
	body := strings.TrimSpace(in.Body)
	if body == "" {
		return nil, workrepo.ErrValidation
	}
	visibility := "internal"
	if access.IsClientGuest || strings.TrimSpace(in.Visibility) == "client" {
		visibility = "client"
	}
	tx, err := c.WorkTracker.P.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `
INSERT INTO task_comments (task_id, company_id, author_user_id, body, visibility)
VALUES ($1,$2,$3,$4,$5)`, in.TaskID.String(), access.CompanyID.String(), userID.String(), body, visibility); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `UPDATE tasks SET comments_count = comments_count + 1, revision = revision + 1 WHERE task_id = $1`, in.TaskID.String()); err != nil {
		return nil, err
	}
	if err := insertActivity(ctx, tx, access.CompanyID, in.TaskID, userID, "comment.added", map[string]any{"visibility": visibility}); err != nil {
		return nil, err
	}
	if c.WorkTracker.TaskEvents != nil {
		if err := c.WorkTracker.TaskEvents.EnqueueTaskModuleTriggers(ctx, tx, access.CompanyID, access.ProjectID, uuid.MustParse(task.WorkflowID), in.TaskID, "task.comment_added", map[string]any{"visibility": visibility}); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	detail, _, err := c.WorkTracker.Task(ctx, in.TaskID, userID)
	return detail, err
}

func (c *Commands) AddChecklistItem(ctx context.Context, userID uuid.UUID, in workrepo.AddChecklistInput) (*workrepo.TaskDetail, error) {
	access, _, err := c.editableTask(ctx, in.TaskID, userID)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, workrepo.ErrValidation
	}
	tx, err := c.WorkTracker.P.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var position int
	_ = tx.QueryRow(ctx, `SELECT COALESCE(MAX(position), 0) + 10 FROM task_checklist_items WHERE task_id = $1`, in.TaskID.String()).Scan(&position)
	if _, err := tx.Exec(ctx, `
INSERT INTO task_checklist_items (task_id, title, position)
VALUES ($1,$2,$3)`, in.TaskID.String(), title, position); err != nil {
		return nil, err
	}
	if err := refreshChecklistCounts(ctx, tx, in.TaskID); err != nil {
		return nil, err
	}
	if err := insertActivity(ctx, tx, access.CompanyID, in.TaskID, userID, "checklist.added", map[string]any{"title": title}); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	detail, _, err := c.WorkTracker.Task(ctx, in.TaskID, userID)
	return detail, err
}

func (c *Commands) PatchChecklistItem(ctx context.Context, userID uuid.UUID, in workrepo.PatchChecklistInput) (*workrepo.TaskDetail, error) {
	access, _, err := c.editableTask(ctx, in.TaskID, userID)
	if err != nil {
		return nil, err
	}
	tx, err := c.WorkTracker.P.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `
UPDATE task_checklist_items
   SET title = COALESCE($3, title),
       is_done = COALESCE($4, is_done),
       completed_at = CASE
         WHEN COALESCE($4, is_done) THEN COALESCE(completed_at, now())
         ELSE NULL
       END
 WHERE task_id = $1
   AND task_checklist_item_id = $2`, in.TaskID.String(), in.ItemID.String(), textPtr(in.Title), in.IsDone)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, workrepo.ErrNotFound
	}
	if err := refreshChecklistCounts(ctx, tx, in.TaskID); err != nil {
		return nil, err
	}
	if err := insertActivity(ctx, tx, access.CompanyID, in.TaskID, userID, "checklist.updated", map[string]any{}); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	detail, _, err := c.WorkTracker.Task(ctx, in.TaskID, userID)
	return detail, err
}

func (c *Commands) UpsertCustomFieldValue(ctx context.Context, userID uuid.UUID, in workrepo.UpsertCustomFieldValueInput) (*workrepo.TaskDetail, error) {
	access, _, err := c.editableTask(ctx, in.TaskID, userID)
	if err != nil {
		return nil, err
	}
	var ok bool
	if err := c.WorkTracker.P.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1
    FROM task_custom_fields f
    JOIN tasks t ON t.task_id = $2
   WHERE f.task_custom_field_id = $1
     AND f.company_id = t.company_id
     AND (f.project_id IS NULL OR f.project_id = t.project_id)
     AND (f.task_workflow_id IS NULL OR f.task_workflow_id = t.task_workflow_id)
)`, in.FieldID.String(), in.TaskID.String()).Scan(&ok); err != nil {
		return nil, err
	}
	if !ok {
		return nil, workrepo.ErrValidation
	}
	raw, _ := json.Marshal(in.Value)
	tx, err := c.WorkTracker.P.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `
INSERT INTO task_custom_field_values (task_custom_field_id, task_id, value_json)
VALUES ($1,$2,$3::jsonb)
ON CONFLICT (task_custom_field_id, task_id)
DO UPDATE SET value_json = EXCLUDED.value_json, updated_at = now()`, in.FieldID.String(), in.TaskID.String(), string(raw)); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `UPDATE tasks SET revision = revision + 1 WHERE task_id = $1`, in.TaskID.String()); err != nil {
		return nil, err
	}
	if err := insertActivity(ctx, tx, access.CompanyID, in.TaskID, userID, "custom_field.updated", map[string]any{"field_id": in.FieldID.String()}); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	detail, _, err := c.WorkTracker.Task(ctx, in.TaskID, userID)
	return detail, err
}

func (c *Commands) editableTask(ctx context.Context, taskID, userID uuid.UUID) (*workrepo.Access, *workrepo.Task, error) {
	if c == nil || c.WorkTracker == nil || c.WorkTracker.P == nil {
		return nil, nil, workrepo.ErrNotFound
	}
	access, _, err := c.WorkTracker.AccessForTask(ctx, taskID, userID)
	if err != nil {
		return nil, nil, err
	}
	task, err := c.WorkTracker.TaskRow(ctx, taskID)
	if err != nil {
		return nil, nil, err
	}
	if !access.CanEditTaskInWorkflow(task.WorkflowID) {
		return nil, nil, workrepo.ErrForbidden
	}
	return access, task, nil
}
