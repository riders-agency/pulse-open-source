package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	airepo "riders.agency/pulse/api/internal/modules/ai/repository"
	aisvc "riders.agency/pulse/api/internal/modules/ai/service"
	"riders.agency/pulse/api/internal/platform/problem"
)

func ModelProviders() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		problem.JSON(w, http.StatusOK, map[string]any{"items": airepo.ModelProviders()})
	}
}

func GetEmployee(ai *aisvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, userID, ok := ids(w, r)
		if !ok {
			return
		}
		out, err := ai.EnsureEmployee(r.Context(), projectID, userID)
		if err == nil {
			out["capability"] = evaluateAICapability(r.Context(), capabilityOptions{CheckPythonGRPC: true})
		}
		write(w, http.StatusOK, out, err)
	}
}

func PatchEmployee(ai *aisvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, userID, ok := ids(w, r)
		if !ok {
			return
		}
		var in airepo.EmployeePatch
		if !decode(w, r, &in) {
			return
		}
		out, err := ai.PatchEmployee(r.Context(), projectID, userID, in)
		write(w, http.StatusOK, out, err)
	}
}

func ListModels(ai *aisvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, userID, ok := ids(w, r)
		if !ok {
			return
		}
		depID, err := uuid.Parse(chi.URLParam(r, "action_dependency_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "action_dependency_id", "")
			return
		}
		out, err := ai.ListModels(r.Context(), projectID, depID, userID, false)
		write(w, http.StatusOK, out, err)
	}
}

func RefreshModels(ai *aisvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, userID, ok := ids(w, r)
		if !ok {
			return
		}
		depID, err := uuid.Parse(chi.URLParam(r, "action_dependency_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "action_dependency_id", "")
			return
		}
		out, err := ai.ListModels(r.Context(), projectID, depID, userID, true)
		write(w, http.StatusOK, out, err)
	}
}

func EvaluateModels(ai *aisvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, userID, ok := ids(w, r)
		if !ok {
			return
		}
		depID, err := uuid.Parse(chi.URLParam(r, "action_dependency_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "action_dependency_id", "")
			return
		}
		out, err := ai.EvaluateModels(r.Context(), projectID, depID, userID)
		write(w, http.StatusOK, out, err)
	}
}
