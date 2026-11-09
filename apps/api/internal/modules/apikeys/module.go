package apikeys

import (
	"github.com/go-chi/chi/v5"
	apikeyssvc "riders.agency/pulse/api/internal/modules/apikeys/service"

	apihttp "riders.agency/pulse/api/internal/modules/apikeys/transport/http"
	"riders.agency/pulse/api/internal/platform/httpserver"
)

func RegisterAPIV1(scope httpserver.APIV1Scope, deps httpserver.Dependencies) {
	if scope.Kind == httpserver.APIV1Protected {
		MountHTTP(scope.Router, deps.APIKeys)
	}
}

func MountHTTP(r chi.Router, apiKeys *apikeyssvc.Commands) {
	apihttp.MountRoutes(r, apiKeys)
}
