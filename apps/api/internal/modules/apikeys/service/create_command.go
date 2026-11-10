package service

import (
	"context"
	time "time"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	repo "riders.agency/pulse/api/internal/modules/apikeys/repository"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) Create(ctx context.Context, userID uuid.UUID, name string, grants []repo.GrantInput, expiresAt *time.Time) (*repo.APIKeyRow, string, error) {
	var result0 *repo.APIKeyRow
	var result1 string
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.APIKeys.WithTx(tx)
		var callErr error
		result0, result1, callErr = bound.CreateTx(ctx, userID, name, grants, expiresAt)
		return callErr
	})
	return result0, result1, err
}
