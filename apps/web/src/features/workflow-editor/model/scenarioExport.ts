import type { Edge } from "@xyflow/react";
import { type IntegrationDocSearchResult, type Project, type Workflow } from "../../../lib/api";
import { actionCatalogForAI, compactIntegrationDoc } from "../ai/promptContext";
import { aiWorkflowDocsURL, normalizeEmailConfig, parseJSONText } from "../registry/actions";
import type { ActionNodeType } from "./types";
import { actionTypeMeta } from "./actionPresentation";

export function parseJSONForExport(text: string) {
  try {
    return parseJSONText(text, {});
  } catch {
    return { _invalid_json_text: text };
  }
}

export function connectionsFrom(edges: Edge[]) {
  return edges.map((edge) => ({
    from: edge.source,
    to: edge.target,
    branch: edge.sourceHandle || edge.label || "",
    description: edge.sourceHandle || edge.label ? `Route ${edge.sourceHandle || edge.label}` : "Default route",
  }));
}

export function actionsFrom(nodes: ActionNodeType[]) {
  return nodes.map((node) => {
    const meta = actionTypeMeta(node.data.action_type_key);
    return {
      action_key: node.id,
      label: node.data.label,
      action_type_key: node.data.action_type_key,
      category: meta.group,
      type_label: meta.label,
      config: node.data.action_type_key === "email.send" ? normalizeEmailConfig(node.data.config) : node.data.config,
      ui: {
        position: node.position,
      },
    };
  });
}

export function scenarioExportObject({
  workflow,
  project,
  nodes,
  edges,
  testInputText,
  templateVariables,
  relevantDocs,
}: {
  workflow: Workflow | null;
  project: Project | null;
  nodes: ActionNodeType[];
  edges: Edge[];
  testInputText: string;
  templateVariables: string[];
  relevantDocs: IntegrationDocSearchResult[];
}) {
  const actions = actionsFrom(nodes);
  const connections = connectionsFrom(edges);
  const catalog = actionCatalogForAI();
  return {
    format: "riders-pulse-workflow-context-v1",
    schema_version: 1,
    purpose: "Compact workflow context for AI editing. Do not return this whole object; return only riders-pulse-workflow-patch-v1.",
    documentation_url: aiWorkflowDocsURL,
    instructions_for_ai: {
      output: "Return only a riders-pulse-workflow-patch-v1 JSON object with operations, without markdown or prose.",
      primary_editing_targets: ["operations[].action", "operations[].set", "operations[].connections", "operations[].test_input_json"],
      keep_stable: ["workflow.workflow_id", "project.project_id", "actions[].action_key when the action keeps the same role"],
      do_not_invent: ["integration UUIDs", "secrets", "unsupported action_type_key values in actions"],
      unsupported_future_actions: "Use planned_extension_points only in notes/architecture, not in importable actions until runtime support exists.",
      patch_contract: {
        format: "riders-pulse-workflow-patch-v1",
        schema_version: 1,
        base_workflow_id: workflow?.workflow_id ?? "",
        change_description: "Кратко что поменялось. Можно оставить пустым.",
        operations: [
          { op: "update_action", action_key: "existing_action_key", set: { label: "Новое название", config: { text: "Новый шаблон" } } },
          { op: "add_action", action: { action_key: "new_action_key", label: "Новое действие", action_type_key: "telegram.send_message", config: {} } },
          { op: "add_connection", from: "source_action_key", to: "target_action_key", branch: "" },
          { op: "delete_connection", from: "source_action_key", to: "target_action_key", branch: "" },
          { op: "delete_action", action_key: "old_action_key", delete_connections: true },
          { op: "set_test_input", test_input_json: {} },
        ],
      },
    },
    workflow: {
      workflow_id: workflow?.workflow_id ?? "",
      name: workflow?.name ?? "Сценарий",
      slug: workflow?.slug ?? "",
      status: workflow?.status ?? "",
    },
    project: {
      project_id: project?.project_id ?? workflow?.project_id ?? "",
      name: project?.name ?? "",
      hook_key: project?.hook_key ?? "",
    },
    actions,
    connections,
    relevant_integration_docs: relevantDocs.map(compactIntegrationDoc),
    test_input_json: parseJSONForExport(testInputText),
    variable_contract: {
      webhook_payload: "{{ <webhook_endpoint_key>.<field.path> }}",
      schedule_payload: "{{ <schedule_key>.scheduled_at }}",
      action_outputs: "{{ steps.<action_key>.json.<field> }} or {{ steps.<action_key>.body }}",
      system: ["{{ system.link }}", "{{ system.id }}"],
      visible_examples: templateVariables,
    },
    action_catalog: {
      supported_now: catalog.supported_now.map((item) => ({
        action_type_key: item.action_type_key,
        config_fields: item.config_fields,
        variables: "variables" in item ? item.variables : undefined,
        outputs: "outputs" in item ? item.outputs : undefined,
      })),
      planned_extension_points: catalog.planned_extension_points.map((item) => ({
        action_type_key: item.action_type_key,
        status: item.status,
        expected_config_fields: item.expected_config_fields,
      })),
    },
    architecture_notes: {
      graph_model: "Workflow is a directed graph. actions are nodes. connections are directed edges. condition.branch edges may use branch true/false.",
      data_model: "Each action receives the run input and prior steps output. HTTP JSON responses are exposed under steps.<action_key>.json.",
      triggers: "webhook.trigger, telegram.relay.trigger, schedule.trigger, manual.trigger, and clients.manual_payment.trigger start runs. clients.payment_due.trigger can also be first, but its main role is to load all due client payments and build the digest for notifications. Saved payments enqueue module triggers so a payment workflow runs again immediately.",
      future_internal_api: "The same patch format is intended for future direct API import so manual clipboard copy can be replaced by an internal scenario editing system.",
    },
    ai_editing_rules: [
      "Return a patch, not a full scenario.",
      "Fill change_description with a short human-readable summary when possible; empty string is allowed.",
      "Keep action_key stable when editing existing actions.",
      "Use action_type_key values from the exported nodes or the documented catalog.",
      "Do not invent secrets. Reuse existing dependency ids or leave them empty with a note.",
      "Use webhook/schedule/telegram_relay namespaces for incoming data; do not use hardcoded input.order aliases.",
      "For HTTP response data use steps.<http_action_key>.json.* after the http.request action.",
      "Every connection target and source must reference an existing actions[].action_key.",
    ],
  };
}
