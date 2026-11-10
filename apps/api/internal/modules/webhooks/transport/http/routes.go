package http

import (
	"github.com/go-chi/chi/v5"
	webhookssvc "riders.agency/pulse/api/internal/modules/webhooks/service"

	whrepo "riders.agency/pulse/api/internal/modules/webhooks/repository"
)

func MountWorkflowRoutes(r chi.Router, webhooks *whrepo.Webhooks) {
	r.Post("/workflows/{workflow_id}/webhook-endpoints", Create(webhooks))
	r.Get("/workflows/{workflow_id}/webhook-endpoints", ListByWorkflow(webhooks))
}

func MountEndpointRoutes(r chi.Router, webhooks *whrepo.Webhooks) {
	r.Get("/webhook-endpoints/{webhook_endpoint_id}", GetEndpoint(webhooks))
	r.Patch("/webhook-endpoints/{webhook_endpoint_id}", PatchEndpoint(webhooks))
	r.Post("/webhook-endpoints/{webhook_endpoint_id}/rotate-secret", RotateEndpointSecret(webhooks))
	r.Get("/webhook-endpoints/{webhook_endpoint_id}/events", EndpointEvents(webhooks))
	r.Get("/webhook-endpoints/{webhook_endpoint_id}/snippets", Snippets(webhooks))
}

func MountEventRoutes(r chi.Router, events *webhookssvc.Commands) {
	r.Get("/webhook-events/{webhook_event_id}", GetWebhookEvent(events))
	r.Post("/webhook-events/{webhook_event_id}/replay", ReplayWebhookEvent(events))
}
