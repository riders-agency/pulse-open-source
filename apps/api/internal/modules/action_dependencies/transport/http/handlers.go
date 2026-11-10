package http

import (
	"encoding/json"
	"errors"
	"net/http"

	actiondependenciessvc "riders.agency/pulse/api/internal/modules/action_dependencies/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	deprepo "riders.agency/pulse/api/internal/modules/action_dependencies/repository"
	"riders.agency/pulse/api/internal/platform/errs"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	problem "riders.agency/pulse/api/internal/platform/problem"
)

func List(d *actiondependenciessvc.Commands) http.HandlerFunc {
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
		items, err := d.List(r.Context(), pid, c.UserID, r.URL.Query().Get("include_archived") == "true")
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "list", "")
			return
		}
		problem.JSON(w, http.StatusOK, items)
	}
}

func Create(d *actiondependenciessvc.Commands) http.HandlerFunc {
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
		var in deprepo.DepCreate
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		if in.Name == "" || in.DependencyTypeKey == "" {
			problem.Write(w, http.StatusBadRequest, "required", "")
			return
		}
		id, err := d.Create(r.Context(), pid, c.UserID, in)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if errors.Is(err, deprepo.ErrSMTPMetadata) {
			problem.Write(w, http.StatusBadRequest, "metadata_json", err.Error())
			return
		}
		if errors.Is(err, deprepo.ErrTelegramRelayMetadata) {
			problem.Write(w, http.StatusBadRequest, "metadata_json", err.Error())
			return
		}
		if err != nil {
			problem.Write(w, http.StatusConflict, "create", "")
			return
		}
		problem.JSON(w, http.StatusCreated, map[string]string{"action_dependency_id": id})
	}
}

func Get(d *actiondependenciessvc.Commands) http.HandlerFunc {
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
		did, err := uuid.Parse(chi.URLParam(r, "action_dependency_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "action_dependency_id", "")
			return
		}
		row, err := d.Get(r.Context(), pid, did, c.UserID)
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

func Patch(d *actiondependenciessvc.Commands) http.HandlerFunc {
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
		did, err := uuid.Parse(chi.URLParam(r, "action_dependency_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "action_dependency_id", "")
			return
		}
		var body deprepo.DepPatch
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		if err := d.Patch(r.Context(), pid, did, c.UserID, body); err != nil {
			if errors.Is(err, errs.ErrForbidden) {
				problem.Write(w, http.StatusForbidden, "forbidden", "")
				return
			}
			if errors.Is(err, deprepo.ErrSMTPMetadata) {
				problem.Write(w, http.StatusBadRequest, "metadata_json", err.Error())
				return
			}
			if errors.Is(err, deprepo.ErrTelegramRelayMetadata) {
				problem.Write(w, http.StatusBadRequest, "metadata_json", err.Error())
				return
			}
			problem.Write(w, http.StatusBadRequest, "patch", "")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func Delete(d *actiondependenciessvc.Commands) http.HandlerFunc {
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
		did, err := uuid.Parse(chi.URLParam(r, "action_dependency_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "action_dependency_id", "")
			return
		}
		if err := d.Delete(r.Context(), pid, did, c.UserID); err != nil {
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

type rotateDepBody struct {
	PlaintextSecret string `json:"plaintext_secret"`
}

type relayPollingBody struct {
	Enabled            bool  `json:"enabled"`
	TimeoutSeconds     int   `json:"timeout_seconds"`
	Limit              int   `json:"limit"`
	DeleteWebhook      *bool `json:"delete_webhook"`
	DropPendingUpdates bool  `json:"drop_pending_updates"`
}

func RotateSecret(d *actiondependenciessvc.Commands) http.HandlerFunc {
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
		did, err := uuid.Parse(chi.URLParam(r, "action_dependency_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "action_dependency_id", "")
			return
		}
		var body rotateDepBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		if err := d.RotateSecret(r.Context(), pid, did, c.UserID, body.PlaintextSecret); err != nil {
			if errors.Is(err, errs.ErrForbidden) {
				problem.Write(w, http.StatusForbidden, "forbidden", "")
				return
			}
			if errors.Is(err, deprepo.ErrPlaintextRequired) {
				problem.Write(w, http.StatusBadRequest, "plaintext_secret", "")
				return
			}
			problem.Write(w, http.StatusConflict, "rotate", "")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func RotateTelegramRelayWebhookSecret(d *actiondependenciessvc.Commands) http.HandlerFunc {
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
		did, err := uuid.Parse(chi.URLParam(r, "action_dependency_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "action_dependency_id", "")
			return
		}
		out, err := d.RotateTelegramRelayWebhookSecret(r.Context(), pid, did, c.UserID)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if errors.Is(err, deprepo.ErrTelegramRelayMetadata) {
			problem.Write(w, http.StatusBadRequest, "telegram_relay", err.Error())
			return
		}
		if err != nil {
			problem.Write(w, http.StatusConflict, "rotate relay webhook secret", "")
			return
		}
		problem.JSON(w, http.StatusOK, out)
	}
}

func SetTelegramRelayPolling(d *actiondependenciessvc.Commands) http.HandlerFunc {
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
		did, err := uuid.Parse(chi.URLParam(r, "action_dependency_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "action_dependency_id", "")
			return
		}
		var body relayPollingBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		deleteWebhook := true
		if body.DeleteWebhook != nil {
			deleteWebhook = *body.DeleteWebhook
		}
		out, err := d.SetTelegramRelayPolling(r.Context(), pid, did, c.UserID, deprepo.TelegramRelayPollingPatch{
			Enabled:            body.Enabled,
			TimeoutSeconds:     body.TimeoutSeconds,
			Limit:              body.Limit,
			DeleteWebhook:      deleteWebhook,
			DropPendingUpdates: body.DropPendingUpdates,
		})
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if errors.Is(err, deprepo.ErrTelegramRelayMetadata) {
			problem.Write(w, http.StatusBadRequest, "telegram_relay", err.Error())
			return
		}
		if err != nil {
			problem.Write(w, http.StatusConflict, "telegram relay polling", "")
			return
		}
		problem.JSON(w, http.StatusOK, out)
	}
}

func Test(d *actiondependenciessvc.Commands) http.HandlerFunc {
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
		did, err := uuid.Parse(chi.URLParam(r, "action_dependency_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "action_dependency_id", "")
			return
		}
		out, err := d.TestConnection(r.Context(), pid, did, c.UserID)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "test", "")
			return
		}
		problem.JSON(w, http.StatusOK, out)
	}
}
