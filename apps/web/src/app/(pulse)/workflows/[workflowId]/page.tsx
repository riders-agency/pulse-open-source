import { WorkflowRedirectPage } from "../../../../features/navigation/ui/WorkflowRedirectPage";

export default async function WorkflowPage({ params }: { params: Promise<{ workflowId: string }> }) {
  const { workflowId } = await params;
  return <WorkflowRedirectPage workflowId={workflowId} />;
}
