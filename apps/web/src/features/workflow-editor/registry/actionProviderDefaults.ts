import type { ActionConfig, ActionKind, RetryPolicy } from "../model/types";

export function providerActionDefault(type: ActionKind, retryPolicy: RetryPolicy): ActionConfig | undefined {
  if (type === "http.request") return {
    retry_policy: retryPolicy,
    method: "GET",
    url: "https://example.com/api/data",
    headers_json: "{\n  \"Accept\": \"application/json\"\n}",
    body: "",
    timeout_seconds: 30,
  };
  if (type === "email.send") return {
    retry_policy: retryPolicy,
    smtp_action_dependency_id: "",
    to: "",
    subject: "Событие {{ system.id }}",
    body: "Запуск: {{ system.link }}",
  };
  if (type === "telegram.send_message") return {
    retry_policy: retryPolicy,
    action_dependency_id: "",
    chat_id: "{{ telegram_relay.chat.id }}",
    text: "Принято: {{ telegram_relay.command }}\nЗапуск: {{ system.link }}",
    reply_to_message_id: "",
    parse_mode: "plain",
    reply_markup_json: "",
    disable_notification: false,
  };
  if (type === "telegram.edit_message_text") return {
    retry_policy: retryPolicy,
    action_dependency_id: "",
    chat_id: "{{ telegram_relay.chat.id }}",
    message_id: "{{ steps.send_processing.message_id }}",
    text: "Готово",
    parse_mode: "plain",
    reply_markup_json: "",
  };
  if (type === "telegram.answer_callback_query") return {
    retry_policy: { ...retryPolicy, timeout: "15s" },
    action_dependency_id: "",
    callback_query_id: "{{ telegram_relay.raw_update.callback_query.id }}",
    text: "",
    show_alert: false,
  };
  if (type === "ai.apinet.normalize_text") return {
    retry_policy: { ...retryPolicy, timeout: "45s" },
    action_dependency_id: "",
    source_text: "{{ telegram_relay.text }}",
    instructions: "Ты редактор срочных сообщений Демо. Исправь орфографию, пунктуацию и формулировки; сохрани факты, числа, имена, телефоны, адреса и смысл; не добавляй неподтвержденные данные; добавь 1-3 уместных эмодзи максимум; сделай текст понятным для публикации в Telegram-канале. Верни только готовый текст без пояснений и markdown.",
    model: "gpt-5-chat-latest",
    temperature: 0.2,
    max_tokens: 800,
  };
  if (type === "max.send_message") return {
    retry_policy: retryPolicy,
    action_dependency_id: "",
    chat_id: "",
    user_id: "",
    text: "Событие {{ system.id }}\nЗапуск: {{ system.link }}",
    format: "",
    notify: true,
    disable_link_preview: false,
  };
  if (type === "telegram.send_document") return {
    retry_policy: { ...retryPolicy, timeout: "60s" },
    action_dependency_id: "",
    chat_id: "",
    document_url: "{{ ag-gift-certificate.pdf_url }}",
    caption: "{{ ag-gift-certificate.body }}",
    disable_notification: false,
  };
  if (type === "uon.create_lead") return {
    retry_policy: retryPolicy,
    action_dependency_id: "",
    channel: "посетитель на сайте",
    note: "Посетитель на сайте\n\nСтраница: {{ current_title }}\nURL: {{ current_url }}\nСессия: {{ visit_session_id }}",
    u_name: "{{ steps.fio_normalize.u_name }}",
    u_surname: "{{ steps.fio_normalize.u_surname }}",
    u_sname: "{{ steps.fio_normalize.u_sname }}",
    extra_fields_json: '{\n  "extended_fields": {\n    "193320": "{{ current_title }}"\n  }\n}',
  };
  if (type === "yclients.certificate.issue") return {
    retry_policy: { ...retryPolicy, max_attempts: 5, timeout: "30s" },
    action_dependency_id: "",
    operation_key: "{{ webhook.operationKey }}",
    order_id: "{{ webhook.orderId }}",
    order_item_id: "{{ webhook.orderItemId }}",
    instance_id: "{{ webhook.instanceId }}",
    certificate_number: "{{ webhook.certificateNumber }}",
    product_kind: "{{ webhook.productKind }}",
    product_title: "{{ webhook.productTitle }}",
    amount_rub: "{{ webhook.amountRub }}",
    service_ids_json: "{{ webhook.serviceIdsJson }}",
  };
  return undefined;
}
