package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/modules/workflows/definitionvalidate"
)

type Snapshot struct {
	ProjectID             string
	VersionID             uuid.UUID
	Definition, TestInput []byte
}

func CreateTestSnapshot(ctx context.Context, tx pgx.Tx, wfID, uid uuid.UUID) (Snapshot, error) {
	var projectID string
	var def []byte
	var graph []byte
	var dsl *string
	var storedInput []byte
	if err := tx.QueryRow(ctx, `
SELECT project_id::text, draft_definition_json, draft_graph_json, draft_dsl_yaml, test_input_json
  FROM workflows WHERE workflow_id = $1 FOR UPDATE`, wfID.String()).
		Scan(&projectID, &def, &graph, &dsl, &storedInput); err != nil {
		return Snapshot{}, err
	}
	if err := definitionvalidate.Validate(def); err != nil {
		return Snapshot{}, err
	}

	var next int
	if err := tx.QueryRow(ctx, `
SELECT COALESCE(MAX(version_no),0)+1 FROM workflow_versions WHERE workflow_id = $1`, wfID.String()).
		Scan(&next); err != nil {
		return Snapshot{}, err
	}

	versionID := uuid.New()
	graphJSON := graph
	if graphJSON == nil {
		graphJSON = []byte("null")
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO workflow_versions (
  workflow_version_id, project_id, workflow_id, version_no, status,
  definition_json, graph_json, dsl_yaml, created_by_user_id)
VALUES ($1,$2,$3,$4,'test',$5::jsonb,$6::jsonb,$7,$8)`,
		versionID.String(), projectID, wfID.String(), next, string(def), string(graphJSON), dsl, uid.String()); err != nil {
		return Snapshot{}, err
	}

	return Snapshot{projectID, versionID, def, storedInput}, nil
}
