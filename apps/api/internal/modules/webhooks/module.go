package webhooks

import (
	"github.com/go-chi/chi/v5"
	webhookssvc "riders.agency/pulse/api/internal/modules/webhooks/service"

	whrepo "riders.agency/pulse/api/internal/modules/webhooks/repository"
	whhttp "riders.agency/pulse/api/internal/modules/webhooks/transport/http"
	"riders.agency/pulse/api/internal/platform/httpserver"
)

func RegisterAPIV1(scope httpserver.APIV1Scope, deps httpserver.Dependencies) {
	if scope.Kind != httpserver.APIV1Protected {
		return
	}
	MountWorkflowHTTP(scope.Router, deps.Webhooks)
	MountEndpointHTTP(scope.Router, deps.Webhooks)
	MountEventHTTP(scope.Router, deps.WebhookEvents)
}

func MountWorkflowHTTP(r chi.Router, webhooks *whrepo.Webhooks) {
	whhttp.MountWorkflowRoutes(r, webhooks)
}

func MountEndpointHTTP(r chi.Router, webhooks *whrepo.Webhooks) {
	whhttp.MountEndpointRoutes(r, webhooks)
}

func MountEventHTTP(r chi.Router, events *webhookssvc.Commands) {
	whhttp.MountEventRoutes(r, events)
}
