package service

import (
	"context"

	"github.com/jackc/pgx/v5"
	repo "riders.agency/pulse/api/internal/modules/sitechat/repository"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) PublicCallRequest(ctx context.Context, in repo.CallInput) (*repo.CallSessionRow, error) {
	var result0 *repo.CallSessionRow
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.SiteChat.WithTx(tx)
		var callErr error
		result0, callErr = bound.PublicCallRequestTx(ctx, in)
		return callErr
	})
	return result0, err
}
