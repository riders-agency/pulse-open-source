package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	docsrepo "riders.agency/pulse/api/internal/modules/integrationdocs/repository"
	"riders.agency/pulse/api/internal/platform/errs"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	problem "riders.agency/pulse/api/internal/platform/problem"
)

func PublicSearch(d *docsrepo.Docs) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := d.Search(r.Context(), docsrepo.SearchParams{
			Query:         r.URL.Query().Get("q"),
			ProviderKey:   r.URL.Query().Get("provider_key"),
			ActionTypeKey: r.URL.Query().Get("action_type_key"),
			Limit:         limitFromQuery(r),
		})
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "search", "")
			return
		}
		problem.JSON(w, http.StatusOK, items)
	}
}

func RawMarkdown(d *docsrepo.Docs) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := strings.TrimSuffix(chi.URLParam(r, "public_slug"), ".md")
		if slug == "" {
			problem.Write(w, http.StatusBadRequest, "public_slug", "")
			return
		}
		body, err := d.MarkdownBySlug(r.Context(), slug)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusNotFound, "not found", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "markdown", "")
			return
		}
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}
}

func ProjectSearch(d *docsrepo.Docs) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		pid, err := uuid.Parse(chi.URLParam(r, "project_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "project_id", "")
			return
		}
		items, err := d.Search(r.Context(), docsrepo.SearchParams{
			Query:         r.URL.Query().Get("q"),
			ProviderKey:   r.URL.Query().Get("provider_key"),
			ActionTypeKey: r.URL.Query().Get("action_type_key"),
			Limit:         limitFromQuery(r),
			ProjectID:     &pid,
			UserID:        &c.UserID,
		})
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "search", "")
			return
		}
		problem.JSON(w, http.StatusOK, items)
	}
}

func limitFromQuery(r *http.Request) int {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		return 8
	}
	if limit > 20 {
		return 20
	}
	return limit
}
