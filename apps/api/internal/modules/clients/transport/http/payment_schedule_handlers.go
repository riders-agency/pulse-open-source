package http

import (
	"encoding/json"
	"net/http"

	clientssvc "riders.agency/pulse/api/internal/modules/clients/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	clientsrepo "riders.agency/pulse/api/internal/modules/clients/repository"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	problem "riders.agency/pulse/api/internal/platform/problem"
)

func ProjectPaymentSchedule(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		projectID, ok := projectIDParam(w, r)
		if !ok {
			return
		}
		rows, err := cl.ListProjectPaymentSchedule(r.Context(), projectID, claims.UserID)
		writeRowsOrError(w, rows, err)
	}
}

func CreateManualPaymentAction(cl *clientssvc.Commands, eng RunNotifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		projectID, ok := projectIDParam(w, r)
		if !ok {
			return
		}
		var body clientsrepo.CreateManualPaymentActionInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		id, err := cl.CreateManualPaymentAction(r.Context(), projectID, claims.UserID, body)
		if writeClientError(w, err) {
			return
		}
		if eng != nil {
			eng.Notify()
		}
		problem.JSON(w, http.StatusCreated, map[string]string{"payment_action_id": id})
	}
}

func projectIDParam(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	projectID, err := uuid.Parse(chi.URLParam(r, "project_id"))
	if err != nil {
		problem.Write(w, http.StatusBadRequest, "project_id", "")
		return uuid.Nil, false
	}
	return projectID, true
}
