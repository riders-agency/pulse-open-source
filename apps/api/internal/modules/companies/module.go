package companies

import (
	"github.com/go-chi/chi/v5"
	companiessvc "riders.agency/pulse/api/internal/modules/companies/service"

	cohttp "riders.agency/pulse/api/internal/modules/companies/transport/http"
	"riders.agency/pulse/api/internal/platform/httpserver"
)

func RegisterAPIV1(scope httpserver.APIV1Scope, deps httpserver.Dependencies) {
	if scope.Kind == httpserver.APIV1Protected {
		MountHTTP(scope.Router, deps.Companies)
	}
}

func MountHTTP(r chi.Router, companies *companiessvc.Commands) {
	cohttp.MountRoutes(r, companies)
}
