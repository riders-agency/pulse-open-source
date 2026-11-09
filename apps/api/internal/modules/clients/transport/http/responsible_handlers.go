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

func ListResponsibles(cl *clientssvc.Commands) http.HandlerFunc {
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
		rows, err := cl.ListResponsibles(r.Context(), clientID, claims.UserID)
		writeRowsOrError(w, rows, err)
	}
}

func CreateResponsible(cl *clientssvc.Commands) http.HandlerFunc {
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
		var body clientsrepo.ClientResponsibleInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		id, err := cl.CreateResponsible(r.Context(), clientID, claims.UserID, body)
		if writeClientError(w, err) {
			return
		}
		problem.JSON(w, http.StatusCreated, map[string]string{"client_responsible_id": id})
	}
}

func PatchResponsible(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		clientID, responsibleID, ok := clientResponsibleIDs(w, r)
		if !ok {
			return
		}
		var body clientsrepo.ClientResponsiblePatch
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		if writeClientError(w, cl.PatchResponsible(r.Context(), clientID, responsibleID, claims.UserID, body)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteResponsible(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		clientID, responsibleID, ok := clientResponsibleIDs(w, r)
		if !ok {
			return
		}
		if writeClientError(w, cl.DeleteResponsible(r.Context(), clientID, responsibleID, claims.UserID)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
