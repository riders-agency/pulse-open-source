import type { AIChatMessage, AIChatSession, AIConfiguredDependency, AIEmployeeEnvelope } from "../../lib/api";
import type { WidgetChatMessage } from "./aiEmployeeTypes";
import { AI_EMPLOYEE_COPY_LABELS } from "./aiEmployeeConstants";
import { messageTimestampLabel } from "./aiEmployeeFormatters";
import {
  entityLinkFromPayload,
  hasRenderableMessageBody,
  hasRenderableSnippet,
  taskSearchItems,
  textValue,
  toolLabel,
} from "./aiEmployeePayloads";

export function primaryActionLabel(toolKey?: string) {
  switch (toolKey) {
    case "task.create":
      return "Создать задачу";
    case "task.move":
      return "Стартовать задачу";
    case "client.create":
      return "Создать клиента";
    case "workflow.patch":
    case "workflow.patch.apply":
    case "task_process.patch.apply":
      return "Применить изменения";
    case "workflow.run_test":
      return "Запустить тест";
    default:
      return "Выполнить";
  }
}

export function statusLabel(status?: string) {
  switch (status) {
    case "pending_approval":
      return "ждет подтверждения";
    case "succeeded":
      return "выполнено";
    case "failed":
      return "ошибка";
    case "rejected":
      return "отклонено";
    case "draft":
      return "черновик";
    case "running":
      return "выполняется";
    case "cancelled":
      return "отменено";
    default:
      return status ? "ожидает обработки" : "черновик";
  }
}

export function modelFromEmployee(employee: AIEmployeeEnvelope | null): {
  dependency?: AIConfiguredDependency;
  model: string;
} {
  const dependencies = employee?.configured_dependencies ?? [];
  const selectedID = employee?.employee?.action_dependency_id;
  const selected = dependencies.find((item) => item.action_dependency_id === selectedID) ?? dependencies.find((item) => item.dependency_type_key === "ai.apinet") ?? dependencies.find((item) => item.dependency_type_key?.startsWith("ai."));
  const model =
    employee?.employee?.model_id ||
    String(selected?.metadata_json?.default_chat_model_id || selected?.metadata_json?.default_model_id || "");
  return { dependency: selected, model };
}

export function copyableTranscript(messages: WidgetChatMessage[]) {
  return messages
    .filter((message) => hasRenderableMessageBody(message) || hasRenderableSnippet(message))
    .map((message) => {
      const role =
        message.role === "user"
          ? AI_EMPLOYEE_COPY_LABELS.user
          : message.role === "assistant"
            ? AI_EMPLOYEE_COPY_LABELS.assistant
            : message.role;
      const time = messageTimestampLabel(message.created_at);
      const body = String(message.body || "").trim();
      const visibleBody = hasRenderableMessageBody(message) ? body : describeSnippetForTranscript(message);
      return `${role}${time ? ` [${time}]` : ""}: ${visibleBody}`;
    })
    .filter((line) => line.trim().length > 0)
    .join("\n\n");
}

function describeSnippetForTranscript(message: AIChatMessage) {
  if (message.payload_json?.kind === "proposal") {
    const toolKey = textValue(message.payload_json.tool_key) || "tool";
    return `[${toolLabel(toolKey)}]`;
  }
  if (entityLinkFromPayload(message.payload_json)) return "[Результат tool]";
  if (taskSearchItems(message.payload_json).length > 0) return "[Найденные задачи]";
  return "[Системное событие]";
}

export function mergeOptimisticMessages(session: AIChatSession | null, optimisticMessages: WidgetChatMessage[]) {
  return [...(session?.messages ?? []), ...optimisticMessages];
}

export function proposalMessageIndexes(messages: AIChatMessage[] = []) {
  const indexes: Record<string, number> = {};
  messages.forEach((message, index) => {
    if (message.payload_json?.kind !== "proposal") return;
    const callId = typeof message.payload_json.ai_tool_call_id === "string" ? message.payload_json.ai_tool_call_id : "";
    if (callId) indexes[callId] = index;
  });
  return indexes;
}
