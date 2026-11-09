import { apiGet, apiPost, asArray, type WorkAccess, type Workflow } from "../../../lib/api";

export async function listProjectWorkflowsIfAllowed(projectId: string, token: string) {
  const access = await apiGet<WorkAccess>(`/api/v1/work/access?project_id=${projectId}`, token);
  if (!access.can_manage_processes) {
    return { denied: true, workflows: [] as Workflow[] };
  }
  return {
    denied: false,
    workflows: asArray<Workflow>(await apiGet<Workflow[]>(`/api/v1/projects/${projectId}/workflows`, token)),
  };
}

export function createProjectWorkflow(projectId: string, token: string, input: { name: string; slug: string }) {
  return apiPost<Workflow>(`/api/v1/projects/${projectId}/workflows`, token, input);
}
