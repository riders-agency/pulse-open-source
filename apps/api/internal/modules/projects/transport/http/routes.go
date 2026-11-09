package http

import (
	"github.com/go-chi/chi/v5"
	projectssvc "riders.agency/pulse/api/internal/modules/projects/service"
)

func MountCompanyRoutes(r chi.Router, projects *projectssvc.Commands) {
	r.Get("/companies/{company_id}/projects", List(projects))
	r.Post("/companies/{company_id}/projects", Create(projects))
}

func MountProjectRoutes(r chi.Router, projects *projectssvc.Commands) {
	r.Get("/", Get(projects))
	r.Patch("/", Patch(projects))
	r.Post("/hook-secret/rotate", RotateHookSecret(projects))
}
