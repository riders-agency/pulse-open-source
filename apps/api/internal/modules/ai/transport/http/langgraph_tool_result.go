package http

import (
	"strings"

	aitools "riders.agency/pulse/api/internal/modules/ai/tools"
)

type langGraphToolResultMapping struct {
	FinalBody  string
	Payload    map[string]any
	ReadResult map[string]any
	StreamText bool
}

func mapLangGraphToolResult(toolCall map[string]any, message string, currentFinal string) langGraphToolResultMapping {
	toolKey := stringFromAny(toolCall["tool_key"])
	result := nestedMap(toolCall, "output_json")
	if isLangGraphReadTool(toolKey) {
		finalBody := firstNonEmpty(stringFromAny(result["answer"]), strings.TrimSpace(message), currentFinal)
		return langGraphToolResultMapping{
			FinalBody: finalBody,
			Payload: map[string]any{
				"kind":        "tool_result",
				"tool_key":    toolKey,
				"risk_level":  "read",
				"output_json": result,
				"from_memory": false,
				"runtime":     "langgraph",
			},
			ReadResult: map[string]any{"tool_key": toolKey, "output_json": result},
			StreamText: finalBody != "",
		}
	}
	if toolKey == "approval.apply" || toolKey == "approval.reject" {
		finalBody := firstNonEmpty(stringFromAny(result["message"]), strings.TrimSpace(message), currentFinal)
		return langGraphToolResultMapping{
			FinalBody: finalBody,
			Payload: map[string]any{
				"kind":        "approval_result",
				"tool_key":    toolKey,
				"risk_level":  string(aitools.RiskHigh),
				"output_json": result,
				"runtime":     "langgraph",
			},
			StreamText: finalBody != "",
		}
	}
	return langGraphToolResultMapping{FinalBody: currentFinal}
}
