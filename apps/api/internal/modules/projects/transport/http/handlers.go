package http

import (
	"encoding/json"
	"errors"
	"net/http"

	projectssvc "riders.agency/pulse/api/internal/modules/projects/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"riders.agency/pulse/api/internal/platform/errs"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	problem "riders.agency/pulse/api/internal/platform/problem"
)

func List(proj *projectssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		cidStr := chi.URLParam(r, "company_id")
		cid, err := uuid.Parse(cidStr)
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "company_id", "")
			return
		}
		rows, err := proj.List(r.Context(), cid, c.UserID)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "membership required", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "list projects", "")
			return
		}
		problem.JSON(w, http.StatusOK, rows)
	}
}

type createProjectIn struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func Create(proj *projectssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		cidStr := chi.URLParam(r, "company_id")
		cid, err := uuid.Parse(cidStr)
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "company_id", "")
			return
		}
		var body createProjectIn
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		if body.Name == "" || body.Slug == "" {
			problem.Write(w, http.StatusBadRequest, "name and slug required", "")
			return
		}
		row, secret, err := proj.Insert(r.Context(), cid, c.UserID, body.Name, body.Slug)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "membership required", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusConflict, "create project", "")
			return
		}
		problem.JSON(w, http.StatusCreated, map[string]any{
			"project_id":       row.ID,
			"name":             row.Name,
			"slug":             row.Slug,
			"hook_key":         row.HookKey,
			"hook_secret":      secret,
			"hook_secret_note": "Shown once; store securely",
		})
	}
}

func Get(proj *projectssvc.Commands) http.HandlerFunc {
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
		row, err := proj.Get(r.Context(), pid, c.UserID)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "get", "")
			return
		}
		problem.JSON(w, http.StatusOK, row)
	}
}

type patchProjectIn struct {
	Name *string `json:"name"`
	Slug *string `json:"slug"`
}

func Patch(proj *projectssvc.Commands) http.HandlerFunc {
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
		var body patchProjectIn
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		err = proj.Patch(r.Context(), pid, c.UserID, body.Name, body.Slug)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusConflict, "patch", "")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func RotateHookSecret(proj *projectssvc.Commands) http.HandlerFunc {
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
		secret, err := proj.RotateHookSecret(r.Context(), pid, c.UserID)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "rotate", "")
			return
		}
		problem.JSON(w, http.StatusOK, map[string]string{
			"hook_secret": secret,
			"note":        "Shown once; previous secret invalidated",
		})
	}
}
