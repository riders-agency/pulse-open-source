package executor

import (
	"encoding/json"
	"strings"

	providerdiagnostics "riders.agency/pulse/api/internal/modules/providers/diagnostics"
)

func failedAction(inputJSON, outputJSON string, provider map[string]any, retryable bool, errorType string, err error) actionExecution {
	provider["provider_response_redacted"] = true
	providerJSON, _ := json.Marshal(provider)
	report := diagnosticsForFailure("", provider, retryable, errorType, err)
	output := providerdiagnostics.OutputWithReport(outputJSON, report)
	return actionExecution{
		InputJSON:            inputJSON,
		OutputJSON:           output,
		ProviderResponseJSON: string(providerJSON),
		Retryable:            retryable,
		ErrorType:            errorType,
		Err:                  err,
	}
}

type incidentEvent struct {
	Message     string
	PayloadJSON string
}

func incidentFromExecution(final actionExecution) (incidentEvent, bool) {
	var output map[string]any
	if err := json.Unmarshal([]byte(final.OutputJSON), &output); err != nil || output == nil {
		return incidentEvent{}, false
	}
	diagnostics, _ := output["diagnostics"].(map[string]any)
	if diagnostics == nil {
		return incidentEvent{}, false
	}
	reasonCode := stringMapValue(diagnostics, "reason_code")
	if !isIncidentReason(reasonCode) {
		return incidentEvent{}, false
	}
	message := stringMapValue(diagnostics, "summary")
	if message == "" {
		message = "Инцидент: неверный или отозванный токен интеграции"
	}
	payload := map[string]any{
		"reason_code":  reasonCode,
		"provider":     stringMapValue(diagnostics, "provider"),
		"summary":      message,
		"human_reason": stringMapValue(diagnostics, "human_reason"),
		"next_action":  stringMapValue(diagnostics, "next_action"),
		"retryable":    diagnostics["retryable"],
		"error_type":   final.ErrorType,
	}
	var provider map[string]any
	if err := json.Unmarshal([]byte(final.ProviderResponseJSON), &provider); err == nil && provider != nil {
		payload["provider_response"] = provider
	}
	payloadJSON, _ := json.Marshal(payload)
	return incidentEvent{Message: message, PayloadJSON: string(payloadJSON)}, true
}

func isIncidentReason(reasonCode string) bool {
	switch strings.TrimSpace(reasonCode) {
	case "telegram_invalid_token", "max_auth_failed", "smtp_auth_failed":
		return true
	default:
		return false
	}
}

func stringMapValue(values map[string]any, key string) string {
	value, _ := values[key].(string)
	return strings.TrimSpace(value)
}
