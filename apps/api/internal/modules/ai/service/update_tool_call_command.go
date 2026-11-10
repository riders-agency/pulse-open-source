package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) UpdateToolCall(ctx context.Context, projectID uuid.UUID, toolCallID uuid.UUID, userID uuid.UUID, input map[string]any, reason string) (map[string]any, error) {
	var result0 map[string]any
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.inTransaction(ctx, tx)
		var callErr error
		result0, callErr = bound.UpdateToolCallTx(ctx, projectID, toolCallID, userID, input, reason)
		return callErr
	})
	return result0, err
}
