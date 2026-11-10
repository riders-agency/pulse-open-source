package repository

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (w *WorkTracker) AccessForProject(ctx context.Context, projectID, userID uuid.UUID) (*Access, error) {
	row := w.pool().QueryRow(ctx, `
SELECT pr.company_id,
       pr.project_id,
       pr.name,
       co.name,
       pm.role,
       pm.client_id
  FROM pulse_read.projects_v1 pr
  JOIN pulse_read.companies_v1 co ON co.company_id = pr.company_id
  JOIN pulse_read.project_members_v1 pm ON pm.project_id = pr.project_id AND pm.user_id = $2
  JOIN pulse_read.company_memberships_v1 cm ON cm.company_id = pr.company_id AND cm.user_id = $2
 WHERE pr.project_id = $1
   AND pm.status = 'active'`, projectID.String(), userID.String())
	var access Access
	var clientID *uuid.UUID
	if err := row.Scan(&access.CompanyID, &access.ProjectID, &access.ProjectName, &access.CompanyName, &access.Role, &clientID); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrForbidden
		}
		return nil, err
	}
	access.UserID = userID
	access.ClientID = clientID
	applyProjectCapabilities(&access)
	if err := w.loadProcessMemberships(ctx, &access); err != nil {
		return nil, err
	}
	applyProcessCapabilities(&access)
	return &access, nil
}

func (w *WorkTracker) AccessForTask(ctx context.Context, taskID, userID uuid.UUID) (*Access, uuid.UUID, error) {
	var active bool
	if err := w.pool().QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tasks WHERE task_id=$1 AND archived_at IS NULL)`, taskID).Scan(&active); err != nil {
		return nil, uuid.Nil, err
	}
	if !active {
		return nil, uuid.Nil, ErrNotFound
	}
	return w.AccessForStoredTask(ctx, taskID, userID)
}

func (w *WorkTracker) AccessForStoredTask(ctx context.Context, taskID, userID uuid.UUID) (*Access, uuid.UUID, error) {
	var projectID, workflowID uuid.UUID
	if err := w.pool().QueryRow(ctx, `
SELECT project_id, task_workflow_id FROM tasks WHERE task_id = $1`, taskID.String()).Scan(&projectID, &workflowID); err != nil {
		if err == pgx.ErrNoRows {
			return nil, uuid.Nil, ErrNotFound
		}
		return nil, uuid.Nil, err
	}
	access, err := w.AccessForProject(ctx, projectID, userID)
	if err != nil {
		return nil, uuid.Nil, err
	}
	if !access.CanUseWorkflow(workflowID.String()) {
		return nil, uuid.Nil, ErrForbidden
	}
	return access, projectID, nil
}

func applyProjectCapabilities(access *Access) {
	access.CanRead = true
	access.RoleName = ProjectRoleLabel(access.Role)
	access.IsClientGuest = access.Role == "client_guest"
	access.AllWorkflows = isProjectAdminRole(access.Role)
	switch access.Role {
	case "owner", "admin", "operations_manager":
		access.CanWriteTasks = true
		access.CanManageProcesses = true
		access.CanManageMembers = true
		access.CanApproveTime = true
		access.CanSeeReports = true
		access.VisibleModules = []string{"dashboard", "clients", "projects", "tasks", "workflows", "employees", "time", "reports", "automations", "settings"}
	case "project_manager":
		access.CanWriteTasks = true
		access.CanManageProcesses = true
		access.CanManageMembers = true
		access.CanApproveTime = true
		access.CanSeeReports = true
		access.VisibleModules = []string{"dashboard", "clients", "projects", "tasks", "workflows", "employees", "time", "reports"}
	case "account_manager":
		access.CanWriteTasks = true
		access.CanSeeReports = true
		access.VisibleModules = []string{"dashboard", "clients", "projects", "tasks", "time", "reports"}
	case "team_lead":
		access.CanWriteTasks = true
		access.CanApproveTime = true
		access.CanSeeReports = true
		access.VisibleModules = []string{"dashboard", "projects", "tasks", "employees", "time", "reports"}
	case "employee":
		access.CanWriteTasks = true
		access.VisibleModules = []string{"dashboard", "projects", "tasks", "time"}
	case "contractor":
		access.CanWriteTasks = true
		access.VisibleModules = []string{"tasks", "time"}
	case "finance":
		access.CanApproveTime = true
		access.CanSeeReports = true
		access.VisibleModules = []string{"dashboard", "clients", "projects", "time", "reports"}
	case "hr":
		access.CanSeeReports = true
		access.VisibleModules = []string{"dashboard", "employees", "time", "reports"}
	case "read_only_auditor":
		access.CanSeeReports = true
		access.VisibleModules = []string{"dashboard", "tasks", "time", "reports"}
	case "client_guest":
		access.VisibleModules = []string{"tasks"}
	default:
		access.CanWriteTasks = true
		access.VisibleModules = []string{"dashboard", "projects", "tasks", "time"}
	}
}

func (w *WorkTracker) loadProcessMemberships(ctx context.Context, access *Access) error {
	rows, err := w.pool().Query(ctx, `
SELECT pm.task_process_membership_id::text,
       pm.task_workflow_id::text,
       tw.name,
       r.task_process_role_id::text,
       r.role_key,
       r.name,
       r.permissions_json,
       COALESCE(pm.client_id::text, '')
  FROM task_process_memberships pm
  JOIN task_process_roles r ON r.task_process_role_id = pm.task_process_role_id
  JOIN task_workflows tw ON tw.task_workflow_id = pm.task_workflow_id
 WHERE pm.project_id = $1
   AND pm.user_id = $2
   AND pm.status = 'active'
 ORDER BY tw.created_at, r.position`, access.ProjectID.String(), access.UserID.String())
	if err != nil {
		return err
	}
	defer rows.Close()
	seen := map[string]bool{}
	for rows.Next() {
		var membership ProcessMembership
		var raw []byte
		if err := rows.Scan(
			&membership.TaskProcessMembershipID,
			&membership.TaskWorkflowID,
			&membership.WorkflowName,
			&membership.TaskProcessRoleID,
			&membership.RoleKey,
			&membership.RoleName,
			&raw,
			&membership.ClientID,
		); err != nil {
			return err
		}
		_ = json.Unmarshal(raw, &membership.Permissions)
		if membership.Permissions == nil {
			membership.Permissions = map[string]any{}
		}
		access.ProcessRoles = append(access.ProcessRoles, membership)
		if !seen[membership.TaskWorkflowID] {
			access.WorkflowIDs = append(access.WorkflowIDs, membership.TaskWorkflowID)
			seen[membership.TaskWorkflowID] = true
		}
	}
	return rows.Err()
}

func applyProcessCapabilities(access *Access) {
	if access.AllWorkflows {
		return
	}
	access.CanWriteTasks = false
	access.CanManageProcesses = false
	access.CanManageMembers = false
	access.CanApproveTime = false
	access.CanSeeReports = false
	for _, role := range access.ProcessRoles {
		if boolPerm(role.Permissions, "create_tasks") || boolPerm(role.Permissions, "edit_tasks") || boolPerm(role.Permissions, "move_tasks") {
			access.CanWriteTasks = true
		}
		if boolPerm(role.Permissions, "manage_workflow") || boolPerm(role.Permissions, "view_process") {
			access.CanManageProcesses = access.CanManageProcesses || boolPerm(role.Permissions, "manage_workflow")
		}
		if boolPerm(role.Permissions, "manage_roles") || boolPerm(role.Permissions, "invite_members") {
			access.CanManageMembers = true
		}
		if boolPerm(role.Permissions, "approve_time") {
			access.CanApproveTime = true
		}
		if boolPerm(role.Permissions, "view_reports") {
			access.CanSeeReports = true
		}
	}
	if len(access.ProcessRoles) == 0 {
		access.VisibleModules = []string{}
		return
	}
	modules := []string{"tasks"}
	if access.CanWriteTasks || workflowPermAny(access, "log_time") || workflowPermAny(access, "view_time") {
		modules = append(modules, "time")
	}
	if access.CanSeeReports {
		modules = append(modules, "reports")
	}
	if access.CanManageMembers {
		modules = append(modules, "employees")
	}
	if access.CanManageProcesses {
		modules = append(modules, "workflows")
	}
	access.VisibleModules = modules
}

func isProjectAdminRole(role string) bool {
	switch role {
	case "owner", "admin", "operations_manager", "project_manager":
		return true
	default:
		return false
	}
}

func (a Access) CanUse(module string) bool {
	return slices.Contains(a.VisibleModules, module)
}

func (a Access) CanSeeTask(task Task) bool {
	if !a.CanUseWorkflow(task.WorkflowID) {
		return false
	}
	if a.IsClientGuest {
		if a.ClientID == nil || task.ClientID == "" {
			return false
		}
		return task.ClientID == a.ClientID.String() && task.ClientVisible
	}
	if a.Role == "contractor" {
		return task.AssigneeUserID == a.UserID.String() || task.ReporterUserID == a.UserID.String()
	}
	return true
}

func (a Access) CanUseWorkflow(workflowID string) bool {
	if workflowID == "" {
		return false
	}
	if a.AllWorkflows {
		return true
	}
	return slices.Contains(a.WorkflowIDs, workflowID)
}

func (a Access) CanCreateTaskInWorkflow(workflowID string) bool {
	return a.workflowPerm(workflowID, "create_tasks") || a.workflowPerm(workflowID, "edit_tasks")
}

func (a Access) CanEditTaskInWorkflow(workflowID string) bool {
	return a.workflowPerm(workflowID, "edit_tasks")
}

func (a Access) CanMoveTaskInWorkflow(workflowID string) bool {
	return a.workflowPerm(workflowID, "move_tasks")
}

func (a Access) CanCommentTaskInWorkflow(workflowID string) bool {
	return a.workflowPerm(workflowID, "comment_tasks")
}

func (a Access) CanLogTimeInWorkflow(workflowID string) bool {
	return a.workflowPerm(workflowID, "log_time") || a.workflowPerm(workflowID, "start_timer")
}

func (a Access) CanManageRolesInWorkflow(workflowID string) bool {
	return a.workflowPerm(workflowID, "manage_roles")
}

func (a Access) CanInviteMembersInWorkflow(workflowID string) bool {
	return a.workflowPerm(workflowID, "invite_members") || a.workflowPerm(workflowID, "manage_roles")
}

func (a Access) HasTransitionRole(workflowID string, allowed []string) bool {
	if len(allowed) == 0 || a.AllWorkflows {
		return true
	}
	if slices.Contains(allowed, a.Role) {
		return true
	}
	for _, membership := range a.ProcessRoles {
		if membership.TaskWorkflowID == workflowID && slices.Contains(allowed, membership.RoleKey) {
			return true
		}
	}
	return false
}

func (a Access) workflowPerm(workflowID, perm string) bool {
	if a.AllWorkflows {
		return true
	}
	for _, membership := range a.ProcessRoles {
		if membership.TaskWorkflowID == workflowID && boolPerm(membership.Permissions, perm) {
			return true
		}
	}
	return false
}

func workflowPermAny(access *Access, perm string) bool {
	for _, membership := range access.ProcessRoles {
		if boolPerm(membership.Permissions, perm) {
			return true
		}
	}
	return false
}

func boolPerm(perms map[string]any, key string) bool {
	value, ok := perms[key]
	if !ok {
		return false
	}
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v == "true" || v == "own"
	default:
		return false
	}
}

func (a Access) RequireTasksWrite() error {
	if !a.CanWriteTasks || a.IsClientGuest {
		return ErrForbidden
	}
	return nil
}

func (a Access) RequireProcessesAdmin() error {
	if !a.CanManageProcesses {
		return ErrForbidden
	}
	return nil
}

func (a Access) RequireMembersAdmin() error {
	if !a.CanManageMembers {
		return ErrForbidden
	}
	return nil
}
