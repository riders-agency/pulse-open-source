import type { Edge } from "@xyflow/react";
import { jsonText } from "../../../lib/api";
import type { ActionConfig, ActionKind, ActionNodeType, ScenarioActionJSON, ScenarioConnectionJSON } from "./types";
import { defaultConfig, normalizeEmailNode } from "../registry/actions";
import { normalizeActionNodeLayout } from "./nodeLayout";
import { isActionKind, isRecord, persistableNode, positionFromUnknown } from "./editorDraftBase";
import { edgeFromScenarioConnection, isSameConnection, validateScenarioEdges } from "./scenarioEdges";
import { nodeFromScenarioAction } from "./scenarioNodes";

export function updateActionNodeFromPatch(node: ActionNodeType, patch: Record<string, unknown>): ActionNodeType {
  const nextTypeValue = typeof patch.action_type_key === "string" ? patch.action_type_key : "";
  if (nextTypeValue && !isActionKind(nextTypeValue)) {
    throw new Error(`Неподдерживаемый action_type_key: ${nextTypeValue}`);
  }
  const nextType = nextTypeValue ? (nextTypeValue as ActionKind) : node.data.action_type_key;
  const configPatch = isRecord(patch.config) ? (patch.config as ActionConfig) : undefined;
  const baseConfig = nextType === node.data.action_type_key ? node.data.config : defaultConfig(nextType);
  const ui = isRecord(patch.ui) ? patch.ui : null;
  const position = positionFromUnknown(ui?.position);
  const updated: ActionNodeType = {
    ...node,
    position: position ?? node.position,
    data: {
      ...node.data,
      label: typeof patch.label === "string" ? patch.label : node.data.label,
      action_type_key: nextType,
      config: configPatch ? { ...baseConfig, ...configPatch } : baseConfig,
    },
  };
  return updated.data.action_type_key === "email.send" ? normalizeEmailNode(updated) : updated;
}

export function applyScenarioPatchToDraft({
  parsed,
  currentNodes,
  currentEdges,
  workflowId,
}: {
  parsed: unknown;
  currentNodes: ActionNodeType[];
  currentEdges: Edge[];
  workflowId: string;
}) {
  const root = parsed as { base_workflow_id?: string; change_description?: unknown; operations?: unknown[] };
  if (root.base_workflow_id && workflowId && root.base_workflow_id !== workflowId) {
    throw new Error("Патч относится к другому workflow_id.");
  }
  if (!Array.isArray(root.operations) || root.operations.length === 0) {
    throw new Error("В patch нет operations.");
  }
  let nextNodes = currentNodes.map((node) => persistableNode(node));
  let nextEdges = currentEdges.map((edge) => ({ ...edge }));
  let nextTestInputText: string | undefined;
  let changeDescription = typeof root.change_description === "string" ? root.change_description : undefined;

  root.operations.forEach((operation, index) => {
    if (!isRecord(operation) || typeof operation.op !== "string") {
      throw new Error(`Операция #${index + 1} не содержит op.`);
    }
    const op = operation.op;
    if (op === "add_action" || op === "upsert_action") {
      const action = isRecord(operation.action) ? (operation.action as ScenarioActionJSON) : null;
      const node = action ? nodeFromScenarioAction(action, nextNodes.length) : null;
      if (!node) throw new Error(`Операция ${op} должна содержать корректный action.`);
      const existingIndex = nextNodes.findIndex((candidate) => candidate.id === node.id);
      if (existingIndex >= 0) {
        if (op === "add_action") throw new Error(`Действие ${node.id} уже существует.`);
        const patch: Record<string, unknown> = {
          label: node.data.label,
          action_type_key: node.data.action_type_key,
          config: node.data.config,
        };
        if (action?.ui?.position) {
          patch.ui = { position: node.position };
        }
        nextNodes[existingIndex] = updateActionNodeFromPatch(nextNodes[existingIndex], patch);
      } else {
        nextNodes = [...nextNodes, node];
      }
      return;
    }

    if (op === "update_action") {
      const actionKey = typeof operation.action_key === "string" ? operation.action_key : "";
      const patch = isRecord(operation.set) ? operation.set : null;
      if (!actionKey || !patch) throw new Error("update_action требует action_key и set.");
      const nodeIndex = nextNodes.findIndex((node) => node.id === actionKey);
      if (nodeIndex < 0) throw new Error(`Действие ${actionKey} не найдено.`);
      nextNodes[nodeIndex] = updateActionNodeFromPatch(nextNodes[nodeIndex], patch);
      return;
    }

    if (op === "delete_action") {
      const actionKey = typeof operation.action_key === "string" ? operation.action_key : "";
      if (!actionKey) throw new Error("delete_action требует action_key.");
      if (nextNodes.length <= 1) throw new Error("Нельзя удалить последнее действие сценария.");
      if (!nextNodes.some((node) => node.id === actionKey)) throw new Error(`Действие ${actionKey} не найдено.`);
      const deleteConnections = operation.delete_connections !== false;
      nextNodes = nextNodes.filter((node) => node.id !== actionKey);
      if (deleteConnections) {
        nextEdges = nextEdges.filter((edge) => edge.source !== actionKey && edge.target !== actionKey);
      }
      return;
    }

    if (op === "add_connection") {
      const from = typeof operation.from === "string" ? operation.from : "";
      const to = typeof operation.to === "string" ? operation.to : "";
      const branch = typeof operation.branch === "string" ? operation.branch : "";
      const edge = edgeFromScenarioConnection({ from, to, branch }, nextEdges.length, "patch-connection");
      if (!edge) throw new Error("add_connection требует from и to.");
      if (!nextEdges.some((candidate) => isSameConnection(candidate, edge.source, edge.target, branch))) {
        nextEdges = [...nextEdges, edge];
      }
      return;
    }

    if (op === "delete_connection") {
      const from = typeof operation.from === "string" ? operation.from : "";
      const to = typeof operation.to === "string" ? operation.to : "";
      const branch = typeof operation.branch === "string" ? operation.branch : undefined;
      if (!from || !to) throw new Error("delete_connection требует from и to.");
      nextEdges = nextEdges.filter((edge) => !isSameConnection(edge, from, to, branch));
      return;
    }

    if (op === "replace_connections") {
      const connections = Array.isArray(operation.connections) ? (operation.connections as ScenarioConnectionJSON[]) : [];
      nextEdges = connections
        .map((edge, edgeIndex) => edgeFromScenarioConnection(edge, edgeIndex, "patch-replace-connection"))
        .filter((edge): edge is Edge => !!edge);
      return;
    }

    if (op === "set_test_input") {
      nextTestInputText = jsonText(operation.test_input_json ?? {});
      return;
    }

    if (op === "set_change_description") {
      changeDescription = typeof operation.description === "string" ? operation.description : "";
      return;
    }

    throw new Error(`Неподдерживаемая операция patch: ${op}`);
  });

  const layoutNodes = normalizeActionNodeLayout(nextNodes, nextEdges);
  validateScenarioEdges(layoutNodes, nextEdges);
  return { nodes: layoutNodes, edges: nextEdges, testInputText: nextTestInputText, changeDescription };
}
