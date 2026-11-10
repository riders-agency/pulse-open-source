package runtime

import (
	"context"
	"encoding/json"
	"strings"

	"riders.agency/pulse/api/internal/modules/person/phone"
	"riders.agency/pulse/api/internal/runtime/contracts"
)

type phoneNormalizeHandler struct {
	renderer TemplateRenderer
}

func (h *phoneNormalizeHandler) TypeKey() string { return "person.phone.normalize" }

func (h *phoneNormalizeHandler) ValidateConfig(context.Context, json.RawMessage) error { return nil }

func (h *phoneNormalizeHandler) Execute(_ context.Context, req contracts.ActionRequest) contracts.ActionResult {
	var cfg struct {
		Input        string            `json:"input"`
		OutputFields map[string]string `json:"output_fields"`
	}
	_ = json.Unmarshal(req.ConfigJSON, &cfg)

	rawInput := strings.TrimSpace(cfg.Input)
	if rawInput == "" {
		rawInput = "{{ webhook.phone }}"
	}
	rendered := rawInput
	if h.renderer != nil {
		rendered = strings.TrimSpace(h.renderer(rawInput, req.RunInputJSON, req.Steps, req.WorkflowRunID))
	}

	result := phone.Normalize(rendered)
	expose := map[string]any{
		"phone_e164":  result.PhoneE164,
		"phone_raw":   result.PhoneRaw,
		"input_phone": result.InputPhone,
		"mode":        result.Mode,
		"changed":     result.Changed,
		"is_valid":    result.IsValid,
		"warning":     result.Warning,
	}
	for alias, key := range cfg.OutputFields {
		if v, ok := expose[key]; ok {
			expose[alias] = v
		}
	}

	inputJSON, _ := json.Marshal(map[string]any{"input_template": rawInput, "input_rendered": rendered})
	outputJSON, _ := json.Marshal(map[string]any{
		"action_key":      req.ActionKey,
		"action_type_key": req.ActionTypeKey,
		"provider":        "internal",
		"delivered":       true,
		"expose":          expose,
	})
	providerJSON, _ := json.Marshal(map[string]any{"provider": "person.phone", "ok": true})
	return contracts.ActionResult{
		InputJSON:            inputJSON,
		OutputJSON:           outputJSON,
		ProviderResponseJSON: providerJSON,
	}
}
