package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	repo "riders.agency/pulse/api/internal/modules/clients/repository"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) CreateManualPaymentAction(ctx context.Context, projectID uuid.UUID, uid uuid.UUID, in repo.CreateManualPaymentActionInput) (string, error) {
	var result0 string
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.Clients.WithTx(tx)
		var callErr error
		result0, callErr = bound.CreateManualPaymentActionTx(ctx, projectID, uid, in)
		return callErr
	})
	return result0, err
}
