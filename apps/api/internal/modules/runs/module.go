package runs

import (
	"github.com/go-chi/chi/v5"
	runssvc "riders.agency/pulse/api/internal/modules/runs/service"

	runhttp "riders.agency/pulse/api/internal/modules/runs/transport/http"
	"riders.agency/pulse/api/internal/platform/httpserver"
)

func RegisterAPIV1(scope httpserver.APIV1Scope, deps httpserver.Dependencies) {
	if scope.Kind == httpserver.APIV1Protected {
		MountProtectedHTTP(scope.Router, deps.Runs)
	}
}

func MountProtectedHTTP(r chi.Router, runs *runssvc.Commands) {
	runhttp.MountWorkflowRoutes(r, runs)
	runhttp.MountRunRoutes(r, runs)
}
