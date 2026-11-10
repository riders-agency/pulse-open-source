import { apiGet, apiPatch, apiPost, apiPostVoid, asArray, jsonText, type Integration } from "../../../lib/api";

export type ProjectIntegrationDetail = Integration & {
  last_checked_at?: string | null;
};

export async function listProjectIntegrations(projectId: string, token: string) {
  return asArray<Integration>(
    await apiGet<Integration[]>(`/api/v1/projects/${projectId}/action-dependencies?include_archived=true`, token),
  );
}

export async function getProjectIntegration(projectId: string, integrationId: string, token: string) {
  const row = await apiGet<ProjectIntegrationDetail>(`/api/v1/projects/${projectId}/action-dependencies/${integrationId}`, token);
  return {
    row,
    metadataText: jsonText(row.metadata_json),
  };
}

export function createProjectIntegration(projectId: string, token: string, input: Record<string, unknown>) {
  return apiPost(`/api/v1/projects/${projectId}/action-dependencies`, token, input);
}

export function updateProjectIntegration(projectId: string, integrationId: string, token: string, input: Record<string, unknown>) {
  return apiPatch(`/api/v1/projects/${projectId}/action-dependencies/${integrationId}`, token, input);
}

export function rotateProjectIntegrationSecret(projectId: string, integrationId: string, token: string, input: Record<string, unknown>) {
  return apiPostVoid(`/api/v1/projects/${projectId}/action-dependencies/${integrationId}/rotate-secret`, token, input);
}

export function restoreProjectIntegration(projectId: string, integrationId: string, token: string) {
  return apiPatch(`/api/v1/projects/${projectId}/action-dependencies/${integrationId}`, token, { status: "active" });
}
