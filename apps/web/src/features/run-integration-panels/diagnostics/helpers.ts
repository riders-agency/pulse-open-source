import type { ActionRun } from "../../../lib/api";

export function statusCode(provider: Record<string, unknown> | null) {
  const value = provider?.status_code;
  if (typeof value === "number") return value;
  if (typeof value === "string") return Number(value);
  return 0;
}

export function providerText(provider: Record<string, unknown> | null, error?: string | null) {
  for (const key of ["description", "error", "message"]) {
    const value = provider?.[key];
    if (typeof value === "string" && value.trim()) return value.trim();
  }
  return error || "";
}

export function providerName(action: ActionRun, provider: Record<string, unknown> | null) {
  const providerKey = provider?.provider;
  if (typeof providerKey === "string" && providerKey.trim()) return providerKey.trim();
  if (action.action_type_key?.startsWith("telegram.")) return "telegram";
  if (action.action_type_key?.startsWith("max.")) return "max";
  if (action.action_type_key === "email.send") return "smtp";
  return "internal";
}

export function smtpEnvelope(provider: Record<string, unknown> | null) {
  const smtp = provider?.smtp;
  return smtp && typeof smtp === "object" && !Array.isArray(smtp) ? (smtp as Record<string, unknown>) : null;
}

export function integrationConfigDiagnostics(provider: string) {
  return {
    status: "failed",
    provider,
    reason_code: "integration_config_error",
    summary: "Интеграция не выбрана или недоступна.",
    human_reason: "Действие не смогло найти активную интеграцию нужного типа.",
    next_action: "Откройте настройки действия в сценарии и выберите активную интеграцию.",
    retryable: false,
  };
}
