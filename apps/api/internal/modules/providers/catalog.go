package providers

import (
	providerdiagnostics "riders.agency/pulse/api/internal/modules/providers/diagnostics"
	emailprovider "riders.agency/pulse/api/internal/modules/providers/email"
	maxprovider "riders.agency/pulse/api/internal/modules/providers/max"
	telegramprovider "riders.agency/pulse/api/internal/modules/providers/telegram"
	uonprovider "riders.agency/pulse/api/internal/modules/providers/uon"
)

type FailureCategory struct {
	Code       string `json:"code"`
	Label      string `json:"label"`
	Retryable  bool   `json:"retryable"`
	NextAction string `json:"next_action"`
}

type RetryHint struct {
	Summary            string   `json:"summary"`
	RetryableReasonIDs []string `json:"retryable_reason_ids"`
}

type ExposeKey struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

type CatalogEntry struct {
	ActionTypeKey     string                              `json:"action_type_key"`
	ProviderKey       string                              `json:"provider_key"`
	DisplayName       string                              `json:"display_name"`
	Criteria          []providerdiagnostics.CriterionSpec `json:"criteria"`
	FailureCategories []FailureCategory                   `json:"failure_categories"`
	Retry             RetryHint                           `json:"retry"`
	RawResponseSchema map[string]any                      `json:"raw_response_schema"`
	Expose            []ExposeKey                         `json:"expose"`
	TemplateScopes    []string                            `json:"template_scopes"`
	UI                map[string]any                      `json:"ui"`
}

func Catalog() []CatalogEntry {
	entries := []CatalogEntry{
		telegramCatalog(),
		telegramEditMessageCatalog(),
		telegramAnswerCallbackCatalog(),
		telegramPinMessageCatalog(),
		telegramUnpinMessageCatalog(),
		telegramDocumentCatalog(),
		personFioNormalizeCatalog(),
		personPhoneNormalizeCatalog(),
		telegramRelayCatalog(),
		botStateCatalog(),
		manualPaymentTriggerCatalog(),
		paymentDueCatalog(),
		paymentCycleCollectCatalog(),
		aiAPINETNormalizeCatalog(),
		aiAgentRunCatalog(),
		maxCatalog(),
		emailCatalog(),
		uonCatalog(),
		yclientsCatalog(),
		jsonPathCatalog(),
	}
	for i := range entries {
		normalizeCatalogEntry(&entries[i])
	}
	return entries
}

func normalizeCatalogEntry(entry *CatalogEntry) {
	if entry.Criteria == nil {
		entry.Criteria = []providerdiagnostics.CriterionSpec{}
	}
	if entry.FailureCategories == nil {
		entry.FailureCategories = []FailureCategory{}
	}
	if entry.Expose == nil {
		entry.Expose = []ExposeKey{}
	}
	if entry.TemplateScopes == nil {
		entry.TemplateScopes = []string{}
	}
	if entry.RawResponseSchema == nil {
		entry.RawResponseSchema = map[string]any{}
	}
	if entry.UI == nil {
		entry.UI = map[string]any{}
	}
}

func aiAPINETNormalizeCatalog() CatalogEntry {
	return CatalogEntry{
		ActionTypeKey: "ai.apinet.normalize_text",
		ProviderKey:   "ai.apinet",
		DisplayName:   "AI: нормализация текста",
		FailureCategories: []FailureCategory{
			{Code: "ai_apinet_dependency_missing", Label: "APINET-интеграция не подключена", Retryable: false, NextAction: "Выберите активную интеграцию APINET."},
			{Code: "ai_apinet_provider_unavailable", Label: "APINET недоступен или вернул 5xx", Retryable: true, NextAction: "Повторите позже или проверьте доступность APINET."},
			{Code: "ai_apinet_rate_limited", Label: "APINET ограничил запросы", Retryable: true, NextAction: "Повторите после паузы."},
			{Code: "ai_apinet_empty_output", Label: "Модель вернула пустой текст", Retryable: false, NextAction: "Проверьте исходный текст и промпт нормализации."},
		},
		Retry: RetryHint{
			Summary:            "Повторы безопасны для сетевых ошибок, 429 и 5xx. Ошибки интеграции, пустой вход и пустой ответ не повторяются.",
			RetryableReasonIDs: []string{"ai_apinet_provider_unavailable", "ai_apinet_rate_limited"},
		},
		RawResponseSchema: map[string]any{
			"status_code": "number",
			"ok":          "boolean",
			"id":          "string",
			"model":       "string",
			"usage":       "object",
		},
		Expose: []ExposeKey{
			{Key: "original_text", Label: "original_text", Description: "Исходный текст без изменений."},
			{Key: "normalized_text", Label: "normalized_text", Description: "Текст после нормализации моделью."},
			{Key: "model", Label: "model", Description: "Фактически использованная модель."},
		},
		TemplateScopes: []string{"input.*", "telegram_relay.*", "bot_session.*", "steps.<action_key>.<expose_key>"},
		UI:             map[string]any{"card_tone": "indigo", "icon": "sparkles"},
	}
}

func aiAgentRunCatalog() CatalogEntry {
	return CatalogEntry{
		ActionTypeKey: "ai.agent.run",
		ProviderKey:   "ai",
		DisplayName:   "AI: запуск агента",
		FailureCategories: []FailureCategory{
			{Code: "ai_integration_missing", Label: "AI-интеграция не подключена", Retryable: false, NextAction: "Подключите OpenAI, OpenRouter, Gemini или другой провайдер модели в интеграциях проекта."},
			{Code: "ai_model_unsupported", Label: "Модель не поддерживает выбранный режим", Retryable: false, NextAction: "Выберите совместимую модель или измените response_mode."},
			{Code: "ai_provider_rate_limited", Label: "Провайдер ограничил запросы", Retryable: true, NextAction: "Повторите позже или выберите другой route/model."},
			{Code: "ai_tool_approval_required", Label: "Нужно подтверждение tool call", Retryable: false, NextAction: "Подтвердите действие в AI-сотруднике или измените политику tools."},
		},
		Retry: RetryHint{
			Summary:            "Повторы безопасны для чтения и model calls. Записи через tools должны быть идемпотентны и проходить approval.",
			RetryableReasonIDs: []string{"ai_provider_rate_limited"},
		},
		RawResponseSchema: map[string]any{
			"status":      "string",
			"provider":    "string",
			"model":       "string",
			"output_text": "string",
			"output_json": "object",
			"tool_calls":  "array",
			"patches":     "array",
			"usage":       "object",
			"warnings":    "array",
		},
		Expose: []ExposeKey{
			{Key: "output_text", Label: "output_text", Description: "Текстовый результат модели."},
			{Key: "output_json", Label: "output_json", Description: "Структурированный результат при json_schema/auto."},
			{Key: "usage.input_tokens", Label: "usage.input_tokens", Description: "Входные токены."},
			{Key: "usage.output_tokens", Label: "usage.output_tokens", Description: "Выходные токены."},
			{Key: "model", Label: "model", Description: "Фактически использованная модель."},
		},
		TemplateScopes: []string{"input.*", "project.*", "client.*", "task.*", "steps.<action_key>.*"},
		UI: map[string]any{
			"card_tone":     "indigo",
			"icon":          "sparkles",
			"requires_beta": true,
		},
	}
}

func manualPaymentTriggerCatalog() CatalogEntry {
	return CatalogEntry{
		ActionTypeKey: "clients.manual_payment.trigger",
		ProviderKey:   "clients",
		DisplayName:   "Клиенты: счет оплачен",
		FailureCategories: []FailureCategory{
			{Code: "manual_payment_event_missing", Label: "Нет события оплаты", Retryable: false, NextAction: "Передайте payment_event или payment_fact в input запуска."},
		},
		Retry: RetryHint{
			Summary:            "Триггер нормализует факт оплаты счета и передает его дальше в payment_due; сам не отправляет уведомления.",
			RetryableReasonIDs: []string{},
		},
		RawResponseSchema: map[string]any{
			"payment_event":  "object",
			"manual_payment": "boolean",
		},
		Expose: []ExposeKey{
			{Key: "payment_event", Label: "payment_event", Description: "Событие оплаты с источником, периодом, суммой и ссылками."},
			{Key: "payment_event.source", Label: "payment_event.source", Description: "manual, bank, incoming_email, integration или workflow."},
			{Key: "payment_event.next_period_start", Label: "payment_event.next_period_start", Description: "Дата следующего периода после внесения оплаты."},
		},
		TemplateScopes: []string{"payment_event.*", "payment_fact.*", "payment_action.*", "client.*", "project.*", "steps.<action_key>.*"},
		UI:             map[string]any{"card_tone": "slate", "trigger": true},
	}
}

func paymentDueCatalog() CatalogEntry {
	return CatalogEntry{
		ActionTypeKey: "clients.payment_due.trigger",
		ProviderKey:   "clients",
		DisplayName:   "Клиенты: счета к оплате",
		FailureCategories: []FailureCategory{
			{Code: "payment_due_project_missing", Label: "Нет project_id в запуске", Retryable: false, NextAction: "Запускайте из project-scoped trigger или передайте _pulse.project_id."},
			{Code: "payment_due_database_error", Label: "Ошибка чтения оплат", Retryable: true, NextAction: "Повторите запуск после восстановления базы."},
		},
		Retry: RetryHint{
			Summary:            "Payment due ставит счета в работу, собирает очереди “Ждем денег” и “Ждут счета” и может повторяться при временной ошибке базы.",
			RetryableReasonIDs: []string{"payment_due_database_error"},
		},
		RawResponseSchema: map[string]any{
			"payments":  "array",
			"scheduled": "array",
			"sent":      "array",
			"counts":    "object",
			"totals":    "object",
		},
		Expose: []ExposeKey{
			{Key: "digest_subject", Label: "digest_subject", Description: "Тема общего письма."},
			{Key: "digest_html", Label: "digest_html", Description: "HTML-письмо со всей картиной по клиентам."},
			{Key: "digest_telegram", Label: "digest_telegram", Description: "Telegram HTML по всей картине."},
			{Key: "digest_max", Label: "digest_max", Description: "MAX text по всей картине."},
			{Key: "payments", Label: "payments", Description: "Все найденные плановые и отправленные оплаты."},
			{Key: "counts.total", Label: "counts.total", Description: "Общее количество строк цикла."},
		},
		TemplateScopes: []string{"payment_event.*", "client.*", "project.*", "payment_action.*", "totals.*", "steps.<action_key>.*"},
		UI:             map[string]any{"card_tone": "slate", "trigger": true},
	}
}

func paymentCycleCollectCatalog() CatalogEntry {
	return CatalogEntry{
		ActionTypeKey: "clients.payment_cycle.collect",
		ProviderKey:   "clients",
		DisplayName:   "Клиенты: собрать цикл оплат",
		FailureCategories: []FailureCategory{
			{Code: "payment_cycle_project_missing", Label: "Нет project_id в запуске", Retryable: false, NextAction: "Запускайте после project-scoped trigger или передайте _pulse.project_id."},
			{Code: "payment_cycle_database_error", Label: "Ошибка чтения оплат", Retryable: true, NextAction: "Повторите запуск после восстановления базы."},
		},
		Retry: RetryHint{
			Summary:            "Сбор оплат читает внутреннюю БД и может повторяться при временной ошибке базы.",
			RetryableReasonIDs: []string{"payment_cycle_database_error"},
		},
		RawResponseSchema: map[string]any{
			"payments":  "array",
			"scheduled": "array",
			"sent":      "array",
			"counts":    "object",
		},
		Expose: []ExposeKey{
			{Key: "payments", Label: "payments", Description: "Все найденные плановые и отправленные оплаты."},
			{Key: "scheduled", Label: "scheduled", Description: "Оплаты в состоянии запланирован."},
			{Key: "sent", Label: "sent", Description: "Счета, которые уже отправлены."},
			{Key: "counts.total", Label: "counts.total", Description: "Общее количество строк цикла."},
		},
		TemplateScopes: []string{"client.*", "project.*", "payment_action.*", "totals.*", "steps.<action_key>.*"},
		UI:             map[string]any{"card_tone": "slate"},
	}
}

func botStateCatalog() CatalogEntry {
	return CatalogEntry{
		ActionTypeKey: "bot.set_state",
		ProviderKey:   "telegram",
		DisplayName:   "Bot: сохранить состояние",
		FailureCategories: []FailureCategory{
			{Code: "bot_session_missing", Label: "Нет bot session в запуске", Retryable: false, NextAction: "Используйте после telegram.relay.trigger."},
			{Code: "bot_state_validation", Label: "Некорректный state_key или context_json", Retryable: false, NextAction: "Проверьте JSON контекста и имя состояния."},
		},
		Retry: RetryHint{
			Summary:            "Смена состояния не повторяется автоматически; это локальная запись в bot_sessions.",
			RetryableReasonIDs: []string{},
		},
		RawResponseSchema: map[string]any{
			"bot_session_id": "uuid",
			"state_key":      "string",
			"context_json":   "object",
		},
		Expose: []ExposeKey{
			{Key: "bot_session_id", Label: "bot_session_id", Description: "ID сохраненной Telegram bot session."},
			{Key: "state_key", Label: "state_key", Description: "Текущее состояние диалога."},
			{Key: "context_json", Label: "context_json", Description: "Контекст состояния."},
		},
		TemplateScopes: []string{"telegram_relay.*", "bot_session.*", "steps.<action_key>.<expose_key>"},
		UI:             map[string]any{"card_tone": "slate"},
	}
}

func telegramRelayCatalog() CatalogEntry {
	return CatalogEntry{
		ActionTypeKey: "telegram.relay.trigger",
		ProviderKey:   "telegram",
		DisplayName:   "Telegram: relay входящих updates",
		FailureCategories: []FailureCategory{
			{Code: "telegram_relay_auth_failed", Label: "Неверный webhook secret", Retryable: false, NextAction: "Переустановите webhook с корректным secret_token."},
			{Code: "telegram_relay_duplicate_update", Label: "Повтор Telegram update", Retryable: false, NextAction: "Событие уже сохранено; workflow повторно не запускается."},
			{Code: "telegram_relay_no_workflow", Label: "Нет активного сценария", Retryable: false, NextAction: "Добавьте telegram.relay.trigger в активный workflow."},
			{Code: "telegram_relay_forward_failed", Label: "Региональное relay недоступно", Retryable: true, NextAction: "Проверьте HTTPS URL и авторизацию country_relays."},
		},
		Retry: RetryHint{
			Summary:            "Входящие updates сохраняются один раз по update_id; повторы Telegram считаются duplicate. Форвардинг в country_relays логируется отдельно.",
			RetryableReasonIDs: []string{"telegram_relay_forward_failed"},
		},
		RawResponseSchema: map[string]any{
			"telegram_relay_event_id": "uuid",
			"workflow_run_ids":        "array",
			"update_id":               "number",
			"update_type":             "string",
			"command":                 "string",
		},
		Expose: []ExposeKey{
			{Key: "update_id", Label: "update_id", Description: "Telegram update id."},
			{Key: "command", Label: "command", Description: "Команда без slash, например start."},
			{Key: "chat.id", Label: "chat.id", Description: "ID Telegram чата."},
			{Key: "message.text", Label: "message.text", Description: "Текст входящего сообщения."},
		},
		TemplateScopes: []string{"telegram_relay.*", "<relay_key>.*", "telegram.*", "steps.<action_key>.<expose_key>"},
		UI:             map[string]any{"card_tone": "blue"},
	}
}

func telegramCatalog() CatalogEntry {
	return CatalogEntry{
		ActionTypeKey: "telegram.send_message",
		ProviderKey:   "telegram",
		DisplayName:   "Telegram: отправка сообщения",
		Criteria:      telegramprovider.Criteria(),
		FailureCategories: []FailureCategory{
			{Code: "telegram_invalid_token", Label: "Токен бота недействителен или отозван", Retryable: false, NextAction: "Проверьте bot_token в интеграции Telegram."},
			{Code: "telegram_bot_blocked", Label: "Бот заблокирован или исключен из чата", Retryable: false, NextAction: "Проверьте chat_id и права бота."},
			{Code: "telegram_bad_parse_mode", Label: "Неверная разметка сообщения", Retryable: false, NextAction: "Исправьте parse_mode или экранирование."},
			{Code: "telegram_rate_limited", Label: "Rate limit Telegram", Retryable: true, NextAction: "Повторите после retry_after."},
			{Code: "telegram_provider_unavailable", Label: "Ошибка Telegram 5xx", Retryable: true, NextAction: "Повторите позже."},
			{Code: "telegram_network_error", Label: "Сеть, TLS или timeout", Retryable: true, NextAction: "Проверьте сеть сервера."},
		},
		Retry: RetryHint{
			Summary:            "Повторы только для network/timeout, 429 и 5xx. Ошибки токена, чата и разметки не повторяются.",
			RetryableReasonIDs: []string{"telegram_rate_limited", "telegram_provider_unavailable", "telegram_network_error"},
		},
		RawResponseSchema: map[string]any{
			"status_code":       "number",
			"ok":                "boolean",
			"description":       "string",
			"parameters":        "object",
			"message_id":        "number",
			"raw_response_json": "object",
		},
		Expose: []ExposeKey{
			{Key: "chat_id", Label: "chat_id", Description: "Получатель сообщения."},
			{Key: "message_id", Label: "message_id", Description: "ID сообщения в Telegram."},
		},
		TemplateScopes: []string{"input.*", "telegram_relay.*", "bot_session.*", "steps.<action_key>.<expose_key>"},
		UI:             map[string]any{"card_tone": "blue"},
	}
}

func telegramEditMessageCatalog() CatalogEntry {
	return CatalogEntry{
		ActionTypeKey: "telegram.edit_message_text",
		ProviderKey:   "telegram",
		DisplayName:   "Telegram: редактировать сообщение",
		Criteria:      telegramprovider.Criteria(),
		FailureCategories: []FailureCategory{
			{Code: "telegram_invalid_token", Label: "Токен бота недействителен или отозван", Retryable: false, NextAction: "Проверьте bot_token в интеграции Telegram."},
			{Code: "telegram_bad_message_id", Label: "Нельзя отредактировать сообщение", Retryable: false, NextAction: "Проверьте chat_id, message_id и права бота."},
			{Code: "telegram_bad_parse_mode", Label: "Неверная разметка сообщения", Retryable: false, NextAction: "Исправьте parse_mode или экранирование."},
			{Code: "telegram_rate_limited", Label: "Rate limit Telegram", Retryable: true, NextAction: "Повторите после retry_after."},
			{Code: "telegram_provider_unavailable", Label: "Ошибка Telegram 5xx", Retryable: true, NextAction: "Повторите позже."},
		},
		Retry: RetryHint{
			Summary:            "Повторы только для network/timeout, 429 и 5xx. Ошибки message_id, чата и разметки не повторяются.",
			RetryableReasonIDs: []string{"telegram_rate_limited", "telegram_provider_unavailable", "telegram_network_error"},
		},
		RawResponseSchema: map[string]any{
			"status_code": "number",
			"ok":          "boolean",
			"message_id":  "number",
		},
		Expose: []ExposeKey{
			{Key: "chat_id", Label: "chat_id", Description: "Чат сообщения."},
			{Key: "message_id", Label: "message_id", Description: "ID отредактированного сообщения."},
		},
		TemplateScopes: []string{"input.*", "telegram_relay.*", "bot_session.*", "steps.<action_key>.<expose_key>"},
		UI:             map[string]any{"card_tone": "blue"},
	}
}

func telegramAnswerCallbackCatalog() CatalogEntry {
	return CatalogEntry{
		ActionTypeKey: "telegram.answer_callback_query",
		ProviderKey:   "telegram",
		DisplayName:   "Telegram: ответить на callback",
		Criteria:      telegramprovider.CallbackCriteria(),
		FailureCategories: []FailureCategory{
			{Code: "telegram_invalid_token", Label: "Токен бота недействителен или отозван", Retryable: false, NextAction: "Проверьте bot_token в интеграции Telegram."},
			{Code: "telegram_bad_callback_query", Label: "Callback query уже устарел или неверен", Retryable: false, NextAction: "Проверьте callback_query_id."},
			{Code: "telegram_rate_limited", Label: "Rate limit Telegram", Retryable: true, NextAction: "Повторите после retry_after."},
			{Code: "telegram_provider_unavailable", Label: "Ошибка Telegram 5xx", Retryable: true, NextAction: "Повторите позже."},
		},
		Retry: RetryHint{
			Summary:            "Callback answer нужен для inline-кнопок, чтобы Telegram-клиент закрыл индикатор ожидания.",
			RetryableReasonIDs: []string{"telegram_rate_limited", "telegram_provider_unavailable", "telegram_network_error"},
		},
		RawResponseSchema: map[string]any{
			"status_code": "number",
			"ok":          "boolean",
		},
		TemplateScopes: []string{"telegram_relay.*", "steps.<action_key>.<expose_key>"},
		UI:             map[string]any{"card_tone": "blue"},
	}
}

func telegramPinMessageCatalog() CatalogEntry {
	return CatalogEntry{
		ActionTypeKey: "telegram.pin_message",
		ProviderKey:   "telegram",
		DisplayName:   "Telegram: закрепить сообщение",
		Criteria:      telegramprovider.Criteria(),
		FailureCategories: []FailureCategory{
			{Code: "telegram_invalid_token", Label: "Токен бота недействителен или отозван", Retryable: false, NextAction: "Проверьте bot_token в интеграции Telegram."},
			{Code: "telegram_bad_message_id", Label: "Нельзя закрепить сообщение", Retryable: false, NextAction: "Проверьте chat_id, message_id и права бота."},
			{Code: "telegram_rate_limited", Label: "Rate limit Telegram", Retryable: true, NextAction: "Повторите после retry_after."},
			{Code: "telegram_provider_unavailable", Label: "Ошибка Telegram 5xx", Retryable: true, NextAction: "Повторите позже."},
		},
		Retry: RetryHint{
			Summary:            "Повторы только для network/timeout, 429 и 5xx. Ошибки прав бота, чата и message_id не повторяются.",
			RetryableReasonIDs: []string{"telegram_rate_limited", "telegram_provider_unavailable", "telegram_network_error"},
		},
		RawResponseSchema: map[string]any{
			"status_code": "number",
			"ok":          "boolean",
		},
		Expose: []ExposeKey{
			{Key: "chat_id", Label: "chat_id", Description: "Чат закрепленного сообщения."},
			{Key: "message_id", Label: "message_id", Description: "ID закрепленного сообщения."},
		},
		TemplateScopes: []string{"input.*", "telegram_relay.*", "bot_session.*", "steps.<action_key>.<expose_key>"},
		UI:             map[string]any{"card_tone": "blue"},
	}
}

func telegramUnpinMessageCatalog() CatalogEntry {
	return CatalogEntry{
		ActionTypeKey: "telegram.unpin_message",
		ProviderKey:   "telegram",
		DisplayName:   "Telegram: открепить сообщение",
		Criteria:      telegramprovider.Criteria(),
		FailureCategories: []FailureCategory{
			{Code: "telegram_invalid_token", Label: "Токен бота недействителен или отозван", Retryable: false, NextAction: "Проверьте bot_token в интеграции Telegram."},
			{Code: "telegram_bad_message_id", Label: "Нельзя открепить сообщение", Retryable: false, NextAction: "Проверьте chat_id, message_id и права бота."},
			{Code: "telegram_rate_limited", Label: "Rate limit Telegram", Retryable: true, NextAction: "Повторите после retry_after."},
			{Code: "telegram_provider_unavailable", Label: "Ошибка Telegram 5xx", Retryable: true, NextAction: "Повторите позже."},
		},
		Retry: RetryHint{
			Summary:            "Повторы только для network/timeout, 429 и 5xx. Ошибки прав бота, чата и message_id не повторяются.",
			RetryableReasonIDs: []string{"telegram_rate_limited", "telegram_provider_unavailable", "telegram_network_error"},
		},
		RawResponseSchema: map[string]any{
			"status_code": "number",
			"ok":          "boolean",
		},
		Expose: []ExposeKey{
			{Key: "chat_id", Label: "chat_id", Description: "Чат открепленного сообщения."},
			{Key: "message_id", Label: "message_id", Description: "ID открепленного сообщения."},
		},
		TemplateScopes: []string{"input.*", "telegram_relay.*", "bot_session.*", "steps.<action_key>.<expose_key>"},
		UI:             map[string]any{"card_tone": "blue"},
	}
}

func telegramDocumentCatalog() CatalogEntry {
	return CatalogEntry{
		ActionTypeKey: "telegram.send_document",
		ProviderKey:   "telegram",
		DisplayName:   "Telegram: отправка документа",
		Criteria:      telegramprovider.Criteria(),
		FailureCategories: []FailureCategory{
			{Code: "telegram_document_download_failed", Label: "Не удалось скачать document_url", Retryable: true, NextAction: "Проверьте доступность PDF URL из webhook."},
			{Code: "telegram_invalid_token", Label: "Токен бота недействителен", Retryable: false, NextAction: "Проверьте bot_token."},
			{Code: "telegram_bot_blocked", Label: "Бот не может отправить файл в чат", Retryable: false, NextAction: "Проверьте chat_id."},
			{Code: "telegram_rate_limited", Label: "Rate limit Telegram", Retryable: true, NextAction: "Повторите позже."},
		},
		Retry: RetryHint{
			Summary:            "Повторы для сетевых ошибок, download и 429/5xx.",
			RetryableReasonIDs: []string{"telegram_document_download_failed", "telegram_rate_limited", "telegram_provider_unavailable", "telegram_network_error"},
		},
		RawResponseSchema: map[string]any{
			"status_code": "number",
			"ok":          "boolean",
			"message_id":  "number",
		},
		Expose: []ExposeKey{
			{Key: "chat_id", Label: "chat_id", Description: "Получатель."},
			{Key: "message_id", Label: "message_id", Description: "ID сообщения с документом."},
		},
		TemplateScopes: []string{"input.*", "steps.<action_key>.<expose_key>"},
		UI:             map[string]any{"card_tone": "blue"},
	}
}

func personFioNormalizeCatalog() CatalogEntry {
	return CatalogEntry{
		ActionTypeKey: "person.fio.normalize",
		ProviderKey:   "internal",
		DisplayName:   "ФИО: нормализация",
		Retry: RetryHint{
			Summary:            "Локальная нормализация без внешних вызовов; повторы не требуются.",
			RetryableReasonIDs: []string{},
		},
		RawResponseSchema: map[string]any{
			"u_name":     "string",
			"u_surname":  "string",
			"u_sname":    "string",
			"mode":       "string",
			"changed":    "boolean",
			"is_valid":   "boolean",
			"confidence": "string",
			"warnings":   "array",
		},
		Expose: []ExposeKey{
			{Key: "u_name", Label: "u_name", Description: "Имя."},
			{Key: "u_surname", Label: "u_surname", Description: "Фамилия."},
			{Key: "u_sname", Label: "u_sname", Description: "Отчество."},
			{Key: "mode", Label: "mode", Description: "empty | single_name | fio_normalized."},
			{Key: "changed", Label: "changed", Description: "Было ли исправление по словарю."},
			{Key: "is_valid", Label: "is_valid", Description: "Есть ли непустое имя для отправки."},
			{Key: "confidence", Label: "confidence", Description: "high | medium | low | none."},
			{Key: "warnings", Label: "warnings", Description: "Предупреждения о неоднозначном разборе."},
			{Key: "full_name", Label: "full_name", Description: "Собранное ФИО после нормализации."},
		},
		TemplateScopes: []string{"input.*", "steps.<action_key>.<expose_key>"},
		UI:             map[string]any{"card_tone": "slate"},
	}
}

func personPhoneNormalizeCatalog() CatalogEntry {
	return CatalogEntry{
		ActionTypeKey: "person.phone.normalize",
		ProviderKey:   "internal",
		DisplayName:   "Телефон: нормализация",
		Retry: RetryHint{
			Summary:            "Локальная нормализация без внешних вызовов; повторы не требуются.",
			RetryableReasonIDs: []string{},
		},
		RawResponseSchema: map[string]any{
			"phone_e164": "string",
			"phone_raw":  "string",
			"mode":       "string",
			"changed":    "boolean",
			"is_valid":   "boolean",
			"warning":    "string",
		},
		Expose: []ExposeKey{
			{Key: "phone_e164", Label: "phone_e164", Description: "Телефон в формате +7XXXXXXXXXX."},
			{Key: "phone_raw", Label: "phone_raw", Description: "Исходная строка телефона."},
			{Key: "mode", Label: "mode", Description: "empty | ru_11 | ru_10 | international | passthrough."},
			{Key: "changed", Label: "changed", Description: "Было ли изменение формата."},
			{Key: "is_valid", Label: "is_valid", Description: "Телефон похож на валидный номер."},
			{Key: "warning", Label: "warning", Description: "Причина, если телефон не нормализован."},
		},
		TemplateScopes: []string{"input.*", "steps.<action_key>.<expose_key>"},
		UI:             map[string]any{"card_tone": "slate"},
	}
}

func maxCatalog() CatalogEntry {
	return CatalogEntry{
		ActionTypeKey: "max.send_message",
		ProviderKey:   "max",
		DisplayName:   "MAX: отправка сообщения",
		Criteria:      maxprovider.Criteria(),
		FailureCategories: []FailureCategory{
			{Code: "max_auth_failed", Label: "Токен MAX недействителен или отозван", Retryable: false, NextAction: "Проверьте токен MAX Bot: заголовок должен быть Authorization: <token> без Bearer."},
			{Code: "max_recipient_not_found", Label: "Неверный chat_id/user_id или нет доступа", Retryable: false, NextAction: "Проверьте получателя и доступ бота к чату."},
			{Code: "action_validation_error", Label: "Текст длиннее 4000 или неверный формат", Retryable: false, NextAction: "Исправьте настройки действия."},
			{Code: "max_rate_limited", Label: "Rate limit MAX", Retryable: true, NextAction: "Повторите позже."},
			{Code: "max_provider_unavailable", Label: "MAX 503/5xx", Retryable: true, NextAction: "Повторите позже."},
			{Code: "max_network_error", Label: "Сеть, TLS или timeout", Retryable: true, NextAction: "Проверьте доступность API MAX."},
		},
		Retry: RetryHint{
			Summary:            "Повторы только для network/timeout, 429 и 5xx/503. Auth, chat_id/user_id и формат не повторяются.",
			RetryableReasonIDs: []string{"max_rate_limited", "max_provider_unavailable", "max_network_error"},
		},
		RawResponseSchema: map[string]any{
			"status_code":       "number",
			"ok":                "boolean",
			"description":       "string",
			"message_id":        "string",
			"raw_response_json": "object",
		},
		Expose: []ExposeKey{
			{Key: "recipient_type", Label: "recipient_type", Description: "Тип получателя: chat_id или user_id."},
			{Key: "recipient_id", Label: "recipient_id", Description: "ID получателя MAX."},
			{Key: "chat_id", Label: "chat_id", Description: "Получатель-чата, если сообщение отправлено в чат."},
			{Key: "user_id", Label: "user_id", Description: "Получатель-пользователь, если сообщение отправлено пользователю."},
			{Key: "message_id", Label: "message_id", Description: "ID сообщения MAX."},
		},
		TemplateScopes: []string{"input.*", "steps.<action_key>.<expose_key>"},
		UI:             map[string]any{"card_tone": "violet"},
	}
}

func uonCatalog() CatalogEntry {
	return CatalogEntry{
		ActionTypeKey: "uon.create_lead",
		ProviderKey:   "uon",
		DisplayName:   "U-ON: создать обращение",
		Criteria:      uonprovider.Criteria(),
		FailureCategories: []FailureCategory{
			{Code: "uon_invalid_api_key", Label: "API-ключ U-ON недействителен", Retryable: false, NextAction: "Проверьте api_key в интеграции U-ON."},
			{Code: "uon_result_error", Label: "U-ON отклонил обращение", Retryable: false, NextAction: "Проверьте note, source и контактные поля."},
			{Code: "uon_accepted_with_warnings", Label: "U-ON принял с предупреждением", Retryable: false, NextAction: "Проверьте настройки распределения в U-ON. Для Orion Travel обращения должны оставаться без назначенного менеджера."},
			{Code: "uon_provider_unavailable", Label: "U-ON 5xx", Retryable: true, NextAction: "Повторите позже."},
			{Code: "uon_network_error", Label: "Сеть, TLS или timeout", Retryable: true, NextAction: "Проверьте доступность api.u-on.ru."},
		},
		Retry: RetryHint{
			Summary:            "Повторы только для network/timeout и 5xx. Ошибки ключа и валидации полей не повторяются.",
			RetryableReasonIDs: []string{"uon_provider_unavailable", "uon_network_error"},
		},
		RawResponseSchema: map[string]any{
			"status_code":       "number",
			"result":            "number",
			"description":       "string",
			"raw_response_json": "object",
		},
		Expose: []ExposeKey{
			{Key: "source", Label: "source", Description: "Источник обращения в U-ON."},
			{Key: "note", Label: "note", Description: "Текст обращения."},
			{Key: "u_name", Label: "u_name", Description: "Имя, отправленное в U-ON."},
			{Key: "u_surname", Label: "u_surname", Description: "Фамилия, отправленная в U-ON."},
			{Key: "u_sname", Label: "u_sname", Description: "Отчество, отправленное в U-ON."},
			{Key: "r_co_id", Label: "r_co_id", Description: "ID офиса U-ON для назначения обращения."},
		},
		TemplateScopes: []string{"input.*", "steps.<action_key>.<expose_key>"},
		UI:             map[string]any{"card_tone": "emerald"},
	}
}

func emailCatalog() CatalogEntry {
	return CatalogEntry{
		ActionTypeKey: "email.send",
		ProviderKey:   "email",
		DisplayName:   "Email: SMTP отправка",
		Criteria:      emailprovider.Criteria(),
		FailureCategories: []FailureCategory{
			{Code: "smtp_auth_failed", Label: "AUTH отклонен", Retryable: false, NextAction: "Проверьте username/password и пароль приложения."},
			{Code: "smtp_sender_rejected", Label: "MAIL FROM отклонен", Retryable: false, NextAction: "Проверьте from_email."},
			{Code: "smtp_recipient_rejected", Label: "RCPT TO отклонен", Retryable: false, NextAction: "Проверьте адрес получателя."},
			{Code: "smtp_tls_failed", Label: "TLS/STARTTLS не установлен", Retryable: true, NextAction: "Проверьте host/port/encryption."},
			{Code: "smtp_connection_failed", Label: "DNS/connect/no answer", Retryable: true, NextAction: "Проверьте SMTP host, port и firewall."},
			{Code: "smtp_transient_error", Label: "Временная SMTP 4xx", Retryable: true, NextAction: "Повторите позже."},
			{Code: "smtp_provider_error", Label: "Постоянная SMTP 5xx или policy/spam", Retryable: false, NextAction: "Смотрите enhanced status code и ответ сервера."},
		},
		Retry: RetryHint{
			Summary:            "Повторы для SMTP 4xx, connect/timeout и временных сбоев DATA/QUIT. AUTH и permanent 5xx не повторяются.",
			RetryableReasonIDs: []string{"smtp_tls_failed", "smtp_connection_failed", "smtp_transient_error"},
		},
		RawResponseSchema: map[string]any{
			"smtp.accepted":             "boolean",
			"smtp.stages":               "array",
			"smtp.failed_stage":         "string",
			"smtp.smtp_code":            "string",
			"smtp.enhanced_status_code": "string",
			"smtp.last_server_reply":    "string",
		},
		Expose: []ExposeKey{
			{Key: "recipient_count", Label: "recipient_count", Description: "Количество получателей."},
			{Key: "smtp_host", Label: "smtp_host", Description: "SMTP host."},
			{Key: "smtp_port", Label: "smtp_port", Description: "SMTP port."},
		},
		TemplateScopes: []string{"input.*", "steps.<action_key>.<expose_key>"},
		UI:             map[string]any{"card_tone": "amber"},
	}
}

func jsonPathCatalog() CatalogEntry {
	return CatalogEntry{
		ActionTypeKey: "json_path",
		ProviderKey:   "internal",
		DisplayName:   "JSON path проверка",
		Criteria: []providerdiagnostics.CriterionSpec{
			{
				Type:        "json_path_equals",
				Label:       "JSON path равен ожидаемому значению",
				Description: "Backend сравнивает path/operator/expected/source без произвольных типов.",
				Expected:    "expected",
				Source:      "input или provider_response_json",
			},
		},
		Retry: RetryHint{
			Summary:            "JSON path проверка не повторяется автоматически.",
			RetryableReasonIDs: []string{},
		},
		RawResponseSchema: map[string]any{"source": "input|provider_response_json", "path": "string", "operator": "equals", "expected": "string"},
		TemplateScopes:    []string{"input.*", "steps.<action_key>.<expose_key>"},
		UI:                map[string]any{"card_tone": "slate"},
	}
}
