package commands

import (
	"context"

	"github.com/google/uuid"
	dbpkg "riders.agency/pulse/api/internal/platform/db"
)

type Membership struct {
	ProjectID, MemberID, ActorID uuid.UUID
	Role, Permissions            string
	ClientID                     any
}

func EnsureMember(ctx context.Context, tx dbpkg.Executor, in Membership) error {
	if _, err := tx.Exec(ctx, `
INSERT INTO project_members (project_id, user_id, role, permissions, client_id, invited_by_user_id, invited_at, status)
VALUES ($1,$2,$3,$4::jsonb,$5::uuid,$6,now(),'active')
ON CONFLICT (project_id, user_id) DO UPDATE SET
  client_id = COALESCE(EXCLUDED.client_id, project_members.client_id),
  status = 'active'`,
		in.ProjectID, in.MemberID, in.Role, in.Permissions, in.ClientID, in.ActorID); err != nil {
		return err
	}
	return nil
}
