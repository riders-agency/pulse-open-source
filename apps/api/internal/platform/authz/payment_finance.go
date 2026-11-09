package authz

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	dbpkg "riders.agency/pulse/api/internal/platform/db"
	"riders.agency/pulse/api/internal/platform/errs"
)

// RequirePaymentFinance additionally restricts session writes to finance/managers.
// API keys retain the explicitly granted payment/client/project scope.
func RequirePaymentFinance(ctx context.Context, pool dbpkg.Executor, actionID uuid.UUID, permission string) error {
	if err := RequirePaymentAction(ctx, pool, actionID, permission); err != nil {
		return err
	}
	claims, _ := SessionClaimsFrom(ctx)
	if claims.ActorType == ActorAPIKey {
		return nil
	}
	_, projectID, _, err := scopeForPaymentAction(ctx, pool, actionID)
	if err != nil {
		return err
	}
	return requireFinanceMember(ctx, pool, projectID, claims.UserID, permission)
}

func RequireProjectFinance(ctx context.Context, pool dbpkg.Executor, projectID uuid.UUID) error {
	if err := RequireProject(ctx, pool, projectID, PermWrite); err != nil {
		return err
	}
	claims, _ := SessionClaimsFrom(ctx)
	if claims.ActorType == ActorAPIKey {
		return nil
	}
	return requireFinanceMember(ctx, pool, projectID, claims.UserID, PermWrite)
}

func requireFinanceMember(ctx context.Context, pool dbpkg.Executor, projectID, userID uuid.UUID, permission string) error {
	var role string
	err := pool.QueryRow(ctx, `SELECT pm.role FROM project_members pm
WHERE pm.project_id=$1 AND pm.user_id=$2 AND pm.status='active'`, projectID, userID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrForbidden
	}
	if err != nil {
		return err
	}
	if permission == PermRead {
		return nil
	}
	switch role {
	case "owner", "admin", "operations_manager", "project_manager", "finance":
		return nil
	default:
		return errs.ErrForbidden
	}
}
