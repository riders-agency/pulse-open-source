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

func ListServiceTypes(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := cl.ListServiceTypes(r.Context())
		writeRowsOrError(w, rows, err)
	}
}

func ListProjectServiceTypes(cl *clientssvc.Commands) http.HandlerFunc {
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
		rows, err := cl.ListProjectServiceTypes(r.Context(), projectID, claims.UserID)
		writeRowsOrError(w, rows, err)
	}
}

func CreateProjectServiceType(cl *clientssvc.Commands) http.HandlerFunc {
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
		var body clientsrepo.ServiceTypeInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		id, err := cl.CreateProjectServiceType(r.Context(), projectID, claims.UserID, body)
		if writeClientError(w, err) {
			return
		}
		problem.JSON(w, http.StatusCreated, map[string]string{"client_service_type_id": id})
	}
}

func PatchProjectServiceType(cl *clientssvc.Commands) http.HandlerFunc {
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
		serviceTypeID, err := uuid.Parse(chi.URLParam(r, "client_service_type_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "client_service_type_id", "")
			return
		}
		var body clientsrepo.ServiceTypePatch
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		if writeClientError(w, cl.PatchProjectServiceType(r.Context(), projectID, serviceTypeID, claims.UserID, body)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func ArchiveProjectServiceType(cl *clientssvc.Commands) http.HandlerFunc {
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
		serviceTypeID, err := uuid.Parse(chi.URLParam(r, "client_service_type_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "client_service_type_id", "")
			return
		}
		if writeClientError(w, cl.ArchiveProjectServiceType(r.Context(), projectID, serviceTypeID, claims.UserID)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
