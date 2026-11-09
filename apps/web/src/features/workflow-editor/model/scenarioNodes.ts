import type { ActionConfig, ActionKind, ActionNodeType, ScenarioActionJSON } from "./types";
import { defaultConfig, normalizeEmailNode } from "../registry/actions";
import { actionLabel } from "./actionPresentation";
import { normalizeActionNodeLayout } from "./nodeLayout";
import { defaultNodePosition, isActionKind, normalizeGraphNode } from "./editorDraftBase";

export function nodeFromScenarioAction(action: ScenarioActionJSON, index: number): ActionNodeType | null {
  if (!action.action_key || !action.action_type_key || !isActionKind(action.action_type_key)) return null;
  const baseConfig = defaultConfig(action.action_type_key);
  const node: ActionNodeType = {
    id: action.action_key,
    type: "action",
    position: action.ui?.position ?? defaultNodePosition(index),
    data: {
      label: action.label || actionLabel(action.action_type_key),
      action_type_key: action.action_type_key,
      config: action.config ? { ...baseConfig, ...action.config } : baseConfig,
    },
  };
  return node.data.action_type_key === "email.send" ? normalizeEmailNode(node) : node;
}

export function nodesFromScenarioJSON(parsed: unknown): ActionNodeType[] {
  const root = parsed as {
    actions?: ScenarioActionJSON[];
    draft_definition_json?: { nodes?: Array<{ action_key?: string; label?: string; action_type_key?: string; config?: ActionConfig }> };
    draft_graph_json?: { nodes?: ActionNodeType[] };
  };
  if (Array.isArray(root.actions) && root.actions.length > 0) {
    return normalizeActionNodeLayout(root.actions
      .map((action, index) => nodeFromScenarioAction(action, index))
      .filter((node): node is ActionNodeType => !!node));
  }
  if (Array.isArray(root.draft_graph_json?.nodes) && root.draft_graph_json.nodes.length > 0) {
    const definitionNodes = root.draft_definition_json?.nodes ?? [];
    return normalizeActionNodeLayout(root.draft_graph_json.nodes
      .map((node, index) => normalizeGraphNode(node, index, definitionNodes))
      .filter((node): node is ActionNodeType => !!node)
      .map((node) => (node.data.action_type_key === "email.send" ? normalizeEmailNode(node) : node)));
  }
  const defNodes = root.draft_definition_json?.nodes ?? [];
  return normalizeActionNodeLayout(defNodes
    .filter((node) => node.action_key && node.action_type_key && isActionKind(node.action_type_key))
    .map((node, index) => ({
      id: node.action_key!,
      type: "action" as const,
      position: defaultNodePosition(index),
      data: {
        label: node.label || actionLabel(node.action_type_key as ActionKind),
        action_type_key: node.action_type_key as ActionKind,
        config: node.config ?? defaultConfig(node.action_type_key as ActionKind),
      },
    })));
}
