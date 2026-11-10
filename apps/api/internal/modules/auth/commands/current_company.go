package commands

import (
	"context"

	"github.com/google/uuid"
	"riders.agency/pulse/api/internal/platform/authz"
	dbpkg "riders.agency/pulse/api/internal/platform/db"
	"riders.agency/pulse/api/internal/platform/errs"
)

func SetCurrentCompany(ctx context.Context, db dbpkg.Executor, userID, companyID uuid.UUID) error {
	claims, ok := authz.SessionClaimsFrom(ctx)
	if !ok || claims.UserID != userID || claims.ActorType == authz.ActorAPIKey {
		return errs.ErrForbidden
	}
	if err := authz.RequireCompany(ctx, db, companyID, authz.PermRead); err != nil {
		return err
	}
	result, err := db.Exec(ctx, `UPDATE users SET current_company_id=$2 WHERE user_id=$1`, userID, companyID)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return errs.ErrForbidden
	}
	return nil
}
