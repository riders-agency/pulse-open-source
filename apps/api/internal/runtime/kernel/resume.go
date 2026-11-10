package kernel

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// Versions are immutable within a run. A reclaimed run resumes from its durable
// successful outputs, so approval cannot silently regenerate the reviewed draft.
func (k *Kernel) completedOutput(ctx context.Context, claim runClaim, node Node) (string, bool, error) {
	if claim.ClaimCount < 2 {
		return "", false, nil
	}
	var output string
	err := k.pool.QueryRow(ctx, `SELECT COALESCE(output_json,'{}'::jsonb)::text FROM action_runs
	 WHERE workflow_run_id=$1 AND action_key=$2 AND action_type_key=$3 AND status='succeeded'`,
		claim.RunID, node.ActionKey, node.ActionTypeKey).Scan(&output)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	return output, err == nil, err
}
