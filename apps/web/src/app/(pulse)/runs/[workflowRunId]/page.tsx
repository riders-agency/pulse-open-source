import { WorkflowRunDetailScreen } from "../../../../features/workflow-run-detail/WorkflowRunDetailScreen";

export default async function RunPage({ params }: { params: Promise<{ workflowRunId: string }> }) {
  const { workflowRunId } = await params;
  return <WorkflowRunDetailScreen workflowRunId={workflowRunId} />;
}
