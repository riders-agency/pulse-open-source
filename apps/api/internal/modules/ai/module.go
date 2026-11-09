package ai

import (
	"github.com/go-chi/chi/v5"

	aisvc "riders.agency/pulse/api/internal/modules/ai/service"
	aihttp "riders.agency/pulse/api/internal/modules/ai/transport/http"
	"riders.agency/pulse/api/internal/platform/httpserver"
)

func RegisterAPIV1(scope httpserver.APIV1Scope, deps httpserver.Dependencies) {
	switch scope.Kind {
	case httpserver.APIV1Public:
		MountPublicHTTP(scope.Router)
	case httpserver.APIV1Project:
		MountProjectHTTP(scope.Router, deps.AI)
	}
}

func MountPublicHTTP(r chi.Router) {
	aihttp.MountReadinessRoutes(r)
}

func MountProjectHTTP(r chi.Router, runtime *aisvc.Commands) {
	aihttp.MountProjectRoutes(r, runtime)
	aihttp.MountActionDependencyModelRoutes(r, runtime)
}
