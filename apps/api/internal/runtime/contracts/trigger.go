package contracts

import (
	"context"
	"encoding/json"
)

type TriggerEvent struct {
	ProjectID         string
	WorkflowID        string
	WorkflowVersionID string
	SourceType        string
	ActionKey         string
	IdempotencyKey    string
	InputJSON         json.RawMessage
}

type TriggerHandler interface {
	TypeKey() string
	Sync(ctx context.Context) error
	ClaimDue(ctx context.Context, limit int) ([]TriggerEvent, error)
	MarkEnqueued(ctx context.Context, ev TriggerEvent, workflowRunID string) error
}
