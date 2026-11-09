package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	tpl "riders.agency/pulse/api/internal/modules/templates/repository"
	"riders.agency/pulse/api/internal/platform/errs"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	problem "riders.agency/pulse/api/internal/platform/problem"
)

func ListEmail(t *tpl.EmailTemplates) http.HandlerFunc {
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
		rows, err := t.List(r.Context(), pid, c.UserID)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "list", "")
			return
		}
		problem.JSON(w, http.StatusOK, rows)
	}
}

func CreateEmail(t *tpl.EmailTemplates) http.HandlerFunc {
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
		var in tpl.EmailCreate
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		if in.Name == "" || in.Slug == "" || in.SubjectTemplate == "" || in.BodyTemplate == "" {
			problem.Write(w, http.StatusBadRequest, "required fields", "")
			return
		}
		id, err := t.Create(r.Context(), pid, c.UserID, in)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusConflict, "create", "")
			return
		}
		problem.JSON(w, http.StatusCreated, map[string]string{"email_template_id": id})
	}
}

func GetEmail(t *tpl.EmailTemplates) http.HandlerFunc {
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
		tid, err := uuid.Parse(chi.URLParam(r, "email_template_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "email_template_id", "")
			return
		}
		row, err := t.Get(r.Context(), pid, tid, c.UserID)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusNotFound, "not found", "")
			return
		}
		problem.JSON(w, http.StatusOK, row)
	}
}

func PatchEmail(t *tpl.EmailTemplates) http.HandlerFunc {
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
		tid, err := uuid.Parse(chi.URLParam(r, "email_template_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "email_template_id", "")
			return
		}
		var body tpl.EmailPatch
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		if err := t.Patch(r.Context(), pid, tid, c.UserID, body); err != nil {
			if errors.Is(err, errs.ErrForbidden) {
				problem.Write(w, http.StatusForbidden, "forbidden", "")
				return
			}
			problem.Write(w, http.StatusConflict, "patch", "")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteEmail(t *tpl.EmailTemplates) http.HandlerFunc {
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
		tid, err := uuid.Parse(chi.URLParam(r, "email_template_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "email_template_id", "")
			return
		}
		if err := t.Delete(r.Context(), pid, tid, c.UserID); err != nil {
			if errors.Is(err, errs.ErrForbidden) {
				problem.Write(w, http.StatusForbidden, "forbidden", "")
				return
			}
			problem.Write(w, http.StatusBadRequest, "delete", "")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
