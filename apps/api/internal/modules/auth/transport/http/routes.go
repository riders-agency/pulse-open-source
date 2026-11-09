package http

import (
	"github.com/go-chi/chi/v5"

	authsvc "riders.agency/pulse/api/internal/modules/auth/service"
)

func MountPublicRoutes(r chi.Router, auth *authsvc.Auth) {
	r.Post("/auth/register", RegisterJSON(auth))
	r.Post("/auth/login", LoginJSON(auth))
	r.Get("/auth/refresh/csrf", RefreshCSRFJSON(auth))
	r.Post("/auth/refresh", RefreshJSON(auth))
}

func MountProtectedRoutes(r chi.Router, auth *authsvc.Auth) {
	r.Get("/auth/csrf", CSRFJSON(auth))
	r.Get("/auth/me", MeJSON(auth))
	r.Get("/auth/workspace-context", WorkspaceContextJSON(auth))
	r.Put("/auth/workspace-context", PutWorkspaceContextJSON(auth))
	r.Post("/auth/logout", LogoutJSON(auth))
	r.Get("/auth/sessions", SessionsJSON(auth))
	r.Delete("/auth/sessions/{session_id}", EndSessionJSON(auth))
	r.Patch("/auth/profile", PatchProfileJSON(auth))
}
