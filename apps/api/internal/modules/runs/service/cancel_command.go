package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) Cancel(ctx context.Context, runID uuid.UUID, uid uuid.UUID) error {

	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.Runs.WithTx(tx)
		return bound.CancelTx(ctx, runID, uid)
	})
	return err
}
