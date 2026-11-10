package service

import (
	"context"

	"github.com/jackc/pgx/v5"
	repo "riders.agency/pulse/api/internal/modules/sitechat/repository"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) RecordLiveEvents(ctx context.Context, scope repo.LiveAssistScope, events []repo.EventInput) error {

	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.SiteChat.WithTx(tx)
		return bound.RecordLiveEventsTx(ctx, scope, events)
	})
	return err
}
