import type { Edge } from "@xyflow/react";

import type { Workflow } from "../../../lib/api";
import { normalizeEmailConfig, parseJSONText } from "../registry/actions";
import { mergeCanvasNodesWithDefinition, persistableNode } from "./editorDraft";
import { normalizeActionNodeLayout } from "./nodeLayout";
import type { ActionNodeType } from "./types";

export function buildWorkflowDraftSave({
  edges,
  nodes,
  testInputText,
  workflow,
}: {
  edges: Edge[];
  nodes: ActionNodeType[];
  testInputText: string;
  workflow: Workflow | null;
}) {
  const testInput = parseJSONText(testInputText, {});
  const nodesToSave = normalizeActionNodeLayout(mergeCanvasNodesWithDefinition(nodes, workflow), edges);
  const definition = {
    nodes: nodesToSave.map((node) => ({
      action_key: node.id,
      action_type_key: node.data.action_type_key,
      config: node.data.action_type_key === "email.send" ? normalizeEmailConfig(node.data.config) : node.data.config,
      label: node.data.label,
    })),
    edges: edges.map((edge) => ({ from: edge.source, label: edge.sourceHandle || edge.label || "", to: edge.target })),
  };
  const graphNodes = nodesToSave.map(persistableNode);

  return {
    definition,
    graphNodes,
    nodesToSave,
    payload: {
      draft_definition_json: definition,
      draft_dsl_yaml: `name: ${workflow?.name ?? "Сценарий"}\n# visual graph stored in draft_definition_json`,
      draft_graph_json: { nodes: graphNodes, edges },
      test_input_json: testInput,
    },
    testInput,
  };
}
