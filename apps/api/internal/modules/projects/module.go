package projects

import (
	"github.com/go-chi/chi/v5"
	projectssvc "riders.agency/pulse/api/internal/modules/projects/service"

	prhttp "riders.agency/pulse/api/internal/modules/projects/transport/http"
	"riders.agency/pulse/api/internal/platform/httpserver"
)

func RegisterAPIV1(scope httpserver.APIV1Scope, deps httpserver.Dependencies) {
	switch scope.Kind {
	case httpserver.APIV1Protected:
		MountCompanyHTTP(scope.Router, deps.Projects)
	case httpserver.APIV1Project:
		MountProjectHTTP(scope.Router, deps.Projects)
	}
}

func MountCompanyHTTP(r chi.Router, projects *projectssvc.Commands) {
	prhttp.MountCompanyRoutes(r, projects)
}

func MountProjectHTTP(r chi.Router, projects *projectssvc.Commands) {
	prhttp.MountProjectRoutes(r, projects)
}
