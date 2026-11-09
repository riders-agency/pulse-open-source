package http

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	airepo "riders.agency/pulse/api/internal/modules/ai/repository"
)

func TestAIRolloutPolicyScopesByProjectAllowlist(t *testing.T) {
	clearCapabilityEnv(t)
	projectID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	t.Setenv("PULSE_AI_V3_PROJECT_ALLOWLIST", "22222222-2222-2222-2222-222222222222")
	t.Setenv("PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP", "true")

	policy := evaluateAIRolloutPolicy("company-1", projectID, airepo.MessageInput{}, readyRolloutCapability())
	if policy.Enabled {
		t.Fatalf("policy should be disabled outside project allowlist: %#v", policy)
	}
	if !containsString(policy.Reasons, "project_scope_not_enabled") {
		t.Fatalf("reasons = %#v", policy.Reasons)
	}
}

func TestAIRolloutPolicyRollbackDowngradesWrites(t *testing.T) {
	clearCapabilityEnv(t)
	t.Setenv("PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP", "true")
	t.Setenv("PULSE_AI_V3_CURRENT_ERROR_RATE", "0.08")
	t.Setenv("PULSE_AI_V3_MAX_ERROR_RATE", "0.01")

	policy := evaluateAIRolloutPolicy("company-1", uuid.New(), airepo.MessageInput{}, readyRolloutCapability())
	if !policy.Rollback || policy.Stage != AIRolloutReadOnly || policy.AllowsPropose || policy.AllowsWrite {
		t.Fatalf("rollback policy = %#v", policy)
	}
	if err := policy.ValidateExecution(airepo.MessageInput{Body: "покажи задачи"}); err != nil {
		t.Fatalf("read-only traffic should remain available during rollback: %v", err)
	}
	err := policy.ValidateExecution(airepo.MessageInput{Body: "создай задачу"})
	if err == nil || !strings.Contains(err.Error(), "ai_v3_proposal_gate_not_ready") {
		t.Fatalf("write-like traffic should be gated during rollback: %v", err)
	}
}

func TestAIRolloutPolicyShadowRejectsExecution(t *testing.T) {
	clearCapabilityEnv(t)
	t.Setenv("PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP", "true")
	t.Setenv("PULSE_AI_V3_STAGE", "shadow")

	policy := evaluateAIRolloutPolicy("company-1", uuid.New(), airepo.MessageInput{}, readyRolloutCapability())
	if !policy.Shadow || policy.Stage != AIRolloutShadow {
		t.Fatalf("shadow policy = %#v", policy)
	}
	err := policy.ValidateExecution(airepo.MessageInput{Body: "покажи задачи"})
	if err == nil || !strings.Contains(err.Error(), "ai_v3_shadow_mode") {
		t.Fatalf("shadow mode should not execute: %v", err)
	}
}

func TestAIRolloutPolicyStageOverrideControlsWriteIntent(t *testing.T) {
	clearCapabilityEnv(t)
	t.Setenv("PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP", "true")
	t.Setenv("PULSE_AI_V3_STAGE", "read_only")

	in := airepo.MessageInput{Body: "создай задачу проверить оплату"}
	policy := evaluateAIRolloutPolicy("company-1", uuid.New(), in, readyRolloutCapability())
	if policy.Stage != AIRolloutReadOnly || policy.AllowsPropose {
		t.Fatalf("read-only policy = %#v", policy)
	}
	if err := policy.ValidateExecution(in); err == nil || !strings.Contains(err.Error(), "ai_v3_proposal_gate_not_ready") {
		t.Fatalf("read-only stage should reject write intent: %v", err)
	}

	t.Setenv("PULSE_AI_V3_STAGE", "approved_writes")
	policy = evaluateAIRolloutPolicy("company-1", uuid.New(), in, readyRolloutCapability())
	if policy.Stage != AIRolloutApprovedWrites || !policy.AllowsPropose || !policy.AllowsWrite {
		t.Fatalf("approved-writes policy = %#v", policy)
	}
	if err := policy.ValidateExecution(in); err != nil {
		t.Fatalf("approved-writes stage should allow proposal flow: %v", err)
	}
}

func TestAIRolloutPolicyRequiresGraphToolLoop(t *testing.T) {
	clearCapabilityEnv(t)
	policy := evaluateAIRolloutPolicy("company-1", uuid.New(), airepo.MessageInput{}, readyRolloutCapability())

	err := policy.ValidateExecution(airepo.MessageInput{Body: "покажи задачи"})
	if err == nil || !strings.Contains(err.Error(), "langgraph_v3_tool_gateway_required") {
		t.Fatalf("missing graph tool loop should be fatal: %v", err)
	}
}

func TestAttachAIRolloutPayloadPreservesExistingPayload(t *testing.T) {
	clearCapabilityEnv(t)
	policy := evaluateAIRolloutPolicy("company-1", uuid.New(), airepo.MessageInput{}, readyRolloutCapability())
	payload := attachAIRolloutPayload(map[string]any{"kind": "answer"}, policy)

	if payload["kind"] != "answer" {
		t.Fatalf("payload kind lost: %#v", payload)
	}
	rollout, ok := payload["rollout"].(map[string]any)
	if !ok || rollout["schema"] != "riders-pulse-ai-rollout-policy-v1" {
		t.Fatalf("rollout payload missing: %#v", payload)
	}
}

func readyRolloutCapability() AICapability {
	return AICapability{
		State:         AICapabilityReady,
		Runtime:       "langgraph",
		AllowsRead:    true,
		AllowsPropose: true,
		AllowsWrite:   true,
	}
}
