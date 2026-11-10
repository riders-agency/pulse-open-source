package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	workrepo "riders.agency/pulse/api/internal/modules/worktracker/repository"
	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) AddReviewStage(ctx context.Context, projectID, userID, workflowID uuid.UUID) error {
	if err := authz.RequireProjectProcess(ctx, c.P, projectID); err != nil {
		return err
	}
	access, err := c.WorkTracker.AccessForProject(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if !access.CanManageProcesses {
		return workrepo.ErrForbidden
	}
	return uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		var belongs bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM task_workflows WHERE task_workflow_id=$1 AND project_id=$2)`, workflowID, projectID).Scan(&belongs); err != nil {
			return err
		}
		if !belongs {
			return workrepo.ErrForbidden
		}
		var existingReview uuid.UUID
		err = tx.QueryRow(ctx, `
SELECT task_workflow_status_id
  FROM task_workflow_statuses
 WHERE task_workflow_id = $1
   AND (lower(name) IN ('review','ревью','на ревью') OR category = 'review')
 ORDER BY position LIMIT 1`, workflowID).Scan(&existingReview)
		if err == pgx.ErrNoRows {
			existingReview = uuid.New()
			_, err = tx.Exec(ctx, `
INSERT INTO task_workflow_statuses (
  task_workflow_status_id, task_workflow_id, name, category, color, position,
  is_start, is_done, is_archived, requires_assignee, requires_estimate
) VALUES ($1,$2,'Review','review','violet',30,false,false,false,false,false)`,
				existingReview, workflowID)
		}
		if err != nil {
			return err
		}
		var fromID uuid.UUID
		err = tx.QueryRow(ctx, `
SELECT task_workflow_status_id
  FROM task_workflow_statuses
 WHERE task_workflow_id = $1
   AND (category = 'in_progress' OR lower(name) IN ('in progress','в работе'))
 ORDER BY position LIMIT 1`, workflowID).Scan(&fromID)
		if err == pgx.ErrNoRows {
			err = tx.QueryRow(ctx, `
SELECT task_workflow_status_id
  FROM task_workflow_statuses
 WHERE task_workflow_id = $1
 ORDER BY position LIMIT 1`, workflowID).Scan(&fromID)
		}
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
INSERT INTO task_workflow_transitions (
  task_workflow_id, from_status_id, to_status_id, name, allowed_roles, position
)
SELECT $1,$2,$3,'Передать на Review', ARRAY['project_manager','team_lead']::text[],
       COALESCE((SELECT max(position) + 10 FROM task_workflow_transitions WHERE task_workflow_id = $1), 10)
WHERE NOT EXISTS (
  SELECT 1 FROM task_workflow_transitions
   WHERE task_workflow_id = $1 AND from_status_id = $2 AND to_status_id = $3
)`, workflowID, fromID, existingReview)
		if err != nil {
			return err
		}
		return nil
	})
}
