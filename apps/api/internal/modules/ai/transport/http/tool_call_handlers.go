package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	aisvc "riders.agency/pulse/api/internal/modules/ai/service"
	"riders.agency/pulse/api/internal/platform/problem"
)

func ApproveToolCall(ai *aisvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, userID, ok := ids(w, r)
		if !ok {
			return
		}
		if !requireAIWriteCapability(w, r) {
			return
		}
		toolCallID, err := uuid.Parse(chi.URLParam(r, "tool_call_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "tool_call_id", "")
			return
		}
		resumeCtx, err := ai.ToolCallRuntimeContext(r.Context(), projectID, toolCallID)
		if err != nil {
			write(w, http.StatusOK, nil, err)
			return
		}
		if out, handled, err := resumeLangGraphToolDecision(r.Context(), ai, resumeCtx, userID, "approve", "approved from UI"); handled {
			write(w, http.StatusOK, out, err)
			return
		}
		out, err := ai.ApproveToolCall(r.Context(), projectID, toolCallID, userID)
		write(w, http.StatusOK, out, err)
	}
}

func UpdateToolCall(ai *aisvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, userID, ok := ids(w, r)
		if !ok {
			return
		}
		if !requireAIProposalCapability(w, r) {
			return
		}
		toolCallID, err := uuid.Parse(chi.URLParam(r, "tool_call_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "tool_call_id", "")
			return
		}
		var in struct {
			InputJSON  map[string]any `json:"input_json"`
			EditReason string         `json:"edit_reason"`
		}
		if !decode(w, r, &in) {
			return
		}
		out, err := ai.UpdateToolCall(r.Context(), projectID, toolCallID, userID, in.InputJSON, strings.TrimSpace(in.EditReason))
		write(w, http.StatusOK, out, err)
	}
}

func RejectToolCall(ai *aisvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, userID, ok := ids(w, r)
		if !ok {
			return
		}
		if !requireAIWriteCapability(w, r) {
			return
		}
		toolCallID, err := uuid.Parse(chi.URLParam(r, "tool_call_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "tool_call_id", "")
			return
		}
		var in struct {
			Reason string `json:"reason"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		resumeCtx, err := ai.ToolCallRuntimeContext(r.Context(), projectID, toolCallID)
		if err != nil {
			write(w, http.StatusOK, nil, err)
			return
		}
		if out, handled, err := resumeLangGraphToolDecision(r.Context(), ai, resumeCtx, userID, "reject", in.Reason); handled {
			write(w, http.StatusOK, out, err)
			return
		}
		out, err := ai.RejectToolCall(r.Context(), projectID, toolCallID, userID, strings.TrimSpace(in.Reason))
		write(w, http.StatusOK, out, err)
	}
}
