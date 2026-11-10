package kernel_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"riders.agency/pulse/api/internal/executor"
	"riders.agency/pulse/api/internal/runtime/contracts"
	"riders.agency/pulse/api/internal/runtime/kernel"
	"riders.agency/pulse/api/internal/runtime/runqueue"
)

type resumeProbe struct {
	calls    int
	approved bool
	applied  any
}
type resumeAction struct {
	kind  string
	probe *resumeProbe
}

func (a resumeAction) TypeKey() string                                       { return a.kind }
func (a resumeAction) ValidateConfig(context.Context, json.RawMessage) error { return nil }
func (a resumeAction) Execute(_ context.Context, req contracts.ActionRequest) contracts.ActionResult {
	switch a.kind {
	case "test.prepare":
		a.probe.calls++
		return contracts.ActionResult{OutputJSON: json.RawMessage(fmt.Sprintf(`{"expose":{"draft":%d}}`, a.probe.calls)), ProviderResponseJSON: json.RawMessage(`{"usage":{"inputTokens":11}}`)}
	case "test.approve":
		return contracts.ActionResult{OutputJSON: json.RawMessage(`{}`), Waiting: !a.probe.approved, WaitingReason: "office_approval_required"}
	default:
		a.probe.applied = req.Steps["prepare"]
		return contracts.ActionResult{OutputJSON: json.RawMessage(`{}`)}
	}
}

func TestResumePreservesReviewedOutputAndDoesNotCallModelAgain(t *testing.T) {
	dsn := os.Getenv("OFFICE_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("OFFICE_TEST_DATABASE_URL required")
	}
	if !strings.Contains(dsn, "office_test") {
		t.Fatal("isolated office_test database required")
	}
	t.Setenv("PULSE_PUBLIC_DEMO", "false")
	ctx := context.Background()
	p, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := p.Exec(ctx, q, args...); err != nil {
			t.Fatal(err)
		}
	}
	uid, cid, pid, wid, vid, rid := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	exec(`INSERT INTO users(user_id,email,password_hash,is_activated) VALUES($1,$2,'test-only',true)`, uid, uid.String()+"@example.test")
	exec(`INSERT INTO companies(company_id,owner_user_id,name,slug) VALUES($1,$2,'Resume test',$3)`, cid, uid, cid.String())
	exec(`INSERT INTO company_memberships(company_id,user_id,role) VALUES($1,$2,'owner')`, cid, uid)
	exec(`INSERT INTO projects(project_id,company_id,name,slug,hook_key,hook_secret_hash) VALUES($1,$2,'Resume test','resume',$3,'test-only')`, pid, cid, uuid.NewString())
	definition := `{"nodes":[{"action_key":"prepare","action_type_key":"test.prepare","config":{}},{"action_key":"approval","action_type_key":"test.approve","config":{}},{"action_key":"apply","action_type_key":"test.apply","config":{}}],"edges":[{"from":"prepare","to":"approval"},{"from":"approval","to":"apply"}]}`
	exec(`INSERT INTO workflows(workflow_id,project_id,name,slug,status) VALUES($1,$2,'Resume','resume','draft')`, wid, pid)
	exec(`INSERT INTO workflow_versions(workflow_version_id,project_id,workflow_id,version_no,status,definition_json,graph_json) VALUES($1,$2,$3,1,'published',$4,'{}')`, vid, pid, wid, definition)
	if _, _, err := runqueue.RequestRun(ctx, p, runqueue.Request{WorkflowRunID: rid, ProjectID: pid.String(), WorkflowID: wid.String(), WorkflowVersionID: vid.String(), SourceType: runqueue.SourceManual, IdempotencyKey: rid.String(), InputJSON: json.RawMessage(`{}`)}); err != nil {
		t.Fatal(err)
	}
	engine, probe := executor.New(p), &resumeProbe{}
	for _, kind := range []string{"test.prepare", "test.approve", "test.apply"} {
		engine.Registry.MustRegisterAction(resumeAction{kind, probe})
	}
	k := kernel.New(p, engine, kernel.WithRunClaimAllowlist(rid.String()))
	if !k.ProcessOne(ctx) {
		t.Fatal("first execution not claimed")
	}
	var status string
	p.QueryRow(ctx, `SELECT status FROM workflow_runs WHERE workflow_run_id=$1`, rid).Scan(&status)
	if status != "waiting" || probe.calls != 1 || probe.applied != nil {
		t.Fatalf("before approval: %s %+v", status, probe)
	}
	probe.approved = true
	exec(`UPDATE workflow_runs SET status='pending',waiting_reason='',wake_reason='human_decision' WHERE workflow_run_id=$1`, rid)
	if !k.ProcessOne(ctx) {
		t.Fatal("resume not claimed")
	}
	p.QueryRow(ctx, `SELECT status FROM workflow_runs WHERE workflow_run_id=$1`, rid).Scan(&status)
	if status != "succeeded" || probe.calls != 1 {
		t.Fatalf("resume regenerated the draft: %s %+v", status, probe)
	}
	applied, _ := probe.applied.(map[string]any)
	if applied["draft"] != float64(1) {
		t.Fatalf("reviewed draft was replaced: %#v", applied)
	}
	var calls int
	p.QueryRow(ctx, `SELECT count(*) FROM run_events WHERE workflow_run_id=$1 AND action_key='prepare' AND event_type='action.started'`, rid).Scan(&calls)
	if calls != 1 {
		t.Fatalf("duplicate preparation event: %d", calls)
	}
}
