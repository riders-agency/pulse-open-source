"use client";

import { useCallback, useEffect, useState, type Dispatch, type SetStateAction } from "react";
import type { Edge } from "@xyflow/react";

import type { Integration, Project, ProviderCatalogEntry, Workflow, WorkflowVersion } from "../../../lib/api";
import {
  fetchProviderCatalog,
  fetchWorkflow,
  fetchWorkflowIntegrations,
  fetchWorkflowProject,
  fetchWorkflowVersions,
} from "../api/workflowEditorApi";
import {
  edgesFromScenarioJSON,
  filterEdgesToNodes,
  graphNodesFromWorkflow,
} from "./editorDraft";
import { normalizeActionNodeLayout } from "./nodeLayout";
import type { ActionNodeType } from "./types";
import { prettyInput, triggerSampleFrom } from "../registry/actions";

export function useWorkflowEditorData({
  selectAction,
  setEdges,
  setError,
  setGraphLoaded,
  setNodes,
  setTestInputText,
  token,
  workflowId,
}: {
  selectAction: (nodeId: string) => void;
  setEdges: Dispatch<SetStateAction<Edge[]>>;
  setError: (value: string) => void;
  setGraphLoaded: (value: boolean) => void;
  setNodes: Dispatch<SetStateAction<ActionNodeType[]>>;
  setTestInputText: (value: string) => void;
  token: string;
  workflowId: string;
}) {
  const [workflow, setWorkflow] = useState<Workflow | null>(null);
  const [project, setProject] = useState<Project | null>(null);
  const [integrations, setIntegrations] = useState<Integration[]>([]);
  const [providerCatalog, setProviderCatalog] = useState<ProviderCatalogEntry[]>([]);
  const [versions, setVersions] = useState<WorkflowVersion[]>([]);

  const applyWorkflowDraft = useCallback(
    (wf: Workflow) => {
      const graph = wf.draft_graph_json as { nodes?: ActionNodeType[]; edges?: Edge[] } | undefined;
      const draft = { draft_definition_json: wf.draft_definition_json, draft_graph_json: wf.draft_graph_json };
      const loadedNodes = graphNodesFromWorkflow(wf);
      const definitionEdges = edgesFromScenarioJSON(draft);
      const graphEdges = Array.isArray(graph?.edges) && graph.edges.length > 0 ? graph.edges : definitionEdges;
      const loadedEdges = filterEdgesToNodes(loadedNodes, graphEdges.length > 0 ? graphEdges : definitionEdges);
      const effectiveEdges = loadedEdges.length > 0 ? loadedEdges : definitionEdges;
      const layoutNodes = normalizeActionNodeLayout(loadedNodes, effectiveEdges);
      setNodes(layoutNodes);
      setEdges(effectiveEdges);
      setTestInputText(prettyInput(wf.test_input_json, triggerSampleFrom(layoutNodes, wf.project_id)));
      selectAction(layoutNodes[0]?.id ?? "");
      setGraphLoaded(true);
    },
    [selectAction, setEdges, setGraphLoaded, setNodes, setTestInputText],
  );

  const loadVersions = useCallback(async () => {
    if (!token) return;
    setVersions(await fetchWorkflowVersions(token, workflowId));
  }, [token, workflowId]);

  const loadWorkflow = useCallback(async () => {
    if (!token) return;
    setGraphLoaded(false);
    const wf = await fetchWorkflow(token, workflowId);
    setWorkflow(wf);
    if (wf.project_id) {
      const [pr, deps] = await Promise.all([
        fetchWorkflowProject(token, wf.project_id),
        fetchWorkflowIntegrations(token, wf.project_id),
      ]);
      setProject(pr);
      setIntegrations(deps);
    }
    applyWorkflowDraft(wf);
  }, [applyWorkflowDraft, setGraphLoaded, token, workflowId]);

  useEffect(() => {
    if (!token) return;
    fetchProviderCatalog(token)
      .then(setProviderCatalog)
      .catch(() => setProviderCatalog([]));
    loadWorkflow()
      .then(loadVersions)
      .catch((err) => {
        setError(err.message);
        setGraphLoaded(true);
      });
  }, [loadVersions, loadWorkflow, setError, setGraphLoaded, token]);

  return {
    integrations,
    loadVersions,
    loadWorkflow,
    project,
    providerCatalog,
    setWorkflow,
    versions,
    workflow,
  };
}
