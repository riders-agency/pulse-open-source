package integrationdocs

import (
	"github.com/go-chi/chi/v5"

	docsrepo "riders.agency/pulse/api/internal/modules/integrationdocs/repository"
	docshttp "riders.agency/pulse/api/internal/modules/integrationdocs/transport/http"
	"riders.agency/pulse/api/internal/platform/httpserver"
)

func RegisterAPIV1(scope httpserver.APIV1Scope, deps httpserver.Dependencies) {
	switch scope.Kind {
	case httpserver.APIV1Public:
		MountPublicHTTP(scope.Router, deps.IntegrationDocs)
	case httpserver.APIV1Project:
		MountProjectHTTP(scope.Router, deps.IntegrationDocs)
	}
}

func MountPublicHTTP(r chi.Router, docs *docsrepo.Docs) {
	docshttp.MountPublicRoutes(r, docs)
}

func MountProjectHTTP(r chi.Router, docs *docsrepo.Docs) {
	docshttp.MountProjectRoutes(r, docs)
}
