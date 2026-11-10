import type { WorkProcessMembership } from "./work-process";

export type WorkEmployee = {
  employee_profile_id: string;
  user_id: string;
  full_name: string;
  email: string;
  position: string;
  department: string;
  role: string;
  employment_type: string;
  status: string;
  capacity_hours_per_week: number;
  skills: string[];
  tasks_in_progress: number;
  overdue_tasks: number;
  logged_week_minutes: number;
  billable_week_minutes: number;
  process_accesses?: WorkProcessMembership[];
};

export type WorkReportBucket = {
  id?: string;
  name: string;
  tasks: number;
  logged_minutes: number;
  billable_minutes: number;
};

export type WorkReport = {
  created_tasks: number;
  completed_tasks: number;
  overdue_tasks: number;
  blocked_tasks: number;
  logged_minutes: number;
  billable_minutes: number;
  by_client: WorkReportBucket[];
  by_project: WorkReportBucket[];
  by_employee: WorkReportBucket[];
  by_status: WorkReportBucket[];
};
