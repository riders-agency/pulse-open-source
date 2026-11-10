package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	clientsrepo "riders.agency/pulse/api/internal/modules/clients/repository"
	clientssvc "riders.agency/pulse/api/internal/modules/clients/service"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	"riders.agency/pulse/api/internal/platform/problem"
)

type DaDataFinder interface {
	FindParty(context.Context, uuid.UUID, uuid.UUID, string) (map[string]any, error)
	FindPartyByProject(context.Context, uuid.UUID, uuid.UUID, string) (map[string]any, error)
}

type DaDataIntegrationSaver interface {
	SaveCompanyIntegration(context.Context, uuid.UUID, uuid.UUID, clientsrepo.DaDataIntegrationInput) (*clientsrepo.DaDataIntegrationRow, error)
	SaveProjectIntegration(context.Context, uuid.UUID, uuid.UUID, clientsrepo.DaDataIntegrationInput) (*clientsrepo.DaDataIntegrationRow, error)
}

type DaDataService interface {
	DaDataFinder
	DaDataIntegrationSaver
}

func GetDaDataIntegration(cl *clientssvc.Commands) http.HandlerFunc {
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
		row, err := cl.GetDaDataIntegration(r.Context(), companyID, claims.UserID)
		if writeClientError(w, err) {
			return
		}
		problem.JSON(w, http.StatusOK, row)
	}
}

func GetProjectDaDataIntegration(cl *clientssvc.Commands) http.HandlerFunc {
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
		row, err := cl.GetDaDataProjectIntegration(r.Context(), projectID, claims.UserID)
		if writeClientError(w, err) {
			return
		}
		problem.JSON(w, http.StatusOK, row)
	}
}

func SaveDaDataIntegration(saver DaDataIntegrationSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if saver == nil {
			problem.Write(w, http.StatusInternalServerError, "dadata_saver", "")
			return
		}
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
		var body clientsrepo.DaDataIntegrationInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		row, err := saver.SaveCompanyIntegration(r.Context(), companyID, claims.UserID, body)
		if writeClientError(w, err) {
			return
		}
		problem.JSON(w, http.StatusOK, row)
	}
}

func SaveProjectDaDataIntegration(saver DaDataIntegrationSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if saver == nil {
			problem.Write(w, http.StatusInternalServerError, "dadata_saver", "")
			return
		}
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
		var body clientsrepo.DaDataIntegrationInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		row, err := saver.SaveProjectIntegration(r.Context(), projectID, claims.UserID, body)
		if writeClientError(w, err) {
			return
		}
		problem.JSON(w, http.StatusOK, row)
	}
}

func FindDaDataParty(finder DaDataFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if finder == nil {
			problem.Write(w, http.StatusInternalServerError, "dadata_finder", "")
			return
		}
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
		var body struct {
			Query string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		out, err := finder.FindParty(r.Context(), companyID, claims.UserID, body.Query)
		if writeClientError(w, err) {
			return
		}
		problem.JSON(w, http.StatusOK, out)
	}
}

func FindProjectDaDataParty(finder DaDataFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if finder == nil {
			problem.Write(w, http.StatusInternalServerError, "dadata_finder", "")
			return
		}
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
		var body struct {
			Query string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		out, err := finder.FindPartyByProject(r.Context(), projectID, claims.UserID, body.Query)
		if writeClientError(w, err) {
			return
		}
		problem.JSON(w, http.StatusOK, out)
	}
}
