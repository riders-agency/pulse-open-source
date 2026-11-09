package desktop

import (
	"github.com/go-chi/chi/v5"

	deskhttp "riders.agency/pulse/api/internal/modules/desktop/transport/http"
	"riders.agency/pulse/api/internal/platform/httpserver"
)

type Handlers = deskhttp.Handlers

func RegisterAPIV1(scope httpserver.APIV1Scope, deps httpserver.Dependencies) {
	if scope.Kind == httpserver.APIV1Protected {
		MountHTTP(scope.Router, deps.Desktop)
	}
}

func MountHTTP(r chi.Router, desktop *Handlers) {
	deskhttp.MountRoutes(r, desktop)
}
