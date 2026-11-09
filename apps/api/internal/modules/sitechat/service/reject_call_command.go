package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	repo "riders.agency/pulse/api/internal/modules/sitechat/repository"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) RejectCall(ctx context.Context, callID uuid.UUID, actor string, in repo.CallInput) (*repo.CallSessionRow, error) {
	var result0 *repo.CallSessionRow
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.SiteChat.WithTx(tx)
		var callErr error
		result0, callErr = bound.RejectCallTx(ctx, callID, actor, in)
		return callErr
	})
	return result0, err
}
