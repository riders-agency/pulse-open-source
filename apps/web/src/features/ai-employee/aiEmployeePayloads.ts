import type { AIChatMessage, AIToolCall } from "../../lib/api";
import { dateLabel } from "./aiEmployeeFormatters";
export { toolLabel, toolResultTitle } from "./aiEmployeeToolPayloads";

export function messageKey(message: AIChatMessage, index: number) {
  return message.ai_chat_message_id || `${message.role}-${index}`;
}

export function jsonPreview(value: unknown) {
  try {
    return JSON.stringify(value ?? {}, null, 2);
  } catch {
    return "{}";
  }
}

export function asRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" && !Array.isArray(value) ? value as Record<string, unknown> : {};
}

export function textValue(value: unknown) {
  return typeof value === "string" ? value : "";
}

export function entityLinkFromPayload(payload?: Record<string, unknown>) {
  const safe = asRecord(payload);
  const output = asRecord(safe.output_json || safe.output);
  const url = textValue(output.url) || textValue(safe.url);
  if (!url) return null;
  const label =
    textValue(output.task_key) ||
    textValue(output.title) ||
    textValue(output.name) ||
    textValue(output.workflow_id) ||
    textValue(output.client_name) ||
    textValue(output.payment_action_id) ||
    "Открыть сущность";
  return { url, label };
}

export function taskSearchItems(payload?: Record<string, unknown>) {
  const outputs = payload?.kind === "tool_results"
    ? recordList(payload.results_json).map((result) => asRecord(result.output_json))
    : [asRecord(payload?.output_json)];
  const rawTasks = outputs.flatMap((output) => Array.isArray(output.tasks) ? output.tasks : []);
  return rawTasks.map((item) => asRecord(item)).filter((item) => textValue(item.task_id) || textValue(item.url));
}

export function hasRenderableSnippet(message: AIChatMessage) {
  if (entityLinkFromPayload(message.payload_json)) return true;
  if (taskSearchItems(message.payload_json).length > 0) return true;
  if (message.payload_json?.kind === "tool_result") return true;
  if (message.payload_json?.kind === "tool_results") return true;
  if (message.payload_json?.kind === "card") return true;
  return message.payload_json?.kind === "proposal";
}

export function isInternalPlanText(value: unknown) {
  const body = typeof value === "string" ? value.trim() : "";
  if (!body) return false;
  return /^план действия[:\s]*$/i.test(body);
}

export function hasRenderableMessageBody(message: AIChatMessage) {
  const body = String(message.body || "").trim();
  return body.length > 0 && !isInternalPlanText(body);
}

export function pendingCallForMessage(message: AIChatMessage, calls: AIToolCall[]) {
  const id = typeof message.payload_json?.ai_tool_call_id === "string" ? message.payload_json.ai_tool_call_id : "";
  if (!id) return undefined;
  return calls.find((call) => call.ai_tool_call_id === id);
}

export function activePendingCall(calls: AIToolCall[]) {
  return [...calls].reverse().find((call) => call.status === "pending_approval");
}

export function lastAssistantMessageID(messages: AIChatMessage[]) {
  return [...messages].reverse().find((message) => message.role === "assistant")?.ai_chat_message_id || "";
}

export function boolLabel(value: unknown) {
  return value === true ? "Да" : "Нет";
}

export function priorityLabel(value: unknown) {
  switch (textValue(value)) {
    case "high":
      return "Высокий";
    case "urgent":
      return "Срочный";
    case "low":
      return "Низкий";
    default:
      return "Средний";
  }
}

export function stringList(value: unknown) {
  if (Array.isArray(value)) return value.map(String).filter(Boolean);
  return [];
}

export function proposalInput(call: AIToolCall | undefined, message: AIChatMessage) {
  return asRecord(call?.input_json || message.payload_json?.fields_json || message.payload_json?.fields);
}

export function proposalDisplay(call: AIToolCall | undefined, message: AIChatMessage): Record<string, unknown> {
  const fromPayload = asRecord(message.payload_json?.display_json);
  const input = proposalInput(call, message);
  return {
    ...fromPayload,
    ...input,
    client_name: textValue(input.client_name) || textValue(input.client_query) || (!call ? textValue(fromPayload.client_name) : ""),
    due_date: input.due_date,
    due_date_label: dateLabel(input.due_date || fromPayload.due_date),
    priority: input.priority,
    billable: input.billable,
    assignee_name: input.assignee_name,
    labels: stringList(input.labels),
    checklist: stringList(input.checklist),
    start_timer_after_create: input.start_timer_after_create,
  };
}

export function proposalMissing(message: AIChatMessage) {
  const missing = message.payload_json?.missing_fields_json;
  return Array.isArray(missing) ? missing.map(asRecord) : [];
}

export function recordList(value: unknown) {
  return Array.isArray(value) ? value.map(asRecord).filter((item) => Object.keys(item).length > 0) : [];
}

export function proposalEntitySnippets(call: AIToolCall | undefined, message: AIChatMessage) {
  const payloadSnippets = recordList(message.payload_json?.entity_snippets_json);
  if (payloadSnippets.length > 0) return payloadSnippets;
  const displaySnippets = recordList(asRecord(message.payload_json?.display_json).entity_snippets);
  if (displaySnippets.length > 0) return displaySnippets;
  const input = proposalInput(call, message);
  const inputSnippets = recordList(input.entity_snippets);
  if (inputSnippets.length > 0) return inputSnippets;
  const clientSnippet = asRecord(input.client_search_snippet);
  return Object.keys(clientSnippet).length > 0 ? [clientSnippet] : [];
}

export function entitySnippetTitle(snippet: Record<string, unknown>) {
  switch (textValue(snippet.entity_type)) {
    case "client":
      return "Поиск клиента";
    case "task":
      return "Поиск задачи";
    case "workflow":
      return "Поиск сценария";
    default:
      return "Поиск сущности";
  }
}

export function entitySnippetStatus(snippet: Record<string, unknown>) {
  const selected = textValue(snippet.selected_name);
  if (selected) return `Найдено: ${selected}`;
  const matches = recordList(snippet.matches);
  if (matches.length > 0) return `Найдено вариантов: ${matches.length}`;
  return "Совпадений пока нет";
}
