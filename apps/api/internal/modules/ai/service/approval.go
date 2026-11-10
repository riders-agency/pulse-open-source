package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/errs"
	"riders.agency/pulse/api/internal/platform/uow"
)

func (c *Commands) ApproveToolCall(ctx context.Context, projectID, toolCallID, userID uuid.UUID) (map[string]any, error) {
	var result map[string]any
	err := uow.Run(ctx, c.P, func(ctx context.Context, tx pgx.Tx) error {
		bound := c.inTransaction(ctx, tx)
		if err := authz.RequireProject(ctx, tx, projectID, authz.PermWrite); err != nil {
			return err
		}
		proposal, err := bound.LockApproval(ctx, tx, projectID, toolCallID, userID)
		if err != nil {
			return err
		}
		if proposal.Status == "succeeded" || proposal.Status == "failed" {
			result, err = bound.GetSession(ctx, projectID, proposal.SessionID, userID)
			return err
		}
		if proposal.Status != "pending_approval" {
			return errs.ErrForbidden
		}
		if err := bound.RequireEmployeeTool(ctx, projectID, proposal.ToolKey); err != nil {
			return err
		}
		var output map[string]any
		executionErr := uow.Savepoint(ctx, func(ctx context.Context, commandTx pgx.Tx) error {
			operation := c.inTransaction(ctx, commandTx)
			var callErr error
			output, callErr = operation.ExecuteApprovedToolTx(ctx, projectID, userID, proposal.ToolKey, proposal.Input)
			return callErr
		})
		result, err = bound.RecordApprovalTx(ctx, tx, projectID, toolCallID, userID, proposal, output, executionErr)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
