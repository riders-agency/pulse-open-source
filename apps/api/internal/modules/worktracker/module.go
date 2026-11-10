package worktracker

import (
	"github.com/go-chi/chi/v5"

	worksvc "riders.agency/pulse/api/internal/modules/worktracker/service"
	workhttp "riders.agency/pulse/api/internal/modules/worktracker/transport/http"
	"riders.agency/pulse/api/internal/platform/httpserver"
)

func RegisterAPIV1(scope httpserver.APIV1Scope, deps httpserver.Dependencies) {
	if scope.Kind == httpserver.APIV1Work {
		MountHTTP(scope.Router, deps.WorkTracker)
	}
}

func MountHTTP(r chi.Router, workTracker *worksvc.Commands) {
	workhttp.Mount(r, workTracker)
}
