export type Workflow = {
  workflow_id: string;
  project_id?: string;
  name: string;
  slug: string;
  status?: string;
  run_count?: number;
  draft_definition_json?: unknown;
  draft_graph_json?: unknown;
  draft_dsl_yaml?: string;
  test_input_json?: unknown;
  current_workflow_version_id?: string;
};

export type WorkflowVersion = {
  workflow_version_id: string;
  workflow_id?: string;
  version_no: number;
  status: string;
  description?: string;
  published_at?: string;
  created_at?: string;
  is_current?: boolean;
  definition_json?: unknown;
  graph_json?: unknown;
  dsl_yaml?: string | null;
};

export type Integration = {
  action_dependency_id: string;
  name: string;
  status: string;
  dependency_type_key: string;
  provider_key?: string;
  dependency_type_display_name?: string;
  metadata_json?: unknown;
};

export type Run = {
  workflow_run_id: string;
  project_id?: string;
  workflow_id?: string;
  company_id?: string;
  status: string;
  source_type?: string;
  started_at?: string;
  finished_at?: string;
  input_json?: unknown;
  output_json?: unknown;
  error?: string | null;
};

export type ActionRun = {
  action_run_id?: string;
  action_key?: string;
  action_type_key?: string;
  /** Resolved from linked action_dependencies (same as integration_profile). */
  dependency_type_key?: string;
  integration_profile?: string;
  status?: string;
  attempt_count?: number;
  max_attempts?: number;
  input_json?: unknown;
  output_json?: unknown;
  provider_response_json?: unknown;
  diagnostics?: unknown;
  error?: string | null;
  retry_class?: string | null;
  retry_policy?: unknown;
  next_retry_at?: string | null;
  retry_exhausted_at?: string | null;
  dead_lettered_at?: string | null;
  dead_letter_reason?: string | null;
  dead_letter_owner?: string | null;
  replayable?: boolean;
  started_at?: string;
  finished_at?: string;
  attempts?: ActionAttempt[];
};

export type ActionAttempt = {
  action_attempt_id?: string;
  attempt_no?: number;
  status?: string;
  input_json?: unknown;
  output_json?: unknown;
  provider_response_json?: unknown;
  diagnostics?: unknown;
  error?: string | null;
  error_type?: string | null;
  started_at?: string;
  finished_at?: string;
  duration_ms?: number | null;
};

export type RunEvent = {
  run_event_id?: string;
  action_key?: string;
  event_type?: string;
  message?: string;
  created_at?: string;
  payload_json?: unknown;
};
