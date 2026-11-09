package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"riders.agency/pulse/api/internal/runtime/contracts"
	"riders.agency/pulse/api/internal/runtime/registry"
)

var errRuntimeUnavailable = errors.New("client runtime unavailable")

type Runner struct {
	P *pgxpool.Pool
}

type Notifier interface {
	Notify()
}

type PaymentDueTrigger struct {
	P        *pgxpool.Pool
	Notifier Notifier
}

type actionPlan struct {
	ActionKey     string       `json:"action_key"`
	ActionTypeKey string       `json:"action_type_key"`
	Config        actionConfig `json:"config"`
}

type actionConfig struct {
	Enabled         *bool  `json:"enabled"`
	ActiveFrom      string `json:"active_from"`
	ActiveUntil     string `json:"active_until"`
	RemindEveryDays int    `json:"remind_every_days"`
	DaysBeforeDue   int    `json:"days_before_due"`
	Timezone        string `json:"timezone"`
}

type actionExecution struct {
	InputJSON            string
	OutputJSON           string
	ProviderResponseJSON string
	Retryable            bool
	ErrorType            string
	Err                  error
}

type workflowDefinition struct {
	Nodes []actionPlan `json:"nodes"`
}

func RegisterHandlers(reg *registry.Registry, pool *pgxpool.Pool, notifier Notifier) error {
	runner := &Runner{P: pool}
	for _, adapter := range clientActionAdapters {
		if err := reg.RegisterAction(handler{runner: runner, adapter: adapter}); err != nil {
			return err
		}
	}
	return reg.RegisterTrigger(&PaymentDueTrigger{P: pool, Notifier: notifier})
}

type handler struct {
	runner  *Runner
	adapter clientActionAdapter
}

func (h handler) TypeKey() string { return h.adapter.typeKey }

func (h handler) ValidateConfig(context.Context, json.RawMessage) error { return nil }

func (h handler) Execute(ctx context.Context, req contracts.ActionRequest) contracts.ActionResult {
	if h.runner == nil || h.runner.P == nil {
		return contracts.ActionResult{
			InputJSON:            json.RawMessage(actionInput(actionPlan{ActionKey: req.ActionKey, ActionTypeKey: req.ActionTypeKey}, req.RunInputJSON)),
			OutputJSON:           json.RawMessage(`{}`),
			ProviderResponseJSON: json.RawMessage(`{"provider":"clients","ok":false}`),
			ErrorType:            "runtime",
			Err:                  errRuntimeUnavailable,
		}
	}
	if h.adapter.execute == nil {
		err := fmt.Errorf("unsupported client runtime action: %s", h.adapter.typeKey)
		return failedAction(actionInput(actionPlan{ActionKey: req.ActionKey, ActionTypeKey: req.ActionTypeKey}, req.RunInputJSON), "{}", map[string]any{
			"provider": "clients",
			"ok":       false,
			"error":    err.Error(),
		}, false, "missing_handler", err).contractResult()
	}
	return h.adapter.execute(ctx, h.runner, req)
}

type clientActionAdapter struct {
	typeKey string
	execute func(context.Context, *Runner, contracts.ActionRequest) contracts.ActionResult
}

type clientExecutor func(context.Context, actionPlan, []byte) actionExecution

func (r *Runner) executeClientAction(ctx context.Context, req contracts.ActionRequest, execute clientExecutor) contracts.ActionResult {
	node := actionPlan{ActionKey: req.ActionKey, ActionTypeKey: req.ActionTypeKey}
	_ = json.Unmarshal(req.ConfigJSON, &node.Config)
	return execute(ctx, node, req.RunInputJSON).contractResult()
}

func (final actionExecution) contractResult() contracts.ActionResult {
	return contracts.ActionResult{
		InputJSON:            json.RawMessage(final.InputJSON),
		OutputJSON:           json.RawMessage(final.OutputJSON),
		ProviderResponseJSON: json.RawMessage(final.ProviderResponseJSON),
		Retryable:            final.Retryable,
		ErrorType:            final.ErrorType,
		Err:                  final.Err,
	}
}
