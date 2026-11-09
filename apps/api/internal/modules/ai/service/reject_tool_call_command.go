package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) RejectToolCall(ctx context.Context, projectID uuid.UUID, toolCallID uuid.UUID, userID uuid.UUID, reason string) (map[string]any, error) {
	var result0 map[string]any
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.inTransaction(ctx, tx)
		var callErr error
		result0, callErr = bound.RejectToolCallTx(ctx, projectID, toolCallID, userID, reason)
		return callErr
	})
	return result0, err
}
