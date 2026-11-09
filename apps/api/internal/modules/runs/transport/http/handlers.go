package http

import (
	"encoding/json"
	"errors"
	"net/http"

	runssvc "riders.agency/pulse/api/internal/modules/runs/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	runsrepo "riders.agency/pulse/api/internal/modules/runs/repository"
	"riders.agency/pulse/api/internal/platform/errs"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	problem "riders.agency/pulse/api/internal/platform/problem"
)

func Manual(ro *runssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		wfid, err := uuid.Parse(chi.URLParam(r, "workflow_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "workflow_id", "")
			return
		}
		var body struct {
			Input map[string]any `json:"input_json"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil && r.ContentLength != 0 {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		rid, err := ro.Manual(r.Context(), wfid, c.UserID, body.Input)
		if errors.Is(err, runsrepo.ErrNeedsPublish) {
			problem.Write(w, http.StatusBadRequest, runsrepo.ErrNeedsPublish.Error(), "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		problem.JSON(w, http.StatusAccepted, map[string]string{"workflow_run_id": rid})
	}
}

func TestRun(ro *runssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		wfid, err := uuid.Parse(chi.URLParam(r, "workflow_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "workflow_id", "")
			return
		}
		var body runsrepo.TestRunRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil && r.ContentLength != 0 {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		rid, err := ro.Test(r.Context(), wfid, c.UserID, body)
		if errors.Is(err, runsrepo.ErrInvalidTestMode) {
			problem.Write(w, http.StatusBadRequest, runsrepo.ErrInvalidTestMode.Error(), "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "test_run", err.Error())
			return
		}
		problem.JSON(w, http.StatusAccepted, map[string]string{"workflow_run_id": rid})
	}
}

func ListByWorkflow(ro *runssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		wfid, err := uuid.Parse(chi.URLParam(r, "workflow_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "workflow_id", "")
			return
		}
		rows, err := ro.ListByWorkflow(r.Context(), wfid, c.UserID)
		if err != nil {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		problem.JSON(w, http.StatusOK, rows)
	}
}

func Get(ro *runssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		rid, err := uuid.Parse(chi.URLParam(r, "workflow_run_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "workflow_run_id", "")
			return
		}
		row, err := ro.Get(r.Context(), rid, c.UserID)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusNotFound, "not found", "")
			return
		}
		problem.JSON(w, http.StatusOK, row)
	}
}

func Graph(ro *runssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		rid, err := uuid.Parse(chi.URLParam(r, "workflow_run_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "workflow_run_id", "")
			return
		}
		g, err := ro.Graph(r.Context(), rid, c.UserID)
		if err != nil {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		problem.JSON(w, http.StatusOK, g)
	}
}

func Actions(ro *runssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		rid, err := uuid.Parse(chi.URLParam(r, "workflow_run_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "workflow_run_id", "")
			return
		}
		ar, err := ro.ActionRuns(r.Context(), rid, c.UserID)
		if err != nil {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		problem.JSON(w, http.StatusOK, ar)
	}
}

func Events(ro *runssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		rid, err := uuid.Parse(chi.URLParam(r, "workflow_run_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "workflow_run_id", "")
			return
		}
		ev, err := ro.Events(r.Context(), rid, c.UserID)
		if err != nil {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		problem.JSON(w, http.StatusOK, ev)
	}
}

func Stream(ro *runssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		rid, err := uuid.Parse(chi.URLParam(r, "workflow_run_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "workflow_run_id", "")
			return
		}
		if err := streamRunEvents(ro, w, r, rid, c.UserID); err != nil {
			problem.Write(w, http.StatusBadRequest, "stream", "")
		}
	}
}

func Cancel(ro *runssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		rid, err := uuid.Parse(chi.URLParam(r, "workflow_run_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "workflow_run_id", "")
			return
		}
		if err := ro.Cancel(r.Context(), rid, c.UserID); err != nil {
			problem.Write(w, http.StatusBadRequest, "cancel", "")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func Replay(ro *runssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		rid, err := uuid.Parse(chi.URLParam(r, "workflow_run_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "workflow_run_id", "")
			return
		}
		nid, err := ro.Replay(r.Context(), rid, c.UserID)
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "replay", "")
			return
		}
		problem.JSON(w, http.StatusAccepted, map[string]string{"workflow_run_id": nid})
	}
}

func ReplayAction(ro *runssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		rid, err := uuid.Parse(chi.URLParam(r, "workflow_run_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "workflow_run_id", "")
			return
		}
		var body struct {
			ActionKey string `json:"action_key"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		nid, err := ro.ReplayAction(r.Context(), rid, c.UserID, body.ActionKey)
		if err != nil {
			if errors.Is(err, runsrepo.ErrReplayActionKeyRequired) {
				problem.Write(w, http.StatusBadRequest, "action_key", "")
				return
			}
			problem.Write(w, http.StatusBadRequest, "retry_action", "")
			return
		}
		problem.JSON(w, http.StatusAccepted, map[string]string{"workflow_run_id": nid})
	}
}
