export type TaskTab = "card" | "comments" | "time" | "activity" | "access" | "attachments";

export const taskTypes = ["task", "bug", "feature", "request", "idea", "support", "content", "design", "legal", "finance", "hr", "document"];
export const taskPriorities = ["none", "low", "medium", "high", "urgent"];

export type WorkTaskDraft = {
  title: string;
  description: string;
  type: string;
  priority: string;
  status_id: string;
  client_id: string;
  due_date: string;
  estimate_minutes: string;
};

export function emptyWorkTaskDraft(statusId = ""): WorkTaskDraft {
  return { title: "", description: "", type: "task", priority: "medium", status_id: statusId, client_id: "", due_date: "", estimate_minutes: "60" };
}
