package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	workrepo "riders.agency/pulse/api/internal/modules/worktracker/repository"
)

func (c *Commands) CreateTask(ctx context.Context, userID uuid.UUID, in workrepo.CreateTaskInput) (*workrepo.Task, error) {
	if c == nil || c.WorkTracker == nil || c.WorkTracker.P == nil {
		return nil, workrepo.ErrNotFound
	}
	access, err := c.WorkTracker.AccessForProject(ctx, in.ProjectID, userID)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, workrepo.ErrValidation
	}
	var workflowID uuid.UUID
	if err := c.WorkTracker.P.QueryRow(ctx, `
SELECT s.task_workflow_id
  FROM task_workflow_statuses s
  JOIN task_workflows tw ON tw.task_workflow_id = s.task_workflow_id
 WHERE s.task_workflow_status_id = $1::uuid
   AND tw.project_id = $2::uuid`,
		in.StatusID.String(), in.ProjectID.String()).Scan(&workflowID); err != nil {
		if err == pgx.ErrNoRows {
			return nil, workrepo.ErrValidation
		}
		return nil, fmt.Errorf("create task workflow lookup: %w", err)
	}
	if !access.CanCreateTaskInWorkflow(workflowID.String()) {
		return nil, workrepo.ErrForbidden
	}

	taskID := uuid.New()
	tx, err := c.WorkTracker.P.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("create task begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	taskKey, err := nextTaskKey(ctx, tx, in.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("create task next key: %w", err)
	}
	labels := cleanLabels(in.Labels)
	if _, err := tx.Exec(ctx, `
INSERT INTO tasks (
  task_id, company_id, project_id, client_id, task_key, title, description, type,
  task_workflow_status_id, task_workflow_id, priority, severity, assignee_user_id,
  reporter_user_id, reviewer_user_id, owner_user_id, created_by_user_id,
  start_date, due_date, estimate_minutes, remaining_minutes, story_points,
  rank, client_visible, billable, labels
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$14,$14,$16,$17,$18,$18,$19,$20,$21,$22,$23
)`,
		taskID.String(),
		access.CompanyID.String(),
		in.ProjectID.String(),
		nullableUUID(in.ClientID),
		taskKey,
		title,
		strings.TrimSpace(in.Description),
		defaultText(in.Type, "task"),
		in.StatusID.String(),
		workflowID.String(),
		defaultText(in.Priority, "medium"),
		defaultText(in.Severity, "medium"),
		nullableUUID(in.AssigneeUserID),
		userID.String(),
		nullableUUID(in.ReviewerUserID),
		dateArg(in.StartDate),
		dateArg(in.DueDate),
		in.EstimateMinutes,
		in.StoryPoints,
		rankValue(in.StatusID),
		in.ClientVisible,
		in.Billable,
		labels,
	); err != nil {
		return nil, fmt.Errorf("create task insert: %w", err)
	}
	if err := insertActivity(ctx, tx, access.CompanyID, taskID, userID, "task.created", map[string]any{"task_key": taskKey}); err != nil {
		return nil, fmt.Errorf("create task activity: %w", err)
	}
	if c.WorkTracker.TaskEvents != nil {
		if err := c.WorkTracker.TaskEvents.EnqueueTaskModuleTriggers(ctx, tx, access.CompanyID, in.ProjectID, workflowID, taskID, "task.created", map[string]any{"task_key": taskKey}); err != nil {
			return nil, fmt.Errorf("create task triggers: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("create task commit: %w", err)
	}
	return c.WorkTracker.TaskRow(ctx, taskID)
}
