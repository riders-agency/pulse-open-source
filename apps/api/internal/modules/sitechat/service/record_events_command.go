package service

import (
	"context"

	"github.com/jackc/pgx/v5"
	repo "riders.agency/pulse/api/internal/modules/sitechat/repository"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) RecordEvents(ctx context.Context, in repo.EventBatchInput) error {

	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.SiteChat.WithTx(tx)
		return bound.RecordEventsTx(ctx, in)
	})
	return err
}
