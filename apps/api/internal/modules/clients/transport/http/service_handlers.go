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

func ListServices(cl *clientssvc.Commands) http.HandlerFunc {
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
		rows, err := cl.ListServices(r.Context(), clientID, claims.UserID, projectID)
		writeRowsOrError(w, rows, err)
	}
}

func CreateService(cl *clientssvc.Commands) http.HandlerFunc {
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
		var body clientsrepo.ServiceInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		id, err := cl.CreateService(r.Context(), clientID, claims.UserID, body)
		if writeClientError(w, err) {
			return
		}
		problem.JSON(w, http.StatusCreated, map[string]string{"client_service_id": id})
	}
}

func GetService(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		serviceID, err := uuid.Parse(chi.URLParam(r, "client_service_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "client_service_id", "")
			return
		}
		row, err := cl.GetService(r.Context(), serviceID, claims.UserID)
		if writeClientError(w, err) {
			return
		}
		problem.JSON(w, http.StatusOK, row)
	}
}

func PatchService(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		serviceID, err := uuid.Parse(chi.URLParam(r, "client_service_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "client_service_id", "")
			return
		}
		var body clientsrepo.ServicePatch
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		if writeClientError(w, cl.PatchService(r.Context(), serviceID, claims.UserID, body)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteService(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		serviceID, err := uuid.Parse(chi.URLParam(r, "client_service_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "client_service_id", "")
			return
		}
		if writeClientError(w, cl.DeleteService(r.Context(), serviceID, claims.UserID)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
