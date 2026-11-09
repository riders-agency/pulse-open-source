"use client";

import { useMemo, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Background, Controls, MiniMap, ReactFlow } from "@xyflow/react";

import { usePulseStore } from "../../store/pulse-store";
import { useWorkflowEditorCommands } from "./model/useWorkflowEditorCommands";
import { useWorkflowEditorData } from "./model/useWorkflowEditorData";
import { useWorkflowEditorDocs } from "./model/useWorkflowEditorDocs";
import { useWorkflowEditorGraph } from "./model/useWorkflowEditorGraph";
import { normalizeActionNodeLayout } from "./model/nodeLayout";
import { templateVariablesFromText, webhookSlugFromNodes } from "./registry/actions";
import { ActionNode } from "./ui/action-node";
import { EditorPaletteSidebar } from "./ui/editor-palette-sidebar";
import { EditorSettingsSheet } from "./ui/editor-settings-sheet";
import { EditorToolbar } from "./ui/editor-toolbar";

const nodeTypes = { action: ActionNode };

export function WorkflowEditorScreen({ workflowId }: { workflowId: string }) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = usePulseStore((s) => s.token);
  const [testInputText, setTestInputText] = useState("{}");
  const [stepsOverrideText, setStepsOverrideText] = useState(`{\n  "previous_action": {\n    "message_id": "manual-1"\n  }\n}`);
  const [notice, setNotice] = useState("");
  const [error, setError] = useState("");

  const graph = useWorkflowEditorGraph({
    requestedActionId: searchParams.get("action") ?? "",
    setError,
    setNotice,
  });
  const webhookSlug = useMemo(() => webhookSlugFromNodes(graph.nodes), [graph.nodes]);
  const templateVariables = useMemo(() => templateVariablesFromText(webhookSlug, testInputText), [testInputText, webhookSlug]);
  const data = useWorkflowEditorData({
    selectAction: graph.selectAction,
    setEdges: graph.setEdges,
    setError,
    setGraphLoaded: graph.setGraphLoaded,
    setNodes: graph.setNodes,
    setTestInputText,
    token,
    workflowId,
  });
  const docs = useWorkflowEditorDocs({
    edges: graph.edges,
    nodes: graph.nodes,
    project: data.project,
    setError,
    setNotice,
    templateVariables,
    testInputText,
    token,
    workflow: data.workflow,
  });
  const commands = useWorkflowEditorCommands({
    edges: graph.edges,
    loadVersions: data.loadVersions,
    loadWorkflow: data.loadWorkflow,
    nodes: graph.nodes,
    scenarioImportText: docs.scenarioImportText,
    selectAction: graph.selectAction,
    selectedNode: graph.selectedNode,
    setEdges: graph.setEdges,
    setError,
    setNodes: graph.setNodes,
    setNotice,
    setScenarioImportText: docs.setScenarioImportText,
    setTestInputText,
    setWorkflow: data.setWorkflow,
    stepsOverrideText,
    testInputText,
    token,
    workflow: data.workflow,
    workflowId,
  });

  return (
    <main className="flow-page" data-feature="workflow">
      <EditorToolbar
        workflow={data.workflow}
        project={data.project}
        notice={notice}
        error={error}
        versionDescription={commands.versionDescription}
        descriptionOpen={commands.descriptionOpen}
        testBusy={commands.testBusy}
        setVersionDescription={commands.setVersionDescription}
        setDescriptionOpen={commands.setDescriptionOpen}
        save={commands.save}
        validate={commands.validate}
        publish={commands.publish}
        runTest={commands.runTest}
        openRuns={() => router.push(`/workflows/${workflowId}/runs`)}
      />

      <div className="flow-workspace">
        <EditorPaletteSidebar
          exportedScenarioJSON={docs.exportedScenarioJSON}
          exportedAIPrompt={docs.exportedAIPrompt}
          aiTaskText={docs.aiTaskText}
          integrationDocResults={docs.integrationDocResults}
          selectedIntegrationDocKeys={docs.selectedIntegrationDocKeys}
          integrationDocsBusy={docs.integrationDocsBusy}
          scenarioImportText={docs.scenarioImportText}
          versions={data.versions}
          versionsBusy={commands.versionsBusy}
          addActionNode={graph.addActionNode}
          setAiTaskText={docs.setAiTaskText}
          setSelectedIntegrationDocKeys={docs.setSelectedIntegrationDocKeys}
          setScenarioImportText={docs.setScenarioImportText}
          applyScenarioImport={commands.applyScenarioImport}
          copyText={docs.copyText}
          restoreVersion={commands.restoreVersion}
        />

        <div className="flow-canvas" onDrop={graph.onDrop} onDragOver={(event) => event.preventDefault()}>
          <ReactFlow
            nodes={graph.flowNodes}
            edges={graph.edges}
            nodeTypes={nodeTypes}
            onNodesChange={graph.onNodesChange}
            onEdgesChange={graph.onEdgesChange}
            onNodeDragStop={() => graph.setNodes((current) => normalizeActionNodeLayout(current, graph.edges))}
            onConnect={graph.onConnect}
            onNodeClick={(_, node) => graph.selectAction(node.id)}
            fitView
          >
            <Background />
            <MiniMap />
            <Controls />
          </ReactFlow>
        </div>

        <EditorSettingsSheet
          selectedNode={graph.selectedNode}
          selectedSubtitle={graph.selectedSubtitle}
          project={data.project}
          token={token}
          integrations={data.integrations}
          webhookSlug={webhookSlug}
          templateVariables={templateVariables}
          testInputText={testInputText}
          stepsOverrideText={stepsOverrideText}
          testBusy={commands.testBusy}
          providerCatalog={data.providerCatalog}
          duplicateActionNode={graph.duplicateActionNode}
          deleteSelectedNode={graph.deleteSelectedNode}
          updateSelectedLabel={graph.updateSelectedLabel}
          updateSelectedConfig={graph.updateSelectedConfig}
          updateRetry={graph.updateRetry}
          setTestInputText={setTestInputText}
          setStepsOverrideText={setStepsOverrideText}
          runTest={commands.runTest}
        />
      </div>
    </main>
  );
}
