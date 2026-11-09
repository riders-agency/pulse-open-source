export function boardGroupLabel(value?: string) {
  return {
    status: "статус",
    assignee: "ответственный",
    priority: "приоритет",
    client: "клиент",
    project: "проект",
    epic: "эпик",
    deadline: "срок",
    none: "нет",
  }[value || "none"] || value || "нет";
}
