package kernel_test

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"riders.agency/pulse/api/internal/modules/workflows/office"
	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/demomode"
	"riders.agency/pulse/api/internal/runtime/contracts"
	"riders.agency/pulse/api/internal/runtime/runqueue"
	"strings"
	"testing"
	"time"
)

func TestOfficePersistencePollingApprovalAndRetention(t *testing.T) {
	dsn := os.Getenv("OFFICE_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("OFFICE_TEST_DATABASE_URL required")
	}
	if !strings.Contains(dsn, "office_test") {
		t.Fatal("isolated office_test database required")
	}
	t.Setenv("PULSE_PUBLIC_DEMO", "false")
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	uid, cid, pid := uuid.New(), uuid.New(), uuid.New()
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, q, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO users(user_id,email,password_hash,is_activated) VALUES($1,$2,'test-only-not-a-password',true)`, uid, "guest-"+uid.String()+"@demo.example.test")
	exec(`INSERT INTO companies(company_id,owner_user_id,name,slug) VALUES($1,$2,'Demo',$3)`, cid, uid, "demo-"+cid.String())
	exec(`INSERT INTO company_memberships(company_id,user_id,role) VALUES($1,$2,'owner')`, cid, uid)
	exec(`INSERT INTO projects(project_id,company_id,name,slug,hook_key,hook_secret_hash) VALUES($1,$2,'Office','office',$3,'test-only')`, pid, cid, uuid.NewString())
	exec(`INSERT INTO project_members(project_id,user_id,role) VALUES($1,$2,'owner')`, pid, uid)
	ctx = authz.WithSessionClaims(ctx, authz.SessionClaims{UserID: uid, ActorType: authz.ActorSession})
	svc := &office.Service{P: pool}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `UPDATE workflow_office_workspaces SET enabled=false WHERE project_id=$1`, pid)
	})
	state, err := svc.Ensure(ctx, pid, uid)
	if err != nil {
		t.Fatal(err)
	}
	same, err := svc.Ensure(ctx, pid, uid)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM workflows WHERE workflow_id=$1`, state["workflow_id"])
	})
	if err != nil || same["workflow_id"] != state["workflow_id"] {
		t.Fatal("duplicate provisioning", err)
	}
	other := authz.WithSessionClaims(context.Background(), authz.SessionClaims{UserID: uuid.New(), ActorType: authz.ActorSession})
	if _, err := svc.State(other, pid, uuid.New()); err == nil {
		t.Fatal("cross tenant state allowed")
	}
	tr := &office.Trigger{P: pool}
	if n, err := tr.EnqueueDue(ctx, 10); err != nil || n != 0 {
		t.Fatal("polled before 30s", n, err)
	}
	scheduled := time.Now().UTC().Add(-time.Second).Truncate(time.Microsecond)
	exec(`UPDATE workflow_office_workspaces SET next_poll_at=$2 WHERE project_id=$1`, pid, scheduled)
	exec(`UPDATE workflow_versions SET definition_json=jsonb_set(definition_json,'{nodes,0,config,disabled}','true') WHERE workflow_id=$1`, state["workflow_id"])
	if n, err := tr.EnqueueDue(ctx, 10); err != nil || n != 0 {
		t.Fatal("disabled inbox trigger enqueued mail", n, err)
	}
	exec(`UPDATE workflow_versions SET definition_json=jsonb_set(definition_json,'{nodes,0,config,disabled}','false') WHERE workflow_id=$1`, state["workflow_id"])
	if n, err := tr.EnqueueDue(ctx, 10); err != nil || n != 1 {
		t.Fatal("mail not enqueued", n, err)
	}
	var nextPoll time.Time
	if err := pool.QueryRow(ctx, `SELECT next_poll_at FROM workflow_office_workspaces WHERE project_id=$1`, pid).Scan(&nextPoll); err != nil {
		t.Fatal(err)
	}
	if !nextPoll.Equal(scheduled.Add(30 * time.Second)) {
		t.Fatalf("poll schedule drifted: got %s, want %s", nextPoll, scheduled.Add(30*time.Second))
	}
	if n, err := tr.EnqueueDue(ctx, 10); err != nil || n != 0 {
		t.Fatal("duplicate poll", n, err)
	}
	overdue := time.Now().UTC().Add(-95 * time.Second).Truncate(time.Microsecond)
	exec(`UPDATE workflow_office_workspaces SET next_poll_at=$2 WHERE project_id=$1`, pid, overdue)
	if n, err := tr.EnqueueDue(ctx, 10); err != nil || n != 0 {
		t.Fatal("empty poll after downtime", n, err)
	}
	if err := pool.QueryRow(ctx, `SELECT next_poll_at FROM workflow_office_workspaces WHERE project_id=$1`, pid).Scan(&nextPoll); err != nil {
		t.Fatal(err)
	}
	if !nextPoll.Equal(overdue.Add(120 * time.Second)) {
		t.Fatalf("missed slots were not skipped: got %s, want %s", nextPoll, overdue.Add(120*time.Second))
	}
	var rid uuid.UUID
	var vid uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT workflow_run_id,workflow_version_id FROM workflow_runs WHERE workflow_id=$1`, state["workflow_id"]).Scan(&rid, &vid); err != nil {
		t.Fatal(err)
	}
	h := office.Handler{P: pool, Kind: "office.approve"}
	if allowed, err := demomode.RunAllowed(ctx, pool, rid.String(), pid.String()); err != nil || !allowed {
		t.Fatal("native demo run denied", err)
	}
	if allowed, err := demomode.RunAllowed(ctx, pool, uuid.NewString(), pid.String()); err != nil || allowed {
		t.Fatal("unregistered demo run allowed", err)
	}
	req := contracts.ActionRequest{ProjectID: pid.String(), WorkflowRunID: rid.String(), RunInputJSON: json.RawMessage(`{}`)}
	if result := h.Execute(ctx, req); !result.Waiting || result.Err != nil {
		t.Fatal("approval did not pause", result.Err)
	}
	h.Kind = "office.apply"
	if result := h.Execute(ctx, req); result.Err == nil {
		t.Fatal("write before approval")
	}
	exec(`UPDATE workflow_runs SET status='waiting',waiting_reason='office_approval_required' WHERE workflow_run_id=$1`, rid)
	if _, err := svc.Decide(other, pid, uuid.New(), rid, true); err == nil {
		t.Fatal("cross tenant approval")
	}
	if _, err := svc.Decide(ctx, pid, uid, rid, true); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Decide(ctx, pid, uid, rid, true); err == nil {
		t.Fatal("duplicate decision")
	}
	req.Steps = map[string]any{"website": map[string]any{"title": "Test", "body": "Saved"}, "invoice": map[string]any{"number": "test", "amount_minor": 4500000}}
	if result := h.Execute(ctx, req); result.Err != nil {
		t.Fatal(result.Err)
	}
	if result := h.Execute(ctx, req); result.Err != nil {
		t.Fatal("repeat apply", result.Err)
	}
	var count int
	pool.QueryRow(ctx, `SELECT count(*) FROM workflow_office_artifacts WHERE workflow_run_id=$1`, rid).Scan(&count)
	if count != 1 {
		t.Fatal("idempotency", count)
	}
	exec(`UPDATE workflow_runs SET status='succeeded' WHERE workflow_run_id=$1`, rid)
	wid := uuid.MustParse(state["workflow_id"].(string))
	for i := 0; i < 29; i++ {
		id := uuid.New()
		if _, _, err := runqueue.RequestRun(ctx, pool, runqueue.Request{WorkflowRunID: id, ProjectID: pid.String(), WorkflowID: wid.String(), WorkflowVersionID: vid.String(), SourceType: runqueue.SourceManual}); err != nil {
			t.Fatal(err)
		}
		exec(`UPDATE workflow_runs SET status='succeeded' WHERE workflow_run_id=$1`, id)
		exec(`INSERT INTO workflow_office_messages(project_id,delivery_key,subject,body,sender,amount_minor,workflow_run_id) VALUES($1,$2,'s','b','demo',1,$3)`, pid, uuid.New(), id)
	}
	if _, err := svc.Deliver(ctx, pid, uid, uuid.New()); err != nil {
		t.Fatal(err)
	}
	exec(`UPDATE workflow_office_workspaces SET next_poll_at=now() WHERE project_id=$1`, pid)
	if n, err := tr.EnqueueDue(ctx, 10); err != nil || n != 1 {
		t.Fatal("retention dispatch", n, err)
	}
	pool.QueryRow(ctx, `SELECT count(*) FROM workflow_runs WHERE workflow_id=$1`, wid).Scan(&count)
	if count != 30 {
		t.Fatal("retention count", count)
	}
	pool.QueryRow(ctx, `SELECT count(*) FROM workflow_office_artifacts WHERE workflow_run_id=$1`, rid).Scan(&count)
	if count != 0 {
		t.Fatal("orphan artifact", count)
	}
	exec(`UPDATE workflow_runs SET status='waiting' WHERE workflow_id=$1`, wid)
	if _, err := svc.Deliver(ctx, pid, uid, uuid.New()); err != nil {
		t.Fatal(err)
	}
	exec(`UPDATE workflow_office_workspaces SET next_poll_at=now() WHERE project_id=$1`, pid)
	if n, err := tr.EnqueueDue(ctx, 10); err != nil || n != 0 {
		t.Fatal("waiting run evicted", n, err)
	}
}
