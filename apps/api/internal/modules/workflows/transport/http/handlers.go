package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	workflowssvc "riders.agency/pulse/api/internal/modules/workflows/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"riders.agency/pulse/api/internal/modules/workflows/definitionvalidate"
	"riders.agency/pulse/api/internal/modules/workflows/repository"
	"riders.agency/pulse/api/internal/platform/errs"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	problem "riders.agency/pulse/api/internal/platform/problem"
)

func ListByProject(wf *workflowssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		pid, err := uuid.Parse(chi.URLParam(r, "project_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "project_id", "")
			return
		}
		items, err := wf.ListByProject(r.Context(), pid, c.UserID)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "list", "")
			return
		}
		problem.JSON(w, http.StatusOK, items)
	}
}

type createWF struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func Create(wf *workflowssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		pid, err := uuid.Parse(chi.URLParam(r, "project_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "project_id", "")
			return
		}
		var body createWF
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		if body.Name == "" || body.Slug == "" {
			problem.Write(w, http.StatusBadRequest, "name and slug required", "")
			return
		}
		row, err := wf.Create(r.Context(), pid, c.UserID, body.Name, body.Slug)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusConflict, "create", "")
			return
		}
		problem.JSON(w, http.StatusCreated, row)
	}
}

func Get(wf *workflowssvc.Commands) http.HandlerFunc {
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
		d, err := wf.Get(r.Context(), wfid, c.UserID)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusNotFound, "not found", "")
			return
		}
		problem.JSON(w, http.StatusOK, d)
	}
}

func PutDraft(wf *workflowssvc.Commands) http.HandlerFunc {
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
		var patch repository.DraftPatch
		if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		err = wf.PutDraft(r.Context(), wfid, c.UserID, patch)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "save", "")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func Validate(wf *workflowssvc.Commands) http.HandlerFunc {
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
		detail, err := wf.Get(r.Context(), wfid, c.UserID)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusNotFound, "not found", "")
			return
		}
		if err := definitionvalidate.Validate(detail.DraftDefinition); err != nil {
			problem.Write(w, http.StatusBadRequest, "validation", err.Error())
			return
		}
		problem.JSON(w, http.StatusOK, map[string]any{"valid": true, "warnings": []any{}})
	}
}

func Publish(wf *workflowssvc.Commands) http.HandlerFunc {
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
			Description string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
			problem.Write(w, http.StatusBadRequest, "json", "")
			return
		}
		vid, vn, err := wf.Publish(r.Context(), wfid, c.UserID, body.Description)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "publish", err.Error())
			return
		}
		problem.JSON(w, http.StatusOK, map[string]any{
			"workflow_version_id": vid,
			"version_no":          vn,
		})
	}
}

func RestoreVersion(wf *workflowssvc.Commands) http.HandlerFunc {
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
		wvid, err := uuid.Parse(chi.URLParam(r, "workflow_version_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "workflow_version_id", "")
			return
		}
		row, err := wf.RestoreVersion(r.Context(), wfid, wvid, c.UserID)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "restore", err.Error())
			return
		}
		problem.JSON(w, http.StatusOK, row)
	}
}

func ListVersions(wf *workflowssvc.Commands) http.HandlerFunc {
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
		items, err := wf.ListVersions(r.Context(), wfid, c.UserID)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "list", "")
			return
		}
		problem.JSON(w, http.StatusOK, items)
	}
}

func GetVersion(wf *workflowssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		wvid, err := uuid.Parse(chi.URLParam(r, "workflow_version_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "workflow_version_id", "")
			return
		}
		d, err := wf.VersionByID(r.Context(), wvid, c.UserID)
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, http.StatusForbidden, "forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusNotFound, "not found", "")
			return
		}
		problem.JSON(w, http.StatusOK, d)
	}
}
