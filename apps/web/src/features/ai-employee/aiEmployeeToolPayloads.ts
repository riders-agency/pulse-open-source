function asToolRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" && !Array.isArray(value) ? value as Record<string, unknown> : {};
}

function toolText(value: unknown) {
  return typeof value === "string" ? value : "";
}

export function toolResultTitle(payload: unknown) {
  const data = asToolRecord(payload);
  if (data.kind !== "tool_result") return "Результат";
  switch (toolText(data.tool_key)) {
    case "task.create":
      return "Задача создана";
    case "task.move":
      return "Задача стартована";
    case "client.create":
      return "Клиент создан";
    default:
      return "Результат";
  }
}

export function toolLabel(toolKey?: string) {
  switch (toolKey) {
    case "task.create":
      return "Создание задачи";
    case "task.move":
      return "Запуск задачи";
    case "client.create":
      return "Создание клиента";
    case "workflow.patch":
    case "workflow.patch.apply":
      return "Изменение сценария";
    case "task_process.patch.apply":
      return "Изменение процесса задач";
    case "payment.lookup_due":
      return "Поиск оплаты";
    default:
      return toolKey || "Действие";
  }
}
