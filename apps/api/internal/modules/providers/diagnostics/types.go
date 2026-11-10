package diagnostics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Report struct {
	Summary         string            `json:"summary"`
	Status          string            `json:"status"`
	Provider        string            `json:"provider,omitempty"`
	ReasonCode      string            `json:"reason_code,omitempty"`
	HumanReason     string            `json:"human_reason"`
	NextAction      string            `json:"next_action,omitempty"`
	Retryable       bool              `json:"retryable"`
	CriteriaResults []CriterionResult `json:"criteria_results,omitempty"`
}

type CriterionResult struct {
	Type     string `json:"type"`
	Label    string `json:"label,omitempty"`
	Passed   bool   `json:"passed"`
	Expected string `json:"expected,omitempty"`
	Actual   string `json:"actual,omitempty"`
	Source   string `json:"source,omitempty"`
	Detail   string `json:"detail,omitempty"`
	Message  string `json:"message,omitempty"`
}

type CriterionSpec struct {
	Type        string `json:"type"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Expected    string `json:"expected,omitempty"`
	Source      string `json:"source,omitempty"`
}

func OutputWithReport(outputJSON string, report Report) string {
	var root map[string]any
	if err := json.Unmarshal([]byte(outputJSON), &root); err != nil || root == nil {
		root = map[string]any{}
	}
	root["diagnostics"] = report
	if len(report.CriteriaResults) > 0 {
		root["criteria_results"] = report.CriteriaResults
		root["criteria_passed"] = CriteriaPassed(report.CriteriaResults)
	}
	b, _ := json.Marshal(root)
	return string(b)
}

func CriteriaPassed(results []CriterionResult) bool {
	if len(results) == 0 {
		return false
	}
	for _, result := range results {
		if !result.Passed {
			return false
		}
	}
	return true
}

func ProviderName(provider map[string]any) string {
	if v, ok := provider["provider"].(string); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return "internal"
}

func ProviderDescription(provider map[string]any) string {
	for _, key := range []string{"description", "error", "message"} {
		if v, ok := provider[key].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	if raw, ok := provider["raw_response_json"].(map[string]any); ok {
		for _, key := range []string{"description", "error", "message"} {
			if v, ok := raw[key].(string); ok && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		}
	}
	return ""
}

func StatusCode(provider map[string]any) int {
	switch v := provider["status_code"].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case json.Number:
		i, _ := v.Int64()
		return int(i)
	default:
		return 0
	}
}

func StringifyAtPath(root map[string]any, path string) string {
	cur := any(root)
	for _, part := range strings.Split(path, ".") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		m, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur, ok = m[part]
		if !ok || cur == nil {
			return ""
		}
	}
	return StringifyJSONValue(cur)
}

func StringifyJSONValue(cur any) string {
	switch typed := cur.(type) {
	case string:
		return typed
	case float64:
		return fmt.Sprintf("%v", typed)
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case nil:
		return ""
	default:
		b, _ := json.Marshal(typed)
		return strings.TrimSpace(string(b))
	}
}

func Success(actionTypeKey string, provider map[string]any, criteria []CriterionResult) Report {
	providerKey := ProviderName(provider)
	report := Report{
		Status:          "success",
		Provider:        providerKey,
		ReasonCode:      providerKey + "_accepted",
		HumanReason:     "Провайдер принял действие, все критерии успеха прошли.",
		NextAction:      "Ничего исправлять не нужно.",
		CriteriaResults: criteria,
	}
	switch actionTypeKey {
	case "telegram.send_message":
		report.Summary = "Telegram подтвердил отправку сообщения."
		if mid := provider["message_id"]; mid != nil {
			report.HumanReason = fmt.Sprintf("Telegram API вернул ok=true, message_id=%v.", mid)
		}
	case "max.send_message":
		report.Summary = "MAX принял сообщение."
		if mid := provider["message_id"]; mid != nil {
			report.HumanReason = fmt.Sprintf("MAX API принял запрос, message_id=%v.", mid)
		}
	case "email.send":
		report.Summary = "SMTP сервер принял письмо."
		report.HumanReason = "SMTP сессия дошла до DATA/QUIT, сервер принял письмо к доставке."
	default:
		report.Summary = "Действие выполнено."
	}
	if warnings, ok := provider["warnings"].([]string); ok && len(warnings) > 0 {
		report.Status = "warning"
		report.ReasonCode = providerKey + "_accepted_with_warnings"
		report.Summary = "Действие выполнено с предупреждением."
		report.HumanReason = strings.Join(warnings, "; ")
		report.NextAction = "Проверьте предупреждение провайдера и исправьте операционную настройку, если оно повторяется."
		if providerKey == "uon" {
			report.NextAction = "В U-ON нет активного менеджера для распределения. Включите менеджеров/правила в кабинете или задайте default_manager_id/default_office_id в интеграции U-ON."
		}
	}
	return report
}

func CriteriaFailure(actionTypeKey string, provider map[string]any, criteria []CriterionResult, err error) Report {
	report := BaseFailure(provider, false, err)
	report.Summary = "Провайдер ответил, но критерии успеха не прошли."
	report.ReasonCode = "success_criteria_failed"
	report.HumanReason = "Ответ получен, но хотя бы один критерий успеха не совпал с ожидаемым значением."
	report.NextAction = "Откройте список критериев ниже и исправьте условие или настройки действия."
	report.CriteriaResults = criteria
	return report
}

func BaseFailure(provider map[string]any, retryable bool, err error) Report {
	description := ProviderDescription(provider)
	if description == "" && err != nil {
		description = err.Error()
	}
	status := "failed"
	if retryable {
		status = "retryable"
	}
	providerKey := ProviderName(provider)
	return Report{
		Status:      status,
		Provider:    providerKey,
		ReasonCode:  providerKey + "_error",
		Summary:     "Действие завершилось с ошибкой.",
		HumanReason: "Исполнитель получил ошибку: " + description,
		NextAction:  "Проверьте настройки действия и сырой ответ провайдера.",
		Retryable:   retryable,
	}
}

func DependencyFailure(provider map[string]any) Report {
	report := BaseFailure(provider, false, nil)
	report.ReasonCode = "integration_config_error"
	report.Summary = "Интеграция не выбрана или недоступна."
	report.HumanReason = "Действие не смогло найти активную интеграцию нужного типа."
	report.NextAction = "Откройте настройки действия и выберите активную интеграцию."
	return report
}

func ValidationFailure(provider map[string]any) Report {
	report := BaseFailure(provider, false, nil)
	report.ReasonCode = "action_validation_error"
	report.Summary = "Действие настроено не полностью."
	report.HumanReason = "Не заполнены обязательные поля действия или значение не подходит провайдеру."
	report.NextAction = "Исправьте поля действия в сценарии и запустите проверку заново."
	return report
}

func HTTPStatus2xx(provider map[string]any) (bool, string, string, string, string) {
	code := StatusCode(provider)
	if code == 0 {
		return false, "missing status_code in provider response", "HTTP 200..299", "нет status_code", "provider_response_json.status_code"
	}
	actual := fmt.Sprintf("HTTP %d", code)
	if code >= http.StatusOK && code < http.StatusMultipleChoices {
		return true, actual, "HTTP 200..299", actual, "provider_response_json.status_code"
	}
	return false, fmt.Sprintf("HTTP %d (expected 2xx)", code), "HTTP 200..299", actual, "provider_response_json.status_code"
}
