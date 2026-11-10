import {
  apiBase,
  apiGet,
  apiFetch,
  apiPatch,
  apiPost,
  apiPut,
  apiUpload,
  type WorkBoard,
  type WorkPresets,
  type WorkProcessRole,
  type WorkTask,
  type WorkTaskAttachment,
  type WorkTaskDetail,
  type WorkTimer,
  type WorkWorkflow,
} from "../../../lib/api";
import type {
  CreateTaskWorkflowInput,
  CreateWorkTaskInput,
  ProcessRoleInput,
  WorkBoardEnvelope,
  WorkBoardsEnvelope,
  WorkEmployeesEnvelope,
  WorkOverviewEnvelope,
  WorkReportEnvelope,
  WorkTaskEnvelope,
  WorkTaskQuery,
  WorkTasksEnvelope,
  WorkTimeEnvelope,
  WorkTimersEnvelope,
  WorkWorkflowsEnvelope,
} from "./workApiContracts";
export { listCompanyClients } from "./workClientsApi";

export type {
  CreateTaskWorkflowInput,
  CreateWorkTaskInput,
  ProcessRoleInput,
  WorkBoardEnvelope,
  WorkBoardsEnvelope,
  WorkEmployeesEnvelope,
  WorkOverviewEnvelope,
  WorkReportEnvelope,
  WorkTaskEnvelope,
  WorkTaskQuery,
  WorkTasksEnvelope,
  WorkTimeEnvelope,
  WorkTimersEnvelope,
  WorkWorkflowsEnvelope,
} from "./workApiContracts";

export function getWorkOverview(projectId: string, token: string) {
  return apiGet<WorkOverviewEnvelope>(`/api/v1/work/overview?project_id=${projectId}`, token);
}

export function listWorkBoards(projectId: string, token: string) {
  return apiGet<WorkBoardsEnvelope>(`/api/v1/work/boards?project_id=${projectId}`, token);
}

export function getWorkBoard(boardId: string, token: string) {
  return apiGet<WorkBoardEnvelope>(`/api/v1/work/boards/${boardId}`, token);
}

export function getWorkPresets(token: string) {
  return apiGet<WorkPresets>("/api/v1/work/presets", token);
}

export function createWorkBoard(token: string, input: { project_id: string; template_key: string; name?: string }) {
  return apiPost<WorkBoard>("/api/v1/work/boards", token, input);
}

export function listWorkTasks(projectId: string, token: string, query: WorkTaskQuery = {}) {
  const params = new URLSearchParams({ project_id: projectId, limit: String(query.limit || 120) });
  if (query.archived) params.set("archived", "true");
  if (query.search?.trim()) params.set("search", query.search.trim());
  if (query.assignee) params.set("assignee", query.assignee);
  if (query.due) params.set("due", query.due);
  if (query.taskWorkflowId) params.set("task_workflow_id", query.taskWorkflowId);
  return apiGet<WorkTasksEnvelope>(`/api/v1/work/tasks?${params.toString()}`, token);
}

export function createWorkTask(token: string, input: CreateWorkTaskInput) {
  return apiPost<WorkTask>("/api/v1/work/tasks", token, input);
}

export function moveWorkTask(token: string, taskId: string, statusId: string, rank: number, expectedRevision?: number) {
  return apiPost(`/api/v1/work/tasks/${taskId}/move`, token, { task_workflow_status_id: statusId, rank, expected_revision: expectedRevision });
}

export function getWorkTask(taskId: string, token: string) {
  return apiGet<WorkTaskEnvelope>(`/api/v1/work/tasks/${taskId}`, token);
}

export function patchWorkTask(taskId: string, token: string, body: Record<string, unknown>) {
  return apiPatch<WorkTaskDetail>(`/api/v1/work/tasks/${taskId}`, token, body);
}

export function addWorkTaskComment(taskId: string, token: string, body: string, visibility: string) {
  return apiPost<{ task: WorkTaskDetail }>(`/api/v1/work/tasks/${taskId}/comments`, token, { body, visibility });
}

export function addWorkTaskChecklistItem(taskId: string, token: string, title: string) {
  return apiPost<{ task: WorkTaskDetail }>(`/api/v1/work/tasks/${taskId}/checklist`, token, { title });
}

export function setWorkTaskChecklistDone(taskId: string, itemId: string, token: string, isDone: boolean) {
  return apiPatch<{ task: WorkTaskDetail }>(`/api/v1/work/tasks/${taskId}/checklist/${itemId}`, token, { is_done: isDone });
}

export function setWorkTaskCustomField(taskId: string, fieldId: string, token: string, value: unknown) {
  return apiPut<{ task: WorkTaskDetail }>(`/api/v1/work/tasks/${taskId}/custom-fields/${fieldId}`, token, { value });
}

export function uploadWorkTaskAttachment(taskId: string, token: string, form: FormData) {
  return apiUpload<WorkTaskAttachment>(`/api/v1/work/tasks/${taskId}/attachments`, token, form);
}

export async function fetchWorkTaskAttachment(attachment: WorkTaskAttachment, token: string) {
  const path = attachment.download_url.startsWith("http") ? attachment.download_url : `${apiBase}${attachment.download_url}`;
  const response = await apiFetch(path, { method: "GET", token });
  if (!response.ok) throw new Error("Не удалось скачать вложение");
  return response.blob();
}

export function getActiveWorkTimers(token: string) {
  return apiGet<WorkTimersEnvelope>("/api/v1/work/timer/active", token);
}

export function startWorkTaskTimer(taskId: string, token: string) {
  return apiPost<WorkTimer>(`/api/v1/work/tasks/${taskId}/timer/start`, token);
}

export function stopWorkTimer(token: string, taskId = "") {
  return apiPost("/api/v1/work/timer/stop", token, { task_id: taskId });
}

export function listWorkTime(projectId: string, token: string) {
  return apiGet<WorkTimeEnvelope>(`/api/v1/work/time?project_id=${projectId}`, token);
}

export function createWorkTimeEntry(token: string, input: { task_id: string; duration_minutes: number; description: string; entry_date: string; billable: boolean }) {
  return apiPost("/api/v1/work/time", token, input);
}

export function listTaskWorkflows(projectId: string, token: string) {
  return apiGet<WorkWorkflowsEnvelope>(`/api/v1/work/task-workflows?project_id=${projectId}`, token);
}

export function createTaskWorkflow(token: string, input: CreateTaskWorkflowInput) {
  return apiPost<WorkWorkflow>("/api/v1/work/task-workflows", token, input);
}

export function updateProcessRole(workflowId: string, roleId: string, token: string, input: ProcessRoleInput) {
  return apiPatch<WorkProcessRole>(`/api/v1/work/task-workflows/${workflowId}/roles/${roleId}`, token, input);
}

export function createProcessRole(workflowId: string, token: string, input: ProcessRoleInput) {
  return apiPost<WorkProcessRole>(`/api/v1/work/task-workflows/${workflowId}/roles`, token, input);
}

export function resetProcessRoles(workflowId: string, token: string) {
  return apiPost(`/api/v1/work/task-workflows/${workflowId}/roles/reset`, token, {});
}

export function listWorkEmployees(projectId: string, token: string) {
  return apiGet<WorkEmployeesEnvelope>(`/api/v1/work/employees?project_id=${projectId}`, token);
}

export function inviteWorkProjectMember(projectId: string, token: string, input: { email: string; display_name: string; task_workflow_id: string; task_process_role_id: string; role: string; client_id?: string }) {
  return apiPost<{ temporary_password?: string; role?: WorkProcessRole }>(`/api/v1/work/projects/${projectId}/members`, token, input);
}

export function getWorkReport(projectId: string, token: string) {
  return apiGet<WorkReportEnvelope>(`/api/v1/work/reports?project_id=${projectId}`, token);
}

export { archiveWorkTask, restoreWorkTask } from "./workTaskLifecycleApi";
