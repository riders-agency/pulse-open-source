package service

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	repo "riders.agency/pulse/api/internal/modules/clients/repository"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) RecordManualPayment(ctx context.Context, clientID uuid.UUID, uid uuid.UUID, in repo.ManualPaymentRecordInput) (*repo.ManualPaymentRecordResult, error) {
	var result0 *repo.ManualPaymentRecordResult
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.Clients.WithTx(tx)
		var callErr error
		result0, callErr = bound.RecordManualPaymentTx(ctx, clientID, uid, in)
		return callErr
	})
	return result0, err
}
