package http

import (
	"net/http"

	cat "riders.agency/pulse/api/internal/modules/catalog/repository"
	providers "riders.agency/pulse/api/internal/modules/providers"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	problem "riders.agency/pulse/api/internal/platform/problem"

	"github.com/go-chi/chi/v5"
)

func ListActionTypes(c *cat.Catalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := mware.SessionClaimsFrom(r.Context()); !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		rows, err := c.ListActionTypes(r.Context())
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "catalog", "")
			return
		}
		problem.JSON(w, http.StatusOK, rows)
	}
}

func ProviderCatalog() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := mware.SessionClaimsFrom(r.Context()); !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		problem.JSON(w, http.StatusOK, providers.Catalog())
	}
}

func GetActionType(c *cat.Catalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := mware.SessionClaimsFrom(r.Context()); !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		k := chi.URLParam(r, "action_type_key")
		row, err := c.GetActionType(r.Context(), k)
		if err != nil {
			problem.Write(w, http.StatusNotFound, "not found", "")
			return
		}
		problem.JSON(w, http.StatusOK, row)
	}
}

func ListDepTypes(c *cat.Catalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := mware.SessionClaimsFrom(r.Context()); !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		rows, err := c.ListDepTypes(r.Context())
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "catalog", "")
			return
		}
		problem.JSON(w, http.StatusOK, rows)
	}
}

func GetDepType(c *cat.Catalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := mware.SessionClaimsFrom(r.Context()); !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		k := chi.URLParam(r, "dependency_type_key")
		row, err := c.GetDepType(r.Context(), k)
		if err != nil {
			problem.Write(w, http.StatusNotFound, "not found", "")
			return
		}
		problem.JSON(w, http.StatusOK, row)
	}
}
