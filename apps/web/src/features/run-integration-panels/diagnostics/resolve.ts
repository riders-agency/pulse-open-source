import type { ActionDiagnosticsWire } from "../../../contracts/workflow-step-output";
import type { ActionRun } from "../../../lib/api";
import { normalizeActionOutput, normalizeDiagnostics, normalizeProvider } from "../parseOutputs";
import { integrationConfigDiagnostics, providerName, providerText, statusCode } from "./helpers";
import { analyzeMax } from "./max";
import { analyzeSMTP } from "./smtp";
import { analyzeTelegram } from "./telegram";

export function diagnosticsForActionRun(action?: ActionRun | null): ActionDiagnosticsWire | null {
  if (!action) return null;
  const stored = normalizeDiagnostics(action.diagnostics) ?? normalizeActionOutput(action.output_json).diagnostics;
  if (stored) return stored;

  const provider = normalizeProvider(action.provider_response_json);
  const providerKey = providerName(action, provider);
  const failed = action.status === "failed" || Boolean(action.error);
  const text = providerText(provider, action.error).toLowerCase();
  if (failed && (text.includes("action_dependency_id") || text.includes("dependency") || text.includes("интеграц"))) {
    return integrationConfigDiagnostics(providerKey);
  }
  const ctx = { action, provider, statusCode: statusCode(provider), text, failed };
  if (providerKey === "telegram") return analyzeTelegram(ctx);
  if (providerKey === "max") return analyzeMax(ctx);
  if (providerKey === "smtp") return analyzeSMTP(ctx);
  if (!failed) return null;
  return {
    status: "failed",
    provider: providerKey,
    reason_code: "action_failed",
    summary: "Действие завершилось с ошибкой.",
    human_reason: action.error || "Неизвестная ошибка.",
    next_action: "Проверьте input/output/provider JSON.",
    retryable: false,
  };
}
