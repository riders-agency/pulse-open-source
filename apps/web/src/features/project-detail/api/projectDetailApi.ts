import {
  apiDelete,
  apiGet,
  apiGetProjectClientServiceTypes,
  apiPatch,
  apiPost,
  asArray,
  type ClientServiceType,
  type Integration,
  type Project,
  type Run,
  type Workflow,
} from "../../../lib/api";

export type ServiceTypeInput = {
  service_type_key: string;
  display_name: string;
  description: string;
  sort_order: number;
};

export async function getProjectDashboard(projectId: string, token: string) {
  const [project, workflows, integrations, serviceTypes] = await Promise.all([
    apiGet<Project>(`/api/v1/projects/${projectId}`, token),
    apiGet<Workflow[]>(`/api/v1/projects/${projectId}/workflows`, token),
    apiGet<Integration[]>(`/api/v1/projects/${projectId}/action-dependencies`, token),
    apiGetProjectClientServiceTypes(projectId, token),
  ]);
  const workflowRows = asArray<Workflow>(workflows);
  const runs = workflowRows[0]?.workflow_id
    ? asArray<Run>(await apiGet<Run[]>(`/api/v1/workflows/${workflowRows[0].workflow_id}/runs`, token))
    : [];
  return {
    project,
    workflows: workflowRows,
    integrations: asArray<Integration>(integrations),
    serviceTypes: asArray<ClientServiceType>(serviceTypes),
    runs,
  };
}

export async function listProjectServiceTypes(projectId: string, token: string) {
  return asArray<ClientServiceType>(await apiGetProjectClientServiceTypes(projectId, token));
}

export function updateProjectServiceType(projectId: string, typeId: string, token: string, input: ServiceTypeInput) {
  return apiPatch(`/api/v1/projects/${projectId}/client-service-types/${typeId}`, token, input);
}

export function archiveProjectServiceType(projectId: string, typeId: string, token: string) {
  return apiDelete(`/api/v1/projects/${projectId}/client-service-types/${typeId}`, token);
}

export function createProjectServiceType(projectId: string, token: string, input: ServiceTypeInput) {
  return apiPost(`/api/v1/projects/${projectId}/client-service-types`, token, input);
}
