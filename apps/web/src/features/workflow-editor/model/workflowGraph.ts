import type { Workflow } from "../../../lib/api";
import type { ActionKind, ActionNodeType } from "./types";
import type { DefinitionNodeJSON, RawGraphNodeJSON } from "./editorDraftTypes";
import { defaultConfig, normalizeEmailNode } from "../registry/actions";
import { actionLabel } from "./actionPresentation";
import { normalizeActionNodeLayout } from "./nodeLayout";
import { defaultNodePosition, isActionKind, isRecord, normalizeGraphNode, positionByActionKey, positionFromUnknown } from "./editorDraftBase";
import { nodesFromScenarioJSON } from "./scenarioNodes";

export function mergeCanvasNodesWithDefinition(
  canvasNodes: ActionNodeType[],
  workflow: Workflow | null | undefined,
): ActionNodeType[] {
  if (!workflow?.draft_definition_json) return canvasNodes;
  const definition = workflow.draft_definition_json as { nodes?: DefinitionNodeJSON[] };
  const graph = workflow.draft_graph_json as { nodes?: unknown[] } | undefined;
  const graphNodes = Array.isArray(graph?.nodes) ? graph.nodes : [];
  const byId = new Map(canvasNodes.map((node) => [node.id, node]));
  const merged = [...canvasNodes];
  for (const defNode of definition.nodes ?? []) {
    const id = defNode.action_key?.trim();
    if (!id || byId.has(id) || !defNode.action_type_key || !isActionKind(defNode.action_type_key)) continue;
    const actionType = defNode.action_type_key as ActionKind;
    const node: ActionNodeType = {
      id,
      type: "action",
      position: positionByActionKey(graphNodes, id) ?? defaultNodePosition(merged.length),
      data: {
        label: defNode.label || actionLabel(actionType),
        action_type_key: actionType,
        config: defNode.config ?? defaultConfig(actionType),
      },
    };
    merged.push(node.data.action_type_key === "email.send" ? normalizeEmailNode(node) : node);
    byId.set(id, merged[merged.length - 1]!);
  }
  return normalizeActionNodeLayout(merged);
}

export function graphNodesFromWorkflow(workflow: Workflow): ActionNodeType[] {
  const graph = workflow.draft_graph_json as { nodes?: unknown[] } | undefined;
  const definition = workflow.draft_definition_json as { nodes?: DefinitionNodeJSON[] } | undefined;
  const definitionNodes = (definition?.nodes ?? []).filter((n) => n.action_type_key !== "clients.payment_cycle.collect");
  const graphNodes = Array.isArray(graph?.nodes) ? graph.nodes : [];

  if (definitionNodes.length > 0) {
    const loaded = definitionNodes
      .filter((node) => node.action_key && node.action_type_key)
      .map((node, index) => {
        const id = node.action_key!;
        const graphNode = graphNodes.find((raw) => isRecord(raw) && String((raw as RawGraphNodeJSON).id ?? "") === id) as
          | RawGraphNodeJSON
          | undefined;
        const pos =
          positionByActionKey(graphNodes, id) ??
          (graphNode?.position ? positionFromUnknown(graphNode.position) : null) ??
          defaultNodePosition(index);
        const actionType = node.action_type_key as ActionKind;
        const config =
          node.config ??
          graphNode?.data?.config ??
          graphNode?.config ??
          (isActionKind(node.action_type_key) ? defaultConfig(actionType) : {});
        return {
          id,
          type: "action" as const,
          position: pos ?? defaultNodePosition(index),
          data: {
            label: node.label || (isActionKind(node.action_type_key) ? actionLabel(actionType) : node.action_type_key),
            action_type_key: actionType,
            config,
          },
        };
      })
      .map((node) => (node.data.action_type_key === "email.send" ? normalizeEmailNode(node) : node));
    return normalizeActionNodeLayout(loaded);
  }

  if (graphNodes.length > 0) {
    const normalized = graphNodes
      .map((node, index) => normalizeGraphNode(node, index, definitionNodes))
      .filter((node): node is ActionNodeType => !!node && node.data.action_type_key !== "clients.payment_cycle.collect")
      .map((node) => (node.data.action_type_key === "email.send" ? normalizeEmailNode(node) : node));
    if (normalized.length > 0) return normalizeActionNodeLayout(normalized);
  }

  return nodesFromScenarioJSON({ draft_definition_json: workflow.draft_definition_json });
}
