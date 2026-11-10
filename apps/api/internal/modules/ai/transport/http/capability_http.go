package http

import (
	"net/http"
	"strings"

	airepo "riders.agency/pulse/api/internal/modules/ai/repository"
	"riders.agency/pulse/api/internal/platform/problem"
)

func EmployeeCapability() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, _, ok := ids(w, r); !ok {
			return
		}
		problem.JSON(w, http.StatusOK, map[string]any{
			"capability": evaluateAICapability(r.Context(), capabilityOptions{CheckPythonGRPC: true}),
			"rollout":    aiRolloutRuntimeSettings(),
		})
	}
}

func requireAIMessageCapability(w http.ResponseWriter, r *http.Request, in airepo.MessageInput) bool {
	capability := evaluateAICapability(r.Context(), capabilityOptions{})
	if capability.State == AICapabilityDisabled {
		writeCapabilityProblem(w, http.StatusServiceUnavailable, "ai_capability_disabled", capability)
		return false
	}
	if capability.Production && likelyWriteIntent(in) && !capability.AllowsPropose {
		writeCapabilityProblem(w, http.StatusConflict, "ai_proposal_capability_not_ready", capability)
		return false
	}
	return true
}

func requireAIProposalCapability(w http.ResponseWriter, r *http.Request) bool {
	capability := evaluateAICapability(r.Context(), capabilityOptions{})
	if !capability.Production || capability.AllowsPropose {
		return true
	}
	writeCapabilityProblem(w, http.StatusConflict, "ai_proposal_capability_not_ready", capability)
	return false
}

func writeCapabilityProblem(w http.ResponseWriter, status int, code string, capability AICapability) {
	problem.WriteSpec(w, problem.Spec{
		Status: status,
		Module: "ai",
		Code:   code,
		Title:  "AI capability is not available",
		Detail: "Requested AI employee capability is gated by runtime readiness.",
		Action: "Check /api/v1/ai/runtime/readiness and enable the missing AI dependencies before retrying.",
		Extra:  map[string]any{"capability": capability},
	})
}

func likelyWriteIntent(in airepo.MessageInput) bool {
	text := strings.ToLower(strings.TrimSpace(in.Body + " " + in.ClientCommand + " " + in.ActiveToolCallID + " " + in.ActivePatchProposalID))
	if text == "" {
		return false
	}
	for _, marker := range []string{
		"создай", "создать", "измени", "изменить", "обнови", "обновить", "перенеси", "перемести",
		"подтверди", "подтвердить", "запусти", "запустить", "примени", "применить", "approve",
		"apply", "create", "update", "move", "patch", "confirm",
	} {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}
