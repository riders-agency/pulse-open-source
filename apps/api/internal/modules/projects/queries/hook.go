package queries

import (
	"context"

	dbpkg "riders.agency/pulse/api/internal/platform/db"
)

func HookSecretHash(ctx context.Context, db dbpkg.Executor, projectID string) (string, error) {
	var hash string
	err := db.QueryRow(ctx, `SELECT hook_secret_hash FROM projects WHERE project_id=$1`, projectID).Scan(&hash)
	return hash, err
}
