package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	yclientsprovider "riders.agency/pulse/api/internal/modules/providers/yclients"
)

type yclientsOperation struct {
	ID             string
	Status         string
	ResultJSON     []byte
	ErrorCode      string
	PossibleEffect bool
	Retryable      bool
	Replay         bool
	InProgress     bool
}

func yclientsCommandFingerprint(command yclientsCommand) (string, []byte, error) {
	raw, err := json.Marshal(command)
	if err != nil {
		return "", nil, err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), raw, nil
}

func (r *Runner) reserveYClientsOperation(ctx context.Context, dependencyID string, command yclientsCommand) (yclientsOperation, []byte, error) {
	fingerprint, requestJSON, err := yclientsCommandFingerprint(command)
	if err != nil {
		return yclientsOperation{}, nil, err
	}
	tx, err := r.P.Begin(ctx)
	if err != nil {
		return yclientsOperation{}, nil, err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `
INSERT INTO provider_operation_effects (
  action_dependency_id, provider, operation_key, request_sha256, status, request_json
) VALUES ($1, 'yclients', $2, $3, 'planned', $4::jsonb)
ON CONFLICT (action_dependency_id, provider, operation_key) DO NOTHING`,
		dependencyID, command.OperationKey, fingerprint, string(requestJSON))
	if err != nil {
		return yclientsOperation{}, nil, err
	}
	var operation yclientsOperation
	var storedFingerprint string
	err = tx.QueryRow(ctx, `
SELECT provider_operation_effect_id::text, status, request_sha256, result_json,
       last_error_code, possible_effect, retryable
FROM provider_operation_effects
WHERE action_dependency_id = $1 AND provider = 'yclients' AND operation_key = $2
FOR UPDATE`, dependencyID, command.OperationKey).Scan(
		&operation.ID, &operation.Status, &storedFingerprint, &operation.ResultJSON,
		&operation.ErrorCode, &operation.PossibleEffect, &operation.Retryable,
	)
	if err != nil {
		return yclientsOperation{}, nil, err
	}
	if storedFingerprint != fingerprint {
		return yclientsOperation{}, nil, fmt.Errorf("YClients operation key conflicts with another payload")
	}
	switch operation.Status {
	case "succeeded", "unknown":
		operation.Replay = true
	case "sending":
		operation.InProgress = true
	case "failed":
		if operation.PossibleEffect || !operation.Retryable {
			operation.Replay = true
		} else {
			operation.Status = "sending"
		}
	case "planned":
		operation.Status = "sending"
	default:
		return yclientsOperation{}, nil, fmt.Errorf("unsupported YClients operation status %q", operation.Status)
	}
	if operation.Status == "sending" && !operation.InProgress {
		_, err = tx.Exec(ctx, `
UPDATE provider_operation_effects
SET status = 'sending', attempts = attempts + 1, possible_effect = false,
    retryable = false, last_error_code = '', updated_at = now()
WHERE provider_operation_effect_id = $1`, operation.ID)
		if err != nil {
			return yclientsOperation{}, nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return yclientsOperation{}, nil, err
	}
	return operation, requestJSON, nil
}

func (r *Runner) finishYClientsOperation(ctx context.Context, operationID string, result yclientsprovider.CertificateIssueResult) error {
	raw, err := json.Marshal(result)
	if err != nil {
		return err
	}
	tag, err := r.P.Exec(ctx, `
UPDATE provider_operation_effects
SET status = 'succeeded', result_json = $2::jsonb, possible_effect = true,
    retryable = false, last_error_code = '', finished_at = now(), updated_at = now()
WHERE provider_operation_effect_id = $1 AND status = 'sending'`, operationID, string(raw))
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("YClients operation state changed before success was persisted")
	}
	return nil
}

func (r *Runner) failYClientsOperation(ctx context.Context, operationID string, providerErr error) error {
	status := "failed"
	possible := yclientsprovider.CertificatePossibleEffect(providerErr)
	if possible {
		status = "unknown"
	}
	_, err := r.P.Exec(ctx, `
UPDATE provider_operation_effects
SET status = $2, possible_effect = $3, retryable = $4,
    last_error_code = $5, finished_at = now(), updated_at = now()
WHERE provider_operation_effect_id = $1 AND status = 'sending'`,
		operationID, status, possible, yclientsprovider.CertificateRetryable(providerErr),
		yclientsprovider.CertificateErrorCode(providerErr))
	return err
}

func decodeYClientsResult(raw []byte) (yclientsprovider.CertificateIssueResult, error) {
	var result yclientsprovider.CertificateIssueResult
	if len(raw) == 0 || json.Unmarshal(raw, &result) != nil {
		return result, errors.New("persisted YClients result is invalid")
	}
	return result, nil
}

func yclientsReplayError(operation yclientsOperation) error {
	code := strings.TrimSpace(operation.ErrorCode)
	if code == "" {
		code = "operation_failed"
	}
	return &yclientsprovider.CertificateAPIError{
		Code: code, PossibleEffect: operation.PossibleEffect,
		Retryable: operation.Retryable && !operation.PossibleEffect,
	}
}
