import {
  apiBase,
  apiGet,
  apiPost,
  apiPostVoid,
  apiUpload,
  asArray,
  type PaymentAction,
  type PaymentDocument,
  type Project,
  type ProjectPaymentScheduleRow,
} from "../../../lib/api";

export async function getProjectPayments(projectId: string, token: string) {
  const [project, actions, schedule] = await Promise.all([
    apiGet<Project>(`/api/v1/projects/${projectId}`, token),
    apiGet<PaymentAction[]>(`/api/v1/projects/${projectId}/payment-actions`, token),
    apiGet<ProjectPaymentScheduleRow[]>(`/api/v1/projects/${projectId}/payment-schedule`, token),
  ]);
  return {
    project,
    actions: asArray<PaymentAction>(actions),
    schedule: asArray<ProjectPaymentScheduleRow>(schedule),
  };
}

export function createProjectPaymentAction(projectId: string, token: string, input: { client_id: string; due_date: string; invoice_prepared?: boolean }) {
  return apiPost<{ payment_action_id: string }>(`/api/v1/projects/${projectId}/payment-actions/manual`, token, input);
}

export function uploadPaymentActionDocument(paymentActionId: string, token: string, form: FormData) {
  return apiUpload<PaymentDocument>(`/api/v1/payment-actions/${paymentActionId}/documents`, token, form);
}

export function getPaymentAction(paymentActionId: string, token: string) {
  return apiGet<PaymentAction>(`/api/v1/payment-actions/${paymentActionId}`, token);
}

export function confirmPaymentAction(paymentActionId: string, token: string, input: { amount?: number; paid_at: string; source: string; payment_method: string; next_due_date?: string }) {
  return apiPostVoid(`/api/v1/payment-actions/${paymentActionId}/confirm`, token, input);
}

export function paymentDocumentHref(path: string) {
  if (/^https?:\/\//.test(path)) return path;
  return `${apiBase}${path}`;
}
