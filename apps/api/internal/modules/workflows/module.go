package workflows

import (
	"github.com/go-chi/chi/v5"
	"riders.agency/pulse/api/internal/modules/workflows/office"
	workflowssvc "riders.agency/pulse/api/internal/modules/workflows/service"

	wfhttp "riders.agency/pulse/api/internal/modules/workflows/transport/http"
	"riders.agency/pulse/api/internal/platform/httpserver"
)

func RegisterAPIV1(scope httpserver.APIV1Scope, deps httpserver.Dependencies) {
	switch scope.Kind {
	case httpserver.APIV1Project:
		MountProjectHTTP(scope.Router, deps.Workflows)
		wfhttp.MountOfficeRoutes(scope.Router, &office.Service{P: deps.Pool})
	case httpserver.APIV1Protected:
		MountProtectedHTTP(scope.Router, deps.Workflows)
	}
}

func MountProjectHTTP(r chi.Router, workflows *workflowssvc.Commands) {
	wfhttp.MountProjectRoutes(r, workflows)
}

func MountProtectedHTTP(r chi.Router, workflows *workflowssvc.Commands) {
	wfhttp.MountRoutes(r, workflows)
}
