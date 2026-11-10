import {
  apiGet,
  apiPostVoid,
  apiPut,
  asArray,
  type Company,
  type Project,
  type Run,
  type WorkAccess,
  type WorkspaceContext,
} from "../../../lib/api";

export function getWorkspaceContext(token: string) {
  return apiGet<WorkspaceContext>("/api/v1/auth/workspace-context", token);
}

export function saveWorkspaceContext(token: string, input: { active_company_id: string | null; active_project_id: string | null }) {
  return apiPut<WorkspaceContext>("/api/v1/auth/workspace-context", token, input);
}

export async function listShellCompanies(token: string) {
  return asArray<Company>(await apiGet<Company[]>("/api/v1/companies", token));
}

export function switchCompanyContext(companyId: string, token: string) {
  return apiPostVoid(`/api/v1/companies/${companyId}/switch`, token);
}

export async function listShellCompanyProjects(companyId: string, token: string) {
  return asArray<Project>(await apiGet<Project[]>(`/api/v1/companies/${companyId}/projects`, token));
}

export function getShellWorkAccess(projectId: string, token: string) {
  return apiGet<WorkAccess>(`/api/v1/work/access?project_id=${projectId}`, token);
}

export function getShellProject(projectId: string, token: string) {
  return apiGet<Project>(`/api/v1/projects/${projectId}`, token);
}

export function getShellRun(runId: string, token: string) {
  return apiGet<Run>(`/api/v1/runs/${runId}`, token);
}
