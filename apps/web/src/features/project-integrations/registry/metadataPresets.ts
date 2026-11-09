import { telegramRelayPreset } from "./telegramRelayPreset";

const metadataPresets: Record<string, string> = {
  "max.bot": JSON.stringify(
    {
      default_chat_id: "",
      default_user_id: "",
      default_format: "",
      default_notify: true,
      api_base: "https://platform-api.max.ru",
    },
    null,
    2,
  ),
  "telegram.bot": JSON.stringify({}, null, 2),
  "smtp.server": JSON.stringify(
    {
      host: "smtp.yandex.ru",
      port: 465,
      username: "you@yandex.ru",
      from_email: "you@yandex.ru",
      from_name: "",
      encryption: "tls",
    },
    null,
    2,
  ),
  "uon.account": JSON.stringify(
    {
      api_base: "https://api.u-on.ru",
      lead_source_prefix: "example.test",
      allow_note_only: true,
      default_office_id: "",
      default_lead_fields: {},
      extended_field_ids: {
        tour_name: "193320",
        region: "192647",
        tourist_count_text: "193640",
      },
    },
    null,
    2,
  ),
  "yclients.account": JSON.stringify(
    {
      api_base: "https://api.yclients.com/api/v1",
      mode: "disabled",
      chain_id: 1972020,
      location_id: 0,
      storage_id: 0,
      salon_ids: [],
      allow_type_create: false,
      resolution_ttl_seconds: 600,
      receipt_url: "",
      receipt_audience: "",
      receipt_allowed_hosts: [],
    },
    null,
    2,
  ),
  "email.template": JSON.stringify({}, null, 2),
  "dadata.party": JSON.stringify(
    {
      base_url: "https://suggestions.dadata.ru/suggestions/api/4_1/rs",
    },
    null,
    2,
  ),
  "ai.apinet": JSON.stringify(
    {
      base_url: "https://apinet.cloud/v1",
      default_model_id: "",
      allowed_model_ids: [],
      model_refresh_ttl_seconds: 86400,
      enable_chat_completions: true,
      enable_embeddings: false,
      enable_tool_call_probe: true,
      privacy_profile: "standard",
    },
    null,
    2,
  ),
  "ai.openai": JSON.stringify(
    {
      base_url: "https://api.openai.com/v1",
      default_model_id: "gpt-5",
      allowed_model_ids: [],
      model_refresh_ttl_seconds: 900,
      enable_tools: true,
      enable_patch_tool: true,
      privacy_profile: "standard",
    },
    null,
    2,
  ),
  "ai.openrouter": JSON.stringify(
    {
      base_url: "https://openrouter.ai/api/v1",
      default_model_id: "openrouter/auto",
      allowed_model_ids: [],
      allow_fallbacks: true,
      require_zdr: false,
    },
    null,
    2,
  ),
  "ai.google_gemini": JSON.stringify(
    {
      base_url: "https://generativelanguage.googleapis.com/v1beta",
      default_model_id: "gemini-2.5-pro",
      allowed_model_ids: [],
    },
    null,
    2,
  ),
  "ai.anthropic": JSON.stringify(
    {
      base_url: "https://api.anthropic.com/v1",
      default_model_id: "claude-sonnet-4-20250514",
      allowed_model_ids: [],
    },
    null,
    2,
  ),
  "ai.mistral": JSON.stringify(
    {
      base_url: "https://api.mistral.ai/v1",
      default_model_id: "mistral-large-latest",
      allowed_model_ids: [],
    },
    null,
    2,
  ),
  "ai.groq": JSON.stringify(
    {
      base_url: "https://api.groq.com/openai/v1",
      default_model_id: "llama-3.3-70b-versatile",
      allowed_model_ids: [],
    },
    null,
    2,
  ),
  "ai.xai": JSON.stringify(
    {
      base_url: "https://api.x.ai/v1",
      default_model_id: "grok-4",
      allowed_model_ids: [],
    },
    null,
    2,
  ),
  "network.vless_proxy": JSON.stringify(
    {
      route_key: "ai-primary",
      route_kind: "vless_proxy",
      status: "disabled",
      allow_hosts: ["api.openai.com", "openrouter.ai"],
      fail_mode: "fail_closed",
    },
    null,
    2,
  ),
};

export function metadataPreset(typeKey: string) {
  if (typeKey === "telegram.relay") return telegramRelayPreset();
  return metadataPresets[typeKey] ?? "{}";
}
