package http

import (
	"github.com/go-chi/chi/v5"
	apikeyssvc "riders.agency/pulse/api/internal/modules/apikeys/service"
)

func MountRoutes(r chi.Router, apiKeys *apikeyssvc.Commands) {
	r.Get("/auth/api-keys", List(apiKeys))
	r.Post("/auth/api-keys", Create(apiKeys))
	r.Delete("/auth/api-keys/{api_key_id}", Revoke(apiKeys))
}
