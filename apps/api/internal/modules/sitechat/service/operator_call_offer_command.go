package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	repo "riders.agency/pulse/api/internal/modules/sitechat/repository"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) OperatorCallOffer(ctx context.Context, sessionID uuid.UUID, uid uuid.UUID, in repo.CallInput) (*repo.CallSessionRow, error) {
	var result0 *repo.CallSessionRow
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.SiteChat.WithTx(tx)
		var callErr error
		result0, callErr = bound.OperatorCallOfferTx(ctx, sessionID, uid, in)
		return callErr
	})
	return result0, err
}
