import { apiGet, apiPost, asArray, type Company, type Project, type Run, type Workflow } from "../../../lib/api";

export async function listDashboardCompanies(token: string) {
  return asArray<Company>(await apiGet<Company[]>("/api/v1/companies", token));
}

export async function listDashboardProjects(companyId: string, token: string) {
  return asArray<Project>(await apiGet<Project[]>(`/api/v1/companies/${companyId}/projects`, token));
}

export async function getDashboardWorkflows(projectId: string, token: string) {
  const workflows = asArray<Workflow>(await apiGet<Workflow[]>(`/api/v1/projects/${projectId}/workflows`, token));
  const runs = workflows[0]?.workflow_id
    ? asArray<Run>(await apiGet<Run[]>(`/api/v1/workflows/${workflows[0].workflow_id}/runs`, token))
    : [];
  return { workflows, runs };
}

export function loginDashboard(input: { email: string; password: string }) {
  return apiPost<{ access_token?: string }>("/api/v1/auth/login", "", input);
}
