package action_dependencies

import (
	"github.com/go-chi/chi/v5"
	actiondependenciessvc "riders.agency/pulse/api/internal/modules/action_dependencies/service"

	dehttp "riders.agency/pulse/api/internal/modules/action_dependencies/transport/http"
	"riders.agency/pulse/api/internal/platform/httpserver"
)

func RegisterAPIV1(scope httpserver.APIV1Scope, deps httpserver.Dependencies) {
	if scope.Kind == httpserver.APIV1Project {
		MountProjectHTTP(scope.Router, deps.ActionDeps)
	}
}

func MountProjectHTTP(r chi.Router, deps *actiondependenciessvc.Commands) {
	dehttp.MountProjectRoutes(r, deps)
}
