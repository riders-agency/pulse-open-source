import type { AIRuntimeThreadState } from "./ai-runtime-state";

export type { AIRuntimeThreadState } from "./ai-runtime-state";

export type AIModelProvider = {
  provider_key: string;
  dependency_type_key: string;
  display_name: string;
  status: string;
  default_base_url: string;
  capabilities: string[];
  tool_loop: string;
  docs_url?: string;
};

export type AIConfiguredDependency = {
  action_dependency_id: string;
  name: string;
  status: string;
  dependency_type_key: string;
  provider_key: string;
  metadata_json?: Record<string, unknown>;
};

export type AICapabilityState = "disabled" | "deterministic_preview" | "read_only" | "proposal_only" | "ready";

export type AICapability = {
  state: AICapabilityState;
  runtime: string;
  production: boolean;
  allows_read: boolean;
  allows_proposals: boolean;
  allows_writes: boolean;
  reasons: string[];
  dependencies: Record<string, { status: string; detail?: string }>;
};

export type AIEmployeeEnvelope = {
  employee: {
    project_ai_employee_id: string;
    employee_key: string;
    display_name: string;
    status: string;
    action_dependency_id?: string | null;
    model_id?: string;
  };
  project: {
    project_id: string;
    company_id: string;
    name: string;
    slug: string;
    company_name?: string;
  };
  model_providers: AIModelProvider[];
  configured_dependencies: AIConfiguredDependency[];
  setup_required: boolean;
  capability?: AICapability;
};

export type AIChatMessage = {
  ai_chat_message_id: string;
  role: "system" | "user" | "assistant" | "tool" | "approval" | "event";
  body: string;
  payload_json?: Record<string, unknown>;
  token_usage_json?: Record<string, unknown>;
  created_at: string;
};

export type AIToolCall = {
  ai_tool_call_id: string;
  tool_key: string;
  risk_level: "read" | "propose" | "write" | "critical" | string;
  status: "proposed" | "pending_approval" | "approved" | "rejected" | "running" | "succeeded" | "failed" | "canceled" | string;
  input_json?: Record<string, unknown>;
  output_json?: Record<string, unknown>;
  error_text?: string;
  requires_approval: boolean;
  created_at: string;
  finished_at?: string;
};

export type AIProposalDisplay = {
  title?: string;
  description?: string;
  client_name?: string;
  client_status?: string;
  due_date?: string;
  due_date_label?: string;
  status_id?: string;
  priority?: string;
  billable?: boolean;
  assignee_name?: string;
  labels?: string[];
  checklist?: string[];
  notification_note?: string;
  start_timer_after_create?: boolean;
  [key: string]: unknown;
};

export type AIPageContext = {
  route: string;
  pathname: string;
  pageKind:
    | "dashboard"
    | "project"
    | "tasks"
    | "task_board"
    | "task"
    | "clients"
    | "client"
    | "payments"
    | "workflow_editor"
    | "workflow_run"
    | "integrations"
    | "site_chat"
    | "system_status"
    | "settings"
    | "unknown";
  companyId?: string;
  projectId?: string;
  workflowId?: string;
  workflowRunId?: string;
  taskId?: string;
  clientId?: string;
  selectedNodeId?: string;
  selectedActionKey?: string;
  visiblePanel?: string;
  visibleFilters?: Record<string, unknown>;
  visibleEntitySummary?: Record<string, unknown>;
  timestamp: string;
};

export type AIChatSession = {
  ai_chat_session_id: string;
  company_id?: string;
  project_id: string;
  status: string;
  agent_runtime?: string;
  langgraph_thread_id?: string;
  screen_scope: string;
  screen_entity_type: string;
  screen_entity_id: string;
  model_id?: string;
  session_revision?: number;
  page_context_json?: AIPageContext | Record<string, unknown>;
  last_page_context_json?: AIPageContext | Record<string, unknown>;
  ui_state_json?: Record<string, unknown>;
  pending_context_json?: Record<string, unknown>;
  working_memory_json?: Record<string, unknown>;
  runtime_state_json?: AIRuntimeThreadState;
  active_tool_call_id?: string;
  active_patch_proposal_id?: string;
  messages: AIChatMessage[];
  tool_calls?: AIToolCall[];
};

export type AIMessageInput = {
  body: string;
  page_context_json?: AIPageContext | Record<string, unknown>;
  client_context_json?: Record<string, unknown>;
  ui_state_json?: Record<string, unknown>;
  active_tool_call_id?: string;
  active_patch_proposal_id?: string;
  reply_to_message_id?: string;
  client_command?: string;
  local_date?: string;
  timezone?: string;
};
