package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) Manual(ctx context.Context, wfID uuid.UUID, uid uuid.UUID, input map[string]any) (string, error) {
	var result0 string
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.Runs.WithTx(tx)
		var callErr error
		result0, callErr = bound.ManualTx(ctx, wfID, uid, input)
		if callErr == nil && c.E != nil {
			uow.AfterCommit(ctx, c.E.Notify)
		}
		return callErr
	})
	return result0, err
}
