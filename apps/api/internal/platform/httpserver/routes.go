package httpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func mountRoutes(r chi.Router, deps Dependencies, modules HTTPModules) {
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	if modules != nil {
		modules.RegisterHTTP(r, deps)
	}
}
