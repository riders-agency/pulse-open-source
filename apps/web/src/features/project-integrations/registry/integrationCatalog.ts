type ProjectIntegrationCatalogItem = {
  key: string;
  label: string;
  secret: string;
  editSecret: string;
  secretHint?: string;
  showMetadata: boolean;
};

export type NewProjectIntegrationType = {
  key: string;
  label: string;
  secret: string;
  showMetadata: boolean;
};

export type EditProjectIntegrationType = {
  key: string;
  label: string;
  secret: string;
  secretHint: string;
  showMetadata: boolean;
};

const projectIntegrationCatalog: ProjectIntegrationCatalogItem[] = [
  { key: "max.bot", label: "Бот MAX", secret: "bot_token", editSecret: "Новый токен бота", showMetadata: true },
  { key: "telegram.bot", label: "Бот Telegram", secret: "bot_token", editSecret: "Новый токен бота", showMetadata: false },
  {
    key: "telegram.relay",
    label: "Ретранслятор Telegram",
    secret: "bot_token",
    editSecret: "Новый токен бота",
    secretHint: "Пусто — токен бота не меняем. Новый webhook_secret можно записать в JSON-настройки для ротации проверки.",
    showMetadata: true,
  },
  {
    key: "smtp.server",
    label: "SMTP-сервер",
    secret: "password",
    editSecret: "Новый пароль",
    secretHint: "Пусто — пароль не меняем.",
    showMetadata: true,
  },
  {
    key: "uon.account",
    label: "U-ON.Travel CRM",
    secret: "api_key",
    editSecret: "Новый API-ключ",
    secretHint: "Ключ из личного кабинета U-ON. Пусто — ключ не меняем.",
    showMetadata: true,
  },
  {
    key: "yclients.account",
    label: "YClients · сертификаты",
    secret: "credential_profile_json",
    editSecret: "Новый профиль доступа JSON",
    secretHint: "Пусто — профиль не меняем. Токены, fallback login/password, receipt secret и optional stage callback Basic Auth хранятся только в этом зашифрованном JSON.",
    showMetadata: true,
  },
  { key: "email.template", label: "Шаблон письма", secret: "", editSecret: "", showMetadata: true },
  {
    key: "dadata.party",
    label: "DaData",
    secret: "api_key",
    editSecret: "Новый API-ключ",
    secretHint: "Пусто — ключ DaData не меняем.",
    showMetadata: true,
  },
  { key: "ai.apinet", label: "APINET / Riders Agency AI Gateway", secret: "api_key", editSecret: "Новый API-ключ", secretHint: "Пусто — ключ APINET не меняем.", showMetadata: true },
  { key: "ai.openai", label: "OpenAI", secret: "api_key", editSecret: "Новый API-ключ", secretHint: "Пусто — ключ OpenAI не меняем.", showMetadata: true },
  { key: "ai.openrouter", label: "OpenRouter", secret: "api_key", editSecret: "Новый API-ключ", secretHint: "Пусто — ключ OpenRouter не меняем.", showMetadata: true },
  { key: "ai.google_gemini", label: "Google Gemini", secret: "api_key", editSecret: "Новый API-ключ", secretHint: "Пусто — ключ Gemini не меняем.", showMetadata: true },
  { key: "ai.anthropic", label: "Anthropic Claude", secret: "api_key", editSecret: "Новый API-ключ", secretHint: "Пусто — ключ Anthropic не меняем.", showMetadata: true },
  { key: "ai.mistral", label: "Mistral AI", secret: "api_key", editSecret: "Новый API-ключ", secretHint: "Пусто — ключ Mistral не меняем.", showMetadata: true },
  { key: "ai.groq", label: "Groq", secret: "api_key", editSecret: "Новый API-ключ", secretHint: "Пусто — ключ Groq не меняем.", showMetadata: true },
  { key: "ai.xai", label: "xAI", secret: "api_key", editSecret: "Новый API-ключ", secretHint: "Пусто — ключ xAI не меняем.", showMetadata: true },
  { key: "network.vless_proxy", label: "VLESS route", secret: "vless_url", editSecret: "Новая VLESS URL", secretHint: "Пусто — маршрут не меняем.", showMetadata: true },
];

export const newProjectIntegrationTypes: NewProjectIntegrationType[] = projectIntegrationCatalog.map((item) => ({
  key: item.key,
  label: item.label,
  secret: item.secret,
  showMetadata: item.showMetadata,
}));

export const editProjectIntegrationTypes: EditProjectIntegrationType[] = projectIntegrationCatalog.map((item) => ({
  key: item.key,
  label: item.label,
  secret: item.editSecret,
  secretHint: item.secretHint ?? "",
  showMetadata: item.showMetadata,
}));
