package executor

import (
	"encoding/json"
	"errors"
	"strings"

	providerdiagnostics "riders.agency/pulse/api/internal/modules/providers/diagnostics"
	emailprovider "riders.agency/pulse/api/internal/modules/providers/email"
	maxprovider "riders.agency/pulse/api/internal/modules/providers/max"
	telegramprovider "riders.agency/pulse/api/internal/modules/providers/telegram"
	uonprovider "riders.agency/pulse/api/internal/modules/providers/uon"
)

func diagnosticsForFailure(actionTypeKey string, provider map[string]any, retryable bool, errorType string, err error) providerdiagnostics.Report {
	if errorType == "dependency" || errorType == "dependency_type" || errorType == "dependency_status" {
		return providerdiagnostics.DependencyFailure(provider)
	}
	if errorType == "validation" {
		return providerdiagnostics.ValidationFailure(provider)
	}
	base := providerdiagnostics.BaseFailure(provider, retryable, err)
	providerKey := providerdiagnostics.ProviderName(provider)
	switch {
	case actionTypeKey == "telegram.send_message" || providerKey == "telegram":
		return telegramprovider.Analyze(provider, retryable, base)
	case actionTypeKey == "max.send_message" || providerKey == "max":
		return maxprovider.Analyze(provider, retryable, base)
	case actionTypeKey == "email.send" || providerKey == "smtp":
		return emailprovider.AnalyzeSMTP(provider, retryable, base)
	case actionTypeKey == "uon.create_lead" || providerKey == "uon":
		return uonprovider.Analyze(provider, retryable, base)
	default:
		if errorType == "network" {
			base.ReasonCode = "network_error"
			base.HumanReason = "Не удалось выполнить сетевой запрос к провайдеру."
			base.NextAction = "Проверьте доступность провайдера и повторите действие."
		}
		return base
	}
}

func DiagnosticsFromStored(actionTypeKey string, outputJSON []byte, providerJSON []byte, errMsg *string, status string) map[string]any {
	var output map[string]any
	_ = json.Unmarshal(outputJSON, &output)
	if existing, ok := output["diagnostics"].(map[string]any); ok && existing != nil {
		return existing
	}
	var provider map[string]any
	_ = json.Unmarshal(providerJSON, &provider)
	if provider == nil {
		provider = map[string]any{}
	}
	var report providerdiagnostics.Report
	if status == "succeeded" {
		report = providerdiagnostics.Success(actionTypeKey, provider, criteriaResultsFromOutput(output))
	} else {
		message := "action failed"
		if errMsg != nil && strings.TrimSpace(*errMsg) != "" {
			message = strings.TrimSpace(*errMsg)
		}
		report = diagnosticsForFailure(actionTypeKey, provider, false, "stored", errors.New(message))
	}
	b, _ := json.Marshal(report)
	var out map[string]any
	_ = json.Unmarshal(b, &out)
	if out == nil {
		return map[string]any{}
	}
	return out
}

func criteriaResultsFromOutput(output map[string]any) []providerdiagnostics.CriterionResult {
	raw, ok := output["criteria_results"]
	if !ok || raw == nil {
		return nil
	}
	b, _ := json.Marshal(raw)
	var out []providerdiagnostics.CriterionResult
	_ = json.Unmarshal(b, &out)
	return out
}
