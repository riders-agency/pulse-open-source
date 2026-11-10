package yclients

import (
	"errors"
	"fmt"

	providerdiagnostics "riders.agency/pulse/api/internal/modules/providers/diagnostics"
)

func Criteria() []providerdiagnostics.CriterionSpec {
	return []providerdiagnostics.CriterionSpec{
		{
			Type:        "yclients_certificate_issued",
			Label:       "Сертификат выпущен в YClients",
			Description: "YClients вернул и подтвердил отдельную складскую операцию сертификата.",
			Expected:    "certificate_issued=true",
			Source:      "provider_response_json.certificate_issued",
		},
		{
			Type:        "siam_receipt_delivered",
			Label:       "DemoShop принял результат выпуска",
			Description: "Подписанный receipt доставлен в DemoShop до завершения действия.",
			Expected:    "receipt_delivered=true",
			Source:      "provider_response_json.receipt_delivered",
		},
	}
}

func Evaluate(provider map[string]any) ([]providerdiagnostics.CriterionResult, error) {
	issued, _ := provider["certificate_issued"].(bool)
	receipt, _ := provider["receipt_delivered"].(bool)
	results := []providerdiagnostics.CriterionResult{
		criterion("yclients_certificate_issued", "Сертификат выпущен в YClients", issued, "true", fmt.Sprint(issued), "provider_response_json.certificate_issued"),
		criterion("siam_receipt_delivered", "DemoShop принял результат выпуска", receipt, "true", fmt.Sprint(receipt), "provider_response_json.receipt_delivered"),
	}
	for _, result := range results {
		if !result.Passed {
			return results, errors.New("YClients certificate criteria failed")
		}
	}
	return results, nil
}

func Analyze(provider map[string]any, retryable bool, base providerdiagnostics.Report) providerdiagnostics.Report {
	base.Provider = "yclients"
	base.Summary = "Выпуск сертификата в YClients не подтвержден."
	base.Retryable = retryable
	if possible, _ := provider["possible_effect"].(bool); possible {
		base.ReasonCode = "yclients_effect_unknown"
		base.HumanReason = "YClients мог принять складскую операцию, поэтому автоматический повтор остановлен."
		base.NextAction = "Сверьте операцию в YClients по operation_key и выполните ручное согласование."
		base.Retryable = false
		return base
	}
	if provider["receipt_delivered"] == false {
		base.ReasonCode = "siam_receipt_failed"
		base.HumanReason = "Результат сохранен в Pulse, но DemoShop не принял подписанный receipt."
		base.NextAction = "Повторите действие: YClients повторно вызван не будет."
		return base
	}
	base.ReasonCode = "yclients_certificate_failed"
	base.HumanReason = "YClients отклонил выпуск сертификата."
	base.NextAction = "Проверьте mapping, права филиала и настройки склада."
	return base
}

func criterion(kind, label string, passed bool, expected, actual, source string) providerdiagnostics.CriterionResult {
	detail := "проверка пройдена"
	if !passed {
		detail = "проверка не пройдена"
	}
	return providerdiagnostics.CriterionResult{
		Type: kind, Label: label, Passed: passed, Expected: expected,
		Actual: actual, Source: source, Detail: detail,
	}
}

func FailureResponse(err error) map[string]any {
	return map[string]any{
		"provider":                   "yclients",
		"ok":                         false,
		"error_code":                 CertificateErrorCode(err),
		"possible_effect":            CertificatePossibleEffect(err),
		"retryable":                  CertificateRetryable(err),
		"provider_response_redacted": true,
	}
}
