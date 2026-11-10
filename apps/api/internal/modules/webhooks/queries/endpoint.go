package queries

import (
	"context"
	"time"

	projectqueries "riders.agency/pulse/api/internal/modules/projects/queries"
	dbpkg "riders.agency/pulse/api/internal/platform/db"
)

type Endpoint struct {
	ProjectID, EndpointID, WorkflowID, WorkflowVersionID, ProjectHookSecretHash, AuthType string
	SecretHash, PreviousSecretHash                                                        *string
	SecretCiphertext, PreviousSecretCiphertext                                            []byte
	PreviousSecretExpiresAt                                                               *time.Time
	ReplayWindowSeconds                                                                   int
}

func LookupEndpoint(ctx context.Context, db dbpkg.Executor, hookKey, endpointKey string) (Endpoint, error) {
	var e Endpoint
	err := db.QueryRow(ctx, `SELECT p.project_id::text,we.webhook_endpoint_id::text,we.workflow_id::text,we.workflow_version_id::text,we.secret_hash::text,we.secret_ciphertext,we.previous_secret_hash::text,we.previous_secret_ciphertext,we.previous_secret_expires_at,we.auth_type::text,we.replay_window_seconds::int FROM pulse_read.projects_v1 p JOIN webhook_endpoints we ON we.project_id=p.project_id WHERE p.hook_key=$1 AND we.endpoint_key=$2 AND we.status='active'`, hookKey, endpointKey).Scan(&e.ProjectID, &e.EndpointID, &e.WorkflowID, &e.WorkflowVersionID, &e.SecretHash, &e.SecretCiphertext, &e.PreviousSecretHash, &e.PreviousSecretCiphertext, &e.PreviousSecretExpiresAt, &e.AuthType, &e.ReplayWindowSeconds)
	if err != nil {
		return e, err
	}
	e.ProjectHookSecretHash, err = projectqueries.HookSecretHash(ctx, db, e.ProjectID)
	return e, err
}
