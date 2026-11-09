package effects

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	StatusStarted   = "started"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
	StatusUnknown   = "unknown"
)

type Ledger struct {
	P *pgxpool.Pool
}

type StartRequest struct {
	EffectID               string
	WorkflowRunID          string
	ActionRunID            string
	ActionKey              string
	ActionTypeKey          string
	AttemptNo              int
	IdempotencyKey         string
	ProviderIdempotencyKey string
	InputJSON              string
	NonIdempotent          bool
}

type Record struct {
	EffectID             string
	Status               string
	Existing             bool
	OutputJSON           string
	ProviderResponseJSON string
	ErrorType            string
	Error                string
}

func RequiresLedger(actionTypeKey string) bool {
	actionTypeKey = strings.TrimSpace(actionTypeKey)
	switch {
	case strings.HasPrefix(actionTypeKey, "telegram."):
		return true
	case strings.HasPrefix(actionTypeKey, "max."):
		return true
	case actionTypeKey == "email.send":
		return true
	case actionTypeKey == "http.request":
		return true
	case actionTypeKey == "uon.create_lead":
		return true
	case actionTypeKey == "yclients.certificate.issue":
		return true
	case strings.HasPrefix(actionTypeKey, "ai."):
		return true
	case strings.HasPrefix(actionTypeKey, "clients."):
		return true
	default:
		return false
	}
}

func StableEffectID(workflowRunID, actionKey string, attemptNo int) string {
	seed := strings.Join([]string{
		strings.TrimSpace(workflowRunID),
		strings.TrimSpace(actionKey),
		fmt.Sprintf("%d", attemptNo),
	}, ":")
	sum := sha256.Sum256([]byte(seed))
	return "effect:" + hex.EncodeToString(sum[:])
}

func (l *Ledger) Start(ctx context.Context, req StartRequest) (Record, error) {
	if l == nil || l.P == nil {
		return Record{}, fmt.Errorf("effect ledger is not configured")
	}
	req = normalizeStartRequest(req)
	var rec Record
	var output, provider []byte
	err := l.P.QueryRow(ctx, `
WITH inserted AS (
  INSERT INTO effect_ledger (
    effect_id, workflow_run_id, action_run_id, action_key, action_type_key,
    attempt_no, idempotency_key, provider_idempotency_key, input_fingerprint,
    status, non_idempotent)
  VALUES ($1,$2::uuid,NULLIF($3, '')::uuid,$4,$5,$6,$7,$8,$9,'started',$10)
  ON CONFLICT DO NOTHING
  RETURNING effect_id, status, false AS existing, output_json, provider_response_json, error_type, error
)
SELECT effect_id, status, existing, output_json, provider_response_json, error_type, error
  FROM inserted
UNION ALL
SELECT effect_id, status, true AS existing, output_json, provider_response_json, error_type, error
  FROM effect_ledger
 WHERE effect_id = $1
   AND NOT EXISTS (SELECT 1 FROM inserted)
LIMIT 1
`, req.EffectID, req.WorkflowRunID, req.ActionRunID, req.ActionKey, req.ActionTypeKey,
		req.AttemptNo, req.IdempotencyKey, req.ProviderIdempotencyKey, fingerprint(req.InputJSON),
		req.NonIdempotent).Scan(&rec.EffectID, &rec.Status, &rec.Existing, &output, &provider, &rec.ErrorType, &rec.Error)
	if err != nil {
		return Record{}, err
	}
	rec.OutputJSON = rawOrDefault(output)
	rec.ProviderResponseJSON = rawOrDefault(provider)
	return rec, nil
}

func (l *Ledger) Succeeded(ctx context.Context, effectID, outputJSON, providerResponseJSON string) error {
	return l.finish(ctx, effectID, StatusSucceeded, outputJSON, providerResponseJSON, "", "")
}

func (l *Ledger) Failed(ctx context.Context, effectID, outputJSON, providerResponseJSON, errorType, message string) error {
	return l.finish(ctx, effectID, StatusFailed, outputJSON, providerResponseJSON, errorType, message)
}

func (l *Ledger) Unknown(ctx context.Context, effectID, message string) error {
	return l.finish(ctx, effectID, StatusUnknown, "{}", "{}", "unknown_outcome", message)
}

func (l *Ledger) finish(ctx context.Context, effectID, status, outputJSON, providerResponseJSON, errorType, message string) error {
	if l == nil || l.P == nil {
		return fmt.Errorf("effect ledger is not configured")
	}
	outputJSON = validOrDefault(outputJSON)
	providerResponseJSON = validOrDefault(providerResponseJSON)
	tag, err := l.P.Exec(ctx, `
UPDATE effect_ledger
   SET status = $2,
       output_json = $3::jsonb,
       provider_response_json = $4::jsonb,
       response_fingerprint = $5,
       error_type = $6,
       error = $7,
       finished_at = now(),
       updated_at = now()
 WHERE effect_id = $1
   AND status = 'started'`,
		strings.TrimSpace(effectID), status, outputJSON, providerResponseJSON,
		fingerprint(outputJSON+"|"+providerResponseJSON), errorType, message)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		var current string
		if scanErr := l.P.QueryRow(ctx, `SELECT status FROM effect_ledger WHERE effect_id = $1`, strings.TrimSpace(effectID)).Scan(&current); scanErr != nil {
			if scanErr == pgx.ErrNoRows {
				return fmt.Errorf("effect ledger record not found")
			}
			return scanErr
		}
		if current != status {
			return fmt.Errorf("effect ledger status is %s, expected started", current)
		}
	}
	return nil
}

func normalizeStartRequest(req StartRequest) StartRequest {
	req.EffectID = strings.TrimSpace(req.EffectID)
	req.WorkflowRunID = strings.TrimSpace(req.WorkflowRunID)
	req.ActionRunID = strings.TrimSpace(req.ActionRunID)
	req.ActionKey = strings.TrimSpace(req.ActionKey)
	req.ActionTypeKey = strings.TrimSpace(req.ActionTypeKey)
	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	if req.IdempotencyKey == "" {
		req.IdempotencyKey = req.EffectID
	}
	req.ProviderIdempotencyKey = strings.TrimSpace(req.ProviderIdempotencyKey)
	if req.ProviderIdempotencyKey == "" {
		req.ProviderIdempotencyKey = "pulse:" + strings.TrimPrefix(req.EffectID, "effect:")
	}
	if req.AttemptNo < 1 {
		req.AttemptNo = 1
	}
	req.InputJSON = validOrDefault(req.InputJSON)
	return req
}

func validOrDefault(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || !json.Valid([]byte(raw)) {
		return "{}"
	}
	return raw
}

func rawOrDefault(raw []byte) string {
	if len(raw) == 0 {
		return "{}"
	}
	return validOrDefault(string(raw))
}

func fingerprint(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
