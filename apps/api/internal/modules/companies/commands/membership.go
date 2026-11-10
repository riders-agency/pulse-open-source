package commands

import (
	"context"

	"github.com/google/uuid"
	dbpkg "riders.agency/pulse/api/internal/platform/db"
)

func EnsureMember(ctx context.Context, tx dbpkg.Executor, companyID, memberID, actorID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `
INSERT INTO company_memberships (company_id, user_id, role, invited_by, invited_at)
VALUES ($1,$2,'member',$3,now())
ON CONFLICT (company_id, user_id) DO NOTHING`, companyID, memberID, actorID); err != nil {
		return err
	}
	return nil
}
