package contracts

import (
	"context"
	"encoding/json"
	"time"
)

type ActionRequest struct {
	WorkflowRunID string
	ProjectID     string
	ActionKey     string
	ActionTypeKey string
	ConfigJSON    json.RawMessage
	RunInputJSON  json.RawMessage
	Steps         map[string]any
	Attempt       int
	Deadline      time.Time
}

type ActionResult struct {
	InputJSON            json.RawMessage
	OutputJSON           json.RawMessage
	ProviderResponseJSON json.RawMessage
	Waiting              bool
	WaitingReason        string
	WaitingDeadline      time.Time
	Retryable            bool
	ErrorType            string
	Err                  error
}

type ActionHandler interface {
	TypeKey() string
	ValidateConfig(ctx context.Context, cfg json.RawMessage) error
	Execute(ctx context.Context, req ActionRequest) ActionResult
}
