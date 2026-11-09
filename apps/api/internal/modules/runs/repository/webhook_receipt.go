package repository

import (
	"context"

	runscontracts "riders.agency/pulse/api/internal/modules/runs/contracts"
)

var _ runscontracts.RunReceiptReader = (*Runs)(nil)

func (r *Runs) WebhookReceipt(
	ctx context.Context,
	workflowRunID string,
	projectID string,
) (runscontracts.WebhookRunReceipt, error) {
	receipt := runscontracts.WebhookRunReceipt{
		WorkflowRunID: workflowRunID,
		Actions:       []runscontracts.WebhookActionReceipt{},
	}
	if err := r.P.QueryRow(ctx, `
SELECT status::text,
       COALESCE(terminal_reason::text, ''),
       started_at,
       finished_at
  FROM workflow_runs
 WHERE workflow_run_id = $1
   AND project_id = $2`, workflowRunID, projectID).Scan(
		&receipt.Status,
		&receipt.TerminalReason,
		&receipt.StartedAt,
		&receipt.FinishedAt,
	); err != nil {
		return runscontracts.WebhookRunReceipt{}, err
	}

	rows, err := r.P.Query(ctx, `
SELECT action_key::text,
       action_type_key::text,
       status::text,
       attempt_count::int,
       max_attempts::int,
       COALESCE(terminal_reason::text, ''),
       retry_class::text,
       started_at,
       finished_at
  FROM action_runs
 WHERE workflow_run_id = $1
 ORDER BY action_key`, workflowRunID)
	if err != nil {
		return runscontracts.WebhookRunReceipt{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var action runscontracts.WebhookActionReceipt
		if err := rows.Scan(
			&action.ActionKey,
			&action.ActionTypeKey,
			&action.Status,
			&action.AttemptCount,
			&action.MaxAttempts,
			&action.TerminalReason,
			&action.RetryClass,
			&action.StartedAt,
			&action.FinishedAt,
		); err != nil {
			return runscontracts.WebhookRunReceipt{}, err
		}
		receipt.Actions = append(receipt.Actions, action)
	}
	if err := rows.Err(); err != nil {
		return runscontracts.WebhookRunReceipt{}, err
	}
	return receipt, nil
}
