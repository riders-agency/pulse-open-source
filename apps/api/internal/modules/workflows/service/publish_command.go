package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) Publish(ctx context.Context, wfID uuid.UUID, uid uuid.UUID, description string) (string, int, error) {
	var result0 string
	var result1 int
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.Workflows.WithTx(tx)
		var callErr error
		result0, result1, callErr = bound.PublishTx(ctx, wfID, uid, description)
		return callErr
	})
	return result0, result1, err
}
