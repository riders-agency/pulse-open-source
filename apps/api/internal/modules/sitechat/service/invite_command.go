package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	repo "riders.agency/pulse/api/internal/modules/sitechat/repository"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) Invite(ctx context.Context, sessionID uuid.UUID, uid uuid.UUID) (*repo.ConversationRow, error) {
	var result0 *repo.ConversationRow
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.SiteChat.WithTx(tx)
		var callErr error
		result0, callErr = bound.InviteTx(ctx, sessionID, uid)
		return callErr
	})
	return result0, err
}
