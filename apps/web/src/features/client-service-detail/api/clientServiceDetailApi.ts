import {
  apiDelete,
  apiGet,
  apiGetProjectClientServiceTypes,
  apiPatch,
  apiPost,
  asArray,
  type Client,
  type ClientService,
  type ClientServiceType,
  type Project,
} from "../../../lib/api";

export async function getClientServiceCreateContext(clientId: string, token: string) {
  const client = await apiGet<Client>(`/api/v1/clients/${clientId}`, token);
  const projects = await apiGet<Project[]>(`/api/v1/companies/${client.company_id}/projects`, token);
  return {
    client,
    projects: asArray<Project>(projects),
  };
}

export async function getClientServiceDetail(clientServiceId: string, token: string) {
  const service = await apiGet<ClientService>(`/api/v1/client-services/${clientServiceId}`, token);
  const [client, projects, serviceTypes] = await Promise.all([
    apiGet<Client>(`/api/v1/clients/${service.client_id}`, token),
    apiGet<Project[]>(`/api/v1/companies/${service.company_id}/projects`, token),
    apiGetProjectClientServiceTypes(service.project_id, token),
  ]);
  return {
    service,
    client,
    projects: asArray<Project>(projects),
    serviceTypes: asArray<ClientServiceType>(serviceTypes),
  };
}

export async function listClientServiceTypes(projectId: string, token: string) {
  return asArray<ClientServiceType>(await apiGetProjectClientServiceTypes(projectId, token));
}

export function updateClientService(clientServiceId: string, token: string, input: Record<string, unknown>) {
  return apiPatch(`/api/v1/client-services/${clientServiceId}`, token, input);
}

export function createClientService(clientId: string, token: string, input: Record<string, unknown>) {
  return apiPost(`/api/v1/clients/${clientId}/services`, token, input);
}

export function deleteClientService(clientServiceId: string, token: string) {
  return apiDelete(`/api/v1/client-services/${clientServiceId}`, token);
}
