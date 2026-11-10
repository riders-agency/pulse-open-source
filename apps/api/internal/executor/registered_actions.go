package executor

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"

	"riders.agency/pulse/api/internal/runtime/contracts"
)

var errBotStateStoreUnavailable = errors.New("bot state store unavailable")

type botSessionStore interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (e *Engine) registerExecutorRuntimeActions() {
	e.Registry.MustRegisterAction(botSetStateAction{store: e.P})
}

type botSetStateAction struct {
	store botSessionStore
}

func (a botSetStateAction) TypeKey() string {
	return "bot.set_state"
}

func (a botSetStateAction) ValidateConfig(context.Context, json.RawMessage) error {
	return nil
}

func (a botSetStateAction) Execute(ctx context.Context, req contracts.ActionRequest) contracts.ActionResult {
	if a.store == nil {
		return contracts.ActionResult{
			OutputJSON:           json.RawMessage(`{}`),
			ProviderResponseJSON: json.RawMessage(`{"provider":"internal","ok":false}`),
			ErrorType:            "storage",
			Err:                  errBotStateStoreUnavailable,
		}
	}
	node := actionPlan{
		ActionKey:     req.ActionKey,
		ActionTypeKey: req.ActionTypeKey,
	}
	_ = json.Unmarshal(req.ConfigJSON, &node.Config)
	final := executeBotSetState(ctx, a.store, node, req.RunInputJSON, req.Steps, req.WorkflowRunID)
	return contracts.ActionResult{
		InputJSON:            json.RawMessage(final.InputJSON),
		OutputJSON:           json.RawMessage(final.OutputJSON),
		ProviderResponseJSON: json.RawMessage(final.ProviderResponseJSON),
		Waiting:              final.Waiting,
		WaitingReason:        final.WaitingReason,
		WaitingDeadline:      final.WaitingDeadline,
		Retryable:            final.Retryable,
		ErrorType:            final.ErrorType,
		Err:                  final.Err,
	}
}
