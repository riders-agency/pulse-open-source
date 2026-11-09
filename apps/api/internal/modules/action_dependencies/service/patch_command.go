package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	repo "riders.agency/pulse/api/internal/modules/action_dependencies/repository"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) Patch(ctx context.Context, pid uuid.UUID, depID uuid.UUID, uid uuid.UUID, patch repo.DepPatch) error {

	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.Deps.WithTx(tx)
		return bound.PatchTx(ctx, pid, depID, uid, patch)
	})
	return err
}
