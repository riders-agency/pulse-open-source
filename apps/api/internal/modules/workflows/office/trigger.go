package office

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	runcommands "riders.agency/pulse/api/internal/modules/runs/commands"
	"riders.agency/pulse/api/internal/platform/uow"
	"riders.agency/pulse/api/internal/runtime/contracts"
	"riders.agency/pulse/api/internal/runtime/runqueue"
	"time"
)

type Trigger struct{ P *pgxpool.Pool }

func (*Trigger) TypeKey() string                                                    { return "mailbox.incoming.trigger" }
func (*Trigger) Sync(context.Context) error                                         { return nil }
func (*Trigger) ClaimDue(context.Context, int) ([]contracts.TriggerEvent, error)    { return nil, nil }
func (*Trigger) MarkEnqueued(context.Context, contracts.TriggerEvent, string) error { return nil }

func (t *Trigger) EnqueueDue(ctx context.Context, limit int) (int, error) {
	if t.P == nil {
		return 0, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 50
	}
	total := 0
	for i := 0; i < limit; i++ {
		polled, queued, err := t.poll(ctx)
		if err != nil {
			return total, err
		}
		if !polled {
			break
		}
		total += queued
	}
	return total, nil
}

func (t *Trigger) poll(ctx context.Context) (bool, int, error) {
	polled, queued := false, 0
	err := uow.Run(ctx, t.P, func(ctx context.Context, tx pgx.Tx) error {
		var pid, wid, vid uuid.UUID
		var definition []byte
		err := tx.QueryRow(ctx, `SELECT o.project_id,o.workflow_id,w.current_workflow_version_id,v.definition_json FROM workflow_office_workspaces o
  JOIN workflows w USING(workflow_id) JOIN workflow_versions v ON v.workflow_version_id=w.current_workflow_version_id
  WHERE o.enabled AND o.next_poll_at<=now() AND w.status='active'
  AND EXISTS(SELECT 1 FROM jsonb_array_elements(v.definition_json->'nodes') n WHERE n->>'action_type_key'='mailbox.incoming.trigger'
  AND COALESCE((n->'config'->>'enabled')::boolean,true)
  AND NOT COALESCE((n->'config'->>'disabled')::boolean,false))
  ORDER BY o.next_poll_at FOR UPDATE OF o SKIP LOCKED LIMIT 1`).Scan(&pid, &wid, &vid, &definition)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		polled = true
		if err := ValidateDefinition(definition); err != nil {
			_, err = tx.Exec(ctx, `UPDATE workflow_office_workspaces SET enabled=false WHERE project_id=$1`, pid)
			return err
		}
		if _, err = tx.Exec(ctx, `UPDATE workflow_office_workspaces SET last_poll_at=now(),
 next_poll_at=next_poll_at+(floor(extract(epoch FROM (now()-next_poll_at))/30)::integer+1)*interval '30 seconds',
 poll_count=poll_count+1 WHERE project_id=$1`, pid); err != nil {
			return err
		}
		var mid uuid.UUID
		var subject, body, sender string
		var amount int64
		err = tx.QueryRow(ctx, `SELECT workflow_office_message_id,subject,body,sender,amount_minor FROM workflow_office_messages
  WHERE project_id=$1 AND workflow_run_id IS NULL ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 1`, pid).Scan(&mid, &subject, &body, &sender, &amount)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		room, err := runcommands.MakeRetentionRoom(ctx, tx, wid, RetainedRuns)
		if err != nil || !room {
			return err
		}
		input, _ := json.Marshal(map[string]any{"mail": map[string]any{"message_id": mid, "subject": subject, "body": body, "sender": sender}, "order": map[string]any{"amount_minor": amount, "currency": "RUB", "customer": "ООО «Демо»"}, "polled_at": time.Now().UTC(), "_pulse": map[string]any{"source_type": "module_trigger"}})
		rid := uuid.New()
		result, _, err := runqueue.RequestRunTx(ctx, tx, runqueue.Request{WorkflowRunID: rid, ProjectID: pid.String(), WorkflowID: wid.String(), WorkflowVersionID: vid.String(), SourceType: runqueue.SourceModuleTrigger, IdempotencyKey: "office-mail:" + mid.String(), InputJSON: input, EventActionKey: "inbox", EventMessage: "Новое письмо получено при опросе почты"})
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE workflow_office_messages SET workflow_run_id=$2 WHERE workflow_office_message_id=$1`, mid, result)
		if err == nil {
			queued = 1
		}
		return err
	})
	return polled, queued, err
}
