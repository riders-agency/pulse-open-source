package catalog

import (
	"github.com/go-chi/chi/v5"

	catrepo "riders.agency/pulse/api/internal/modules/catalog/repository"
	cathttp "riders.agency/pulse/api/internal/modules/catalog/transport/http"
	"riders.agency/pulse/api/internal/platform/httpserver"
)

func RegisterAPIV1(scope httpserver.APIV1Scope, deps httpserver.Dependencies) {
	if scope.Kind == httpserver.APIV1Protected {
		MountHTTP(scope.Router, deps.Catalog)
	}
}

func MountHTTP(r chi.Router, catalog *catrepo.Catalog) {
	cathttp.MountRoutes(r, catalog)
}
