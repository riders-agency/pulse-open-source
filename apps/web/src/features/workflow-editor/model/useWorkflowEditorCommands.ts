"use client";

import { useState, type Dispatch, type SetStateAction } from "react";
import { useRouter } from "next/navigation";
import type { Edge } from "@xyflow/react";

import { type Workflow, type WorkflowVersion } from "../../../lib/api";
import {
  createWorkflowTestRun,
  publishWorkflow,
  restoreWorkflowVersion,
  saveWorkflowDraft,
  validateWorkflow,
} from "../api/workflowEditorApi";
import type { ActionNodeType, WorkflowTestMode } from "./types";
import { parseJSONText } from "../registry/actions";
import { buildWorkflowDraftSave } from "./workflowDraftSave";
import { applyWorkflowScenarioImport } from "./workflowScenarioImport";

export function useWorkflowEditorCommands({
  edges,
  loadVersions,
  loadWorkflow,
  nodes,
  scenarioImportText,
  selectAction,
  selectedNode,
  setEdges,
  setError,
  setNodes,
  setNotice,
  setScenarioImportText,
  setTestInputText,
  setWorkflow,
  stepsOverrideText,
  testInputText,
  token,
  workflow,
  workflowId,
}: {
  edges: Edge[];
  loadVersions: () => Promise<void>;
  loadWorkflow: () => Promise<void>;
  nodes: ActionNodeType[];
  scenarioImportText: string;
  selectAction: (nodeId: string) => void;
  selectedNode: ActionNodeType | null;
  setEdges: Dispatch<SetStateAction<Edge[]>>;
  setError: (value: string) => void;
  setNodes: Dispatch<SetStateAction<ActionNodeType[]>>;
  setNotice: (value: string) => void;
  setScenarioImportText: (value: string) => void;
  setTestInputText: (value: string) => void;
  setWorkflow: Dispatch<SetStateAction<Workflow | null>>;
  stepsOverrideText: string;
  testInputText: string;
  token: string;
  workflow: Workflow | null;
  workflowId: string;
}) {
  const router = useRouter();
  const [versionDescription, setVersionDescription] = useState("");
  const [descriptionOpen, setDescriptionOpen] = useState(false);
  const [versionsBusy, setVersionsBusy] = useState(false);
  const [testBusy, setTestBusy] = useState<WorkflowTestMode | "">("");

  const applyScenarioImport = () => {
    applyWorkflowScenarioImport({
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
    });
  };

  const save = async () => {
    setError("");
    setNotice("");
    try {
      const draft = buildWorkflowDraftSave({ edges, nodes, testInputText, workflow });
      await saveWorkflowDraft(token, workflowId, draft.payload);
      setWorkflow((current) => current ? { ...current, draft_definition_json: draft.definition, draft_graph_json: { nodes: draft.graphNodes, edges }, test_input_json: draft.testInput } : current);
      setNodes(draft.graphNodes);
      setNotice("Сценарий сохранен");
      return true;
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось сохранить");
      return false;
    }
  };

  const validate = async () => {
    setError("");
    try {
      await validateWorkflow(token, workflowId);
      setNotice("Проверка пройдена");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Сценарий не прошел проверку");
    }
  };

  const publish = async () => {
    setError("");
    try {
      if (!(await save())) return;
      const out = await publishWorkflow(token, workflowId, versionDescription);
      await loadVersions();
      setVersionDescription("");
      setDescriptionOpen(false);
      setNotice(`Опубликовано v${out.version_no}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось опубликовать");
    }
  };

  const restoreVersion = async (version: WorkflowVersion) => {
    const ok = window.confirm(`Откатить сценарий к версии v${version.version_no}? Текущий draft будет заменен этой версией.`);
    if (!ok) return;
    setVersionsBusy(true);
    setError("");
    try {
      await restoreWorkflowVersion(token, workflowId, version.workflow_version_id);
      await loadWorkflow();
      await loadVersions();
      setNotice(`Сценарий откатан к v${version.version_no}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось откатить версию");
    } finally {
      setVersionsBusy(false);
    }
  };

  const runTest = async (mode: WorkflowTestMode) => {
    setError("");
    setNotice("");
    setTestBusy(mode);
    try {
      const input = parseJSONText(testInputText, {});
      const steps = mode === "selected_only" ? parseJSONText(stepsOverrideText, {}) : undefined;
      if (!(await save())) return;
      const out = await createWorkflowTestRun(token, workflowId, {
        input_json: input,
        mode,
        action_key: mode === "full" ? "" : selectedNode?.id ?? "",
        steps_overrides: steps,
      });
      router.push(`/runs/${out.workflow_run_id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось запустить проверку");
    } finally {
      setTestBusy("");
    }
  };

  return {
    applyScenarioImport,
    descriptionOpen,
    publish,
    restoreVersion,
    runTest,
    save,
    setDescriptionOpen,
    setVersionDescription,
    testBusy,
    validate,
    versionDescription,
    versionsBusy,
  };
}
