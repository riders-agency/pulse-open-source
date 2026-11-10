import { apiDelete, apiPost } from "../../../lib/api";

export function archiveWorkTask(taskId: string, token: string) {
  return apiDelete(`/api/v1/work/tasks/${taskId}`, token);
}
export function restoreWorkTask(taskId: string, token: string) {
  return apiPost<{ archived: boolean }>(`/api/v1/work/tasks/${taskId}/restore`, token, {});
}
