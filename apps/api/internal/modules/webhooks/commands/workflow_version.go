package commands

import (
	"context"

	"github.com/google/uuid"
	"riders.agency/pulse/api/internal/platform/authz"
	dbpkg "riders.agency/pulse/api/internal/platform/db"
)

func FollowWorkflowVersion(ctx context.Context, db dbpkg.Executor, workflowID uuid.UUID, versionID string) error {
	if err := authz.RequireWorkflow(ctx, db, workflowID, authz.PermWrite); err != nil {
		return err
	}
	_, err := db.Exec(ctx, `UPDATE webhook_endpoints SET workflow_version_id=$2,updated_at=now() WHERE workflow_id=$1`, workflowID, versionID)
	return err
}
