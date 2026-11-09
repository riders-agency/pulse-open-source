package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	workrepo "riders.agency/pulse/api/internal/modules/worktracker/repository"
)

type roleSeed struct {
	Key         string
	Name        string
	Description string
	Permissions map[string]any
	Position    int
}

func insertCustomFieldSpecs(ctx context.Context, tx pgx.Tx, companyID, projectID, workflowID uuid.UUID, fields []workrepo.CustomFieldSpec) error {
	for i, field := range fields {
		if strings.TrimSpace(field.Key) == "" || strings.TrimSpace(field.Name) == "" {
			continue
		}
		optionsRaw, _ := json.Marshal(field.Options)
		if _, err := tx.Exec(ctx, `
INSERT INTO task_custom_fields (
  company_id, project_id, task_workflow_id, name, field_key, type, scope, required,
  options_json, position, visible_by_default, applies_to_task_types, is_system
) VALUES ($1,$2,$3,$4,$5,$6,'task',$7,$8::jsonb,$9,$10,$11,$12)
ON CONFLICT (company_id, field_key) DO UPDATE SET
  project_id = COALESCE(task_custom_fields.project_id, EXCLUDED.project_id),
  task_workflow_id = COALESCE(task_custom_fields.task_workflow_id, EXCLUDED.task_workflow_id),
  options_json = EXCLUDED.options_json,
  applies_to_task_types = EXCLUDED.applies_to_task_types`,
			companyID.String(), projectID.String(), workflowID.String(), strings.TrimSpace(field.Name),
			strings.TrimSpace(field.Key), defaultText(field.Type, "text"), field.Required,
			string(optionsRaw), (i+1)*10, field.VisibleByDefault, field.AppliesToTaskTypes, true); err != nil {
			return err
		}
	}
	return nil
}

func insertBoardFromTemplate(ctx context.Context, tx pgx.Tx, companyID, projectID, workflowID, userID uuid.UUID, board workrepo.BoardTemplate, isDefault bool) error {
	if strings.TrimSpace(board.Name) == "" {
		return nil
	}
	fieldsRaw, _ := json.Marshal(board.VisibleFields)
	layoutRaw, _ := json.Marshal(map[string]any{"density": defaultText(board.Density, "normal"), "template": board.Key})
	filterRaw, _ := json.Marshal(map[string]any{})
	sortRaw, _ := json.Marshal(map[string]any{"field": "rank", "direction": "asc"})
	_, err := tx.Exec(ctx, `
INSERT INTO task_boards (
  company_id, project_id, name, type, task_workflow_id, group_by, swimlane_by,
  filters_json, sort_json, visible_fields_json, card_layout_json, created_by_user_id, is_default
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9::jsonb,$10::jsonb,$11::jsonb,$12,$13)
ON CONFLICT DO NOTHING`,
		companyID.String(), projectID.String(), board.Name, defaultText(board.Type, "kanban"), workflowID.String(),
		defaultText(board.GroupBy, "status"), defaultText(board.SwimlaneBy, "none"), string(filterRaw), string(sortRaw),
		string(fieldsRaw), string(layoutRaw), userID.String(), isDefault)
	return err
}

func insertDefaultProcessRoles(ctx context.Context, tx pgx.Tx, companyID, projectID, workflowID uuid.UUID) error {
	for _, role := range defaultRoleSeeds() {
		raw, _ := json.Marshal(role.Permissions)
		if _, err := tx.Exec(ctx, `
INSERT INTO task_process_roles (
  company_id, project_id, task_workflow_id, role_key, name, description,
  permissions_json, position, is_system
) VALUES ($1,$2,$3,$4,$5,$6,$7::jsonb,$8,true)
ON CONFLICT (task_workflow_id, role_key) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  permissions_json = EXCLUDED.permissions_json,
  position = EXCLUDED.position,
  is_system = true,
  updated_at = now()`,
			companyID.String(), projectID.String(), workflowID.String(), role.Key, role.Name, role.Description, string(raw), role.Position); err != nil {
			return err
		}
	}
	return nil
}

func defaultRoleSeeds() []roleSeed {
	return []roleSeed{
		{"process_owner", "Владелец процесса", "Полный доступ к задачам, ролям, участникам, отчетам и настройкам процесса.", map[string]any{"read_tasks": true, "create_tasks": true, "edit_tasks": true, "move_tasks": true, "comment_tasks": true, "log_time": true, "view_time": true, "approve_time": true, "view_reports": true, "manage_workflow": true, "manage_roles": true, "invite_members": true, "view_process": true, "start_timer": true}, 10},
		{"process_manager", "Менеджер процесса", "Настраивает процесс, приглашает участников и управляет задачами.", map[string]any{"read_tasks": true, "create_tasks": true, "edit_tasks": true, "move_tasks": true, "comment_tasks": true, "log_time": true, "view_time": true, "approve_time": true, "view_reports": true, "manage_workflow": true, "manage_roles": true, "invite_members": true, "view_process": true, "start_timer": true}, 20},
		{"executor", "Исполнитель", "Работает с задачами процесса, ведет время и комментирует.", map[string]any{"read_tasks": true, "create_tasks": true, "edit_tasks": true, "move_tasks": true, "comment_tasks": true, "log_time": true, "view_time": "own", "view_reports": false, "manage_workflow": false, "manage_roles": false, "invite_members": false, "view_process": false, "start_timer": true}, 30},
		{"reviewer", "Проверяющий", "Проверяет задачи, двигает review/acceptance и комментирует.", map[string]any{"read_tasks": true, "create_tasks": false, "edit_tasks": true, "move_tasks": true, "comment_tasks": true, "log_time": true, "view_time": "own", "view_reports": true, "manage_workflow": false, "manage_roles": false, "invite_members": false, "view_process": false, "start_timer": true}, 40},
		{"observer", "Наблюдатель", "Читает задачи процесса и может оставлять внутренние комментарии.", map[string]any{"read_tasks": true, "create_tasks": false, "edit_tasks": false, "move_tasks": false, "comment_tasks": true, "log_time": false, "view_time": false, "view_reports": true, "manage_workflow": false, "manage_roles": false, "invite_members": false, "view_process": false, "start_timer": false}, 50},
		{"client", "Клиент", "Видит только client-visible задачи своего клиента и пишет клиентские комментарии.", map[string]any{"read_tasks": true, "read_client_visible_only": true, "create_tasks": false, "edit_tasks": false, "move_tasks": false, "comment_tasks": true, "client_comments_only": true, "log_time": false, "view_time": false, "view_reports": false, "manage_workflow": false, "manage_roles": false, "invite_members": false, "view_process": false, "start_timer": false}, 60},
	}
}

func nullableUUIDValue(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}
	return id.String()
}

func statusCategory(index, total int) string {
	switch {
	case index == 0:
		return "backlog"
	case index == total-1:
		return "done"
	case index == total-2:
		return "review"
	case index >= 2:
		return "in_progress"
	default:
		return "todo"
	}
}

func statusColor(category string) string {
	switch category {
	case "done":
		return "green"
	case "review":
		return "violet"
	case "in_progress":
		return "amber"
	case "todo":
		return "blue"
	default:
		return "slate"
	}
}
