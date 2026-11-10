export type WorkRolePermissions = {
  read_tasks?: boolean;
  read_client_visible_only?: boolean;
  create_tasks?: boolean;
  edit_tasks?: boolean;
  move_tasks?: boolean;
  comment_tasks?: boolean;
  client_comments_only?: boolean;
  log_time?: boolean;
  start_timer?: boolean;
  view_time?: boolean | "own";
  approve_time?: boolean;
  view_reports?: boolean;
  manage_workflow?: boolean;
  manage_roles?: boolean;
  invite_members?: boolean;
  view_process?: boolean;
  [key: string]: unknown;
};

export type WorkProcessRole = {
  task_process_role_id: string;
  task_workflow_id: string;
  role_key: string;
  name: string;
  description: string;
  permissions: WorkRolePermissions;
  position: number;
  is_system: boolean;
};

export type WorkProcessMembership = {
  task_process_membership_id?: string;
  task_workflow_id: string;
  workflow_name?: string;
  task_process_role_id: string;
  role_key: string;
  role_name: string;
  permissions: WorkRolePermissions;
  client_id?: string;
};

export type WorkAccess = {
  role: string;
  role_name?: string;
  project_name: string;
  company_name: string;
  can_read: boolean;
  can_write_tasks: boolean;
  can_manage_processes: boolean;
  can_manage_members: boolean;
  can_approve_time: boolean;
  can_see_reports: boolean;
  is_client_guest: boolean;
  visible_modules: string[];
  task_workflow_ids?: string[];
  process_roles?: WorkProcessMembership[];
  all_workflows?: boolean;
};
