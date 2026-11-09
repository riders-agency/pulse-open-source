package auth

import (
	"github.com/go-chi/chi/v5"

	authsvc "riders.agency/pulse/api/internal/modules/auth/service"
	authhttp "riders.agency/pulse/api/internal/modules/auth/transport/http"
	"riders.agency/pulse/api/internal/platform/httpserver"
)

func RegisterAPIV1(scope httpserver.APIV1Scope, deps httpserver.Dependencies) {
	switch scope.Kind {
	case httpserver.APIV1Public:
		MountPublicHTTP(scope.Router, deps.Auth)
	case httpserver.APIV1Protected:
		MountProtectedHTTP(scope.Router, deps.Auth)
	}
}

func MountPublicHTTP(r chi.Router, auth *authsvc.Auth) {
	authhttp.MountPublicRoutes(r, auth)
}

func MountProtectedHTTP(r chi.Router, auth *authsvc.Auth) {
	authhttp.MountProtectedRoutes(r, auth)
}
