package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/google/uuid"

	workrepo "riders.agency/pulse/api/internal/modules/worktracker/repository"
)

func (c *Commands) UpsertProcessRole(ctx context.Context, userID uuid.UUID, in workrepo.UpsertProcessRoleInput) (*workrepo.ProcessRole, error) {
	if c == nil || c.WorkTracker == nil || c.WorkTracker.P == nil {
		return nil, workrepo.ErrNotFound
	}
	access, err := c.WorkTracker.AccessForWorkflow(ctx, in.WorkflowID, userID)
	if err != nil {
		return nil, err
	}
	if !access.CanManageRolesInWorkflow(in.WorkflowID.String()) {
		return nil, workrepo.ErrForbidden
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, workrepo.ErrValidation
	}
	perms := normalizeRolePermissions(in.Permissions)
	raw, _ := json.Marshal(perms)
	roleID := in.RoleID
	if roleID == uuid.Nil {
		roleID = uuid.New()
	}
	roleKey := "custom_" + shortRoleKey(roleID)
	var execErr error
	if in.RoleID == uuid.Nil {
		_, execErr = c.WorkTracker.P.Exec(ctx, `
INSERT INTO task_process_roles (
  task_process_role_id, company_id, project_id, task_workflow_id, role_key, name,
  description, permissions_json, position, is_system
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9,false)`,
			roleID.String(), access.CompanyID.String(), access.ProjectID.String(), in.WorkflowID.String(),
			roleKey, name, strings.TrimSpace(in.Description), string(raw), in.Position)
	} else {
		_, execErr = c.WorkTracker.P.Exec(ctx, `
UPDATE task_process_roles
   SET name = $3,
       description = $4,
       permissions_json = $5::jsonb,
       position = $6,
       is_system = false,
       updated_at = now()
 WHERE task_process_role_id = $1
   AND task_workflow_id = $2`,
			in.RoleID.String(), in.WorkflowID.String(), name, strings.TrimSpace(in.Description), string(raw), in.Position)
	}
	if execErr != nil {
		return nil, execErr
	}
	return c.WorkTracker.ProcessRole(ctx, roleID)
}

func (c *Commands) ResetProcessRoles(ctx context.Context, workflowID, userID uuid.UUID) ([]workrepo.ProcessRole, error) {
	if c == nil || c.WorkTracker == nil || c.WorkTracker.P == nil {
		return nil, workrepo.ErrNotFound
	}
	access, err := c.WorkTracker.AccessForWorkflow(ctx, workflowID, userID)
	if err != nil {
		return nil, err
	}
	if !access.CanManageRolesInWorkflow(workflowID.String()) {
		return nil, workrepo.ErrForbidden
	}
	tx, err := c.WorkTracker.P.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM task_process_roles WHERE task_workflow_id = $1 AND is_system = false`, workflowID.String()); err != nil {
		return nil, err
	}
	if err := insertDefaultProcessRoles(ctx, tx, access.CompanyID, access.ProjectID, workflowID); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	roles, _, err := c.WorkTracker.ListProcessRoles(ctx, workflowID, userID)
	return roles, err
}

func normalizeRolePermissions(in map[string]any) map[string]any {
	out := map[string]any{}
	keys := []string{
		"read_tasks", "read_client_visible_only", "create_tasks", "edit_tasks", "move_tasks",
		"comment_tasks", "client_comments_only", "log_time", "view_time", "approve_time",
		"view_reports", "manage_workflow", "manage_roles", "invite_members", "view_process", "start_timer",
	}
	for _, key := range keys {
		if value, ok := in[key]; ok {
			out[key] = value
		} else {
			out[key] = false
		}
	}
	return out
}

func shortRoleKey(id uuid.UUID) string {
	raw := make([]byte, 6)
	if _, err := rand.Read(raw); err == nil {
		return strings.TrimRight(base64.RawURLEncoding.EncodeToString(raw), "=")
	}
	return strings.ReplaceAll(id.String()[:8], "-", "")
}
