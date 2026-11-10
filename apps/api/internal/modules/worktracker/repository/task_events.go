package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TaskEventPublisher interface {
	EnqueueTaskModuleTriggers(ctx context.Context, tx pgx.Tx, companyID, projectID, workflowID, taskID uuid.UUID, eventKey string, payload map[string]any) error
}

func (w *WorkTracker) enqueueTaskModuleTriggers(ctx context.Context, tx pgx.Tx, companyID, projectID, workflowID, taskID uuid.UUID, eventKey string, payload map[string]any) error {
	if w == nil || w.TaskEvents == nil {
		return nil
	}
	return w.TaskEvents.EnqueueTaskModuleTriggers(ctx, tx, companyID, projectID, workflowID, taskID, eventKey, payload)
}
