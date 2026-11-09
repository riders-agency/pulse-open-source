import type {
  WorkAccess,
  WorkBoard,
  WorkEmployee,
  WorkOverview,
  WorkReport,
  WorkTaskDetail,
  WorkTaskList,
  WorkTimeEntry,
  WorkTimer,
  WorkWorkflow,
} from "../../../lib/api";

export type WorkBoardEnvelope = { access: WorkAccess; board: WorkBoard };
export type WorkBoardsEnvelope = { access: WorkAccess; items: WorkBoard[] };
export type WorkTaskEnvelope = { access: WorkAccess; task: WorkTaskDetail };
export type WorkTasksEnvelope = { access: WorkAccess; tasks: WorkTaskList };
export type WorkTimeEnvelope = { access: WorkAccess; items: WorkTimeEntry[] };
export type WorkTimersEnvelope = { timer: WorkTimer | null; timers: WorkTimer[] };
export type WorkWorkflowsEnvelope = { access: WorkAccess; items: WorkWorkflow[] };
export type WorkEmployeesEnvelope = { access: WorkAccess; items: WorkEmployee[] };
export type WorkReportEnvelope = { access: WorkAccess; report: WorkReport };
export type WorkOverviewEnvelope = { access: WorkAccess; overview: WorkOverview };

export type WorkTaskQuery = {
  archived?: boolean;
  limit?: number;
  search?: string;
  assignee?: "me" | "unassigned";
  due?: "overdue";
  taskWorkflowId?: string;
};

export type CreateWorkTaskInput = {
  project_id: string;
  task_workflow_status_id?: string;
  title: string;
  description?: string;
  type: string;
  priority: string;
  severity: string;
  client_id?: string;
  due_date?: string;
  estimate_minutes?: number;
  billable: boolean;
  labels?: string[];
};

export type CreateTaskWorkflowInput = {
  project_id: string;
  name: string;
  template_key?: string;
  status_names: string[];
  create_default_boards: boolean;
};

export type ProcessRoleInput = {
  name: string;
  description: string;
  permissions: Record<string, unknown>;
  position: number;
};
