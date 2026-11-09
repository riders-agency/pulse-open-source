package commands

import (
	"context"

	"github.com/google/uuid"
	"riders.agency/pulse/api/internal/platform/authz"
	dbpkg "riders.agency/pulse/api/internal/platform/db"
)

func SetInitialTaskWorkflow(ctx context.Context, db dbpkg.Executor, projectID, workflowID uuid.UUID) error {
	if err := authz.RequireProjectProcess(ctx, db, projectID); err != nil {
		return err
	}
	_, err := db.Exec(ctx, `UPDATE projects SET task_workflow_id=$2 WHERE project_id=$1 AND task_workflow_id IS NULL`, projectID, workflowID)
	return err
}
