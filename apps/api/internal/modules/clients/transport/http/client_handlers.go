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

func List(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		companyID, err := uuid.Parse(chi.URLParam(r, "company_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "company_id", "")
			return
		}
		rows, err := cl.List(r.Context(), companyID, claims.UserID, r.URL.Query().Get("search"))
		writeRowsOrError(w, rows, err)
	}
}

func ListByProject(cl *clientssvc.Commands) http.HandlerFunc {
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
		rows, err := cl.ListByProject(r.Context(), projectID, claims.UserID)
		writeRowsOrError(w, rows, err)
	}
}

func Create(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		companyID, err := uuid.Parse(chi.URLParam(r, "company_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "company_id", "")
			return
		}
		var body clientsrepo.ClientInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		id, err := cl.Create(r.Context(), companyID, claims.UserID, body)
		if writeClientError(w, err) {
			return
		}
		problem.JSON(w, http.StatusCreated, map[string]string{"client_id": id})
	}
}

func CreateForProject(cl *clientssvc.Commands) http.HandlerFunc {
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
		var body clientsrepo.ClientInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		id, err := cl.CreateForProject(r.Context(), projectID, claims.UserID, body)
		if writeClientError(w, err) {
			return
		}
		problem.JSON(w, http.StatusCreated, map[string]string{"client_id": id})
	}
}

func Get(cl *clientssvc.Commands) http.HandlerFunc {
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
		row, err := cl.Get(r.Context(), clientID, claims.UserID)
		if writeClientError(w, err) {
			return
		}
		problem.JSON(w, http.StatusOK, row)
	}
}

func Patch(cl *clientssvc.Commands) http.HandlerFunc {
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
		var body clientsrepo.ClientPatch
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		if writeClientError(w, cl.Patch(r.Context(), clientID, claims.UserID, body)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func Delete(cl *clientssvc.Commands) http.HandlerFunc {
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
		if writeClientError(w, cl.Delete(r.Context(), clientID, claims.UserID)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
