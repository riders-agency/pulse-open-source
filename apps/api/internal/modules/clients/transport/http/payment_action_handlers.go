package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	clientsrepo "riders.agency/pulse/api/internal/modules/clients/repository"
	clientssvc "riders.agency/pulse/api/internal/modules/clients/service"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	"riders.agency/pulse/api/internal/platform/problem"
)

func ProjectPaymentActions(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		projectID, err := uuid.Parse(chi.URLParam(r, "project_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "project_id", "")
			return
		}
		rows, err := cl.ListPaymentActions(r.Context(), projectID, claims.UserID)
		writeRowsOrError(w, rows, err)
	}
}

func ClientPaymentActions(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		clientID, err := uuid.Parse(chi.URLParam(r, "client_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "client_id", "")
			return
		}
		var projectID *uuid.UUID
		if value := r.URL.Query().Get("project_id"); value != "" {
			parsed, err := uuid.Parse(value)
			if err != nil {
				problem.Write(w, http.StatusBadRequest, "project_id", "")
				return
			}
			projectID = &parsed
		}
		rows, err := cl.ListClientPaymentActions(r.Context(), clientID, claims.UserID, projectID)
		writeRowsOrError(w, rows, err)
	}
}

func RecordManualPayment(cl *clientssvc.Commands, eng RunNotifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		clientID, err := uuid.Parse(chi.URLParam(r, "client_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "client_id", "")
			return
		}
		var body clientsrepo.ManualPaymentRecordInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		if body.IdempotencyKey == "" {
			body.IdempotencyKey = r.Header.Get("Idempotency-Key")
		}
		out, err := cl.RecordManualPayment(r.Context(), clientID, claims.UserID, body)
		if writeClientError(w, err) {
			return
		}
		if eng != nil && out != nil && !out.Duplicate {
			eng.Notify()
		}
		status := http.StatusCreated
		if out != nil && out.Duplicate {
			status = http.StatusOK
		}
		problem.JSON(w, status, out)
	}
}

func GetPaymentAction(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		actionID, err := uuid.Parse(chi.URLParam(r, "payment_action_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "payment_action_id", "")
			return
		}
		row, err := cl.GetPaymentAction(r.Context(), actionID, claims.UserID)
		if writeClientError(w, err) {
			return
		}
		problem.JSON(w, http.StatusOK, row)
	}
}

func ConfirmPaymentAction(cl *clientssvc.Commands, eng RunNotifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		actionID, err := uuid.Parse(chi.URLParam(r, "payment_action_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "payment_action_id", "")
			return
		}
		var body clientsrepo.ConfirmPaymentInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil && r.ContentLength != 0 {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		if writeClientError(w, cl.ConfirmPaymentAction(r.Context(), actionID, claims.UserID, body)) {
			return
		}
		if eng != nil {
			eng.Notify()
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func SnoozePaymentAction(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		actionID, err := uuid.Parse(chi.URLParam(r, "payment_action_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "payment_action_id", "")
			return
		}
		var body clientsrepo.SnoozePaymentInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil && r.ContentLength != 0 {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		if writeClientError(w, cl.SnoozePaymentAction(r.Context(), actionID, claims.UserID, body)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
