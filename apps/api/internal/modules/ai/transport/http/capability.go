package http

import (
	"context"
	"net/http"
	"os"
	"strings"
)

type AICapabilityState string

const (
	AICapabilityDisabled             AICapabilityState = "disabled"
	AICapabilityDeterministicPreview AICapabilityState = "deterministic_preview"
	AICapabilityReadOnly             AICapabilityState = "read_only"
	AICapabilityProposalOnly         AICapabilityState = "proposal_only"
	AICapabilityReady                AICapabilityState = "ready"
)

type AICapability struct {
	State         AICapabilityState               `json:"state"`
	Runtime       string                          `json:"runtime"`
	Production    bool                            `json:"production"`
	AllowsRead    bool                            `json:"allows_read"`
	AllowsPropose bool                            `json:"allows_proposals"`
	AllowsWrite   bool                            `json:"allows_writes"`
	Reasons       []string                        `json:"reasons"`
	Dependencies  map[string]CapabilityDependency `json:"dependencies"`
}

type CapabilityDependency struct {
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type capabilityOptions struct {
	CheckPythonGRPC bool
}

func evaluateAICapability(ctx context.Context, opts capabilityOptions) AICapability {
	runtime := aiEmployeeRuntime()
	capability := AICapability{
		State:         AICapabilityReadOnly,
		Runtime:       runtime,
		Production:    strings.EqualFold(os.Getenv("RP_ENV"), "production"),
		AllowsRead:    true,
		AllowsPropose: false,
		AllowsWrite:   false,
		Reasons:       []string{},
		Dependencies:  map[string]CapabilityDependency{},
	}
	if disabledRuntime(runtime) {
		capability.State = AICapabilityDisabled
		capability.AllowsRead = false
		capability.Dependencies["runtime"] = dep("disabled", "AI employee runtime is disabled")
		capability.Reasons = append(capability.Reasons, "runtime_disabled")
		return capability
	}
	capability.Dependencies["runtime"] = dep("ok", runtime)
	if runtime != "langgraph" {
		capability.State = AICapabilityDisabled
		capability.AllowsRead = false
		capability.Reasons = append(capability.Reasons, "unsupported_runtime")
		return capability
	}
	pythonOK := capability.addPythonDependency(ctx, opts.CheckPythonGRPC)
	if !pythonOK {
		capability.State = AICapabilityDisabled
		capability.AllowsRead = false
		capability.Reasons = append(capability.Reasons, "python_grpc_unavailable")
		return capability
	}
	modelOK := capability.addOptionalFlagDependency("model_gateway", "PULSE_MODEL_GATEWAY_READY")
	toolLoopOK := capability.addFlagDependency("tool_gateway_graph_loop", "PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP")
	checkpointOK := capability.addFlagDependency("checkpointer", "PULSE_AI_CHECKPOINTER_READY")
	resumeOK := capability.addFlagDependency("resume_lifecycle", "PULSE_AI_RESUME_READY")
	if !toolLoopOK {
		capability.State = AICapabilityDisabled
		capability.AllowsRead = false
		capability.Reasons = append(capability.Reasons, "tool_gateway_graph_loop_required")
		return capability
	}
	capability.AllowsPropose = true
	if !modelOK {
		capability.Reasons = append(capability.Reasons, "model_gateway_not_required_for_deterministic_tools")
	}
	if !checkpointOK || !resumeOK {
		capability.State = AICapabilityProposalOnly
		capability.Reasons = append(capability.Reasons, "write_lifecycle_not_ready")
		return capability
	}
	capability.State = AICapabilityReady
	capability.AllowsWrite = true
	return capability
}

func (c *AICapability) addPythonDependency(ctx context.Context, check bool) bool {
	target := aiAgentGRPCTarget()
	security := aiAgentGRPCSecurityReadiness(target)
	if security["status"] == "unavailable" {
		c.Dependencies["python_grpc_security"] = dep("unavailable", stringFromAny(security["error"]))
		return false
	}
	if !check {
		c.Dependencies["python_grpc"] = dep("not_checked", target)
		c.Dependencies["python_grpc_security"] = dep(stringFromAny(security["status"]), stringFromAny(security["mode"]))
		return true
	}
	ok, status, errText := checkPythonGRPC(ctx, target)
	c.Dependencies["python_grpc"] = dep(status, errText)
	c.Dependencies["python_grpc_security"] = dep(stringFromAny(security["status"]), stringFromAny(security["mode"]))
	return ok
}

func (c *AICapability) addFlagDependency(name string, key string) bool {
	ok := boolEnv(key)
	status := "missing"
	if ok {
		status = "ok"
	} else {
		c.Reasons = append(c.Reasons, strings.ToLower(name)+"_missing")
	}
	c.Dependencies[name] = dep(status, key)
	return ok
}

func (c *AICapability) addOptionalFlagDependency(name string, key string) bool {
	ok := boolEnv(key)
	status := "optional_missing"
	if ok {
		status = "ok"
	}
	c.Dependencies[name] = dep(status, key)
	return ok
}

func disabledRuntime(runtime string) bool {
	if strings.EqualFold(runtime, "disabled") || strings.EqualFold(runtime, "off") {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(os.Getenv("AI_EMPLOYEE_ENABLED")), "false")
}

func dep(status string, detail string) CapabilityDependency {
	return CapabilityDependency{Status: strings.TrimSpace(status), Detail: strings.TrimSpace(detail)}
}

func requireAIWriteCapability(w http.ResponseWriter, r *http.Request) bool {
	capability := evaluateAICapability(r.Context(), capabilityOptions{})
	if !capability.Production || capability.AllowsWrite {
		return true
	}
	writeCapabilityProblem(w, http.StatusConflict, "ai_write_capability_not_ready", capability)
	return false
}
