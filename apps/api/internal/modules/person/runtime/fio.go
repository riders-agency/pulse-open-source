package runtime

import (
	"context"
	"encoding/json"
	"strings"

	"riders.agency/pulse/api/internal/modules/person/fio"
	"riders.agency/pulse/api/internal/runtime/contracts"
	"riders.agency/pulse/api/internal/runtime/registry"
)

type TemplateRenderer func(template string, input json.RawMessage, steps map[string]any, workflowRunID string) string

func RegisterHandlers(reg *registry.Registry, renderer TemplateRenderer) error {
	if err := reg.RegisterAction(&fioNormalizeHandler{renderer: renderer}); err != nil {
		return err
	}
	return reg.RegisterAction(&phoneNormalizeHandler{renderer: renderer})
}

type fioNormalizeHandler struct {
	renderer TemplateRenderer
}

func (h *fioNormalizeHandler) TypeKey() string { return "person.fio.normalize" }

func (h *fioNormalizeHandler) ValidateConfig(context.Context, json.RawMessage) error { return nil }

func (h *fioNormalizeHandler) Execute(_ context.Context, req contracts.ActionRequest) contracts.ActionResult {
	var cfg struct {
		Input        string            `json:"input"`
		OutputFields map[string]string `json:"output_fields"`
	}
	_ = json.Unmarshal(req.ConfigJSON, &cfg)

	rawInput := strings.TrimSpace(cfg.Input)
	if rawInput == "" {
		rawInput = "{{ webhook.name }}"
	}
	rendered := rawInput
	if h.renderer != nil {
		rendered = strings.TrimSpace(h.renderer(rawInput, req.RunInputJSON, req.Steps, req.WorkflowRunID))
	}

	result := fio.Normalize(rendered)
	expose := map[string]any{
		"u_name":     result.UName,
		"u_surname":  result.USurname,
		"u_sname":    result.USname,
		"mode":       result.Mode,
		"changed":    result.Changed,
		"comment":    result.Comment,
		"split":      result.Split,
		"full_name":  result.FullName,
		"input_name": result.InputName,
		"is_valid":   result.IsValid,
		"confidence": result.Confidence,
		"warnings":   result.Warnings,
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
	providerJSON, _ := json.Marshal(map[string]any{"provider": "person.fio", "ok": true})
	return contracts.ActionResult{
		InputJSON:            inputJSON,
		OutputJSON:           outputJSON,
		ProviderResponseJSON: providerJSON,
	}
}
