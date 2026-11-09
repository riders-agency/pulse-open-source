package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	aitools "riders.agency/pulse/api/internal/modules/ai/tools"
	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/errs"
)

func (a *AI) EmployeePolicy(ctx context.Context, projectID uuid.UUID) (aitools.EmployeePolicy, error) {
	var policy aitools.EmployeePolicy
	if err := authz.RequireProject(ctx, a.P, projectID, authz.PermWrite); err != nil {
		return policy, err
	}
	var status string
	var raw []byte
	err := a.P.QueryRow(ctx, `SELECT status,tool_policy_json FROM project_ai_employees WHERE project_id=$1 AND employee_key='default' FOR SHARE`, projectID).Scan(&status, &raw)
	if err == pgx.ErrNoRows || (err == nil && status != "active") {
		return policy, errs.ErrForbidden
	}
	if err != nil {
		return policy, err
	}
	policy, err = aitools.ParseEmployeePolicy(raw)
	if err != nil {
		return policy, errs.ErrForbidden
	}
	return policy, nil
}

func (a *AI) RequireEmployeeTool(ctx context.Context, projectID uuid.UUID, tool string) error {
	policy, err := a.EmployeePolicy(ctx, projectID)
	if err != nil {
		return err
	}
	if !policy.Allows(tool) {
		return errs.ErrForbidden
	}
	return nil
}
