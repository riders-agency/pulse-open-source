export type WorkTemplateStatus = {
  name: string;
  category: string;
  color: string;
  wip_limit?: number;
  requires_assignee: boolean;
  requires_estimate: boolean;
  auto_start_timer: boolean;
  auto_stop_timer: boolean;
};

export type WorkBoardTemplate = {
  key: string;
  name: string;
  description: string;
  type: string;
  group_by: string;
  swimlane_by: string;
  visible_fields: string[];
  density: string;
};

export type WorkWorkflowTemplate = {
  key: string;
  name: string;
  description: string;
  task_types: string[];
  statuses: WorkTemplateStatus[];
  boards: WorkBoardTemplate[];
};

export type WorkPresets = {
  workflow_templates: WorkWorkflowTemplate[];
  board_templates: WorkBoardTemplate[];
};
