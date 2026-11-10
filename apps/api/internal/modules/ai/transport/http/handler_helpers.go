package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	airepo "riders.agency/pulse/api/internal/modules/ai/repository"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	"riders.agency/pulse/api/internal/platform/problem"
)

func ids(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	c, ok := mware.SessionClaimsFrom(r.Context())
	if !ok {
		problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
		return uuid.Nil, uuid.Nil, false
	}
	projectID, err := uuid.Parse(chi.URLParam(r, "project_id"))
	if err != nil {
		problem.Write(w, http.StatusBadRequest, "project_id", "")
		return uuid.Nil, uuid.Nil, false
	}
	return projectID, c.UserID, true
}

func decode(w http.ResponseWriter, r *http.Request, out any) bool {
	if err := json.NewDecoder(r.Body).Decode(out); err != nil {
		problem.Write(w, http.StatusBadRequest, "json", "")
		return false
	}
	return true
}

func write(w http.ResponseWriter, status int, value any, err error) {
	if err != nil {
		switch {
		case airepo.IsForbidden(err):
			problem.Write(w, http.StatusForbidden, "forbidden", "")
		case airepo.IsNotFound(err):
			problem.Write(w, http.StatusNotFound, "not_found", "")
		default:
			problem.Write(w, http.StatusInternalServerError, "ai", err.Error())
		}
		return
	}
	problem.JSON(w, status, value)
}
