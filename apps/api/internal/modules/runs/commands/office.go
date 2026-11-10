package commands

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrNotWaiting = errors.New("run_not_waiting_for_decision")

func LockWaiting(ctx context.Context, tx pgx.Tx, projectID, runID uuid.UUID, reason string) error {
	var status, actual string
	err := tx.QueryRow(ctx, `SELECT status,waiting_reason FROM workflow_runs WHERE workflow_run_id=$1 AND project_id=$2 FOR UPDATE`, runID, projectID).Scan(&status, &actual)
	if err != nil {
		return err
	}
	if status != "waiting" || actual != reason {
		return ErrNotWaiting
	}
	return nil
}

func ResumeAfterDecision(ctx context.Context, tx pgx.Tx, runID, userID uuid.UUID, approved bool) error {
	if _, err := tx.Exec(ctx, `UPDATE workflow_runs SET status='pending',waiting_reason='',waiting_deadline_at=NULL,wake_reason='human_decision',last_state_transition_at=now() WHERE workflow_run_id=$1 AND status='waiting'`, runID); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `INSERT INTO run_events(workflow_run_id,action_key,event_type,message,payload_json) VALUES($1,'approval','office.decision','Человек принял решение',jsonb_build_object('approved',$2::boolean,'user_id',$3::text))`, runID, approved, userID.String())
	return err
}

// Called under the caller's workflow lock, before enqueueing a new run.
// Backpressure preserves unresolved human decisions while enforcing a hard cap.
func MakeRetentionRoom(ctx context.Context, tx pgx.Tx, workflowID uuid.UUID, limit int) (bool, error) {
	if limit < 1 {
		return false, errors.New("invalid_retention_limit")
	}
	var count int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM workflow_runs WHERE workflow_id=$1`, workflowID).Scan(&count); err != nil {
		return false, err
	}
	for count >= limit {
		var runID uuid.UUID
		var status string
		if err := tx.QueryRow(ctx, `SELECT workflow_run_id,status FROM workflow_runs r WHERE workflow_id=$1
  ORDER BY COALESCE(started_at,(SELECT min(created_at) FROM run_events e WHERE e.workflow_run_id=r.workflow_run_id)),workflow_run_id LIMIT 1 FOR UPDATE`, workflowID).Scan(&runID, &status); err != nil {
			return false, err
		}
		if status != "succeeded" && status != "failed" && status != "cancelled" {
			return false, nil
		}
		if _, err := tx.Exec(ctx, `DELETE FROM workflow_runs WHERE workflow_run_id=$1 AND workflow_id=$2`, runID, workflowID); err != nil {
			return false, err
		}
		count--
	}
	return true, nil
}
