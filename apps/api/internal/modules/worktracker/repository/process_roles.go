package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type rowsQueryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func ProjectRoleLabel(role string) string {
	switch role {
	case "owner":
		return "Владелец"
	case "admin":
		return "Администратор"
	case "operations_manager":
		return "Операционный менеджер"
	case "project_manager":
		return "Руководитель проекта"
	case "account_manager":
		return "Аккаунт-менеджер"
	case "team_lead":
		return "Тимлид"
	case "employee":
		return "Сотрудник"
	case "contractor":
		return "Подрядчик"
	case "client_guest":
		return "Клиент"
	case "finance":
		return "Финансы"
	case "hr":
		return "HR"
	case "read_only_auditor":
		return "Аудитор"
	default:
		return "Участник"
	}
}

func (w *WorkTracker) ListProcessRoles(ctx context.Context, workflowID, userID uuid.UUID) ([]ProcessRole, *Access, error) {
	access, err := w.AccessForWorkflow(ctx, workflowID, userID)
	if err != nil {
		return nil, nil, err
	}
	if !access.CanUseWorkflow(workflowID.String()) && !access.CanManageProcesses {
		return nil, nil, ErrForbidden
	}
	rows, err := w.pool().Query(ctx, `
SELECT task_process_role_id::text, task_workflow_id::text, role_key, name,
       COALESCE(description, ''), permissions_json, position, is_system
  FROM task_process_roles
 WHERE task_workflow_id = $1
 ORDER BY position, name`, workflowID.String())
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	roles, err := scanProcessRoles(rows)
	return roles, access, err
}

func (w *WorkTracker) AccessForWorkflow(ctx context.Context, workflowID, userID uuid.UUID) (*Access, error) {
	var projectID uuid.UUID
	if err := w.pool().QueryRow(ctx, `SELECT project_id FROM task_workflows WHERE task_workflow_id = $1`, workflowID.String()).Scan(&projectID); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return w.AccessForProject(ctx, projectID, userID)
}

func (w *WorkTracker) ProcessRole(ctx context.Context, roleID uuid.UUID) (*ProcessRole, error) {
	rows, err := w.pool().Query(ctx, `
SELECT task_process_role_id::text, task_workflow_id::text, role_key, name,
       COALESCE(description, ''), permissions_json, position, is_system
  FROM task_process_roles
 WHERE task_process_role_id = $1`, roleID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	roles, err := scanProcessRoles(rows)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return nil, ErrNotFound
	}
	return &roles[0], nil
}

func processRolesForWorkflow(ctx context.Context, pool rowsQueryer, workflowID uuid.UUID) ([]ProcessRole, error) {
	rows, err := pool.Query(ctx, `
SELECT task_process_role_id::text, task_workflow_id::text, role_key, name,
       COALESCE(description, ''), permissions_json, position, is_system
  FROM task_process_roles
 WHERE task_workflow_id = $1
 ORDER BY position, name`, workflowID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProcessRoles(rows)
}

func scanProcessRoles(rows pgx.Rows) ([]ProcessRole, error) {
	out := []ProcessRole{}
	for rows.Next() {
		var role ProcessRole
		var raw []byte
		if err := rows.Scan(
			&role.TaskProcessRoleID,
			&role.TaskWorkflowID,
			&role.RoleKey,
			&role.Name,
			&role.Description,
			&raw,
			&role.Position,
			&role.IsSystem,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(raw, &role.Permissions)
		if role.Permissions == nil {
			role.Permissions = map[string]any{}
		}
		out = append(out, role)
	}
	return out, rows.Err()
}
