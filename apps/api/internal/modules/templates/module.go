package templates

import (
	"github.com/go-chi/chi/v5"

	tplrepo "riders.agency/pulse/api/internal/modules/templates/repository"
	tplhttp "riders.agency/pulse/api/internal/modules/templates/transport/http"
	"riders.agency/pulse/api/internal/platform/httpserver"
)

func RegisterAPIV1(scope httpserver.APIV1Scope, deps httpserver.Dependencies) {
	if scope.Kind == httpserver.APIV1Project {
		MountProjectHTTP(scope.Router, deps.EmailTemplates, deps.MessageTemplates)
	}
}

func MountProjectHTTP(r chi.Router, email *tplrepo.EmailTemplates, message *tplrepo.MessageTemplates) {
	tplhttp.MountProjectRoutes(r, email, message)
}
