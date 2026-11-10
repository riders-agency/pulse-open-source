package kernel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"riders.agency/pulse/api/internal/runtime/effects"
	runtimeretry "riders.agency/pulse/api/internal/runtime/retry"
)

type effectAttempt struct {
	Enabled  bool
	EffectID string
}

func (k *Kernel) beginEffect(ctx context.Context, claim runClaim, actionRunID string, node Node, attemptNo int, inputJSON string, runInputJSON []byte) (effectAttempt, ActionExecution, bool) {
	if k == nil || k.effectLedger == nil || !effects.RequiresLedger(node.ActionTypeKey) {
		return effectAttempt{}, ActionExecution{}, false
	}
	anchorRunID := effectAnchorRunID(claim, runInputJSON)
	effectID := effects.StableEffectID(anchorRunID, node.ActionKey, attemptNo)
	record, err := k.effectLedger.Start(ctx, effects.StartRequest{
		EffectID:       effectID,
		WorkflowRunID:  claim.RunID,
		ActionRunID:    actionRunID,
		ActionKey:      node.ActionKey,
		ActionTypeKey:  node.ActionTypeKey,
		AttemptNo:      attemptNo,
		IdempotencyKey: effectID,
		InputJSON:      inputJSON,
		NonIdempotent:  true,
	})
	attempt := effectAttempt{Enabled: true, EffectID: effectID}
	if err != nil {
		return attempt, ActionExecution{
			InputJSON:            inputJSON,
			OutputJSON:           "{}",
			ProviderResponseJSON: `{"provider":"effect_ledger","ok":false}`,
			ErrorType:            "effect_ledger",
			Err:                  fmt.Errorf("effect ledger start failed: %w", err),
		}, true
	}
	switch record.Status {
	case effects.StatusSucceeded:
		return attempt, ActionExecution{
			InputJSON:            inputJSON,
			OutputJSON:           record.OutputJSON,
			ProviderResponseJSON: record.ProviderResponseJSON,
		}, true
	case effects.StatusStarted:
		if record.Existing {
			message := "effect outcome is unknown after previous started record; manual reconciliation required"
			_ = k.effectLedger.Unknown(ctx, effectID, message)
			return attempt, unknownEffectExecution(inputJSON, message), true
		}
	case effects.StatusUnknown:
		message := "effect outcome is unknown; manual reconciliation required"
		return attempt, unknownEffectExecution(inputJSON, message), true
	case effects.StatusFailed:
		message := record.Error
		if message == "" {
			message = "previous effect attempt failed"
		}
		errorType := record.ErrorType
		if errorType == "" {
			errorType = "effect_failed"
		}
		return attempt, ActionExecution{
			InputJSON:            inputJSON,
			OutputJSON:           record.OutputJSON,
			ProviderResponseJSON: record.ProviderResponseJSON,
			ErrorType:            errorType,
			Err:                  errors.New(message),
		}, true
	}
	return attempt, ActionExecution{}, false
}

func (k *Kernel) finishEffect(ctx context.Context, attempt effectAttempt, final ActionExecution) ActionExecution {
	if k == nil || k.effectLedger == nil || !attempt.Enabled {
		return final
	}
	if final.Waiting {
		return final
	}
	if final.Err == nil {
		if err := k.effectLedger.Succeeded(ctx, attempt.EffectID, final.OutputJSON, final.ProviderResponseJSON); err != nil {
			slog.Error("effect ledger success update failed", slog.String("effect_id", attempt.EffectID), slog.Any("err", err))
			return unknownEffectExecution(final.InputJSON, "provider outcome needs manual reconciliation after ledger success update failed")
		}
		return final
	}
	if runtimeretry.Classify(final.ErrorType, final.Err, final.Retryable, final.ProviderResponseJSON) == runtimeretry.ClassTimeoutUnknown {
		message := "provider timeout left effect outcome unknown; manual reconciliation required"
		if err := k.effectLedger.Unknown(ctx, attempt.EffectID, message); err != nil {
			slog.Warn("effect ledger unknown update failed", slog.String("effect_id", attempt.EffectID), slog.Any("err", err))
		}
		return unknownEffectExecution(final.InputJSON, message)
	}
	if err := k.effectLedger.Failed(ctx, attempt.EffectID, final.OutputJSON, final.ProviderResponseJSON, final.ErrorType, final.Err.Error()); err != nil {
		slog.Warn("effect ledger failure update failed", slog.String("effect_id", attempt.EffectID), slog.Any("err", err))
	}
	return final
}

func unknownEffectExecution(inputJSON, message string) ActionExecution {
	return ActionExecution{
		InputJSON:            inputJSON,
		OutputJSON:           `{"status":"manual_reconciliation_required"}`,
		ProviderResponseJSON: `{"provider":"effect_ledger","ok":false,"outcome":"unknown"}`,
		ErrorType:            "unknown_outcome",
		Err:                  errors.New(message),
	}
}

func effectAnchorRunID(claim runClaim, runInputJSON []byte) string {
	if value := strings.TrimSpace(claim.ReplayedFromWorkflowRunID); value != "" {
		return value
	}
	var root map[string]any
	if err := json.Unmarshal(runInputJSON, &root); err != nil {
		return claim.RunID
	}
	pulse, _ := root["_pulse"].(map[string]any)
	if pulse == nil {
		return claim.RunID
	}
	for _, key := range []string{"replay_idempotency_anchor_workflow_run_id", "base_workflow_run_id"} {
		if value, ok := pulse[key].(string); ok {
			value = strings.TrimSpace(value)
			if value != "" {
				return value
			}
		}
	}
	return claim.RunID
}
