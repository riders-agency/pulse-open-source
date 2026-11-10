package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	dbpkg "riders.agency/pulse/api/internal/platform/db"
	"riders.agency/pulse/api/internal/platform/uow"
	"riders.agency/pulse/api/internal/runtime/runqueue"
)

type WebhookInput struct {
	ProjectID, EndpointID, WorkflowID, VersionID                                                    string
	Method, Path, ContentType, UserAgent, IdempotencyKey, EventID, BodySHA, AuthScheme, EndpointKey string
	Timestamp                                                                                       time.Time
	SignatureVerified                                                                               bool
	Payload                                                                                         json.RawMessage
}

var errDuplicate = errors.New("duplicate webhook")

func Receive(ctx context.Context, database dbpkg.DB, input WebhookInput) (map[string]string, error) {
	eventID, runID := uuid.New(), uuid.New()
	err := uow.Run(ctx, database, func(ctx context.Context, tx pgx.Tx) error {
		var inserted bool
		err := tx.QueryRow(ctx, `INSERT INTO webhook_events (
 webhook_event_id,project_id,webhook_endpoint_id,workflow_id,workflow_version_id,method,path,payload_json,raw_body,content_type,source_ip,user_agent,status,idempotency_key,event_id,received_timestamp,body_sha256,auth_scheme,signature_verified)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,NULL::bytea,$9,NULL::inet,$10,'received',$11,$12,$13,$14,$15,$16)
 ON CONFLICT DO NOTHING RETURNING true`, eventID, input.ProjectID, input.EndpointID, input.WorkflowID, input.VersionID, input.Method, input.Path, input.Payload, input.ContentType, input.UserAgent, nullableString(input.IdempotencyKey), nullableString(input.EventID), nullableTime(input.Timestamp), input.BodySHA, input.AuthScheme, input.SignatureVerified).Scan(&inserted)
		if errors.Is(err, pgx.ErrNoRows) {
			return errDuplicate
		}
		if err != nil {
			return err
		}
		body, err := json.Marshal(map[string]any{"webhook": input.Payload, "_pulse": map[string]any{"source_type": "webhook", "webhook_slug": input.EndpointKey, "webhook_endpoint_key": input.EndpointKey, "webhook_endpoint_id": input.EndpointID, "webhook_event_id": eventID.String(), "webhook_path": input.Path}})
		if err != nil {
			return err
		}
		if _, _, err := runqueue.RequestRunTx(ctx, tx, runqueue.Request{WorkflowRunID: runID, ProjectID: input.ProjectID, WorkflowID: input.WorkflowID, WorkflowVersionID: input.VersionID, SourceType: "webhook", SourceWebhookEventID: eventID.String(), IdempotencyKey: "webhook:" + input.EndpointID + ":" + input.IdempotencyKey, InputJSON: body, EventMessage: "webhook ingest queued"}); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE webhook_events SET workflow_run_id=$1,status='queued',queued_at=now() WHERE webhook_event_id=$2`, runID, eventID)
		return err
	})
	if errors.Is(err, errDuplicate) {
		var existing string
		err = database.QueryRow(ctx, `SELECT COALESCE(workflow_run_id::text,'') FROM webhook_events WHERE webhook_endpoint_id=$1 AND (idempotency_key=$2 OR ($3<>'' AND event_id=$3)) ORDER BY received_at DESC LIMIT 1`, input.EndpointID, input.IdempotencyKey, input.EventID).Scan(&existing)
		if err != nil {
			return nil, err
		}
		result := map[string]string{"duplicate": "true"}
		if existing != "" {
			result["workflow_run_id"] = existing
		}
		return result, nil
	}
	if err != nil {
		return nil, err
	}
	return map[string]string{"workflow_run_id": runID.String(), "webhook_event_id": eventID.String()}, nil
}
func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}
