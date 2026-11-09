import type { WorkTask } from "./work-board";

export type WorkChecklistItem = {
  task_checklist_item_id: string;
  title: string;
  is_done: boolean;
  position: number;
  completed_at?: string;
};

export type WorkComment = {
  task_comment_id: string;
  author_name: string;
  body: string;
  visibility: string;
  created_at: string;
};

export type WorkCustomField = {
  task_custom_field_id: string;
  name: string;
  key: string;
  type: string;
  required: boolean;
  options: string[];
  value?: unknown;
  visible_by_default: boolean;
};

export type WorkActivityEvent = {
  event_type: string;
  actor_name: string;
  payload_json?: unknown;
  created_at: string;
};

export type WorkTimeEntry = {
  time_entry_id: string;
  task_id?: string;
  task_key?: string;
  task_title?: string;
  project_id?: string;
  project_name?: string;
  client_id?: string;
  client_name?: string;
  user_id: string;
  user_name: string;
  entry_date: string;
  duration_minutes: number;
  description?: string;
  billable: boolean;
  source: string;
  approved_at?: string;
};

export type WorkTaskAttachment = {
  task_attachment_id: string;
  task_id: string;
  file_name: string;
  content_type: string;
  size_bytes: number;
  uploaded_by_name: string;
  download_url: string;
  created_at: string;
};

export type WorkTaskDetail = WorkTask & {
  comments: WorkComment[];
  checklist: WorkChecklistItem[];
  custom_fields: WorkCustomField[];
  activity: WorkActivityEvent[];
  time_entries: WorkTimeEntry[];
  attachments: WorkTaskAttachment[];
};

export type WorkTaskList = {
  items: WorkTask[];
  total: number;
  limit: number;
  offset: number;
};

export type WorkOverview = {
  today_tasks: number;
  overdue_tasks: number;
  unassigned_tasks: number;
  unestimated_tasks: number;
  blocked_tasks: number;
  review_tasks: number;
  client_acceptance_tasks: number;
  logged_today_minutes: number;
  logged_week_minutes: number;
};

export type WorkTimer = {
  timer_session_id: string;
  task_id: string;
  task_key: string;
  task_title: string;
  started_at: string;
  state: string;
  elapsed_minutes: number;
};
