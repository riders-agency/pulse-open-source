package uon

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	providerdiagnostics "riders.agency/pulse/api/internal/modules/providers/diagnostics"
)

func Criteria() []providerdiagnostics.CriterionSpec {
	return []providerdiagnostics.CriterionSpec{
		{
			Type:        "uon_https_request",
			Label:       "HTTPS запрос до U-ON API",
			Description: "Исполнитель получил HTTP ответ от api.u-on.ru.",
			Expected:    "HTTP ответ получен",
			Source:      "provider_response_json.status_code",
		},
		{
			Type:        "uon_json_parsed",
			Label:       "JSON ответ распарсен",
			Description: "Ответ U-ON распознан как JSON.",
			Expected:    "raw_response_json object",
			Source:      "provider_response_json.raw_response_json",
		},
		{
			Type:        "uon_result_200",
			Label:       "U-ON вернул result=200",
			Description: "CRM подтвердила создание обращения.",
			Expected:    "200",
			Source:      "provider_response_json.result",
		},
	}
}

func Evaluate(provider map[string]any) ([]providerdiagnostics.CriterionResult, error) {
	results := []providerdiagnostics.CriterionResult{
		criterion("uon_https_request", "HTTPS запрос до U-ON API", providerdiagnostics.StatusCode(provider) > 0, "HTTP ответ получен", statusActual(provider), "provider_response_json.status_code"),
		criterion("uon_json_parsed", "JSON ответ распарсен", hasMap(provider, "raw_response_json"), "raw_response_json object", rawJSONActual(provider), "provider_response_json.raw_response_json"),
		criterion("uon_result_200", "U-ON вернул result=200", intFromAny(provider["result"]) == http.StatusOK, "200", providerdiagnostics.StringifyJSONValue(provider["result"]), "provider_response_json.result"),
	}
	return results, firstFailure(results, "u-on lead criteria failed")
}

func Analyze(provider map[string]any, retryable bool, base providerdiagnostics.Report) providerdiagnostics.Report {
	code := providerdiagnostics.StatusCode(provider)
	base.Provider = "uon"
	base.Summary = "U-ON не подтвердил создание обращения."
	switch {
	case code == http.StatusNotFound:
		base.ReasonCode = "uon_invalid_api_key"
		base.HumanReason = "U-ON не принял API-ключ: HTTP 404."
		base.NextAction = "Проверьте api_key в интеграции U-ON и сохраните его заново."
		base.Retryable = false
	case code >= 500:
		base.ReasonCode = "uon_provider_unavailable"
		base.HumanReason = "У U-ON временная серверная ошибка."
		base.NextAction = "Повторите действие позже."
		base.Retryable = true
	case code == 0:
		base.ReasonCode = "uon_network_error"
		base.HumanReason = "Не удалось получить HTTP ответ от U-ON."
		base.NextAction = "Проверьте сеть сервера и повторите действие."
		base.Retryable = retryable
	default:
		result := intFromAny(provider["result"])
		if result != 0 && result != http.StatusOK {
			base.ReasonCode = "uon_result_error"
			base.HumanReason = fmt.Sprintf("U-ON вернул result=%d.", result)
			base.NextAction = "Проверьте поля обращения и ответ CRM."
			base.Retryable = false
		} else {
			base.ReasonCode = "uon_provider_error"
			base.HumanReason = "U-ON вернул ошибку: " + providerdiagnostics.ProviderDescription(provider)
			base.NextAction = "Сверьте source, note и контактные поля."
			base.Retryable = retryable
		}
	}
	return base
}

func SanitizeResponse(statusCode int, raw []byte) map[string]any {
	out := map[string]any{
		"provider":                   "uon",
		"status_code":                statusCode,
		"provider_response_redacted": true,
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		out["description"] = strings.TrimSpace(string(raw))
		return out
	}
	out["raw_response_json"] = payload
	if v, ok := payload["result"]; ok {
		out["result"] = v
	}
	if v, ok := payload["id"]; ok {
		out["id"] = v
	}
	if v, ok := payload["error"]; ok {
		out["description"] = v
	}
	if v, ok := payload["comment"]; ok && strings.TrimSpace(fmt.Sprint(v)) != "" {
		out["description"] = v
	}
	if warnings := ManagerDistributionWarnings(payload); len(warnings) > 0 {
		out["warnings"] = warnings
	}
	return out
}

func ManagerDistributionWarnings(payload map[string]any) []string {
	if payload == nil {
		return nil
	}
	var out []string
	for _, key := range []string{"comment", "description", "message", "error"} {
		text := strings.TrimSpace(fmt.Sprint(payload[key]))
		if text == "" || text == "<nil>" {
			continue
		}
		lower := strings.ToLower(text)
		if strings.Contains(lower, "no active") && strings.Contains(lower, "manager") {
			out = append(out, text)
		}
	}
	return out
}

func criterion(kind, label string, passed bool, expected, actual, source string) providerdiagnostics.CriterionResult {
	detail := "проверка пройдена"
	if !passed {
		detail = "проверка не пройдена"
	}
	return providerdiagnostics.CriterionResult{
		Type:     kind,
		Label:    label,
		Passed:   passed,
		Expected: expected,
		Actual:   actual,
		Source:   source,
		Detail:   detail,
	}
}

func firstFailure(results []providerdiagnostics.CriterionResult, fallback string) error {
	for _, result := range results {
		if !result.Passed {
			if fallback != "" {
				return errors.New(fallback)
			}
			return errors.New(result.Type + ": " + result.Detail)
		}
	}
	return nil
}

func hasMap(provider map[string]any, key string) bool {
	_, ok := provider[key].(map[string]any)
	return ok
}

func statusActual(provider map[string]any) string {
	code := providerdiagnostics.StatusCode(provider)
	if code == 0 {
		return "нет ответа"
	}
	return fmt.Sprintf("HTTP %d", code)
}

func rawJSONActual(provider map[string]any) string {
	if hasMap(provider, "raw_response_json") {
		return "object"
	}
	return "не распарсен"
}
