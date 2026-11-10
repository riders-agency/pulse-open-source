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

func ListMessage(t *tpl.MessageTemplates) http.HandlerFunc {
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

func CreateMessage(t *tpl.MessageTemplates) http.HandlerFunc {
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
		var in tpl.MessageCreate
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		if in.Name == "" || in.Slug == "" || in.ProviderKey == "" || in.BodyTemplate == "" {
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
		problem.JSON(w, http.StatusCreated, map[string]string{"message_template_id": id})
	}
}

func GetMessage(t *tpl.MessageTemplates) http.HandlerFunc {
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
		mid, err := uuid.Parse(chi.URLParam(r, "message_template_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "message_template_id", "")
			return
		}
		row, err := t.Get(r.Context(), pid, mid, c.UserID)
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

func PatchMessage(t *tpl.MessageTemplates) http.HandlerFunc {
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
		mid, err := uuid.Parse(chi.URLParam(r, "message_template_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "message_template_id", "")
			return
		}
		var body tpl.MessagePatch
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		if err := t.Patch(r.Context(), pid, mid, c.UserID, body); err != nil {
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

func DeleteMessage(t *tpl.MessageTemplates) http.HandlerFunc {
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
		mid, err := uuid.Parse(chi.URLParam(r, "message_template_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "message_template_id", "")
			return
		}
		if err := t.Delete(r.Context(), pid, mid, c.UserID); err != nil {
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
