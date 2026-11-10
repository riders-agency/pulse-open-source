import type { Edge } from "@xyflow/react";

import {
  apiGet,
  apiPost,
  apiPut,
  asArray,
  type Integration,
  type IntegrationDocSearchResult,
  type Project,
  type ProviderCatalogEntry,
  type Workflow,
  type WorkflowVersion,
} from "../../../lib/api";
import type { ActionNodeType } from "../model/types";

export type WorkflowDraftPayload = {
  draft_definition_json: unknown;
  draft_graph_json: { nodes: ActionNodeType[]; edges: Edge[] };
  draft_dsl_yaml: string;
  test_input_json: unknown;
};

export type WorkflowTestRunPayload = {
  input_json: unknown;
  mode: "full" | "through_action" | "selected_only";
  action_key: string;
  steps_overrides?: unknown;
};

export function fetchWorkflow(token: string, workflowId: string) {
  return apiGet<Workflow>(`/api/v1/workflows/${workflowId}`, token);
}

export async function fetchWorkflowVersions(token: string, workflowId: string) {
  return asArray<WorkflowVersion>(await apiGet<WorkflowVersion[]>(`/api/v1/workflows/${workflowId}/versions`, token));
}

export function fetchWorkflowProject(token: string, projectId: string) {
  return apiGet<Project>(`/api/v1/projects/${projectId}`, token);
}

export async function fetchWorkflowIntegrations(token: string, projectId: string) {
  return asArray<Integration>(await apiGet<Integration[]>(`/api/v1/projects/${projectId}/action-dependencies`, token));
}

export async function fetchProviderCatalog(token: string) {
  return asArray<ProviderCatalogEntry>(await apiGet<ProviderCatalogEntry[]>(`/api/v1/providers/catalog`, token));
}

export async function searchIntegrationDocs(token: string, query: string, projectId?: string, limit = 8) {
  const params = new URLSearchParams();
  params.set("q", query.trim());
  params.set("limit", String(limit));
  const path = projectId
    ? `/api/v1/projects/${projectId}/integration-docs/search?${params.toString()}`
    : `/api/v1/docs/integrations/search?${params.toString()}`;
  return asArray<IntegrationDocSearchResult>(await apiGet<IntegrationDocSearchResult[]>(path, token));
}

export function saveWorkflowDraft(token: string, workflowId: string, payload: WorkflowDraftPayload) {
  return apiPut(`/api/v1/workflows/${workflowId}/draft`, token, payload);
}

export function validateWorkflow(token: string, workflowId: string) {
  return apiPost(`/api/v1/workflows/${workflowId}/validate`, token, {});
}

export function publishWorkflow(token: string, workflowId: string, description: string) {
  return apiPost<{ workflow_version_id: string; version_no: number }>(`/api/v1/workflows/${workflowId}/publish`, token, {
    description,
  });
}

export function restoreWorkflowVersion(token: string, workflowId: string, versionId: string) {
  return apiPost(`/api/v1/workflows/${workflowId}/versions/${versionId}/restore`, token, {});
}

export function createWorkflowTestRun(token: string, workflowId: string, payload: WorkflowTestRunPayload) {
  return apiPost<{ workflow_run_id: string }>(`/api/v1/workflows/${workflowId}/test-runs`, token, payload);
}

export function rotateTelegramRelayWebhookSecret(token: string, projectId: string, actionDependencyId: string) {
  return apiPost<{ webhook_secret?: string }>(
    `/api/v1/projects/${projectId}/action-dependencies/${actionDependencyId}/telegram-relay-webhook-secret/rotate`,
    token,
    {},
  );
}

export type TelegramRelayPollingPayload = {
  enabled: boolean;
  timeout_seconds: number;
  limit: number;
  delete_webhook: boolean;
  drop_pending_updates: boolean;
};

export function saveTelegramRelayPolling(token: string, projectId: string, actionDependencyId: string, payload: TelegramRelayPollingPayload) {
  return apiPost<Record<string, unknown>>(
    `/api/v1/projects/${projectId}/action-dependencies/${actionDependencyId}/telegram-relay-polling`,
    token,
    payload,
  );
}
