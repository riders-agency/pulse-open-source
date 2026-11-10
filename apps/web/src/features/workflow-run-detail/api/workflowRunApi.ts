import { ActionRun, apiGet, apiPost, asArray, Run, RunEvent } from "../../../lib/api";

export type WorkflowRunGraph = {
  graph?: {
    nodes?: Array<{ id: string; position?: { x: number; y: number }; data?: any }>;
    edges?: any[];
  };
};

export async function getWorkflowRunDetail(workflowRunId: string, token: string) {
  const [run, actions, events, graph] = await Promise.all([
    apiGet<Run>(`/api/v1/runs/${workflowRunId}`, token),
    apiGet<ActionRun[]>(`/api/v1/runs/${workflowRunId}/actions`, token),
    apiGet<RunEvent[]>(`/api/v1/runs/${workflowRunId}/events`, token),
    apiGet<WorkflowRunGraph>(`/api/v1/runs/${workflowRunId}/graph`, token),
  ]);
  return { run, actions, events, graph };
}

export async function listWorkflowRuns(workflowId: string, token: string) {
  return asArray<Run>(await apiGet<Run[]>(`/api/v1/workflows/${workflowId}/runs`, token));
}

export async function startManualWorkflowRun(workflowId: string, token: string) {
  return apiPost(`/api/v1/workflows/${workflowId}/runs/manual`, token, {
    input_json: { order: { id: "test-1001", total: 12500, customer: { email: "client@example.com" } } },
  });
}

export async function replayWorkflowRun(workflowRunId: string, token: string) {
  return apiPost<{ workflow_run_id: string }>(`/api/v1/runs/${workflowRunId}/replay`, token, {});
}

export async function retryWorkflowRunAction(workflowRunId: string, token: string, actionKey: string) {
  return apiPost<{ workflow_run_id: string }>(
    `/api/v1/runs/${workflowRunId}/retry-action`,
    token,
    { action_key: actionKey },
  );
}
