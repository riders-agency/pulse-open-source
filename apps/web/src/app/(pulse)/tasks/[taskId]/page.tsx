import { WorkTaskScreen } from "../../../../features/work/work-task-drawer";

export default async function TaskDetailPage({ params }: { params: Promise<{ taskId?: string }> }) {
  const { taskId = "" } = await params;
  return <WorkTaskScreen taskId={taskId} closeHref="/tasks/boards" />;
}
