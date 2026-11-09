package http

import (
	"os"
	"strconv"
	"strings"
)

func rolloutStageForCapability(capability AICapability) AIRolloutStage {
	switch capability.State {
	case AICapabilityReady:
		return AIRolloutApprovedWrites
	case AICapabilityProposalOnly:
		return AIRolloutProposalOnly
	case AICapabilityReadOnly:
		return AIRolloutReadOnly
	default:
		return AIRolloutShadow
	}
}

func normalizeAIRolloutStage(raw string) (AIRolloutStage, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "":
		return "", false
	case string(AIRolloutShadow):
		return AIRolloutShadow, true
	case string(AIRolloutReadOnly), "read":
		return AIRolloutReadOnly, true
	case string(AIRolloutProposalOnly), "proposal":
		return AIRolloutProposalOnly, true
	case string(AIRolloutApprovedWrites), "write", "writes":
		return AIRolloutApprovedWrites, true
	default:
		return "", false
	}
}

func aiRolloutRuntimeSettings() map[string]any {
	stageRaw := strings.TrimSpace(os.Getenv("PULSE_AI_V3_STAGE"))
	stage, validStage := normalizeAIRolloutStage(stageRaw)
	if stageRaw == "" {
		validStage = false
	}
	return map[string]any{
		"stage":                   string(stage),
		"stage_raw":               stageRaw,
		"stage_valid":             validStage,
		"shadow":                  boolEnv("PULSE_AI_V3_SHADOW"),
		"rollback":                boolEnv("PULSE_AI_V3_ROLLBACK"),
		"tool_gateway_graph_loop": boolEnv("PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP"),
		"checkpointer":            boolEnv("PULSE_AI_CHECKPOINTER_READY"),
		"resume_lifecycle":        boolEnv("PULSE_AI_RESUME_READY"),
	}
}

func rolloutAllowlistContains(key string, needle string) (bool, bool) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return false, true
	}
	needle = strings.ToLower(strings.TrimSpace(needle))
	for _, item := range strings.Split(raw, ",") {
		if strings.ToLower(strings.TrimSpace(item)) == needle {
			return true, true
		}
	}
	return true, false
}

func rolloutVersions() map[string]string {
	return map[string]string{
		"release": os.Getenv("PULSE_RELEASE_ID"),
		"schema":  firstNonEmpty(os.Getenv("PULSE_AI_V3_SCHEMA_VERSION"), "riders-pulse-ai-v3"),
		"prompt":  firstNonEmpty(os.Getenv("PULSE_AI_PROMPT_VERSION"), "turn-understanding-v1"),
	}
}

func envFloat(key string) (float64, bool) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return 0, false
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, false
	}
	return value, true
}
