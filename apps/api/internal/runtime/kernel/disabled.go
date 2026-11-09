package kernel

import "context"

func (k *Kernel) recordDisabledNode(ctx context.Context, claim *runClaim, node Node, input string) error {
	tx, err := k.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	// Serialize against reclaim/cancel so a stale worker cannot rewrite a result.
	var runID string
	if err := tx.QueryRow(ctx, `SELECT workflow_run_id::text FROM workflow_runs WHERE workflow_run_id=$1 AND run_claim_owner=$2 AND run_fencing_token=$3 AND status='running' FOR UPDATE`, claim.RunID, claim.Owner, claim.FencingToken).Scan(&runID); err != nil {
		return err
	}
	output := k.runtime.SkippedOutputWithReason(node, "disabled")
	_, err = tx.Exec(ctx, `
INSERT INTO action_runs (workflow_run_id, action_key, action_type_key, status, attempt_count, max_attempts, input_json, output_json, provider_response_json, started_at, finished_at)
VALUES ($1,$2,$3,'skipped',0,1,$4::jsonb,$5::jsonb,'{}'::jsonb,now(),now())
ON CONFLICT (workflow_run_id, action_key) DO UPDATE SET status='skipped', attempt_count=0, output_json=EXCLUDED.output_json, finished_at=now(), replayable=false, error=NULL`, runID, node.ActionKey, node.ActionTypeKey, input, output)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO run_events (workflow_run_id,action_key,event_type,message,payload_json,created_at) VALUES ($1,$2,'action.skipped','Действие отключено в сценарии',$3::jsonb,clock_timestamp())`, runID, node.ActionKey, output)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
