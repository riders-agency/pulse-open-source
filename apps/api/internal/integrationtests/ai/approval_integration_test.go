package integrationtests

import (
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	airepo "riders.agency/pulse/api/internal/modules/ai/repository"
	commands "riders.agency/pulse/api/internal/modules/ai/service"
	clientssvc "riders.agency/pulse/api/internal/modules/clients/service"
	"riders.agency/pulse/api/internal/platform/dbtest"
	"riders.agency/pulse/api/internal/platform/errs"
)

func seedApproval(t *testing.T) (dbtest.Tenant, *commands.Commands, uuid.UUID) {
	t.Helper()
	f := dbtest.NewTenant(t)
	c := commands.NewCommands(&airepo.AI{P: f.Pool, Clients: clientssvc.New(f.Pool, "")})
	session, err := c.CreateSession(f.Context, f.ProjectID, f.UserID, airepo.SessionInput{})
	if err != nil {
		t.Fatal(err)
	}
	callID := uuid.New()
	f.Exec(t, `INSERT INTO ai_tool_calls(ai_tool_call_id,ai_chat_session_id,project_id,tool_key,status,risk_level,requires_approval,input_json) VALUES($1,$2,$3,'client.create','pending_approval','medium',true,'{"name":"Approval race test","client_type":"company","status":"active"}')`, callID, session["ai_chat_session_id"], f.ProjectID)
	return f, c, callID
}

func TestConcurrentApprovalCreatesExactlyOneClient(t *testing.T) {
	f, c, callID := seedApproval(t)
	var wg sync.WaitGroup
	results := make(chan error, 8)
	start := make(chan struct{})
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := c.ApproveToolCall(f.Context, f.ProjectID, callID, f.UserID)
			results <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err := f.Pool.QueryRow(f.Context, `SELECT count(*) FROM clients WHERE company_id=$1 AND name='Approval race test'`, f.CompanyID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("clients=%d, want exactly one", count)
	}
	var status string
	if err := f.Pool.QueryRow(f.Context, `SELECT status FROM ai_tool_calls WHERE ai_tool_call_id=$1`, callID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "succeeded" {
		t.Fatalf("status=%s", status)
	}
}

func TestApprovalRechecksEmployeePolicyAndRole(t *testing.T) {
	f, c, callID := seedApproval(t)
	f.Exec(t, `UPDATE project_ai_employees SET status='disabled' WHERE project_id=$1`, f.ProjectID)
	if _, err := c.ApproveToolCall(f.Context, f.ProjectID, callID, f.UserID); !errors.Is(err, errs.ErrForbidden) {
		t.Fatalf("disabled: %v", err)
	}
	f.Exec(t, `UPDATE project_ai_employees SET status='active',tool_policy_json='{"allowed_tools":[]}' WHERE project_id=$1`, f.ProjectID)
	if _, err := c.ApproveToolCall(f.Context, f.ProjectID, callID, f.UserID); !errors.Is(err, errs.ErrForbidden) {
		t.Fatalf("tool denied: %v", err)
	}
	f.Exec(t, `UPDATE project_ai_employees SET tool_policy_json='{}' WHERE project_id=$1`, f.ProjectID)
	f.Exec(t, `UPDATE project_members SET role='read_only_auditor' WHERE project_id=$1`, f.ProjectID)
	if _, err := c.ApproveToolCall(f.Context, f.ProjectID, callID, f.UserID); !errors.Is(err, errs.ErrForbidden) {
		t.Fatalf("read-only actor: %v", err)
	}
	var count int
	if err := f.Pool.QueryRow(f.Context, `SELECT count(*) FROM clients WHERE company_id=$1`, f.CompanyID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("forbidden calls created %d clients", count)
	}
}

func TestApprovalRollsBackDomainWriteWhenResultCannotBeRecorded(t *testing.T) {
	f, c, callID := seedApproval(t)
	// A database trigger simulates failure after the domain write, before the approval receipt.
	f.Exec(t, `CREATE OR REPLACE FUNCTION public.test_fail_approval_receipt() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.input_json->>'test_fail_receipt' = 'yes' THEN RAISE EXCEPTION 'test receipt failure'; END IF; RETURN NEW; END $$`)
	f.Exec(t, `UPDATE ai_tool_calls SET input_json=input_json || '{"test_fail_receipt":"yes"}'::jsonb WHERE ai_tool_call_id=$1`, callID)
	f.Exec(t, `CREATE TRIGGER test_fail_approval_receipt BEFORE UPDATE ON ai_tool_calls FOR EACH ROW EXECUTE FUNCTION public.test_fail_approval_receipt()`)
	t.Cleanup(func() {
		f.Pool.Exec(f.Context, `DROP TRIGGER IF EXISTS test_fail_approval_receipt ON ai_tool_calls`)
		f.Pool.Exec(f.Context, `DROP FUNCTION IF EXISTS public.test_fail_approval_receipt()`)
	})
	if _, err := c.ApproveToolCall(f.Context, f.ProjectID, callID, f.UserID); err == nil {
		t.Fatal("expected injected receipt failure")
	}
	var count int
	if err := f.Pool.QueryRow(f.Context, `SELECT count(*) FROM clients WHERE company_id=$1`, f.CompanyID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("partial business write survived rollback: %d", count)
	}
}
