package repository

import (
	"encoding/json"
	"fmt"
	"strings"
)

type workflowSummary struct {
	ActionsCount int
	EdgesCount   int
	TriggerCount int
	Actions      []map[string]any
}

func workflowJSONSummary(raw json.RawMessage) workflowSummary {
	if len(raw) == 0 {
		return workflowSummary{}
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return workflowSummary{}
	}
	nodes := sliceFromAny(payload["nodes"])
	actions := sliceFromAny(payload["actions"])
	triggers := sliceFromAny(payload["triggers"])
	edges := sliceFromAny(payload["edges"])
	if len(actions) == 0 {
		actions = nodes
	}
	outActions := make([]map[string]any, 0, len(actions))
	for index, rawAction := range actions {
		action := mapFromAny(rawAction)
		actionKey := firstNonEmptyString(
			stringValue(action["action_key"]),
			stringValue(action["key"]),
			stringValue(action["id"]),
			fmt.Sprintf("action_%d", index+1),
		)
		actionType := firstNonEmptyString(stringValue(action["action_type_key"]), stringValue(action["type"]), stringValue(action["type_key"]))
		title := firstNonEmptyString(stringValue(action["title"]), stringValue(action["name"]), actionKey)
		outActions = append(outActions, map[string]any{
			"action_key":          actionKey,
			"action_type_key":     actionType,
			"title":               title,
			"position":            mapFromAny(action["position"]),
			"dependency_refs":     sliceFromAny(action["dependency_refs"]),
			"duplicate_signature": strings.ToLower(strings.TrimSpace(actionType + ":" + title)),
		})
	}
	return workflowSummary{
		ActionsCount: len(outActions),
		EdgesCount:   len(edges),
		TriggerCount: len(triggers),
		Actions:      outActions,
	}
}

func duplicateWorkflowActions(actions []map[string]any) []map[string]any {
	seen := map[string]int{}
	for _, action := range actions {
		signature := stringValue(action["duplicate_signature"])
		if signature != "" {
			seen[signature]++
		}
	}
	duplicates := make([]map[string]any, 0)
	for _, action := range actions {
		signature := stringValue(action["duplicate_signature"])
		if signature != "" && seen[signature] > 1 {
			duplicates = append(duplicates, action)
		}
	}
	return duplicates
}

func sliceFromAny(value any) []any {
	if out, ok := value.([]any); ok {
		return out
	}
	return []any{}
}

func firstPositiveInt(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}
