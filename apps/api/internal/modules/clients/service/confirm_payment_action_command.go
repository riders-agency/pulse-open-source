package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	repo "riders.agency/pulse/api/internal/modules/clients/repository"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) ConfirmPaymentAction(ctx context.Context, actionID uuid.UUID, uid uuid.UUID, in repo.ConfirmPaymentInput) error {

	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.Clients.WithTx(tx)
		return bound.ConfirmPaymentActionTx(ctx, actionID, uid, in)
	})
	return err
}
