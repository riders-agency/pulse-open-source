package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	repo "riders.agency/pulse/api/internal/modules/runs/repository"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) Test(ctx context.Context, wfID uuid.UUID, uid uuid.UUID, req repo.TestRunRequest) (string, error) {
	var result0 string
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.Runs.WithTx(tx)
		var callErr error
		result0, callErr = bound.TestTx(ctx, wfID, uid, req)
		return callErr
	})
	return result0, err
}
