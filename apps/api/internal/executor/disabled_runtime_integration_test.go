package executor

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"riders.agency/pulse/api/internal/runtime/kernel"
	"riders.agency/pulse/api/internal/runtime/runqueue"
)

func TestDisabledActionPostgresRuntime(t *testing.T) {
	dsn := os.Getenv("PULSE_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PULSE_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	f, err := createDisabledFixture(ctx, pool)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Exec(ctx, `DELETE FROM users WHERE user_id=$1`, f.UserID)
	defer pool.Exec(ctx, `DELETE FROM companies WHERE company_id=$1`, f.CompanyID)
	runID := uuid.New()
	_, _, err = runqueue.RequestRun(ctx, pool, runqueue.Request{WorkflowRunID: runID, ProjectID: f.ProjectID.String(), WorkflowID: f.WorkflowID.String(), WorkflowVersionID: f.VersionID.String(), SourceType: runqueue.SourceTest, InputJSON: []byte(`{}`), IdempotencyKey: "disabled-test:" + runID.String()})
	if err != nil {
		t.Fatal(err)
	}
	// No mail provider or dependency is registered. Calling it would fail the run.
	k := kernel.New(pool, New(pool), kernel.WithRunClaimAllowlist(runID.String()))
	if !k.ProcessOne(ctx) {
		t.Fatal("run was not processed")
	}
	var status, skipStatus, reason, finishStatus string
	var attempts, effects int
	err = pool.QueryRow(ctx, `SELECT wr.status, ar.status, ar.output_json->>'reason', ar.attempt_count,
 (SELECT status FROM action_runs WHERE workflow_run_id=wr.workflow_run_id AND action_key='finish'),
 (SELECT count(*) FROM effect_ledger WHERE workflow_run_id=wr.workflow_run_id AND action_key='mail')
 FROM workflow_runs wr JOIN action_runs ar USING(workflow_run_id) WHERE wr.workflow_run_id=$1 AND ar.action_key='mail'`, runID).Scan(&status, &skipStatus, &reason, &attempts, &finishStatus, &effects)
	if err != nil {
		t.Fatal(err)
	}
	if status != "succeeded" || skipStatus != "skipped" || reason != "disabled" || attempts != 0 || finishStatus != "succeeded" || effects != 0 {
		t.Fatalf("status=%s skip=%s reason=%s attempts=%d next=%s effects=%d", status, skipStatus, reason, attempts, finishStatus, effects)
	}
}

type disabledFixtureIDs struct {
	UserID     uuid.UUID
	CompanyID  uuid.UUID
	ProjectID  uuid.UUID
	WorkflowID uuid.UUID
	VersionID  uuid.UUID
}

func createDisabledFixture(ctx context.Context, pool *pgxpool.Pool) (disabledFixtureIDs, error) {
	f := disabledFixtureIDs{
		UserID:     uuid.New(),
		CompanyID:  uuid.New(),
		ProjectID:  uuid.New(),
		WorkflowID: uuid.New(),
		VersionID:  uuid.New(),
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO users (user_id, first_name, display_name, is_activated)
VALUES ($1, 'Disabled Action Test', 'Disabled Action Test', true)`, f.UserID); err != nil {
		return disabledFixtureIDs{}, err
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO companies (company_id, owner_user_id, slug, name, created_by_user_id)
VALUES ($1, $2, $3, 'Disabled Action Test', $2)`, f.CompanyID, f.UserID, "disabled-action-"+f.CompanyID.String()); err != nil {
		return disabledFixtureIDs{}, err
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO projects (project_id, company_id, name, slug, hook_key, hook_secret_hash, created_by_user_id)
VALUES ($1, $2, 'Disabled Action Test', $3, $4, 'hash', $5)`,
		f.ProjectID, f.CompanyID, "disabled-action-"+f.ProjectID.String(), "hook-"+f.ProjectID.String(), f.UserID); err != nil {
		return disabledFixtureIDs{}, err
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO company_memberships (company_id, user_id, role)
VALUES ($1, $2, 'owner')`, f.CompanyID, f.UserID); err != nil {
		return disabledFixtureIDs{}, err
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO project_members (project_id, user_id, role, status)
VALUES ($1, $2, 'owner', 'active')`, f.ProjectID, f.UserID); err != nil {
		return disabledFixtureIDs{}, err
	}
	def := `{"nodes":[{"action_key":"start","action_type_key":"manual.trigger","config":{}},{"action_key":"mail","action_type_key":"email.send","config":{"disabled":true}},{"action_key":"finish","action_type_key":"stop","config":{}}],"edges":[{"from":"start","to":"mail"},{"from":"mail","to":"finish"}]}`
	if _, err := pool.Exec(ctx, `
INSERT INTO workflows (workflow_id, project_id, name, slug, status, draft_definition_json, created_by_user_id)
VALUES ($1, $2, 'Disabled Action Test', $3, 'active', $4::jsonb, $5)`,
		f.WorkflowID, f.ProjectID, "disabled-action-"+f.WorkflowID.String(), def, f.UserID); err != nil {
		return disabledFixtureIDs{}, err
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO workflow_versions (workflow_version_id, project_id, workflow_id, version_no, status, definition_json, created_by_user_id)
VALUES ($1, $2, $3, 1, 'published', $4::jsonb, $5)`,
		f.VersionID, f.ProjectID, f.WorkflowID, def, f.UserID); err != nil {
		return disabledFixtureIDs{}, err
	}
	_, err := pool.Exec(ctx, `UPDATE workflows SET current_workflow_version_id = $2 WHERE workflow_id = $1`, f.WorkflowID, f.VersionID)
	return f, err
}
