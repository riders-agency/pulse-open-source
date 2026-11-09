import { apiGet, asArray, type Client } from "../../../lib/api";

export async function listCompanyClients(companyId: string, token: string, search = "") {
  const params = new URLSearchParams();
  if (search.trim()) params.set("search", search.trim());
  const query = params.toString();
  const rows = await apiGet<Client[]>(`/api/v1/companies/${companyId}/clients${query ? `?${query}` : ""}`, token);
  return asArray<Client>(rows);
}
