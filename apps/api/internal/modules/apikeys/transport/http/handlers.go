package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	apikeyssvc "riders.agency/pulse/api/internal/modules/apikeys/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	apirepo "riders.agency/pulse/api/internal/modules/apikeys/repository"
	"riders.agency/pulse/api/internal/platform/authz"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	problem "riders.agency/pulse/api/internal/platform/problem"
)

func List(repo *apikeyssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := userSessionOnly(w, r)
		if !ok {
			return
		}
		rows, err := repo.List(r.Context(), c.UserID)
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "api keys list failed", "")
			return
		}
		problem.JSON(w, http.StatusOK, map[string]any{"api_keys": rows})
	}
}

type createBody struct {
	Name      string            `json:"name"`
	Grants    []createGrantBody `json:"grants"`
	ExpiresAt *string           `json:"expires_at"`
}

type createGrantBody struct {
	ScopeKind   string   `json:"scope_kind"`
	ScopeID     string   `json:"scope_id"`
	Permissions []string `json:"permissions"`
}

func Create(repo *apikeyssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := userSessionOnly(w, r)
		if !ok {
			return
		}
		var body createBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "invalid json", "")
			return
		}
		grants, err := parseGrants(body.Grants)
		if err != nil || len(grants) == 0 {
			problem.Write(w, http.StatusBadRequest, "invalid grants", "Add at least one valid grant")
			return
		}
		var expiresAt *time.Time
		if body.ExpiresAt != nil && *body.ExpiresAt != "" {
			t, err := time.Parse(time.RFC3339, *body.ExpiresAt)
			if err != nil {
				problem.Write(w, http.StatusBadRequest, "invalid expires_at", "")
				return
			}
			expiresAt = &t
		}
		row, plaintext, err := repo.Create(r.Context(), c.UserID, body.Name, grants, expiresAt)
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "api key create failed", err.Error())
			return
		}
		problem.JSON(w, http.StatusCreated, map[string]any{
			"api_key":            row,
			"plaintext_api_key":  plaintext,
			"plaintext_key_note": "Shown once. Store it now; only the hash is saved.",
		})
	}
}

func Revoke(repo *apikeyssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := userSessionOnly(w, r)
		if !ok {
			return
		}
		keyID, err := uuid.Parse(chi.URLParam(r, "api_key_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "bad api_key_id", "")
			return
		}
		err = repo.Revoke(r.Context(), c.UserID, keyID)
		if errors.Is(err, apirepo.ErrAPIKeyNotFound) {
			problem.Write(w, http.StatusNotFound, "api key not found", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "api key revoke failed", "")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func userSessionOnly(w http.ResponseWriter, r *http.Request) (mware.SessionClaims, bool) {
	c, ok := mware.SessionClaimsFrom(r.Context())
	if !ok {
		problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
		return c, false
	}
	if c.ActorType == authz.ActorAPIKey {
		problem.Write(w, http.StatusForbidden, "user session required", "")
		return c, false
	}
	return c, true
}

func parseGrants(rows []createGrantBody) ([]apirepo.GrantInput, error) {
	out := make([]apirepo.GrantInput, 0, len(rows))
	for _, row := range rows {
		grant := apirepo.GrantInput{
			ScopeKind:   row.ScopeKind,
			Permissions: row.Permissions,
		}
		if row.ScopeKind != authz.ScopeSystem {
			id, err := uuid.Parse(row.ScopeID)
			if err != nil || id == uuid.Nil {
				return nil, err
			}
			grant.ScopeID = id
		}
		out = append(out, grant)
	}
	return out, nil
}
