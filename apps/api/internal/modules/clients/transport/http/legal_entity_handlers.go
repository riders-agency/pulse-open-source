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

func ListCompanyLegalEntities(cl *clientssvc.Commands) http.HandlerFunc {
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
		rows, err := cl.ListCompanyLegalEntities(r.Context(), companyID, claims.UserID)
		writeRowsOrError(w, rows, err)
	}
}

func CreateCompanyLegalEntity(cl *clientssvc.Commands) http.HandlerFunc {
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
		var body clientsrepo.CompanyLegalEntityInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		id, err := cl.CreateCompanyLegalEntity(r.Context(), companyID, claims.UserID, body)
		if writeClientError(w, err) {
			return
		}
		problem.JSON(w, http.StatusCreated, map[string]string{"company_legal_entity_id": id})
	}
}

func GetCompanyLegalEntity(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		entityID, err := uuid.Parse(chi.URLParam(r, "company_legal_entity_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "company_legal_entity_id", "")
			return
		}
		row, err := cl.GetCompanyLegalEntity(r.Context(), entityID, claims.UserID)
		if writeClientError(w, err) {
			return
		}
		problem.JSON(w, http.StatusOK, row)
	}
}

func PatchCompanyLegalEntity(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		entityID, err := uuid.Parse(chi.URLParam(r, "company_legal_entity_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "company_legal_entity_id", "")
			return
		}
		var body clientsrepo.CompanyLegalEntityPatch
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		if writeClientError(w, cl.PatchCompanyLegalEntity(r.Context(), entityID, claims.UserID, body)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteCompanyLegalEntity(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		entityID, err := uuid.Parse(chi.URLParam(r, "company_legal_entity_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "company_legal_entity_id", "")
			return
		}
		if writeClientError(w, cl.DeleteCompanyLegalEntity(r.Context(), entityID, claims.UserID)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
