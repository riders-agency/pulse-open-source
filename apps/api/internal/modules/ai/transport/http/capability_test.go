package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	airepo "riders.agency/pulse/api/internal/modules/ai/repository"
)

func TestAICapabilityStateFixtures(t *testing.T) {
	tests := []struct {
		name       string
		env        map[string]string
		wantState  AICapabilityState
		wantRead   bool
		wantWrite  bool
		wantReason string
	}{
		{
			name:       "disabled",
			env:        map[string]string{"AI_EMPLOYEE_RUNTIME": "disabled"},
			wantState:  AICapabilityDisabled,
			wantReason: "runtime_disabled",
		},
		{
			name:       "unsupported runtime fails closed",
			env:        map[string]string{"RP_ENV": "production", "AI_EMPLOYEE_RUNTIME": "deterministic"},
			wantState:  AICapabilityDisabled,
			wantRead:   false,
			wantReason: "unsupported_runtime",
		},
		{
			name:       "langgraph missing python security",
			env:        map[string]string{"AI_EMPLOYEE_RUNTIME": "langgraph"},
			wantState:  AICapabilityDisabled,
			wantReason: "python_grpc_unavailable",
		},
		{
			name:       "langgraph disabled without graph tool loop",
			env:        secureLangGraphEnv(nil),
			wantState:  AICapabilityDisabled,
			wantReason: "tool_gateway_graph_loop_required",
		},
		{
			name: "langgraph proposal only without write lifecycle and model",
			env: secureLangGraphEnv(map[string]string{
				"PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP": "true",
			}),
			wantState:  AICapabilityProposalOnly,
			wantRead:   true,
			wantReason: "write_lifecycle_not_ready",
		},
		{
			name: "langgraph proposal only without write lifecycle",
			env: secureLangGraphEnv(map[string]string{
				"PULSE_MODEL_GATEWAY_READY":        "true",
				"PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP": "true",
			}),
			wantState:  AICapabilityProposalOnly,
			wantRead:   true,
			wantReason: "write_lifecycle_not_ready",
		},
		{
			name: "ready without optional model gateway",
			env: secureLangGraphEnv(map[string]string{
				"PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP": "true",
				"PULSE_AI_CHECKPOINTER_READY":      "true",
				"PULSE_AI_RESUME_READY":            "true",
			}),
			wantState:  AICapabilityReady,
			wantRead:   true,
			wantWrite:  true,
			wantReason: "model_gateway_not_required_for_deterministic_tools",
		},
		{
			name: "ready",
			env: secureLangGraphEnv(map[string]string{
				"PULSE_MODEL_GATEWAY_READY":        "true",
				"PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP": "true",
				"PULSE_AI_CHECKPOINTER_READY":      "true",
				"PULSE_AI_RESUME_READY":            "true",
			}),
			wantState: AICapabilityReady,
			wantRead:  true,
			wantWrite: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearCapabilityEnv(t)
			for key, value := range tt.env {
				t.Setenv(key, value)
			}
			got := evaluateAICapability(context.Background(), capabilityOptions{})
			if got.State != tt.wantState || got.AllowsRead != tt.wantRead || got.AllowsWrite != tt.wantWrite {
				t.Fatalf("capability = %#v", got)
			}
			if tt.wantReason != "" && !containsString(got.Reasons, tt.wantReason) {
				t.Fatalf("reasons = %#v, want %s", got.Reasons, tt.wantReason)
			}
		})
	}
}

func TestRequireAIMessageCapabilityAllowsDeterministicProposalWithoutModelGateway(t *testing.T) {
	clearCapabilityEnv(t)
	for key, value := range secureLangGraphEnv(map[string]string{
		"PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP": "true",
	}) {
		t.Setenv(key, value)
	}
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ai/employee/sessions/s/messages", nil)
	ok := requireAIMessageCapability(recorder, req, airepo.MessageInput{Body: "создай задачу проверить оплату"})
	if !ok {
		t.Fatalf("deterministic proposal should not require model gateway: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestAICapabilityReadyWithCheckedPythonWithoutModelGateway(t *testing.T) {
	clearCapabilityEnv(t)
	server := &contractAgentRuntimeServer{}
	addr, stop := serveContractRuntime(t, server)
	defer stop()
	for key, value := range secureLangGraphEnv(map[string]string{
		"AI_AGENT_GRPC_ADDR":               addr,
		"PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP": "true",
		"PULSE_AI_CHECKPOINTER_READY":      "true",
		"PULSE_AI_RESUME_READY":            "true",
	}) {
		t.Setenv(key, value)
	}
	got := evaluateAICapability(context.Background(), capabilityOptions{CheckPythonGRPC: true})
	if got.State != AICapabilityReady || !got.AllowsPropose || !got.AllowsWrite {
		t.Fatalf("capability = %#v", got)
	}
	if !containsString(got.Reasons, "model_gateway_not_required_for_deterministic_tools") {
		t.Fatalf("reasons = %#v", got.Reasons)
	}
}

func secureLangGraphEnv(extra map[string]string) map[string]string {
	env := map[string]string{
		"RP_ENV":                 "production",
		"AI_EMPLOYEE_RUNTIME":    "langgraph",
		"AI_AGENT_GRPC_ADDR":     "127.0.0.1:9097",
		"AI_AGENT_SERVICE_TOKEN": "service-token",
	}
	for key, value := range extra {
		env[key] = value
	}
	return env
}

func clearCapabilityEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"RP_ENV", "AI_EMPLOYEE_ENABLED", "AI_EMPLOYEE_RUNTIME", "AI_AGENT_GRPC_ADDR",
		"AI_AGENT_SERVICE_TOKEN", "AI_AGENT_GRPC_SERVICE_TOKEN", "AI_AGENT_GRPC_INSECURE_DEV",
		"AI_AGENT_INSECURE_DEV", "PULSE_MODEL_GATEWAY_READY", "PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP",
		"PULSE_AI_CHECKPOINTER_READY", "PULSE_AI_RESUME_READY",
		"PULSE_AI_V3_STAGE", "PULSE_AI_V3_SHADOW", "PULSE_AI_V3_COMPANY_ALLOWLIST",
		"PULSE_AI_V3_PROJECT_ALLOWLIST", "PULSE_AI_V3_ROLLBACK", "PULSE_AI_V3_CURRENT_EVAL_SCORE",
		"PULSE_AI_V3_MIN_EVAL_SCORE", "PULSE_AI_V3_CURRENT_ERROR_RATE", "PULSE_AI_V3_MAX_ERROR_RATE",
		"PULSE_AI_V3_CURRENT_P95_LATENCY_MS", "PULSE_AI_V3_MAX_P95_LATENCY_MS",
	} {
		t.Setenv(key, "")
	}
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
