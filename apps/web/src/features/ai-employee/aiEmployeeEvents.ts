import type { AIToolCall } from "../../lib/api";
import { asRecord, textValue } from "./aiEmployeePayloads";

export function notifyEntityChanged(call?: AIToolCall) {
  if (!call || call.status !== "succeeded") return;
  const output = asRecord(call.output_json);
  const entity = call.tool_key.startsWith("task.") || call.tool_key.startsWith("task_process.")
    ? "task"
    : call.tool_key.startsWith("client.")
      ? "client"
      : call.tool_key.startsWith("workflow.")
        ? "workflow"
        : "entity";
  window.dispatchEvent(new CustomEvent("pulse:ai-entity-changed", {
    detail: {
      entity,
      id: textValue(output.task_id) || textValue(output.client_id) || textValue(output.workflow_id) || textValue(output.task_workflow_id),
      output,
      toolKey: call.tool_key,
      url: textValue(output.url),
    },
  }));
}
