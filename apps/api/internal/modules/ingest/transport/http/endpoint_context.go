package http

import (
	"context"

	webhookqueries "riders.agency/pulse/api/internal/modules/webhooks/queries"

	"github.com/jackc/pgx/v5/pgxpool"
)

type webhookEndpointContext struct {
	ProjectID         string
	EndpointID        string
	WorkflowID        string
	WorkflowVersionID string
	Auth              webhookAuthContext
}

func lookupWebhookEndpoint(ctx context.Context, pool *pgxpool.Pool, hookKey, endpointKey string) (webhookEndpointContext, error) {
	e, err := webhookqueries.LookupEndpoint(ctx, pool, hookKey, endpointKey)
	return webhookEndpointContext{ProjectID: e.ProjectID, EndpointID: e.EndpointID, WorkflowID: e.WorkflowID, WorkflowVersionID: e.WorkflowVersionID, Auth: webhookAuthContext{projectHookSecretHash: e.ProjectHookSecretHash, endpointSecretHash: e.SecretHash, endpointSecretCiphertext: e.SecretCiphertext, previousSecretHash: e.PreviousSecretHash, previousSecretCiphertext: e.PreviousSecretCiphertext, previousSecretExpiresAt: e.PreviousSecretExpiresAt, authType: e.AuthType, replayWindowSeconds: e.ReplayWindowSeconds}}, err
}
