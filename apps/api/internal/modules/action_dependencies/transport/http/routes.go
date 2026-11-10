package http

import (
	"github.com/go-chi/chi/v5"
	actiondependenciessvc "riders.agency/pulse/api/internal/modules/action_dependencies/service"
)

func MountProjectRoutes(r chi.Router, deps *actiondependenciessvc.Commands) {
	r.Get("/action-dependencies", List(deps))
	r.Post("/action-dependencies", Create(deps))
	r.Get("/action-dependencies/{action_dependency_id}", Get(deps))
	r.Patch("/action-dependencies/{action_dependency_id}", Patch(deps))
	r.Delete("/action-dependencies/{action_dependency_id}", Delete(deps))
	r.Post("/action-dependencies/{action_dependency_id}/rotate-secret", RotateSecret(deps))
	r.Post("/action-dependencies/{action_dependency_id}/telegram-relay-webhook-secret/rotate", RotateTelegramRelayWebhookSecret(deps))
	r.Post("/action-dependencies/{action_dependency_id}/telegram-relay-polling", SetTelegramRelayPolling(deps))
	r.Post("/action-dependencies/{action_dependency_id}/test", Test(deps))
}
