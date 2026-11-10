package service

import (
	"context"

	"github.com/jackc/pgx/v5"
	repo "riders.agency/pulse/api/internal/modules/sitechat/repository"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) RecordPresence(ctx context.Context, in repo.PresenceInput) (*repo.PresenceResponse, error) {
	var result0 *repo.PresenceResponse
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.SiteChat.WithTx(tx)
		var callErr error
		result0, callErr = bound.RecordPresenceTx(ctx, in)
		return callErr
	})
	return result0, err
}
