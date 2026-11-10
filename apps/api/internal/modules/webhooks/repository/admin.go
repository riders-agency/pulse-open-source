package repository

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/base64"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	dbpkg "riders.agency/pulse/api/internal/platform/db"

	"riders.agency/pulse/api/internal/executor"

	"riders.agency/pulse/api/internal/platform/authz"
	cryptoapp "riders.agency/pulse/api/internal/platform/crypto"
	"riders.agency/pulse/api/internal/platform/errs"
	"riders.agency/pulse/api/internal/runtime/runqueue"
)

func (w *Webhooks) assertEndpoint(ctx context.Context, weID, uid uuid.UUID) (projectID string, err error) {
	if err := authz.RequireWebhookEndpoint(ctx, w.P, weID, authz.PermRead); err != nil {
		return "", err
	}
	row := w.P.QueryRow(ctx, `
SELECT pr.project_id::text
  FROM webhook_endpoints we
  JOIN pulse_read.projects_v1 pr ON pr.project_id = we.project_id
  JOIN pulse_read.company_memberships_v1 m ON m.company_id = pr.company_id AND m.user_id = $2
 WHERE we.webhook_endpoint_id = $1`, weID.String(), uid.String())
	if err := row.Scan(&projectID); err != nil {
		if err == pgx.ErrNoRows {
			return "", errs.ErrForbidden
		}
		return "", err
	}
	return projectID, nil
}

func (w *Webhooks) GetEndpointDetail(ctx context.Context, weID, uid uuid.UUID) (map[string]any, error) {
	if _, err := w.assertEndpoint(ctx, weID, uid); err != nil {
		return nil, err
	}
	row := w.P.QueryRow(ctx, `
SELECT webhook_endpoint_id::text, workflow_id::text, workflow_version_id::text,
       name::text, endpoint_key::text, status::text, auth_type::text
  FROM webhook_endpoints WHERE webhook_endpoint_id = $1`, weID.String())
	var id, wf, wv, nm, ek, st, at string
	if err := row.Scan(&id, &wf, &wv, &nm, &ek, &st, &at); err != nil {
		return nil, err
	}
	return map[string]any{
		"webhook_endpoint_id": id,
		"workflow_id":         wf,
		"workflow_version_id": wv,
		"name":                nm,
		"endpoint_key":        ek,
		"status":              st,
		"auth_type":           at,
	}, nil
}

type EndpointPatch struct {
	Name   *string `json:"name"`
	Status *string `json:"status"`
}

func optStr(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

func (w *Webhooks) PatchEndpoint(ctx context.Context, weID, uid uuid.UUID, patch EndpointPatch) error {
	if _, err := w.assertEndpoint(ctx, weID, uid); err != nil {
		return err
	}
	_, err := w.P.Exec(ctx, `
UPDATE webhook_endpoints SET
  name = COALESCE($2, name),
  status = COALESCE($3, status),
  updated_at = now()
 WHERE webhook_endpoint_id = $1`,
		weID.String(), optStr(patch.Name), optStr(patch.Status))
	return err
}

func (w *Webhooks) RotateEndpointSecret(ctx context.Context, weID, uid uuid.UUID) (plain string, err error) {
	if _, err := w.assertEndpoint(ctx, weID, uid); err != nil {
		return "", err
	}
	raw := make([]byte, 24)
	if _, err := cryptorand.Read(raw); err != nil {
		return "", err
	}
	plain = base64.RawURLEncoding.EncodeToString(raw)
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	ciphertext, err := cryptoapp.Seal([]byte(plain), w.MasterHEX)
	if err != nil {
		return "", err
	}
	if _, err := w.P.Exec(ctx, `
UPDATE webhook_endpoints
   SET previous_secret_hash = secret_hash,
       previous_secret_ciphertext = secret_ciphertext,
       previous_secret_expires_at = now() + interval '15 minutes',
       secret_hash = $2,
       secret_ciphertext = $3,
       updated_at = now()
 WHERE webhook_endpoint_id = $1`,
		weID.String(), string(h), ciphertext); err != nil {
		return "", err
	}
	return plain, nil
}

func (w *Webhooks) ListWebhookEvents(ctx context.Context, weID, uid uuid.UUID) ([]map[string]any, error) {
	if _, err := w.assertEndpoint(ctx, weID, uid); err != nil {
		return nil, err
	}
	rows, err := w.P.Query(ctx, `
SELECT webhook_event_id::text, status::text, method::text, path::text, received_at
  FROM pulse_read.webhook_events_v1 WHERE webhook_endpoint_id = $1
 ORDER BY received_at DESC LIMIT 200`, weID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, st, meth, pt string
		var rat any
		if err := rows.Scan(&id, &st, &meth, &pt, &rat); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"webhook_event_id": id,
			"status":           st,
			"method":           meth,
			"path":             pt,
			"received_at":      rat,
		})
	}
	return out, rows.Err()
}

// WebhookEvents loads and replays stored webhook payloads (tenant-scoped).
type WebhookEvents struct {
	P dbpkg.DB
	E *executor.Engine
}

func (ev *WebhookEvents) Get(ctx context.Context, eventID, uid uuid.UUID) (map[string]any, error) {
	row := ev.P.QueryRow(ctx, `
SELECT we.webhook_event_id::text, we.status::text, we.method::text, we.path::text,
       we.payload_json, we.workflow_run_id::text, pr.project_id::text
  FROM pulse_read.webhook_events_v1 we
  JOIN pulse_read.projects_v1 pr ON pr.project_id = we.project_id
  JOIN pulse_read.company_memberships_v1 m ON m.company_id = pr.company_id AND m.user_id = $2
 WHERE we.webhook_event_id = $1`, eventID.String(), uid.String())
	var id, st, meth, pt string
	var pj []byte
	var wr *string
	var pid string
	if err := row.Scan(&id, &st, &meth, &pt, &pj, &wr, &pid); err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrForbidden
		}
		return nil, err
	}
	m := map[string]any{
		"webhook_event_id": id,
		"status":           st,
		"method":           meth,
		"path":             pt,
		"project_id":       pid,
	}
	if len(pj) > 0 {
		var payload any
		_ = json.Unmarshal(pj, &payload)
		m["payload"] = payload
	}
	if wr != nil && *wr != "" {
		m["workflow_run_id"] = *wr
	}
	return m, nil
}

func (ev *WebhookEvents) ReplayTx(ctx context.Context, eventID, uid uuid.UUID) (runID string, err error) {
	row := ev.P.QueryRow(ctx, `
SELECT we.project_id::text, we.workflow_id::text, we.workflow_version_id::text,
       COALESCE(we.payload_json::text, '{}'), we.method::text, ep.endpoint_key::text
  FROM pulse_read.webhook_events_v1 we
  JOIN webhook_endpoints ep ON ep.webhook_endpoint_id = we.webhook_endpoint_id
  JOIN pulse_read.projects_v1 pr ON pr.project_id = we.project_id
  JOIN pulse_read.company_memberships_v1 m ON m.company_id = pr.company_id AND m.user_id = $2
 WHERE we.webhook_event_id = $1`, eventID.String(), uid.String())
	var pid, wf, wv, payloadStr, meth, endpointKey string
	if err := row.Scan(&pid, &wf, &wv, &payloadStr, &meth, &endpointKey); err != nil {
		if err == pgx.ErrNoRows {
			return "", errs.ErrForbidden
		}
		return "", err
	}
	var payload any
	_ = json.Unmarshal([]byte(payloadStr), &payload)
	rid := uuid.New()
	in := map[string]any{
		"replay_of_webhook_event": eventID.String(),
		"method":                  meth,
		"payload":                 payload,
		"_pulse": map[string]any{
			"source_type":          "replay",
			"webhook_slug":         endpointKey,
			"webhook_endpoint_key": endpointKey,
		},
	}
	inb, _ := json.Marshal(in)
	tx, err := ev.transaction()
	if err != nil {
		return "", err
	}
	if _, _, err := runqueue.RequestRunTx(ctx, tx, runqueue.Request{
		WorkflowRunID:        rid,
		ProjectID:            pid,
		WorkflowID:           wf,
		WorkflowVersionID:    wv,
		SourceType:           "replay",
		SourceWebhookEventID: eventID.String(),
		InputJSON:            json.RawMessage(inb),
		EventMessage:         "replay from pulse_read.webhook_events_v1",
	}); err != nil {
		return "", err
	}
	return rid.String(), nil
}
