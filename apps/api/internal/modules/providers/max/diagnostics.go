package max

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	providerdiagnostics "riders.agency/pulse/api/internal/modules/providers/diagnostics"
)

func Criteria() []providerdiagnostics.CriterionSpec {
	return []providerdiagnostics.CriterionSpec{
		{
			Type:        "max_http_success",
			Label:       "HTTP ответ успешный",
			Description: "MAX POST /messages вернул код 2xx.",
			Expected:    "HTTP 200..299",
			Source:      "provider_response_json.status_code",
		},
		{
			Type:        "max_json_parsed",
			Label:       "JSON ответ распарсен",
			Description: "Ответ MAX распознан как JSON.",
			Expected:    "raw_response_json object",
			Source:      "provider_response_json.raw_response_json",
		},
		{
			Type:        "max_message_present",
			Label:       "Ответ содержит message",
			Description: "MAX вернул объект message для отправленного сообщения.",
			Expected:    "message object",
			Source:      "provider_response_json.raw_response_json.message",
		},
	}
}

func Evaluate(provider map[string]any) ([]providerdiagnostics.CriterionResult, error) {
	httpOK, detail, expected, actual, source := providerdiagnostics.HTTPStatus2xx(provider)
	results := []providerdiagnostics.CriterionResult{
		{
			Type:     "max_http_success",
			Label:    "HTTP ответ успешный",
			Passed:   httpOK,
			Expected: expected,
			Actual:   actual,
			Source:   source,
			Detail:   detail,
		},
		criterion("max_json_parsed", "JSON ответ распарсен", hasMap(provider, "raw_response_json"), "raw_response_json object", rawJSONActual(provider), "provider_response_json.raw_response_json"),
		criterion("max_message_present", "Ответ содержит message", maxMessagePresent(provider), "message object", maxMessageActual(provider), "provider_response_json.raw_response_json.message"),
	}
	return results, firstFailure(results)
}

func Analyze(provider map[string]any, retryable bool, base providerdiagnostics.Report) providerdiagnostics.Report {
	code := providerdiagnostics.StatusCode(provider)
	description := strings.ToLower(providerdiagnostics.ProviderDescription(provider))
	base.Provider = "max"
	base.Summary = "MAX не подтвердил отправку."
	switch {
	case code == http.StatusUnauthorized || code == http.StatusForbidden:
		base.ReasonCode = "max_auth_failed"
		base.HumanReason = "MAX отклонил токен или права интеграции."
		base.NextAction = "Проверьте токен MAX Bot: он должен уходить в заголовке Authorization без Bearer и без query-параметра."
		base.Retryable = false
	case code == http.StatusTooManyRequests:
		base.ReasonCode = "max_rate_limited"
		base.HumanReason = "MAX ограничил частоту запросов."
		base.NextAction = "Подождите и повторите действие."
		base.Retryable = true
	case code == http.StatusNotFound || strings.Contains(description, "chat") || strings.Contains(description, "recipient"):
		base.ReasonCode = "max_recipient_not_found"
		base.HumanReason = "MAX не нашел чат или получателя."
		base.NextAction = "Проверьте chat_id или user_id в действии и доступ бота к получателю."
		base.Retryable = false
	case code >= 500:
		base.ReasonCode = "max_provider_unavailable"
		base.HumanReason = "MAX вернул временную серверную ошибку."
		base.NextAction = "Повторите действие позже."
		base.Retryable = true
	case code == 0:
		base.ReasonCode = "max_network_error"
		base.HumanReason = "Не удалось получить HTTP ответ от MAX."
		base.NextAction = "Проверьте доступность platform-api.max.ru и повторите действие."
		base.Retryable = retryable
	default:
		base.ReasonCode = "max_provider_error"
		base.HumanReason = "MAX вернул ошибку: " + providerdiagnostics.ProviderDescription(provider)
		base.NextAction = "Проверьте настройки действия и сырой ответ MAX."
		base.Retryable = retryable
	}
	return base
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

func firstFailure(results []providerdiagnostics.CriterionResult) error {
	for _, result := range results {
		if !result.Passed {
			return errors.New(result.Type + ": " + result.Detail)
		}
	}
	return nil
}

func hasMap(provider map[string]any, key string) bool {
	_, ok := provider[key].(map[string]any)
	return ok
}

func rawJSONActual(provider map[string]any) string {
	if hasMap(provider, "raw_response_json") {
		return "object"
	}
	return "не распарсен"
}

func maxMessagePresent(provider map[string]any) bool {
	if providerdiagnostics.StringifyJSONValue(provider["message_id"]) != "" {
		return true
	}
	raw, _ := provider["raw_response_json"].(map[string]any)
	if raw == nil {
		return false
	}
	if msg, ok := raw["message"].(map[string]any); ok && msg != nil {
		return true
	}
	return raw["message"] != nil
}

func maxMessageActual(provider map[string]any) string {
	if mid := providerdiagnostics.StringifyJSONValue(provider["message_id"]); mid != "" {
		return fmt.Sprintf("message_id=%s", mid)
	}
	if maxMessagePresent(provider) {
		return "message present"
	}
	return "message missing"
}
