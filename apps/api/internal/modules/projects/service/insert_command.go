package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	repo "riders.agency/pulse/api/internal/modules/projects/repository"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) Insert(ctx context.Context, companyID uuid.UUID, uid uuid.UUID, name string, slug string) (*repo.ProjectRow, string, error) {
	var result0 *repo.ProjectRow
	var result1 string
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.Projects.WithTx(tx)
		var callErr error
		result0, result1, callErr = bound.InsertTx(ctx, companyID, uid, name, slug)
		return callErr
	})
	return result0, result1, err
}
