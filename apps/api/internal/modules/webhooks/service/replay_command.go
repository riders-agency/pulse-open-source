package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) Replay(ctx context.Context, eventID uuid.UUID, uid uuid.UUID) (string, error) {
	var result0 string
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.WebhookEvents.WithTx(tx)
		var callErr error
		result0, callErr = bound.ReplayTx(ctx, eventID, uid)
		if callErr == nil && c.E != nil {
			uow.AfterCommit(ctx, c.E.Notify)
		}
		return callErr
	})
	return result0, err
}
