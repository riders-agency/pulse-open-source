package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) Insert(ctx context.Context, cid uuid.UUID, uid uuid.UUID, slug string, name string) error {

	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.Companies.WithTx(tx)
		return bound.InsertTx(ctx, cid, uid, slug, name)
	})
	return err
}
