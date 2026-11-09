package kernel

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	runtimeretry "riders.agency/pulse/api/internal/runtime/retry"
)

type retryExecutor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func evaluateActionFailure(node Node, attemptNo, maxAttempts int, actionStartedAt time.Time, claim runClaim, final ActionExecution) runtimeretry.Evaluation {
	return runtimeretry.Evaluate(runtimeretry.Failure{
		ActionTypeKey:        node.ActionTypeKey,
		AttemptNo:            attemptNo,
		MaxAttempts:          maxAttempts,
		Retryable:            final.Retryable,
		ErrorType:            final.ErrorType,
		Err:                  final.Err,
		ProviderResponseJSON: final.ProviderResponseJSON,
		StartedAt:            actionStartedAt,
		Now:                  time.Now().UTC(),
		Seed:                 claim.RunID + ":" + node.ActionKey + ":" + final.ErrorType,
		InitialDelay:         node.RetryInitialDelay,
		MaxDelay:             node.RetryMaxDelay,
		MaxElapsed:           node.RetryMaxElapsed,
	})
}

func retryPolicyJSON(eval runtimeretry.Evaluation) string {
	payload := map[string]any{
		"class":             eval.Class,
		"retryable":         eval.Retryable,
		"should_retry":      eval.ShouldRetry,
		"exhausted":         eval.Exhausted,
		"attempt_no":        eval.AttemptNo,
		"max_attempts":      eval.MaxAttempts,
		"next_delay_ms":     eval.NextDelayMS,
		"max_elapsed_ms":    eval.MaxElapsedMS,
		"owner":             eval.Owner,
		"redacted_reason":   eval.RedactedReason,
		"dead_letter_ready": !eval.ShouldRetry,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return `{}`
	}
	return string(raw)
}

func retryEventPayload(eval runtimeretry.Evaluation, providerResponseJSON string) string {
	var provider any = map[string]any{}
	_ = json.Unmarshal([]byte(providerResponseJSON), &provider)
	payload := map[string]any{
		"retry": map[string]any{
			"class":           eval.Class,
			"attempt_no":      eval.AttemptNo,
			"max_attempts":    eval.MaxAttempts,
			"next_delay_ms":   eval.NextDelayMS,
			"max_elapsed_ms":  eval.MaxElapsedMS,
			"owner":           eval.Owner,
			"redacted_reason": eval.RedactedReason,
		},
		"provider_response": provider,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return `{}`
	}
	return string(raw)
}

func recordDeadLetter(ctx context.Context, exec retryExecutor, claim runClaim, actionRunID string, node Node, eval runtimeretry.Evaluation) {
	policyJSON := retryPolicyJSON(eval)
	_, err := exec.Exec(ctx, `
	INSERT INTO workflow_action_dead_letters (
	  workflow_run_id, action_run_id, action_key, action_type_key,
	  retry_class, redacted_reason, owner, replayable, retry_policy_json,
	  created_at, updated_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,true,$8::jsonb,clock_timestamp(),clock_timestamp())
	ON CONFLICT (action_run_id) DO UPDATE SET
	  retry_class = EXCLUDED.retry_class,
	  redacted_reason = EXCLUDED.redacted_reason,
	  owner = EXCLUDED.owner,
	  replayable = EXCLUDED.replayable,
	  retry_policy_json = EXCLUDED.retry_policy_json,
	  updated_at = clock_timestamp()`,
		claim.RunID, actionRunID, node.ActionKey, node.ActionTypeKey,
		string(eval.Class), eval.RedactedReason, eval.Owner, policyJSON)
	if err != nil {
		slog.Warn("workflow action dead-letter write failed",
			slog.String("workflow_run_id", claim.RunID),
			slog.String("action_key", node.ActionKey),
			slog.Any("err", err),
		)
	}
}
