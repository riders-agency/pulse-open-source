package http

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"

	airepo "riders.agency/pulse/api/internal/modules/ai/repository"
)

type AIRolloutStage string

const (
	AIRolloutShadow         AIRolloutStage = "shadow"
	AIRolloutReadOnly       AIRolloutStage = "read_only"
	AIRolloutProposalOnly   AIRolloutStage = "proposal_only"
	AIRolloutApprovedWrites AIRolloutStage = "approved_writes"
)

type AIRolloutPolicy struct {
	Schema        string                   `json:"schema"`
	Runtime       string                   `json:"runtime"`
	Stage         AIRolloutStage           `json:"stage"`
	Enabled       bool                     `json:"enabled"`
	Shadow        bool                     `json:"shadow"`
	AllowsRead    bool                     `json:"allows_read"`
	AllowsPropose bool                     `json:"allows_proposals"`
	AllowsWrite   bool                     `json:"allows_writes"`
	Rollback      bool                     `json:"rollback"`
	Reasons       []string                 `json:"reasons"`
	Scope         map[string]any           `json:"scope"`
	Gates         map[string]AIRolloutGate `json:"gates"`
	Versions      map[string]string        `json:"versions"`
}

type AIRolloutGate struct {
	Status  string  `json:"status"`
	Detail  string  `json:"detail,omitempty"`
	Current float64 `json:"current,omitempty"`
	Limit   float64 `json:"limit,omitempty"`
}

func evaluateAIRolloutPolicy(companyID string, projectID uuid.UUID, in airepo.MessageInput, capability AICapability) AIRolloutPolicy {
	policy := AIRolloutPolicy{
		Schema:        "riders-pulse-ai-rollout-policy-v1",
		Runtime:       capability.Runtime,
		Stage:         rolloutStageForCapability(capability),
		Enabled:       capability.State != AICapabilityDisabled,
		AllowsRead:    capability.AllowsRead,
		AllowsPropose: capability.AllowsPropose,
		AllowsWrite:   capability.AllowsWrite,
		Reasons:       append([]string{}, capability.Reasons...),
		Scope: map[string]any{
			"company_id": companyID,
			"project_id": projectID.String(),
		},
		Gates:    map[string]AIRolloutGate{},
		Versions: rolloutVersions(),
	}
	policy.applyScope(companyID, projectID)
	policy.applyStageOverride()
	policy.applyCapabilityBounds(capability)
	policy.applyRollbackGates()
	if likelyWriteIntent(in) && !policy.AllowsPropose {
		policy.Reasons = append(policy.Reasons, "write_intent_not_enabled_for_stage")
	}
	if strings.TrimSpace(in.ActiveToolCallID) != "" && !policy.AllowsWrite {
		policy.Reasons = append(policy.Reasons, "approval_write_not_enabled_for_stage")
	}
	return policy
}

func (p AIRolloutPolicy) ValidateExecution(in airepo.MessageInput) error {
	if !p.Enabled {
		return p.gateError("ai_v3_rollout_not_enabled")
	}
	if p.Shadow || p.Stage == AIRolloutShadow {
		return p.gateError("ai_v3_shadow_mode")
	}
	if !p.AllowsRead {
		return p.gateError("ai_v3_read_gate_not_ready")
	}
	if !boolEnv("PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP") {
		return p.gateError("langgraph_v3_tool_gateway_required")
	}
	if likelyWriteIntent(in) && !p.AllowsPropose {
		return p.gateError("ai_v3_proposal_gate_not_ready")
	}
	if strings.TrimSpace(in.ActiveToolCallID) != "" && !p.AllowsWrite {
		return p.gateError("ai_v3_write_gate_not_ready")
	}
	if p.Rollback && (likelyWriteIntent(in) || strings.TrimSpace(in.ActiveToolCallID) != "") {
		return p.gateError("ai_v3_rollback_gate_active")
	}
	return nil
}

func (p AIRolloutPolicy) gateError(code string) error {
	return fmt.Errorf("%s: stage=%s enabled=%t read=%t propose=%t write=%t rollback=%t reasons=%s",
		code,
		p.Stage,
		p.Enabled,
		p.AllowsRead,
		p.AllowsPropose,
		p.AllowsWrite,
		p.Rollback,
		strings.Join(p.Reasons, ","),
	)
}

func attachAIRolloutPayload(payload map[string]any, policy AIRolloutPolicy) map[string]any {
	if payload == nil {
		payload = map[string]any{}
	}
	payload["rollout"] = map[string]any{
		"schema":            policy.Schema,
		"runtime":           policy.Runtime,
		"stage":             string(policy.Stage),
		"enabled":           policy.Enabled,
		"shadow":            policy.Shadow,
		"allows_read":       policy.AllowsRead,
		"allows_proposals":  policy.AllowsPropose,
		"allows_writes":     policy.AllowsWrite,
		"rollback":          policy.Rollback,
		"reasons":           policy.Reasons,
		"scope":             policy.Scope,
		"gates":             policy.Gates,
		"versions":          policy.Versions,
		"old_planner_state": "removed_from_langgraph_stream",
	}
	return payload
}

func (p *AIRolloutPolicy) applyScope(companyID string, projectID uuid.UUID) {
	companyScoped, companyAllowed := rolloutAllowlistContains("PULSE_AI_V3_COMPANY_ALLOWLIST", companyID)
	projectScoped, projectAllowed := rolloutAllowlistContains("PULSE_AI_V3_PROJECT_ALLOWLIST", projectID.String())
	p.Scope["company_allowlist_enabled"] = companyScoped
	p.Scope["project_allowlist_enabled"] = projectScoped
	if companyScoped && !companyAllowed {
		p.Enabled = false
		p.AllowsRead = false
		p.AllowsPropose = false
		p.AllowsWrite = false
		p.Reasons = append(p.Reasons, "company_scope_not_enabled")
	}
	if projectScoped && !projectAllowed {
		p.Enabled = false
		p.AllowsRead = false
		p.AllowsPropose = false
		p.AllowsWrite = false
		p.Reasons = append(p.Reasons, "project_scope_not_enabled")
	}
}

func (p *AIRolloutPolicy) applyStageOverride() {
	stage, ok := normalizeAIRolloutStage(os.Getenv("PULSE_AI_V3_STAGE"))
	if !ok && strings.TrimSpace(os.Getenv("PULSE_AI_V3_STAGE")) != "" {
		p.Reasons = append(p.Reasons, "invalid_stage")
	}
	if ok {
		p.Stage = stage
	}
	if boolEnv("PULSE_AI_V3_SHADOW") {
		p.Stage = AIRolloutShadow
	}
	p.applyStagePermissions()
}

func (p *AIRolloutPolicy) applyCapabilityBounds(capability AICapability) {
	p.AllowsRead = p.AllowsRead && capability.AllowsRead
	p.AllowsPropose = p.AllowsPropose && capability.AllowsPropose
	p.AllowsWrite = p.AllowsWrite && capability.AllowsWrite
	if p.Stage == AIRolloutApprovedWrites && !p.AllowsWrite {
		p.Reasons = append(p.Reasons, "capability_downgraded")
		if p.AllowsPropose {
			p.Stage = AIRolloutProposalOnly
		} else if p.AllowsRead {
			p.Stage = AIRolloutReadOnly
		} else {
			p.Stage = AIRolloutShadow
			p.Enabled = false
		}
	}
	if p.Stage == AIRolloutProposalOnly && !p.AllowsPropose {
		p.Reasons = append(p.Reasons, "capability_downgraded")
		if p.AllowsRead {
			p.Stage = AIRolloutReadOnly
		} else {
			p.Stage = AIRolloutShadow
			p.Enabled = false
		}
	}
	p.applyStagePermissions()
}

func (p *AIRolloutPolicy) applyRollbackGates() {
	if boolEnv("PULSE_AI_V3_ROLLBACK") {
		p.Rollback = true
		p.Gates["manual_rollback"] = AIRolloutGate{Status: "failed", Detail: "PULSE_AI_V3_ROLLBACK"}
	}
	p.addMinGate("eval_score", "PULSE_AI_V3_CURRENT_EVAL_SCORE", "PULSE_AI_V3_MIN_EVAL_SCORE")
	p.addMaxGate("error_rate", "PULSE_AI_V3_CURRENT_ERROR_RATE", "PULSE_AI_V3_MAX_ERROR_RATE")
	p.addMaxGate("p95_latency_ms", "PULSE_AI_V3_CURRENT_P95_LATENCY_MS", "PULSE_AI_V3_MAX_P95_LATENCY_MS")
	if !p.Rollback {
		return
	}
	p.Stage = AIRolloutReadOnly
	p.Shadow = false
	p.Enabled = p.AllowsRead
	p.AllowsPropose = false
	p.AllowsWrite = false
	p.Reasons = append(p.Reasons, "rollback_gate_active")
}

func (p *AIRolloutPolicy) addMinGate(name string, currentKey string, limitKey string) {
	current, currentOK := envFloat(currentKey)
	limit, limitOK := envFloat(limitKey)
	if !currentOK || !limitOK {
		p.Gates[name] = AIRolloutGate{Status: "not_configured"}
		return
	}
	status := "ok"
	detail := ""
	if current < limit {
		status = "failed"
		detail = currentKey + " below " + limitKey
		p.Rollback = true
	}
	p.Gates[name] = AIRolloutGate{Status: status, Detail: detail, Current: current, Limit: limit}
}

func (p *AIRolloutPolicy) addMaxGate(name string, currentKey string, limitKey string) {
	current, currentOK := envFloat(currentKey)
	limit, limitOK := envFloat(limitKey)
	if !currentOK || !limitOK {
		p.Gates[name] = AIRolloutGate{Status: "not_configured"}
		return
	}
	status := "ok"
	detail := ""
	if current > limit {
		status = "failed"
		detail = currentKey + " above " + limitKey
		p.Rollback = true
	}
	p.Gates[name] = AIRolloutGate{Status: status, Detail: detail, Current: current, Limit: limit}
}

func (p *AIRolloutPolicy) applyStagePermissions() {
	switch p.Stage {
	case AIRolloutShadow:
		p.Shadow = true
		p.AllowsRead = false
		p.AllowsPropose = false
		p.AllowsWrite = false
	case AIRolloutReadOnly:
		p.Shadow = false
		p.AllowsPropose = false
		p.AllowsWrite = false
	case AIRolloutProposalOnly:
		p.Shadow = false
		p.AllowsWrite = false
	case AIRolloutApprovedWrites:
		p.Shadow = false
	}
}
