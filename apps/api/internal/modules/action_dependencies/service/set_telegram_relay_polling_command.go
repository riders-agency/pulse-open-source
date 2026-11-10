package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	repo "riders.agency/pulse/api/internal/modules/action_dependencies/repository"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) SetTelegramRelayPolling(ctx context.Context, pid uuid.UUID, depID uuid.UUID, uid uuid.UUID, patch repo.TelegramRelayPollingPatch) (map[string]any, error) {
	var result0 map[string]any
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.Deps.WithTx(tx)
		var callErr error
		result0, callErr = bound.SetTelegramRelayPollingTx(ctx, pid, depID, uid, patch)
		return callErr
	})
	return result0, err
}
