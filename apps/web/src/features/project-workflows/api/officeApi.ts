import { apiGet, apiPost, apiPatch } from "../../../lib/api";
export type OfficeArtifact = { workflow_run_id: string; page_json: { title?: string; body?: string }; invoice_json: { number?: string; amount_minor?: number; customer?: string }; created_at: string };
export type OfficeState = { configured: boolean; workflow_id?: string; enabled: boolean; interval_seconds: number; retained_runs: number; last_poll_at?: string; next_poll_at?: string; poll_count: number; pending_mail: number; waiting: number; artifacts: OfficeArtifact[] };
const path = (project: string) => `/api/v1/projects/${project}/office-demo`;
export const getOffice = (project: string, token: string) => apiGet<OfficeState>(path(project), token);
export const ensureOffice = (project: string, token: string) => apiPost<OfficeState>(path(project), token);
export const deliverOfficeMail = (project: string, token: string, deliveryKey: string) => apiPost<OfficeState>(`${path(project)}/messages`, token, { delivery_key: deliveryKey });
export const setOfficeEnabled = (project: string, token: string, enabled: boolean) => apiPatch<OfficeState>(path(project), token, { enabled });
export const decideOfficeRun = (project: string, token: string, run: string, approved: boolean) => apiPost(`${path(project)}/decisions/${run}`, token, { approved });
