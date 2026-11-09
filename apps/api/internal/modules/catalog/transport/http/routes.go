package http

import (
	"github.com/go-chi/chi/v5"

	catrepo "riders.agency/pulse/api/internal/modules/catalog/repository"
)

func MountRoutes(r chi.Router, catalog *catrepo.Catalog) {
	r.Get("/action-types", ListActionTypes(catalog))
	r.Get("/action-types/{action_type_key}", GetActionType(catalog))
	r.Get("/action-dependency-types", ListDepTypes(catalog))
	r.Get("/action-dependency-types/{dependency_type_key}", GetDepType(catalog))
	r.Get("/providers/catalog", ProviderCatalog())
}
