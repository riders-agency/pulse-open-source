package http

import (
	"encoding/json"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
	aitools "riders.agency/pulse/api/internal/modules/ai/tools"
)

func nestedMap(value map[string]any, path ...string) map[string]any {
	var current any = value
	for _, key := range path {
		item, ok := current.(map[string]any)
		if !ok {
			return map[string]any{}
		}
		current = item[key]
	}
	if out, ok := current.(map[string]any); ok {
		return out
	}
	return map[string]any{}
}

func aggregateReadResultsPayload(payload map[string]any, readResults []map[string]any) map[string]any {
	if len(readResults) <= 1 {
		return payload
	}
	out := map[string]any{
		"kind":         "tool_results",
		"tool_key":     "read.parallel",
		"risk_level":   "read",
		"results_json": readResults,
		"from_memory":  false,
		"runtime":      "langgraph",
	}
	if len(readResults) > 0 {
		out["output_json"] = readResults[len(readResults)-1]["output_json"]
	}
	return out
}

func approvalObservation(session map[string]any, decision, message string) map[string]any {
	return map[string]any{
		"decision":             strings.TrimSpace(decision),
		"message":              strings.TrimSpace(message),
		"active_tool_call_id":  strings.TrimSpace(stringFromAny(session["active_tool_call_id"])),
		"session_revision":     session["session_revision"],
		"ai_chat_session_id":   session["ai_chat_session_id"],
		"active_proposal_done": strings.TrimSpace(stringFromAny(session["active_tool_call_id"])) == "",
	}
}

func isLangGraphReadTool(toolKey string) bool {
	switch toolKey {
	case "context.read",
		"ui_context.read",
		"task.search",
		"task.read",
		"client.search",
		"client.read",
		"client.count",
		"payment.lookup_due",
		"payment.schedule.read",
		"workflow.read_draft",
		"workflow.read_context",
		"workflow.compile_prompt",
		"task_process.read",
		"integration.catalog.list",
		"integration.dependency.read":
		return true
	default:
		return false
	}
}

func proposalPayloadFromCard(card map[string]any) (map[string]any, bool) {
	cardType := stringFromAny(card["card_type"])
	toolKey := toolKeyForCardType(cardType, stringFromAny(card["intent_key"]))
	if toolKey == "" || !boolFromAny(card["requires_approval"]) {
		return nil, false
	}
	inputJSON := mapFromAnyLocal(card["input_json"])
	if len(inputJSON) == 0 {
		inputJSON = nestedMap(card, "debug_json", "input_json")
	}
	displayJSON := mapFromAnyLocal(card["display_json"])
	if len(displayJSON) == 0 {
		displayJSON = cardDisplayJSON(card, inputJSON)
	}
	return map[string]any{
		"kind":              "proposal",
		"schema":            firstNonEmpty(stringFromAny(card["schema"]), "riders-pulse-ai-card-v1"),
		"card_type":         cardType,
		"proposal_type":     toolKey + ".propose",
		"tool_key":          toolKey,
		"risk_level":        string(aitools.NormalizeRiskLevel(firstNonEmpty(stringFromAny(card["risk_level"]), string(aitools.RiskHigh)))),
		"requires_approval": true,
		"input_json":        inputJSON,
		"fields_json":       inputJSON,
		"display_json":      displayJSON,
		"card_json":         card,
		"developer_json": map[string]any{
			"langgraph_card":    card,
			"hidden_by_default": true,
		},
		"plan_json": map[string]any{
			"steps": []map[string]any{
				{"label": "Классифицировать turn", "tool": "intent_router"},
				{"label": "Собрать typed card envelope", "tool": "card_builder"},
				{"label": "Ждать подтверждения пользователя", "tool": "approval_gate"},
			},
		},
	}, true
}

func toolKeyForCardType(cardType, intentKey string) string {
	if intentKey != "" {
		switch intentKey {
		case "task.create", "task.update", "task.move", "client.create", "client.update", "workflow.patch", "task_process.patch":
			return intentKey
		}
	}
	switch cardType {
	case "task_create_proposal":
		return "task.create"
	case "task_update_proposal":
		return "task.update"
	case "task_move_proposal":
		return "task.move"
	case "client_create_proposal":
		return "client.create"
	case "multi_step_plan":
		return "multi_step.plan"
	case "workflow_patch_preview":
		return "workflow.patch"
	case "task_process_patch_preview":
		return "task_process.patch"
	default:
		return ""
	}
}

func cardDisplayJSON(card map[string]any, input map[string]any) map[string]any {
	display := map[string]any{}
	for key, value := range input {
		display[key] = value
	}
	rawFields, _ := card["fields"].([]any)
	for _, raw := range rawFields {
		field, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		key := stringFromAny(field["key"])
		if key == "" {
			continue
		}
		switch key {
		case "client":
			display["client_name"] = field["value"]
		default:
			display[key] = field["value"]
		}
	}
	rawSections, _ := card["sections"].([]any)
	for _, raw := range rawSections {
		section, ok := raw.(map[string]any)
		if !ok || stringFromAny(section["key"]) != "checklist" {
			continue
		}
		display["checklist"] = section["items"]
	}
	return display
}

func mapFromAnyLocal(value any) map[string]any {
	if out, ok := value.(map[string]any); ok {
		return out
	}
	return map[string]any{}
}

func boolFromAny(value any) bool {
	if out, ok := value.(bool); ok {
		return out
	}
	return false
}

func rawStruct(raw json.RawMessage) *structpb.Struct {
	if len(raw) == 0 {
		return &structpb.Struct{Fields: map[string]*structpb.Value{}}
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return &structpb.Struct{Fields: map[string]*structpb.Value{}}
	}
	out, err := structpb.NewStruct(value)
	if err != nil {
		return &structpb.Struct{Fields: map[string]*structpb.Value{}}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
