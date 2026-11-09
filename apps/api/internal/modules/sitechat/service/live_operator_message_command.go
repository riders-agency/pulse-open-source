package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) LiveOperatorMessage(ctx context.Context, conversationID uuid.UUID, body string) (string, error) {
	var result0 string
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.SiteChat.WithTx(tx)
		var callErr error
		result0, callErr = bound.LiveOperatorMessageTx(ctx, conversationID, body)
		return callErr
	})
	return result0, err
}
