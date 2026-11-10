package commands

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"riders.agency/pulse/api/internal/platform/authz"
	dbpkg "riders.agency/pulse/api/internal/platform/db"
)

func SetDefaultModels(ctx context.Context, db dbpkg.Executor, projectID, dependencyID uuid.UUID, metadata json.RawMessage) error {
	if err := authz.RequireProject(ctx, db, projectID, authz.PermAdmin); err != nil {
		return err
	}
	_, err := db.Exec(ctx, `UPDATE action_dependencies SET metadata_json=$3::jsonb,updated_at=now() WHERE project_id=$1 AND action_dependency_id=$2`, projectID, dependencyID, metadata)
	return err
}
