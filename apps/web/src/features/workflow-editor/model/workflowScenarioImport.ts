import type { Edge } from "@xyflow/react";
import type { Dispatch, SetStateAction } from "react";

import { jsonText, type Workflow } from "../../../lib/api";
import { parseScenarioPatchText } from "../registry/actions";
import { applyScenarioPatchToDraft, edgesFromScenarioJSON, nodesFromScenarioJSON, validateScenarioEdges } from "./editorDraft";
import { normalizeActionNodeLayout } from "./nodeLayout";
import type { ActionNodeType } from "./types";

export function applyWorkflowScenarioImport({
  edges,
  nodes,
  scenarioImportText,
  selectAction,
  setDescriptionOpen,
  setEdges,
  setError,
  setNodes,
  setNotice,
  setScenarioImportText,
  setTestInputText,
  setVersionDescription,
  workflow,
}: {
  edges: Edge[];
  nodes: ActionNodeType[];
  scenarioImportText: string;
  selectAction: (nodeId: string) => void;
  setDescriptionOpen: (value: boolean) => void;
  setEdges: Dispatch<SetStateAction<Edge[]>>;
  setError: (value: string) => void;
  setNodes: Dispatch<SetStateAction<ActionNodeType[]>>;
  setNotice: (value: string) => void;
  setScenarioImportText: (value: string) => void;
  setTestInputText: (value: string) => void;
  setVersionDescription: (value: string) => void;
  workflow: Workflow | null;
}) {
  try {
    const parsed = parseScenarioPatchText(scenarioImportText) as { format?: string; operations?: unknown[]; test_input_json?: unknown };
    if (parsed.format === "riders-pulse-workflow-context-v1") throw new Error("Это контекст для ИИ. В импорт нужно вставить patch riders-pulse-workflow-patch-v1.");
    if (parsed.format === "riders-pulse-workflow-patch-v1" || Array.isArray(parsed.operations)) {
      const patched = applyScenarioPatchToDraft({ parsed, currentNodes: nodes, currentEdges: edges, workflowId: workflow?.workflow_id ?? "" });
      setNodes(patched.nodes);
      setEdges(patched.edges);
      if (patched.testInputText !== undefined) setTestInputText(patched.testInputText);
      if (patched.changeDescription !== undefined) {
        setVersionDescription(patched.changeDescription);
        setDescriptionOpen(true);
      }
      selectAction(patched.nodes[0]?.id ?? "");
      setNotice("Патч применен в черновик. Проверьте и нажмите Сохранить.");
      setError("");
      return;
    }
    if (parsed.format && parsed.format !== "riders-pulse-workflow-v1") throw new Error("Неверный format. Нужен riders-pulse-workflow-patch-v1.");
    const importedNodes = nodesFromScenarioJSON(parsed);
    if (!importedNodes.length) throw new Error("В JSON нет operations или действий actions/draft_definition_json.nodes/draft_graph_json.nodes.");
    const importedEdges = edgesFromScenarioJSON(parsed);
    validateScenarioEdges(importedNodes, importedEdges);
    setNodes(normalizeActionNodeLayout(importedNodes, importedEdges));
    setEdges(importedEdges);
    if (parsed.test_input_json !== undefined) setTestInputText(jsonText(parsed.test_input_json));
    selectAction(importedNodes[0].id);
    setNotice("Сценарий применен в черновик. Проверьте и нажмите Сохранить.");
    setError("");
    setScenarioImportText(scenarioImportText);
  } catch (err) {
    setError(err instanceof Error ? err.message : "Не удалось применить JSON patch");
  }
}
