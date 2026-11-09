package http

import (
	"github.com/go-chi/chi/v5"
	companiessvc "riders.agency/pulse/api/internal/modules/companies/service"
)

func MountRoutes(r chi.Router, companies *companiessvc.Commands) {
	r.Get("/companies", List(companies))
	r.Post("/companies", Create(companies))
	r.Get("/companies/{company_id}", Get(companies))
	r.Patch("/companies/{company_id}", Patch(companies))
	r.Post("/companies/{company_id}/switch", Switch(companies))
}
