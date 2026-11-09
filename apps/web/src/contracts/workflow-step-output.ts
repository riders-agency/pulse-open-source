/**
 * Mirrors contracts/workflow_step_output_expose.md — shape of enriched action output_json from the executor.
 */

export type CriterionResultWire = {
  type: string
  label?: string
  passed: boolean
  expected?: string
  actual?: string
  source?: string
  detail?: string
  message?: string
}

export type ActionDiagnosticsWire = {
  summary?: string
  status?: "success" | "failed" | "retryable" | "skipped" | string
  provider?: string
  reason_code?: string
  human_reason?: string
  next_action?: string
  retryable?: boolean
  criteria_results?: CriterionResultWire[]
}

export type ActionOutputExposeTelegramSend = {
  chat_id?: string | number
  message_id?: string | number
}

export type ActionOutputExposeMaxSend = {
  recipient_type?: "chat_id" | "user_id" | string
  recipient_id?: string
  chat_id?: string
  user_id?: string
  message_id?: string | number
}

export type ActionOutputExposeEmailSend = {
  recipient_count?: number
  smtp_host?: string
  smtp_port?: number
}

export type ActionOutputEnvelope = {
  delivered?: boolean
  provider_ok?: boolean
  action_key?: string
  action_type_key?: string
  criteria_passed?: boolean
  criteria_results?: CriterionResultWire[]
  diagnostics?: ActionDiagnosticsWire
  /** Flat, non-sensitive summary for downstream steps and UI (“forward” data lives here). */
  expose?: Record<string, unknown>
}

export const SUCCESS_CRITERION_TYPES = [
  "http_status_2xx",
  "telegram_api_ok",
  "max_api_ok",
  "smtp_session_ok",
  "json_path_equals",
] as const

export type SuccessCriterionType = (typeof SUCCESS_CRITERION_TYPES)[number]

export function defaultCriteriaForActionType(actionTypeKey: string): SuccessCriterionWire[] {
  switch (actionTypeKey) {
    case "telegram.send_message":
      return [{ type: "http_status_2xx" }, { type: "telegram_api_ok" }]
    case "max.send_message":
      return [{ type: "http_status_2xx" }, { type: "max_api_ok" }]
    case "email.send":
      return [{ type: "smtp_session_ok" }]
    default:
      return []
  }
}

export type SuccessCriterionWire = {
  type: SuccessCriterionType | string
  json_path?: string
  equals?: string
}
