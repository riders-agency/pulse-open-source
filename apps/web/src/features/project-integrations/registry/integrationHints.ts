export function newIntegrationMetaHint(typeKey: string): string | undefined {
  switch (typeKey) {
    case "max.bot":
      return "Токен вводите как apiKey без слова Bearer: MAX принимает его в заголовке Authorization: <token>. default_chat_id/default_user_id можно опустить — получатель обычно задаётся в действии сценария. Поле api_base менять только по документации MAX.";
    case "smtp.server":
      return "Для Яндекса задайте username и from_email вашим адресом; порт 465 + encryption tls — обычная схема. Пароль вводится отдельным секретом, не в JSON.";
    case "uon.account":
      return "API-ключ U-ON вводится только в секрет. lead_source_prefix — префикс source. Для Orion Travel не задавайте default_manager_id: обращения должны оставаться без назначенного менеджера.";
    case "yclients.account":
      return "В секрет передайте JSON с partner_token, user_token либо fallback login/password и receipt_hmac_secret. Для закрытого stage callback допустима optional пара receipt_basic_username/receipt_basic_password. Здесь остаются только несекретные ID и allowlist.";
    case "telegram.relay":
      return "webhook_secret будет сохранён как hash и больше не вернётся в интерфейс. Его нужно передать в Telegram setWebhook как secret_token. country_relays — HTTPS-адреса для региональной доставки.";
    case "dadata.party":
      return "Ключ DaData будет доступен только этому проекту. base_url меняйте только если используете совместимый шлюз.";
    case "ai.apinet":
      return "APINET — OpenAI-compatible шлюз Riders Agency. Base URL должен оставаться https://apinet.cloud/v1; API-ключ вводится только в секрет.";
    case "ai.openai":
      return "API-ключ хранится зашифрованно. OpenAI используется как основной tool engine для AI-сотрудника и workflow-ноды.";
    case "ai.openrouter":
      return "OpenRouter подключается как мультимодельный каталог и provider routing. Tool loop первого этапа остается через OpenAI.";
    case "ai.google_gemini":
      return "Gemini добавлен как проектный model provider. Список моделей можно будет обновлять через AI model cache.";
    case "ai.anthropic":
      return "Claude добавлен как model provider для длинного контекста и будущих tool-сценариев.";
    case "ai.mistral":
      return "Mistral добавлен как model provider. API-ключ не пишите в JSON.";
    case "ai.groq":
      return "Groq использует OpenAI-compatible endpoint. API-ключ вводится только в секрет.";
    case "ai.xai":
      return "xAI/Grok добавлен как OpenAI-compatible provider. API-ключ вводится только в секрет.";
    case "network.vless_proxy":
      return "VLESS-ссылка хранится как секрет. Маршрут предназначен только для allowlisted AI egress, не для произвольного proxy.";
    default:
      return undefined;
  }
}

export function editIntegrationMetaHint(typeKey: string): string | undefined {
  switch (typeKey) {
    case "max.bot":
      return "Токен вводите как apiKey без слова Bearer. Поле ниже нужно только при смене токена. default_chat_id/default_user_id, default_format, default_notify и api_base задаются в JSON-настройках.";
    case "smtp.server":
      return "Пароль — только в поле ниже при смене; не добавляйте пароль в JSON.";
    case "uon.account":
      return "api_base по умолчанию https://api.u-on.ru. extended_field_ids — ID доп. полей CRM. default_office_id → r_co_id, default_lead_fields отправляется в lead/create как есть. Для Orion Travel не задавайте default_manager_id.";
    case "yclients.account":
      return "Токены, login/password fallback, receipt_hmac_secret и optional stage callback Basic Auth меняются только полным зашифрованным JSON-профилем. YooKassa и payment credentials здесь недопустимы.";
    case "telegram.relay":
      return "relay_key и relay_url видимы. webhook_secret можно добавить сюда только для ротации: backend сохранит hash и уберёт секрет из настроек.";
    case "dadata.party":
      return "base_url меняйте только если используете совместимый шлюз DaData для этого проекта.";
    case "ai.apinet":
      return "APINET — основной OpenAI-compatible шлюз Riders Agency для AI-сотрудника. Ключ меняется только отдельным секретом ниже.";
    case "ai.openai":
      return "OpenAI — основной provider для tool loop, prompt inspector и AI workflow-ноды. Ключ не хранится в JSON.";
    case "ai.openrouter":
      return "OpenRouter — мультимодельный каталог и routing. Ключ не хранится в JSON.";
    case "ai.google_gemini":
      return "Gemini provider. Меняйте только base_url/default_model_id/allowed_model_ids.";
    case "ai.anthropic":
      return "Anthropic Claude provider. Ключ меняется только отдельным секретом ниже.";
    case "ai.mistral":
      return "Mistral provider. Ключ меняется только отдельным секретом ниже.";
    case "ai.groq":
      return "Groq OpenAI-compatible provider. Ключ меняется только отдельным секретом ниже.";
    case "ai.xai":
      return "xAI provider. Ключ меняется только отдельным секретом ниже.";
    case "network.vless_proxy":
      return "VLESS route хранится как секрет и используется только для allowlisted AI egress.";
    default:
      return undefined;
  }
}
