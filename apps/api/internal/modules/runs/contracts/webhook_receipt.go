package contracts

import (
	"context"
	"time"
)

// RunReceiptReader is the public read boundary exposed by the runs module to
// authenticated transports owned by other modules.
type RunReceiptReader interface {
	WebhookReceipt(ctx context.Context, workflowRunID, projectID string) (WebhookRunReceipt, error)
}

type WebhookRunReceipt struct {
	WorkflowRunID  string                 `json:"workflow_run_id"`
	Status         string                 `json:"status"`
	Terminal       bool                   `json:"terminal"`
	TerminalReason string                 `json:"terminal_reason,omitempty"`
	StartedAt      *time.Time             `json:"started_at,omitempty"`
	FinishedAt     *time.Time             `json:"finished_at,omitempty"`
	Actions        []WebhookActionReceipt `json:"actions"`
}

type WebhookActionReceipt struct {
	ActionKey      string     `json:"action_key"`
	ActionTypeKey  string     `json:"action_type_key"`
	Status         string     `json:"status"`
	AttemptCount   int        `json:"attempt_count"`
	MaxAttempts    *int       `json:"max_attempts,omitempty"`
	TerminalReason string     `json:"terminal_reason,omitempty"`
	RetryClass     *string    `json:"retry_class,omitempty"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
}
