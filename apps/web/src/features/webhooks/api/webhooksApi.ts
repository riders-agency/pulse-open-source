import { apiGet, apiPost } from "../../../lib/api";

export async function getWebhookEndpointDetail(webhookEndpointId: string, token: string) {
  const [detail, snippets, events] = await Promise.all([
    apiGet<any>(`/api/v1/webhook-endpoints/${webhookEndpointId}`, token),
    apiGet<Record<string, string>>(`/api/v1/webhook-endpoints/${webhookEndpointId}/snippets`, token),
    apiGet<any[]>(`/api/v1/webhook-endpoints/${webhookEndpointId}/events`, token),
  ]);
  return { detail, snippets, events };
}

export function getWebhookEvent(webhookEventId: string, token: string) {
  return apiGet<any>(`/api/v1/webhook-events/${webhookEventId}`, token);
}

export function replayWebhookEvent(webhookEventId: string, token: string) {
  return apiPost<{ workflow_run_id: string }>(`/api/v1/webhook-events/${webhookEventId}/replay`, token, {});
}
