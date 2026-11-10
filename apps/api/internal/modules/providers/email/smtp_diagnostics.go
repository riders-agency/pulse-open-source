package email

import (
	"errors"
	"regexp"
	"strings"

	providerdiagnostics "riders.agency/pulse/api/internal/modules/providers/diagnostics"
)

var smtpStatusRE = regexp.MustCompile(`\b([245][0-9]{2})(?:\s+([245]\.[0-9]\.[0-9]+))?`)

func Criteria() []providerdiagnostics.CriterionSpec {
	return []providerdiagnostics.CriterionSpec{
		{
			Type:        "smtp_connected",
			Label:       "SMTP соединение установлено",
			Description: "Исполнитель подключился к SMTP серверу и получил баннер.",
			Expected:    "tcp_connected + smtp_banner",
			Source:      "provider_response_json.smtp.stages",
		},
		{
			Type:        "smtp_mail_from_accepted",
			Label:       "MAIL FROM принят",
			Description: "SMTP сервер принял отправителя письма.",
			Expected:    "mail_from",
			Source:      "provider_response_json.smtp.stages",
		},
		{
			Type:        "smtp_rcpt_to_accepted",
			Label:       "RCPT TO принят",
			Description: "SMTP сервер принял получателей письма.",
			Expected:    "rcpt",
			Source:      "provider_response_json.smtp.stages",
		},
		{
			Type:        "smtp_data_accepted",
			Label:       "DATA принят",
			Description: "SMTP сервер принял тело письма.",
			Expected:    "data_sent",
			Source:      "provider_response_json.smtp.stages",
		},
		{
			Type:        "smtp_quit_completed",
			Label:       "QUIT выполнен",
			Description: "SMTP сессия закрыта штатно.",
			Expected:    "quit",
			Source:      "provider_response_json.smtp.stages",
		},
		{
			Type:        "smtp_session_ok",
			Label:       "Письмо принято SMTP",
			Description: "SMTP envelope помечен как accepted=true.",
			Expected:    "smtp.accepted=true",
			Source:      "provider_response_json.smtp.accepted",
		},
	}
}

func Evaluate(provider map[string]any) ([]providerdiagnostics.CriterionResult, error) {
	envelope := SMTPEnvelope(provider)
	results := []providerdiagnostics.CriterionResult{
		stageCriterion("smtp_connected", "SMTP соединение установлено", hasStages(envelope, "tcp_connected", "smtp_banner"), "tcp_connected + smtp_banner", stagesActual(envelope)),
		stageCriterion("smtp_mail_from_accepted", "MAIL FROM принят", hasStages(envelope, "mail_from"), "mail_from", stagesActual(envelope)),
		stageCriterion("smtp_rcpt_to_accepted", "RCPT TO принят", hasStages(envelope, "rcpt"), "rcpt", stagesActual(envelope)),
		stageCriterion("smtp_data_accepted", "DATA принят", hasStages(envelope, "data_sent"), "data_sent", stagesActual(envelope)),
		stageCriterion("smtp_quit_completed", "QUIT выполнен", hasStages(envelope, "quit"), "quit", stagesActual(envelope)),
		acceptedCriterion(envelope),
	}
	return results, firstFailure(results)
}

func AnalyzeSMTP(provider map[string]any, retryable bool, base providerdiagnostics.Report) providerdiagnostics.Report {
	envelope := SMTPEnvelope(provider)
	stage := stringAt(envelope, "failed_stage")
	if stage == "" {
		stage = lastSMTPStage(envelope)
	}
	code := stringAt(envelope, "smtp_code")
	base.Provider = "smtp"
	base.Summary = "Письмо не принято SMTP сервером."
	base.Retryable = retryable
	switch {
	case stage == "auth" || code == "535" || code == "534":
		base.ReasonCode = "smtp_auth_failed"
		base.HumanReason = "SMTP сервер отклонил логин или пароль."
		base.NextAction = "Проверьте username/password, пароль приложения и способ шифрования."
		base.Retryable = false
	case stage == "rcpt_to" || code == "550" || code == "551" || code == "553":
		base.ReasonCode = "smtp_recipient_rejected"
		base.HumanReason = "SMTP сервер отклонил получателя письма."
		base.NextAction = "Проверьте адрес получателя, домен и ограничения почтового сервера."
		base.Retryable = false
	case stage == "mail_from":
		base.ReasonCode = "smtp_sender_rejected"
		base.HumanReason = "SMTP сервер отклонил отправителя."
		base.NextAction = "Проверьте from_email и разрешенный адрес отправителя в SMTP аккаунте."
		base.Retryable = false
	case stage == "starttls":
		base.ReasonCode = "smtp_tls_failed"
		base.HumanReason = "Не удалось установить защищенное SMTP соединение."
		base.NextAction = "Проверьте host/port/encryption: 465 обычно tls, 587 обычно starttls."
	case stage == "tcp_dial" || stage == "smtp_client" || code == "421":
		base.ReasonCode = "smtp_connection_failed"
		base.HumanReason = "Сервер не смог подключиться к SMTP или получить приветствие."
		base.NextAction = "Проверьте host, port, firewall и доступность SMTP сервера."
	case strings.HasPrefix(code, "4"):
		base.ReasonCode = "smtp_transient_error"
		base.HumanReason = "SMTP сервер вернул временную ошибку."
		base.NextAction = "Повторите действие позже."
		base.Retryable = true
	default:
		base.ReasonCode = "smtp_provider_error"
		base.HumanReason = "SMTP сервер вернул ошибку: " + stringAt(envelope, "last_server_reply")
		base.NextAction = "Смотрите последний ответ сервера и этап SMTP ниже."
	}
	if reply := stringAt(envelope, "last_server_reply"); reply != "" {
		base.HumanReason += " Последний ответ: " + reply
	}
	return base
}

func ExtractSMTPStatus(reply string) (code string, enhanced string) {
	match := smtpStatusRE.FindStringSubmatch(reply)
	if len(match) == 0 {
		return "", ""
	}
	if len(match) > 1 {
		code = match[1]
	}
	if len(match) > 2 {
		enhanced = match[2]
	}
	return code, enhanced
}

func IsRetryableSMTPFailure(provider map[string]any, err error) bool {
	envelope := SMTPEnvelope(provider)
	code := stringAt(envelope, "smtp_code")
	stage := stringAt(envelope, "failed_stage")
	if strings.HasPrefix(code, "4") || code == "421" {
		return true
	}
	if stage == "tcp_dial" || stage == "smtp_client" || stage == "data_write" || stage == "data_close" || stage == "quit" {
		return true
	}
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "timeout") || strings.Contains(msg, "temporary") || strings.Contains(msg, "connection reset")
}

func SMTPEnvelope(provider map[string]any) map[string]any {
	envelope, _ := provider["smtp"].(map[string]any)
	if envelope == nil {
		return map[string]any{}
	}
	return envelope
}

func stringAt(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key].(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}

func stageCriterion(kind, label string, passed bool, expected, actual string) providerdiagnostics.CriterionResult {
	detail := "этап SMTP пройден"
	if !passed {
		detail = "этап SMTP не найден"
	}
	return providerdiagnostics.CriterionResult{
		Type:     kind,
		Label:    label,
		Passed:   passed,
		Expected: expected,
		Actual:   actual,
		Source:   "provider_response_json.smtp.stages",
		Detail:   detail,
	}
}

func acceptedCriterion(envelope map[string]any) providerdiagnostics.CriterionResult {
	accepted, _ := envelope["accepted"].(bool)
	return providerdiagnostics.CriterionResult{
		Type:     "smtp_session_ok",
		Label:    "Письмо принято SMTP",
		Passed:   accepted,
		Expected: "smtp.accepted=true",
		Actual:   providerdiagnostics.StringifyJSONValue(envelope["accepted"]),
		Source:   "provider_response_json.smtp.accepted",
		Detail:   map[bool]string{true: "SMTP accepted=true", false: "SMTP accepted не true"}[accepted],
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

func hasStages(envelope map[string]any, want ...string) bool {
	stages := stageSet(envelope)
	for _, stage := range want {
		if !stages[stage] {
			return false
		}
	}
	return true
}

func stageSet(envelope map[string]any) map[string]bool {
	out := map[string]bool{}
	switch stages := envelope["stages"].(type) {
	case []any:
		for _, stage := range stages {
			if s, ok := stage.(string); ok {
				out[s] = true
			}
		}
	case []string:
		for _, stage := range stages {
			out[stage] = true
		}
	}
	return out
}

func stagesActual(envelope map[string]any) string {
	switch stages := envelope["stages"].(type) {
	case []any:
		parts := make([]string, 0, len(stages))
		for _, stage := range stages {
			if s, ok := stage.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, ",")
	case []string:
		return strings.Join(stages, ",")
	default:
		return ""
	}
}

func lastSMTPStage(envelope map[string]any) string {
	switch stages := envelope["stages"].(type) {
	case []any:
		if len(stages) == 0 {
			return ""
		}
		if v, ok := stages[len(stages)-1].(string); ok {
			return v
		}
	case []string:
		if len(stages) > 0 {
			return stages[len(stages)-1]
		}
	}
	return ""
}
