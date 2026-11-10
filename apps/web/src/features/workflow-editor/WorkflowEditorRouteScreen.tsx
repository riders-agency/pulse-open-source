"use client";

import { ReactFlowProvider } from "@xyflow/react";

import { WorkflowEditorScreen } from "./WorkflowEditorScreen";

export function WorkflowEditorRouteScreen({ workflowId }: { workflowId: string }) {
  return (
    <ReactFlowProvider>
      <WorkflowEditorScreen workflowId={workflowId} />
    </ReactFlowProvider>
  );
}
