import type { WorkProcessRole } from "./work-process";

export type WorkStatus = {
  task_workflow_status_id: string;
  name: string;
  category: "backlog" | "todo" | "in_progress" | "review" | "done" | "canceled" | string;
  color: string;
  position: number;
  is_start: boolean;
  is_done: boolean;
  wip_limit?: number;
  requires_assignee: boolean;
  requires_estimate: boolean;
};

export type WorkTask = {
  task_id: string;
  task_key: string;
  title: string;
  description?: string;
  type: string;
  task_workflow_status_id: string;
  task_workflow_id: string;
  status_name: string;
  status_category: string;
  priority: string;
  severity: string;
  assignee_user_id?: string;
  assignee_name?: string;
  reporter_user_id?: string;
  reviewer_user_id?: string;
  client_id?: string;
  client_name?: string;
  project_id?: string;
  project_name?: string;
  start_date?: string;
  due_date?: string;
  completed_at?: string;
  estimate_minutes?: number;
  spent_minutes_cached: number;
  remaining_minutes?: number;
  story_points?: number;
  rank: number;
  is_blocked: boolean;
  blocked_reason?: string;
  client_visible: boolean;
  billable: boolean;
  labels: string[];
  comments_count: number;
  attachments_count: number;
  checklist_done_count: number;
  checklist_total_count: number;
  revision: number;
  created_at: string;
  updated_at: string;
};

export type WorkBoardColumn = {
  status: WorkStatus;
  count: number;
  tasks: WorkTask[];
};

export type WorkBoard = {
  task_board_id: string;
  name: string;
  description?: string;
  type: string;
  task_workflow_id?: string;
  group_by: string;
  swimlane_by: string;
  visible_fields: string[];
  card_layout?: Record<string, unknown>;
  can_customize: boolean;
  columns?: WorkBoardColumn[];
  summary?: Record<string, number>;
};

export type WorkWorkflow = {
  task_workflow_id: string;
  name: string;
  description?: string;
  scope: string;
  is_default: boolean;
  statuses?: WorkStatus[];
  roles?: WorkProcessRole[];
};
