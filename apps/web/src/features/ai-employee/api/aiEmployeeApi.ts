import { apiFetch, apiGet, apiPatch, apiPost } from "../../../lib/api";
import type { AIChatSession, AIEmployeeEnvelope, AIMessageInput, AIPageContext } from "../../../lib/api";
import type { AIStreamPayload } from "../aiEmployeeTypes";

export type CreateAIEmployeeSessionInput = {
  screen_scope: string;
  screen_entity_type: string;
  screen_entity_id: string;
  workflow_id?: string;
  workflow_run_id?: string;
  task_id?: string;
  client_id?: string;
  action_dependency_id?: string;
  model_id?: string;
  page_context_json: AIPageContext;
};

export function getAIEmployee(projectId: string, token: string) {
  return apiGet<AIEmployeeEnvelope>(`/api/v1/projects/${projectId}/ai/employee`, token);
}

export function getAIEmployeeSession(projectId: string, sessionId: string, token: string) {
  return apiGet<AIChatSession>(`/api/v1/projects/${projectId}/ai/employee/sessions/${sessionId}`, token);
}

export function createAIEmployeeSession(projectId: string, token: string, input: CreateAIEmployeeSessionInput) {
  return apiPost<AIChatSession>(`/api/v1/projects/${projectId}/ai/employee/sessions`, token, input);
}

export function cancelAIEmployeeSession(projectId: string, sessionId: string, token: string) {
  return apiPost<AIChatSession>(`/api/v1/projects/${projectId}/ai/employee/sessions/${sessionId}/cancel`, token, {});
}

export function approveAIToolCall(projectId: string, toolCallId: string, token: string) {
  return apiPost<AIChatSession>(`/api/v1/projects/${projectId}/ai/tool-calls/${toolCallId}/approve`, token, {});
}

export function rejectAIToolCall(projectId: string, toolCallId: string, token: string) {
  return apiPost<AIChatSession>(`/api/v1/projects/${projectId}/ai/tool-calls/${toolCallId}/reject`, token, {
    reason: "Отклонено пользователем",
  });
}

export function patchAIToolCall(
  projectId: string,
  toolCallId: string,
  token: string,
  inputJSON: Record<string, unknown>,
  reason: string,
) {
  return apiPatch<AIChatSession>(`/api/v1/projects/${projectId}/ai/tool-calls/${toolCallId}`, token, {
    input_json: inputJSON,
    edit_reason: reason,
  });
}

export function evaluateAIModels(projectId: string, actionDependencyId: string, token: string) {
  return apiPost<{ selected_model_id?: string }>(
    `/api/v1/projects/${projectId}/action-dependencies/${actionDependencyId}/models/evaluate`,
    token,
    {},
  );
}

export async function postMessageStream(
  projectId: string,
  sessionId: string,
  token: string,
  input: AIMessageInput,
  onEvent: (payload: AIStreamPayload) => void,
  signal?: AbortSignal,
): Promise<AIChatSession> {
  const response = await apiFetch(`/api/v1/projects/${projectId}/ai/employee/sessions/${sessionId}/messages/stream`, {
    method: "POST",
    token,
    headers: { Accept: "text/event-stream" },
    json: input,
    signal,
    timeoutMs: 0,
  });
  if (!response.ok) {
    throw new Error(`Не удалось получить ответ ИИ-сотрудника (код ${response.status}). Повторите попытку.`);
  }
  if (!response.body) {
    throw new Error("Сервис ИИ вернул пустой ответ. Повторите попытку.");
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let finalSession: AIChatSession | null = null;
  let streamError = "";

  for (;;) {
    const { value, done } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    const frames = buffer.split("\n\n");
    buffer = frames.pop() ?? "";
    for (const frame of frames) {
      const payload = parseSSEPayload(frame);
      if (!payload) continue;
      if (payload.session) finalSession = payload.session;
      if (payload.type === "ERROR") streamError = payload.message || payload.code || "Не удалось выполнить ход AI-сотрудника.";
      if (payload.type === "DONE" && payload.ok === false && !streamError) streamError = "Не удалось выполнить ход AI-сотрудника.";
      onEvent(payload);
    }
  }

  if (buffer.trim()) {
    const payload = parseSSEPayload(buffer);
    if (payload?.session) finalSession = payload.session;
    if (payload?.type === "ERROR") streamError = payload.message || payload.code || "Не удалось выполнить ход AI-сотрудника.";
    if (payload?.type === "DONE" && payload.ok === false && !streamError) streamError = "Не удалось выполнить ход AI-сотрудника.";
    if (payload) onEvent(payload);
  }

  if (!finalSession) {
    throw new Error(streamError || "Связь с ИИ прервалась до завершения ответа. Повторите попытку.");
  }
  return finalSession;
}

function parseSSEPayload(frame: string): AIStreamPayload | null {
  const data = frame
    .split(/\r?\n/)
    .filter((line) => line.startsWith("data:"))
    .map((line) => line.slice(5).trimStart())
    .join("\n");
  if (!data) return null;
  try {
    return JSON.parse(data) as AIStreamPayload;
  } catch {
    return null;
  }
}
