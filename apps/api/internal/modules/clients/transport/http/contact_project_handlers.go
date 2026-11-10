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

func ListContacts(cl *clientssvc.Commands) http.HandlerFunc {
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
		rows, err := cl.ListContacts(r.Context(), clientID, claims.UserID)
		writeRowsOrError(w, rows, err)
	}
}

func CreateContact(cl *clientssvc.Commands) http.HandlerFunc {
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
		var body clientsrepo.ContactInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		id, err := cl.CreateContact(r.Context(), clientID, claims.UserID, body)
		if writeClientError(w, err) {
			return
		}
		problem.JSON(w, http.StatusCreated, map[string]string{"client_contact_id": id})
	}
}

func PatchContact(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		clientID, contactID, ok := clientContactIDs(w, r)
		if !ok {
			return
		}
		var body clientsrepo.ContactPatch
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		if writeClientError(w, cl.PatchContact(r.Context(), clientID, contactID, claims.UserID, body)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteContact(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		clientID, contactID, ok := clientContactIDs(w, r)
		if !ok {
			return
		}
		if writeClientError(w, cl.DeleteContact(r.Context(), clientID, contactID, claims.UserID)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func ListProjects(cl *clientssvc.Commands) http.HandlerFunc {
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
		rows, err := cl.ListClientProjects(r.Context(), clientID, claims.UserID)
		writeRowsOrError(w, rows, err)
	}
}

func LinkProject(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		clientID, projectID, ok := clientProjectIDs(w, r)
		if !ok {
			return
		}
		if writeClientError(w, cl.LinkProject(r.Context(), clientID, projectID, claims.UserID)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func UnlinkProject(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		clientID, projectID, ok := clientProjectIDs(w, r)
		if !ok {
			return
		}
		if writeClientError(w, cl.UnlinkProject(r.Context(), clientID, projectID, claims.UserID)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
