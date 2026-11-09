package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	P *pgxpool.Pool
}

func (r *Repository) Enqueue(ctx context.Context, req EnqueueRequest) (Event, error) {
	if r == nil || r.P == nil {
		return Event{}, fmt.Errorf("outbox repository is not configured")
	}
	return enqueue(ctx, r.P, req)
}

func EnqueueTx(ctx context.Context, tx pgx.Tx, req EnqueueRequest) (Event, error) {
	if tx == nil {
		return Event{}, fmt.Errorf("outbox transaction is not configured")
	}
	return enqueue(ctx, tx, req)
}

type enqueueQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func enqueue(ctx context.Context, q enqueueQuerier, req EnqueueRequest) (Event, error) {
	typeKey := strings.TrimSpace(req.TypeKey)
	idempotencyKey := strings.TrimSpace(req.IdempotencyKey)
	if typeKey == "" {
		return Event{}, fmt.Errorf("outbox event type key is required")
	}
	if idempotencyKey == "" {
		return Event{}, fmt.Errorf("outbox idempotency key is required")
	}
	payload := req.PayloadJSON
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}
	normalizedPayload, err := normalizePayload(typeKey, payload, req.SchemaVersion)
	if err != nil {
		return Event{}, err
	}
	payload = normalizedPayload
	availableAt := req.AvailableAt
	if availableAt.IsZero() {
		availableAt = time.Now().UTC()
	}
	maxAttempts := req.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 10
	}
	row := q.QueryRow(ctx, `
INSERT INTO outbox_events (
  event_type, aggregate_type, aggregate_id, idempotency_key, payload_json, available_at, max_attempts
) VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)
ON CONFLICT (idempotency_key) DO UPDATE
  SET updated_at = outbox_events.updated_at
  WHERE outbox_events.event_type = EXCLUDED.event_type
    AND outbox_events.aggregate_type = EXCLUDED.aggregate_type
    AND outbox_events.aggregate_id = EXCLUDED.aggregate_id
    AND outbox_events.payload_json = EXCLUDED.payload_json
    AND outbox_events.max_attempts = EXCLUDED.max_attempts
RETURNING
  outbox_event_id::text, event_type, aggregate_type, aggregate_id, idempotency_key, status,
  payload_json, attempts, max_attempts, available_at, locked_until, last_error, created_at, updated_at
`, typeKey, req.AggregateType, req.AggregateID, idempotencyKey, payload, availableAt, maxAttempts)
	ev, err := scanEvent(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Event{}, fmt.Errorf("outbox idempotency key collision for %q", idempotencyKey)
	}
	return ev, err
}

func (r *Repository) Claim(ctx context.Context, limit int, lease time.Duration) ([]Event, error) {
	if r == nil || r.P == nil {
		return nil, fmt.Errorf("outbox repository is not configured")
	}
	if limit <= 0 {
		limit = 10
	}
	leaseSeconds := int(lease.Seconds())
	if leaseSeconds <= 0 {
		leaseSeconds = 60
	}
	rows, err := r.P.Query(ctx, `
WITH claim AS (
  SELECT outbox_event_id
  FROM outbox_events
  WHERE attempts < max_attempts
    AND (
      (status = 'pending' AND available_at <= now())
      OR (status = 'processing' AND locked_until IS NOT NULL AND locked_until < now())
    )
  ORDER BY available_at ASC, created_at ASC
  FOR UPDATE SKIP LOCKED
  LIMIT $1
)
UPDATE outbox_events e
SET status = 'processing',
    attempts = e.attempts + 1,
    locked_until = now() + ($2 * interval '1 second'),
    updated_at = now()
FROM claim
WHERE e.outbox_event_id = claim.outbox_event_id
RETURNING
  e.outbox_event_id::text, e.event_type, e.aggregate_type, e.aggregate_id, e.idempotency_key, e.status,
  e.payload_json, e.attempts, e.max_attempts, e.available_at, e.locked_until, e.last_error, e.created_at, e.updated_at
`, limit, leaseSeconds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]Event, 0, limit)
	for rows.Next() {
		ev, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

func (r *Repository) MarkProcessed(ctx context.Context, eventID string) error {
	if r == nil || r.P == nil {
		return fmt.Errorf("outbox repository is not configured")
	}
	_, err := r.P.Exec(ctx, `
UPDATE outbox_events
SET status = 'processed',
    locked_until = NULL,
    last_error = '',
    updated_at = now()
WHERE outbox_event_id = $1
`, eventID)
	return err
}

func (r *Repository) MarkFailed(ctx context.Context, eventID string, message string, retryAt time.Time) error {
	if r == nil || r.P == nil {
		return fmt.Errorf("outbox repository is not configured")
	}
	if retryAt.IsZero() {
		retryAt = time.Now().UTC()
	}
	_, err := r.P.Exec(ctx, `
UPDATE outbox_events
SET status = CASE WHEN attempts >= max_attempts THEN 'failed' ELSE 'pending' END,
    locked_until = NULL,
    last_error = $2,
    available_at = CASE WHEN attempts >= max_attempts THEN available_at ELSE $3 END,
    updated_at = now()
WHERE outbox_event_id = $1
`, eventID, message, retryAt)
	return err
}

type eventScanner interface {
	Scan(dest ...any) error
}

func scanEvent(row eventScanner) (Event, error) {
	var ev Event
	var payload []byte
	if err := row.Scan(
		&ev.ID,
		&ev.TypeKey,
		&ev.AggregateType,
		&ev.AggregateID,
		&ev.IdempotencyKey,
		&ev.Status,
		&payload,
		&ev.Attempts,
		&ev.MaxAttempts,
		&ev.AvailableAt,
		&ev.LockedUntil,
		&ev.LastError,
		&ev.CreatedAt,
		&ev.UpdatedAt,
	); err != nil {
		return Event{}, err
	}
	if len(payload) == 0 {
		ev.PayloadJSON = json.RawMessage(`{}`)
	} else {
		ev.PayloadJSON = append(json.RawMessage(nil), payload...)
	}
	return ev, nil
}
