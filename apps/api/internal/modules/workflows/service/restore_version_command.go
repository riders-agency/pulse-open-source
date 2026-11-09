package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) RestoreVersion(ctx context.Context, wfID uuid.UUID, wvID uuid.UUID, uid uuid.UUID) (map[string]any, error) {
	var result0 map[string]any
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.Workflows.WithTx(tx)
		var callErr error
		result0, callErr = bound.RestoreVersionTx(ctx, wfID, wvID, uid)
		return callErr
	})
	return result0, err
}
