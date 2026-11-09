import { WorkflowEditorRouteScreen } from "../../../../../features/workflow-editor/WorkflowEditorRouteScreen";

export default async function WorkflowEditorPage({ params }: { params: Promise<{ workflowId: string }> }) {
  const { workflowId } = await params;
  return <WorkflowEditorRouteScreen workflowId={workflowId} />;
}
