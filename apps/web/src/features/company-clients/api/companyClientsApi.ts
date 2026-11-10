import {
  apiDelete,
  apiGet,
  apiPost,
  apiPostVoid,
  apiPut,
  asArray,
  type Client,
  type CompanyLegalEntity,
  type DaDataIntegration,
  type DaDataPartySuggestion,
} from "../../../lib/api";

export async function listCompanyClients(companyId: string, token: string) {
  return asArray<Client>(await apiGet<Client[]>(`/api/v1/companies/${companyId}/clients`, token));
}

export function getCompanyDaDataIntegration(companyId: string, token: string) {
  return apiGet<DaDataIntegration>(`/api/v1/companies/${companyId}/integrations/dadata`, token);
}

export function saveCompanyDaDataIntegration(companyId: string, token: string, apiKey: string) {
  return apiPut<DaDataIntegration>(`/api/v1/companies/${companyId}/integrations/dadata`, token, {
    api_key: apiKey.trim(),
  });
}

export async function findCompanyDaDataParty(companyId: string, token: string, query: string) {
  const out = await apiPost<{ suggestions?: DaDataPartySuggestion[] }>(
    `/api/v1/companies/${companyId}/clients/dadata/find-party`,
    token,
    { query: query.trim() },
  );
  return out.suggestions ?? [];
}

export function createCompanyClient(companyId: string, token: string, input: Record<string, unknown>) {
  return apiPost<{ client_id: string }>(`/api/v1/companies/${companyId}/clients`, token, input);
}

export function attachClientToProject(clientId: string, projectId: string, token: string) {
  return apiPostVoid(`/api/v1/clients/${clientId}/projects/${projectId}`, token);
}

export async function listCompanyLegalEntities(companyId: string, token: string) {
  return asArray<CompanyLegalEntity>(await apiGet<CompanyLegalEntity[]>(`/api/v1/companies/${companyId}/legal-entities`, token));
}

export function createCompanyLegalEntity(companyId: string, token: string, input: Record<string, unknown>) {
  return apiPost(`/api/v1/companies/${companyId}/legal-entities`, token, input);
}

export function deleteCompanyLegalEntity(companyLegalEntityId: string, token: string) {
  return apiDelete(`/api/v1/company-legal-entities/${companyLegalEntityId}`, token);
}
