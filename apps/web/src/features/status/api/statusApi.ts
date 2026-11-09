import { apiBase, apiGet } from "../../../lib/api";

export type ReadinessPayload = {
  ok?: boolean;
  runtime?: string;
  checked_at?: string;
  ready_policy?: string;
  endpoint_note?: string;
  go_api?: Record<string, unknown>;
  python_grpc?: Record<string, unknown>;
  [key: string]: unknown;
};

export function getRuntimeReadiness(token?: string) {
  return apiGet<ReadinessPayload>("/api/v1/ai/runtime/readiness", token);
}

export function getFrontendStatus(theme: string) {
  return {
    api_base: apiBase || "same-origin",
    user_agent: typeof navigator === "undefined" ? "" : navigator.userAgent,
    timezone: typeof Intl === "undefined" ? "" : Intl.DateTimeFormat().resolvedOptions().timeZone,
    theme,
    checked_at: new Date().toISOString(),
  };
}
