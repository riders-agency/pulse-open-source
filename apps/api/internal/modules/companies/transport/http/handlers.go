package http

//go:generate go run github.com/vektra/mockery/v2@latest --name=Companies --filename=companies_mock.go --output=../../mocks --outpkg=mocks
import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	companiessvc "riders.agency/pulse/api/internal/modules/companies/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	companysvc "riders.agency/pulse/api/internal/modules/companies/repository"
	"riders.agency/pulse/api/internal/platform/errs"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	problem "riders.agency/pulse/api/internal/platform/problem"
)

func List(repo *companiessvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		rows, err := repo.List(r.Context(), c.UserID)
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "query", "")
			return
		}
		problem.JSON(w, http.StatusOK, rows)
	}
}

type createIn struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func Create(repo *companiessvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		var body createIn
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		body.Name = strings.TrimSpace(body.Name)
		body.Slug = strings.TrimSpace(body.Slug)
		if body.Name == "" || body.Slug == "" {
			problem.Write(w, http.StatusBadRequest, "name and slug required", "")
			return
		}
		id := uuid.New()
		if err := repo.Insert(r.Context(), id, c.UserID, body.Slug, body.Name); err != nil {
			if errors.Is(err, companysvc.ErrCompanySlugTaken) {
				problem.Write(w, http.StatusConflict, "company slug already exists", "Change the slug and retry")
				return
			}
			if errors.Is(err, errs.ErrForbidden) {
				problem.Write(w, http.StatusForbidden, "forbidden", "")
				return
			}
			problem.Write(w, http.StatusConflict, "create company", "")
			return
		}
		problem.JSON(w, http.StatusCreated, map[string]string{"company_id": id.String()})
	}
}

func Get(repo *companiessvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		cid, err := uuid.Parse(chi.URLParam(r, "company_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "company_id", "")
			return
		}
		row, err := repo.GetByID(r.Context(), cid, c.UserID)
		if err != nil {
			problem.Write(w, http.StatusNotFound, "not found", "")
			return
		}
		problem.JSON(w, http.StatusOK, row)
	}
}

type patchCompanyIn struct {
	Name *string `json:"name"`
	Slug *string `json:"slug"`
}

func Patch(repo *companiessvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		cid, err := uuid.Parse(chi.URLParam(r, "company_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "company_id", "")
			return
		}
		var body patchCompanyIn
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		err = repo.Patch(r.Context(), cid, c.UserID, body.Name, body.Slug)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusNotFound, "not found", "")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func Switch(repo *companiessvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		cidStr := chi.URLParam(r, "company_id")
		cid, err := uuid.Parse(cidStr)
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "company_id", "")
			return
		}
		err = repo.Switch(r.Context(), c.UserID, cid)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "not a member", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "switch", "")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
