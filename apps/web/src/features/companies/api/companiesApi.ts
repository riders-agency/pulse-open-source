import { apiGet, apiPost, asArray, type Company, type Project } from "../../../lib/api";

export async function listCompanies(token: string) {
  return asArray<Company>(await apiGet<Company[]>("/api/v1/companies", token));
}

export function getCompany(companyId: string, token: string) {
  return apiGet<Company>(`/api/v1/companies/${companyId}`, token);
}

export async function createCompany(token: string, input: { name: string; slug: string }) {
  return apiPost<{ company_id: string }>("/api/v1/companies", token, input);
}

export async function listCompanyProjects(companyId: string, token: string) {
  return asArray<Project>(await apiGet<Project[]>(`/api/v1/companies/${companyId}/projects`, token));
}

export function createCompanyProject(companyId: string, token: string, input: { name: string; slug: string }) {
  return apiPost<Project>(`/api/v1/companies/${companyId}/projects`, token, input);
}
