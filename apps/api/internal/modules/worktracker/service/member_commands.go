package service

import (
	"context"
	"strings"

	authcommands "riders.agency/pulse/api/internal/modules/auth/commands"
	companycommands "riders.agency/pulse/api/internal/modules/companies/commands"
	projectcommands "riders.agency/pulse/api/internal/modules/projects/commands"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	workrepo "riders.agency/pulse/api/internal/modules/worktracker/repository"
)

func (c *Commands) InviteMember(ctx context.Context, userID uuid.UUID, in workrepo.InviteMemberInput) (map[string]any, error) {
	if c == nil || c.WorkTracker == nil || c.WorkTracker.P == nil {
		return nil, workrepo.ErrNotFound
	}
	access, err := c.WorkTracker.AccessForProject(ctx, in.ProjectID, userID)
	if err != nil {
		return nil, err
	}
	if in.WorkflowID == uuid.Nil {
		return nil, workrepo.ErrValidation
	}
	if !access.CanInviteMembersInWorkflow(in.WorkflowID.String()) {
		return nil, workrepo.ErrForbidden
	}
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" {
		return nil, workrepo.ErrValidation
	}
	roleID := in.RoleID
	if roleID == uuid.Nil {
		roleKey := strings.TrimSpace(in.Role)
		if roleKey == "" {
			roleKey = "executor"
		}
		if err := c.WorkTracker.P.QueryRow(ctx, `
SELECT task_process_role_id
  FROM task_process_roles
 WHERE task_workflow_id = $1
   AND role_key = $2`, in.WorkflowID.String(), roleKey).Scan(&roleID); err != nil {
			if err == pgx.ErrNoRows {
				return nil, workrepo.ErrValidation
			}
			return nil, err
		}
	} else {
		var ok bool
		if err := c.WorkTracker.P.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1
    FROM task_process_roles
   WHERE task_process_role_id = $1
     AND task_workflow_id = $2
     AND project_id = $3
)`, roleID.String(), in.WorkflowID.String(), in.ProjectID.String()).Scan(&ok); err != nil {
			return nil, err
		}
		if !ok {
			return nil, workrepo.ErrValidation
		}
	}
	tx, err := c.WorkTracker.P.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	memberID, temp, created, err := authcommands.UpsertInviteUser(ctx, tx, email, strings.TrimSpace(in.DisplayName))
	if err != nil {
		return nil, err
	}
	if err := companycommands.EnsureMember(ctx, tx, access.CompanyID, memberID, userID); err != nil {
		return nil, err
	}
	projectRole := defaultProjectRoleForProcessInvite(in.Role)
	if err := projectcommands.EnsureMember(ctx, tx, projectcommands.Membership{ProjectID: in.ProjectID, MemberID: memberID, ActorID: userID, Role: projectRole, Permissions: permissionsForRole(projectRole), ClientID: nullableUUID(in.ClientID)}); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO task_process_memberships (
  company_id, project_id, task_workflow_id, user_id, task_process_role_id,
  client_id, invited_by_user_id, invited_at, status
) VALUES ($1,$2,$3,$4,$5,$6::uuid,$7,now(),'active')
ON CONFLICT (task_workflow_id, user_id) DO UPDATE SET
  task_process_role_id = EXCLUDED.task_process_role_id,
  client_id = EXCLUDED.client_id,
  invited_by_user_id = EXCLUDED.invited_by_user_id,
  invited_at = now(),
  status = 'active',
  updated_at = now()`,
		access.CompanyID.String(), in.ProjectID.String(), in.WorkflowID.String(), memberID.String(), roleID.String(), nullableUUID(in.ClientID), userID.String()); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	role, _ := c.WorkTracker.ProcessRole(ctx, roleID)
	out := map[string]any{"user_id": memberID.String(), "role": role, "created_user": created}
	if created {
		out["temporary_password"] = temp
	}
	return out, nil
}

func permissionsForRole(role string) string {
	switch role {
	case "client_guest":
		return `{"tasks":"client_visible","time":"none","reports":"none","processes":"none"}`
	case "employee", "contractor":
		return `{"tasks":"write","time":"own","reports":"none","processes":"none"}`
	case "project_manager":
		return `{"tasks":"admin","time":"approve","reports":"read","processes":"admin"}`
	default:
		return `{"tasks":"read","time":"own","reports":"read","processes":"none"}`
	}
}

func defaultProjectRoleForProcessInvite(role string) string {
	switch strings.TrimSpace(role) {
	case "client", "client_guest":
		return "client_guest"
	case "observer", "read_only_auditor":
		return "read_only_auditor"
	case "process_owner", "process_manager":
		return "employee"
	default:
		return "employee"
	}
}
