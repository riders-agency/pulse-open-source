import {
  apiDelete,
  ApiError,
  apiGet,
  apiPatch,
  apiPost,
  apiPostVoid,
  asArray,
  type Client,
  type ClientContact,
  type ClientProject,
  type ClientResponsible,
  type ClientService,
  type CompanyLegalEntity,
  type ManualPaymentRecordResponse,
  type PaymentAction,
  type Project,
} from "../../../lib/api";
import type { ContactForm, EditForm, ResponsibleForm } from "../model/clientDetailModel";

export async function getClientDetail(clientId: string, activeProjectId: string, token: string) {
  const [client, contacts, responsibles, links, services, payments] = await Promise.all([
    apiGet<Client>(`/api/v1/clients/${clientId}`, token),
    apiGet<ClientContact[]>(`/api/v1/clients/${clientId}/contacts`, token),
    apiGet<ClientResponsible[]>(`/api/v1/clients/${clientId}/responsibles`, token),
    apiGet<ClientProject[]>(`/api/v1/clients/${clientId}/projects`, token),
    apiGet<ClientService[]>(`/api/v1/clients/${clientId}/services${activeProjectId ? `?project_id=${activeProjectId}` : ""}`, token),
    listClientPayments(clientId, activeProjectId, token),
  ]);
  return {
    client,
    contacts: asArray<ClientContact>(contacts),
    responsibles: asArray<ClientResponsible>(responsibles),
    links: asArray<ClientProject>(links),
    services: asArray<ClientService>(services),
    payments: asArray<PaymentAction>(payments),
  };
}

export async function getCompanyClientContext(companyId: string, token: string) {
  const [projects, legalEntities] = await Promise.all([
    apiGet<Project[]>(`/api/v1/companies/${companyId}/projects`, token),
    apiGet<CompanyLegalEntity[]>(`/api/v1/companies/${companyId}/legal-entities`, token),
  ]);
  return {
    projects: asArray<Project>(projects),
    legalEntities: asArray<CompanyLegalEntity>(legalEntities),
  };
}

export function updateClient(clientId: string, token: string, input: EditForm) {
  return apiPatch(`/api/v1/clients/${clientId}`, token, input);
}

export function deleteClient(clientId: string, token: string) {
  return apiDelete(`/api/v1/clients/${clientId}`, token);
}

export function addClientResponsible(clientId: string, token: string, input: ResponsibleForm, isPrimaryFallback: boolean) {
  return apiPost(`/api/v1/clients/${clientId}/responsibles`, token, {
    display_name: input.display_name.trim(),
    role: input.role,
    company_legal_entity_id: input.company_legal_entity_id || undefined,
    is_primary: input.is_primary || isPrimaryFallback,
  });
}

export function deleteClientResponsible(clientId: string, responsibleId: string, token: string) {
  return apiDelete(`/api/v1/clients/${clientId}/responsibles/${responsibleId}`, token);
}

export function addClientContact(clientId: string, token: string, input: ContactForm, isPrimary: boolean) {
  return apiPost(`/api/v1/clients/${clientId}/contacts`, token, { ...input, is_primary: isPrimary });
}

export function linkClientProject(clientId: string, projectId: string, token: string) {
  return apiPost(`/api/v1/clients/${clientId}/projects/${projectId}`, token);
}

export function deleteClientService(clientServiceId: string, token: string) {
  return apiDelete(`/api/v1/client-services/${clientServiceId}`, token);
}

export async function recordClientManualPayment(
  clientId: string,
  activeProjectId: string,
  token: string,
  input: {
    client_service_id: string;
    amount: number;
    currency: string;
    period_start: string;
    period_end?: string;
    paid_at: string;
    payment_method: string;
    next_due_date?: string;
    comment: string;
    idempotency_key: string;
  },
) {
  try {
    return await apiPost<ManualPaymentRecordResponse>(`/api/v1/clients/${clientId}/payments/manual`, token, {
      project_id: activeProjectId || undefined,
      ...input,
    });
  } catch (err) {
    if (!(err instanceof ApiError) || err.status !== 404 || !activeProjectId) {
      throw err;
    }
    const out = await apiPost<{ payment_action_id: string }>(`/api/v1/projects/${activeProjectId}/payment-actions/manual`, token, {
      client_id: clientId,
      due_date: input.period_start,
    });
    await apiPostVoid(`/api/v1/payment-actions/${out.payment_action_id}/confirm`, token, {
      amount: input.amount,
      paid_at: input.paid_at,
      source: "manual",
      payment_method: input.payment_method,
      next_due_date: input.next_due_date || undefined,
      comment: input.comment,
    });
  }
}

async function listClientPayments(clientId: string, activeProjectId: string, token: string) {
  try {
    return asArray<PaymentAction>(
      await apiGet<PaymentAction[]>(
        `/api/v1/clients/${clientId}/payment-actions${activeProjectId ? `?project_id=${activeProjectId}` : ""}`,
        token,
      ),
    );
  } catch (err) {
    if (!(err instanceof ApiError) || err.status !== 404 || !activeProjectId) {
      throw err;
    }
    const projectRows = asArray<PaymentAction>(
      await apiGet<PaymentAction[]>(`/api/v1/projects/${activeProjectId}/payment-actions`, token),
    );
    return projectRows.filter((item) => item.client_id === clientId);
  }
}
