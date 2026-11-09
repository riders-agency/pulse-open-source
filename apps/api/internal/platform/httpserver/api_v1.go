package httpserver

import (
	"github.com/go-chi/chi/v5"

	mid "riders.agency/pulse/api/internal/platform/httpserver/middleware"
)

type APIV1ScopeKind string

const (
	APIV1Public    APIV1ScopeKind = "public"
	APIV1Protected APIV1ScopeKind = "protected"
	APIV1Work      APIV1ScopeKind = "work"
	APIV1Project   APIV1ScopeKind = "project"
)

type APIV1Scope struct {
	Kind   APIV1ScopeKind
	Router chi.Router
}

type APIV1Modules interface {
	RegisterAPIV1(scope APIV1Scope, deps Dependencies)
}

func MountAPIV1(r chi.Router, deps Dependencies, modules APIV1Modules) {
	if modules == nil {
		return
	}
	r.Route("/api/v1", func(ar chi.Router) {
		ar.Use(mid.OptionalAuth(deps.Pool))
		modules.RegisterAPIV1(APIV1Scope{Kind: APIV1Public, Router: ar}, deps)

		ar.Group(func(pr chi.Router) {
			pr.Use(mid.RequireAuth(deps.Pool))
			pr.Use(mid.RequireBrowserCSRF)
			modules.RegisterAPIV1(APIV1Scope{Kind: APIV1Protected, Router: pr}, deps)

			pr.Route("/work", func(wr chi.Router) {
				modules.RegisterAPIV1(APIV1Scope{Kind: APIV1Work, Router: wr}, deps)
			})

			pr.Route("/projects/{project_id}", func(sr chi.Router) {
				modules.RegisterAPIV1(APIV1Scope{Kind: APIV1Project, Router: sr}, deps)
			})
		})
	})
}
