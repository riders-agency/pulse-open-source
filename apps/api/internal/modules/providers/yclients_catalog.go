package providers

import yclientsprovider "riders.agency/pulse/api/internal/modules/providers/yclients"

func yclientsCatalog() CatalogEntry {
	return CatalogEntry{
		ActionTypeKey: "yclients.certificate.issue",
		ProviderKey:   "yclients",
		DisplayName:   "YClients: выпустить сертификат",
		Criteria:      yclientsprovider.Criteria(),
		FailureCategories: []FailureCategory{
			{Code: "yclients_mapping_invalid", Label: "Не найдено однозначное соответствие", Retryable: false, NextAction: "Проверьте тип сертификата, услугу, филиалы и товар YClients."},
			{Code: "yclients_provider_unavailable", Label: "YClients временно недоступен", Retryable: true, NextAction: "Повторите безопасную операцию позже."},
			{Code: "yclients_effect_unknown", Label: "Результат записи неизвестен", Retryable: false, NextAction: "Не повторяйте запись вслепую; выполните ручную сверку."},
			{Code: "siam_receipt_failed", Label: "DemoShop не принял receipt", Retryable: true, NextAction: "Повторите callback без повторного выпуска в YClients."},
		},
		Retry: RetryHint{
			Summary:            "Повтор YClients разрешен только до возможного внешнего эффекта; сохраненный успех повторно отправляет только receipt.",
			RetryableReasonIDs: []string{"yclients_provider_unavailable", "siam_receipt_failed"},
		},
		RawResponseSchema: map[string]any{
			"certificate_issued": "boolean", "receipt_delivered": "boolean",
			"certificate_type_id": "integer", "product_id": "integer",
			"location_id": "integer", "document_id": "integer", "transaction_id": "integer",
		},
		Expose: []ExposeKey{
			{Key: "operation_key", Label: "operation_key", Description: "Стабильный ключ экземпляра DemoShop."},
			{Key: "certificate_code", Label: "certificate_code", Description: "Код выпущенного сертификата YClients."},
			{Key: "transaction_id", Label: "transaction_id", Description: "Подтвержденная складская транзакция YClients."},
		},
		TemplateScopes: []string{"webhook.*", "steps.<action_key>.<expose_key>"},
		UI:             map[string]any{"card_tone": "teal", "icon": "certificate"},
	}
}
