package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	repo "riders.agency/pulse/api/internal/modules/action_dependencies/repository"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) Create(ctx context.Context, pid uuid.UUID, uid uuid.UUID, in repo.DepCreate) (string, error) {
	var result0 string
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.Deps.WithTx(tx)
		var callErr error
		result0, callErr = bound.CreateTx(ctx, pid, uid, in)
		return callErr
	})
	return result0, err
}
