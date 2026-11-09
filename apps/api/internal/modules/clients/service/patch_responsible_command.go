package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	repo "riders.agency/pulse/api/internal/modules/clients/repository"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) PatchResponsible(ctx context.Context, clientID uuid.UUID, responsibleID uuid.UUID, uid uuid.UUID, patch repo.ClientResponsiblePatch) error {

	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.Clients.WithTx(tx)
		return bound.PatchResponsibleTx(ctx, clientID, responsibleID, uid, patch)
	})
	return err
}
