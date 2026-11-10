package telegram

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
			Type:        "telegram_https_request",
			Label:       "HTTPS запрос до Bot API",
			Description: "Исполнитель получил HTTP ответ от Telegram Bot API.",
			Expected:    "HTTP ответ получен",
			Source:      "provider_response_json.status_code",
		},
		{
			Type:        "telegram_json_parsed",
			Label:       "JSON ответ распарсен",
			Description: "Ответ Telegram распознан как JSON.",
			Expected:    "raw_response_json object",
			Source:      "provider_response_json.raw_response_json",
		},
		{
			Type:        "telegram_ok_true",
			Label:       "Telegram вернул ok=true",
			Description: "Bot API подтвердил успешное выполнение метода.",
			Expected:    "true",
			Source:      "provider_response_json.ok",
		},
		{
			Type:        "telegram_message_id_present",
			Label:       "message_id получен",
			Description: "Telegram вернул result.message_id для отправленного сообщения.",
			Expected:    "message_id",
			Source:      "provider_response_json.message_id",
		},
	}
}

func CallbackCriteria() []providerdiagnostics.CriterionSpec {
	return []providerdiagnostics.CriterionSpec{
		{
			Type:        "telegram_https_request",
			Label:       "HTTPS запрос до Bot API",
			Description: "Исполнитель получил HTTP ответ от Telegram Bot API.",
			Expected:    "HTTP ответ получен",
			Source:      "provider_response_json.status_code",
		},
		{
			Type:        "telegram_json_parsed",
			Label:       "JSON ответ распарсен",
			Description: "Ответ Telegram распознан как JSON.",
			Expected:    "raw_response_json object",
			Source:      "provider_response_json.raw_response_json",
		},
		{
			Type:        "telegram_ok_true",
			Label:       "Telegram вернул ok=true",
			Description: "Bot API подтвердил успешное выполнение метода.",
			Expected:    "true",
			Source:      "provider_response_json.ok",
		},
	}
}

func Evaluate(provider map[string]any) ([]providerdiagnostics.CriterionResult, error) {
	results := []providerdiagnostics.CriterionResult{
		criterion("telegram_https_request", "HTTPS запрос до Bot API", providerdiagnostics.StatusCode(provider) > 0, "HTTP ответ получен", statusActual(provider), "provider_response_json.status_code"),
		criterion("telegram_json_parsed", "JSON ответ распарсен", hasMap(provider, "raw_response_json"), "raw_response_json object", rawJSONActual(provider), "provider_response_json.raw_response_json"),
		criterion("telegram_ok_true", "Telegram вернул ok=true", boolValue(provider["ok"]), "true", providerdiagnostics.StringifyJSONValue(provider["ok"]), "provider_response_json.ok"),
		criterion("telegram_message_id_present", "message_id получен", present(provider["message_id"]), "message_id", providerdiagnostics.StringifyJSONValue(provider["message_id"]), "provider_response_json.message_id"),
	}
	return results, firstFailure(results, "telegram delivery criteria failed")
}

func EvaluateCallback(provider map[string]any) ([]providerdiagnostics.CriterionResult, error) {
	results := []providerdiagnostics.CriterionResult{
		criterion("telegram_https_request", "HTTPS запрос до Bot API", providerdiagnostics.StatusCode(provider) > 0, "HTTP ответ получен", statusActual(provider), "provider_response_json.status_code"),
		criterion("telegram_json_parsed", "JSON ответ распарсен", hasMap(provider, "raw_response_json"), "raw_response_json object", rawJSONActual(provider), "provider_response_json.raw_response_json"),
		criterion("telegram_ok_true", "Telegram вернул ok=true", boolValue(provider["ok"]), "true", providerdiagnostics.StringifyJSONValue(provider["ok"]), "provider_response_json.ok"),
	}
	return results, firstFailure(results, "telegram callback criteria failed")
}

func Analyze(provider map[string]any, retryable bool, base providerdiagnostics.Report) providerdiagnostics.Report {
	code := providerdiagnostics.StatusCode(provider)
	description := strings.ToLower(providerdiagnostics.ProviderDescription(provider))
	base.Provider = "telegram"
	base.Summary = "Telegram не подтвердил отправку."
	switch {
	case code == http.StatusUnauthorized:
		base.ReasonCode = "telegram_invalid_token"
		base.HumanReason = "Telegram отклонил токен бота: HTTP 401 Unauthorized."
		base.NextAction = "Проверьте bot_token в интеграции Telegram и сохраните его заново."
		base.Retryable = false
	case code == http.StatusTooManyRequests:
		base.ReasonCode = "telegram_rate_limited"
		base.HumanReason = "Telegram ограничил частоту отправки."
		base.NextAction = retryAfterText(provider)
		base.Retryable = true
	case code == http.StatusForbidden && strings.Contains(description, "blocked"):
		base.ReasonCode = "telegram_bot_blocked"
		base.HumanReason = "Получатель заблокировал бота или бот не может писать в этот чат."
		base.NextAction = "Попросите получателя открыть чат с ботом или проверьте chat_id."
		base.Retryable = false
	case code == http.StatusForbidden:
		base.ReasonCode = "telegram_forbidden"
		base.HumanReason = "Telegram запретил отправку в этот чат."
		base.NextAction = "Проверьте права бота и chat_id."
		base.Retryable = false
	case code == http.StatusBadRequest && strings.Contains(description, "chat not found"):
		base.ReasonCode = "telegram_chat_not_found"
		base.HumanReason = "Telegram не нашел указанный chat_id."
		base.NextAction = "Проверьте chat_id или отправьте боту первое сообщение от пользователя."
		base.Retryable = false
	case code == http.StatusBadRequest && (strings.Contains(description, "parse") || strings.Contains(description, "entity")):
		base.ReasonCode = "telegram_bad_parse_mode"
		base.HumanReason = "Telegram не смог разобрать разметку сообщения."
		base.NextAction = "Исправьте parse_mode или экранирование Markdown/HTML в шаблоне."
		base.Retryable = false
	case code >= 500:
		base.ReasonCode = "telegram_provider_unavailable"
		base.HumanReason = "У Telegram временная серверная ошибка."
		base.NextAction = "Повторите действие позже; повтор безопасен."
		base.Retryable = true
	case code == 0:
		base.ReasonCode = "telegram_network_error"
		base.HumanReason = "Не удалось получить HTTP ответ от Telegram."
		base.NextAction = "Проверьте сеть сервера и повторите действие."
		base.Retryable = retryable
	default:
		base.ReasonCode = "telegram_provider_error"
		base.HumanReason = "Telegram вернул ошибку: " + providerdiagnostics.ProviderDescription(provider)
		base.NextAction = "Сверьте chat_id, текст, parse_mode и сырой ответ Telegram."
		base.Retryable = retryable
	}
	return base
}

func retryAfterText(provider map[string]any) string {
	params, _ := provider["parameters"].(map[string]any)
	if params != nil {
		if retryAfter := params["retry_after"]; retryAfter != nil {
			return fmt.Sprintf("Подождите %v секунд и повторите действие.", retryAfter)
		}
	}
	return "Подождите и повторите действие; Telegram вернул rate limit."
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
			return errors.New(result.Type + ": " + result.Detail)
		}
	}
	if fallback == "" {
		return errors.New("criteria failed")
	}
	return nil
}

func hasMap(provider map[string]any, key string) bool {
	_, ok := provider[key].(map[string]any)
	return ok
}

func boolValue(v any) bool {
	b, ok := v.(bool)
	return ok && b
}

func present(v any) bool {
	return strings.TrimSpace(providerdiagnostics.StringifyJSONValue(v)) != ""
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
