package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) OperatorMessage(ctx context.Context, conversationID uuid.UUID, uid uuid.UUID, body string, attachmentIDs []string) (string, error) {
	var result0 string
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.SiteChat.WithTx(tx)
		var callErr error
		result0, callErr = bound.OperatorMessageTx(ctx, conversationID, uid, body, attachmentIDs)
		return callErr
	})
	return result0, err
}
