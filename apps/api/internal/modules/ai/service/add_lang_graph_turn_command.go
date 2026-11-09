package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	repo "riders.agency/pulse/api/internal/modules/ai/repository"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) AddLangGraphTurn(ctx context.Context, projectID uuid.UUID, sessionID uuid.UUID, userID uuid.UUID, in repo.MessageInput, reply string, payload map[string]any) (map[string]any, error) {
	var result0 map[string]any
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.inTransaction(ctx, tx)
		var callErr error
		result0, callErr = bound.AddLangGraphTurnTx(ctx, projectID, sessionID, userID, in, reply, payload)
		return callErr
	})
	return result0, err
}
