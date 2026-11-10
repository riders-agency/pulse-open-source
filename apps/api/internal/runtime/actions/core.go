package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"riders.agency/pulse/api/internal/runtime/contracts"
)

type TemplateRenderer func(template string, input json.RawMessage, steps map[string]any, workflowRunID string) string

func CoreHandlers(renderer TemplateRenderer) []contracts.ActionHandler {
	return []contracts.ActionHandler{
		conditionBranchHandler{renderer: renderer},
		staticCoreHandler{typeKey: "stop", outputKey: "stopped"},
		staticCoreHandler{typeKey: "delay.wait", outputKey: "waited"},
		triggerActionHandler{typeKey: "webhook.trigger"},
		triggerActionHandler{typeKey: "manual.trigger"},
		triggerActionHandler{typeKey: "schedule.trigger"},
		triggerActionHandler{typeKey: "telegram.relay.trigger"},
		triggerActionHandler{typeKey: "module.event.trigger"},
		triggerActionHandler{typeKey: "site_chat.visitor_first_seen.trigger"},
	}
}

type coreActionConfig struct {
	Text            string `json:"text"`
	ScheduleKey     string `json:"schedule_key"`
	SuccessCriteria any    `json:"success_criteria"`
}

type conditionBranchHandler struct {
	renderer TemplateRenderer
}

func (h conditionBranchHandler) TypeKey() string {
	return "condition.branch"
}

func (h conditionBranchHandler) ValidateConfig(context.Context, json.RawMessage) error {
	return nil
}

func (h conditionBranchHandler) Execute(_ context.Context, req contracts.ActionRequest) contracts.ActionResult {
	var cfg coreActionConfig
	_ = json.Unmarshal(req.ConfigJSON, &cfg)
	raw := strings.TrimSpace(cfg.Text)
	rendered := raw
	if h.renderer != nil {
		rendered = strings.TrimSpace(h.renderer(raw, req.RunInputJSON, req.Steps, req.WorkflowRunID))
	}
	result := evalCondition(rendered)
	branch := "false"
	if result {
		branch = "true"
	}
	inputJSON, _ := json.Marshal(map[string]any{
		"text":     raw,
		"rendered": rendered,
	})
	outputJSON, _ := json.Marshal(map[string]any{
		"action_key":      req.ActionKey,
		"action_type_key": req.ActionTypeKey,
		"provider":        "internal",
		"condition":       rendered,
		"result":          result,
		"branch":          branch,
		"expose": map[string]any{
			"result": result,
			"branch": branch,
		},
	})
	providerJSON, _ := json.Marshal(internalProviderResponse(req.ActionKey))
	return contracts.ActionResult{
		InputJSON:            inputJSON,
		OutputJSON:           outputJSON,
		ProviderResponseJSON: providerJSON,
	}
}

type staticCoreHandler struct {
	typeKey   string
	outputKey string
}

func (h staticCoreHandler) TypeKey() string {
	return h.typeKey
}

func (h staticCoreHandler) ValidateConfig(context.Context, json.RawMessage) error {
	return nil
}

func (h staticCoreHandler) Execute(_ context.Context, req contracts.ActionRequest) contracts.ActionResult {
	var cfg coreActionConfig
	_ = json.Unmarshal(req.ConfigJSON, &cfg)
	inputJSON, _ := json.Marshal(map[string]any{
		"config":           jsonRawObjectOrEmpty(req.ConfigJSON),
		"success_criteria": cfg.SuccessCriteria,
	})
	output := map[string]any{
		"mock":            true,
		"action_key":      req.ActionKey,
		"action_type_key": req.ActionTypeKey,
		h.outputKey:       true,
	}
	outputJSON, _ := json.Marshal(output)
	providerJSON, _ := json.Marshal(internalProviderResponse(req.ActionKey))
	return contracts.ActionResult{
		InputJSON:            inputJSON,
		OutputJSON:           outputJSON,
		ProviderResponseJSON: providerJSON,
	}
}

type triggerActionHandler struct {
	typeKey string
}

func (h triggerActionHandler) TypeKey() string {
	return h.typeKey
}

func (h triggerActionHandler) ValidateConfig(context.Context, json.RawMessage) error {
	return nil
}

func (h triggerActionHandler) Execute(_ context.Context, req contracts.ActionRequest) contracts.ActionResult {
	var cfg coreActionConfig
	_ = json.Unmarshal(req.ConfigJSON, &cfg)
	input := req.RunInputJSON
	if len(strings.TrimSpace(string(input))) == 0 {
		input = json.RawMessage(`{}`)
	}
	output := map[string]any{
		"mock":            true,
		"action_key":      req.ActionKey,
		"action_type_key": req.ActionTypeKey,
	}
	switch req.ActionTypeKey {
	case "schedule.trigger":
		output["scheduled"] = true
		output["schedule_key"] = scheduleKey(req.ActionKey, cfg.ScheduleKey)
	case "telegram.relay.trigger":
		output["received"] = true
		output["provider"] = "telegram"
		output["relay"] = true
	case "site_chat.visitor_first_seen.trigger":
		output["received"] = true
		output["provider"] = "site_chat"
	}
	outputJSON, _ := json.Marshal(output)
	providerJSON, _ := json.Marshal(providerResponse(req.ActionTypeKey, req.ActionKey))
	return contracts.ActionResult{
		InputJSON:            input,
		OutputJSON:           outputJSON,
		ProviderResponseJSON: providerJSON,
	}
}

func jsonRawObjectOrEmpty(raw json.RawMessage) any {
	var out any
	if len(strings.TrimSpace(string(raw))) == 0 || json.Unmarshal(raw, &out) != nil || out == nil {
		return map[string]any{}
	}
	return out
}

func internalProviderResponse(actionKey string) map[string]any {
	return providerResponse("internal", actionKey)
}

func providerResponse(actionTypeKey, actionKey string) map[string]any {
	provider := "internal"
	if actionTypeKey == "telegram.relay.trigger" {
		provider = "telegram"
	}
	return map[string]any{
		"mock":                       true,
		"provider":                   provider,
		"provider_response_redacted": true,
		"message_id":                 "mock-" + actionKey,
	}
}

func scheduleKey(actionKey, configured string) string {
	key := strings.TrimSpace(configured)
	if key == "" {
		key = strings.TrimSpace(actionKey)
	}
	if key == "" {
		key = "schedule"
	}
	return key
}

func evalCondition(expr string) bool {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return false
	}
	lower := strings.ToLower(expr)
	switch lower {
	case "true", "yes", "y", "1", "ok":
		return true
	case "false", "no", "n", "0", "null", "undefined":
		return false
	}
	for _, op := range []string{"==", "!=", ">=", "<=", ">", "<"} {
		if idx := strings.Index(expr, op); idx >= 0 {
			left := normalizeConditionValue(expr[:idx])
			right := normalizeConditionValue(expr[idx+len(op):])
			switch op {
			case "==":
				return left == right
			case "!=":
				return left != right
			case ">", ">=", "<", "<=":
				lf, lerr := strconv.ParseFloat(left, 64)
				rf, rerr := strconv.ParseFloat(right, 64)
				if lerr != nil || rerr != nil {
					return false
				}
				switch op {
				case ">":
					return lf > rf
				case ">=":
					return lf >= rf
				case "<":
					return lf < rf
				case "<=":
					return lf <= rf
				}
			}
		}
	}
	if idx := strings.Index(lower, " contains "); idx >= 0 {
		left := normalizeConditionValue(expr[:idx])
		right := normalizeConditionValue(expr[idx+len(" contains "):])
		if right == "" || strings.Trim(right, ", \t") == "" {
			return false
		}
		return strings.Contains(left, right)
	}
	return normalizeConditionValue(expr) != ""
}

func normalizeConditionValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		first := value[0]
		last := value[len(value)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			value = value[1 : len(value)-1]
		}
	}
	return strings.TrimSpace(value)
}

func validateRegisteredType(handler contracts.ActionHandler, want string) error {
	if strings.TrimSpace(handler.TypeKey()) != want {
		return fmt.Errorf("handler type key = %q, want %q", handler.TypeKey(), want)
	}
	return nil
}
