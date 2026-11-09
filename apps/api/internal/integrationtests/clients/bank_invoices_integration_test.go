package integrationtests

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"riders.agency/pulse/api/internal/modules/clients/invoicing"
	"riders.agency/pulse/api/internal/modules/clients/repository"
	"riders.agency/pulse/api/internal/modules/clients/service"
	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/errs"
)

func bankFixture(t *testing.T) (context.Context, *pgxpool.Pool, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	dsn := os.Getenv("PULSE_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PULSE_TEST_DATABASE_URL is not set")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	user, company, project, client, action := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	ctx := authz.WithSessionClaims(context.Background(), authz.SessionClaims{UserID: user})
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM companies WHERE company_id=$1`, company)
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE user_id=$1`, user)
		pool.Close()
	})
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO users(user_id,email,password_hash,is_activated) VALUES($1,$2,'test-hash',true)`, user, user.String()+"@example.test")
	exec(`INSERT INTO companies(company_id,owner_user_id,slug,name,created_by_user_id) VALUES($1,$2,$3,'Invoice test',$2)`, company, user, company.String())
	exec(`INSERT INTO company_memberships(company_id,user_id,role) VALUES($1,$2,'owner')`, company, user)
	exec(`INSERT INTO projects(project_id,company_id,name,slug,hook_key,hook_secret_hash,created_by_user_id) VALUES($1,$2,'Invoice test',$3,$3,'test-hash',$4)`, project, company, project.String(), user)
	exec(`INSERT INTO project_members(project_id,user_id,role,status) VALUES($1,$2,'owner','active')`, project, user)
	exec(`INSERT INTO clients(client_id,company_id,name) VALUES($1,$2,'Demo payer')`, client, company)
	exec(`INSERT INTO payment_actions(payment_action_id,company_id,project_id,client_id,period_start,period_end,due_date,total_amount) VALUES($1,$2,$3,$4,'2026-11-01','2026-11-30','2026-11-30',1234.56)`, action, company, project, client)
	return ctx, pool, user, project, action
}

func TestBankInvoiceDraftPostgres(t *testing.T) {
	ctx, pool, user, project, action := bankFixture(t)
	invoices := service.NewBankInvoices(&repository.Clients{P: pool}, invoicing.NewRegistry())
	input := invoicing.Input{Kind: "bank_invoice", PayerName: " Demo payer ", Purpose: "Design services", DueDate: "2026-11-30"}
	workspace, err := invoices.Read(ctx, action, user)
	if err != nil || workspace.Draft != nil || !workspace.CanEdit || workspace.AmountMinor != "123456" || len(workspace.Providers) != 0 {
		t.Fatalf("initial workspace: %+v %v", workspace, err)
	}
	key := uuid.New()
	first, err := invoices.Save(ctx, action, user, key, input)
	if err != nil || first.Revision != 1 || first.AmountMinor != "123456" || first.PayerName != "Demo payer" {
		t.Fatalf("first draft: %+v %v", first, err)
	}
	retry, err := invoices.Save(ctx, action, user, key, input)
	if err != nil || retry.Revision != 1 || !retry.UpdatedAt.Equal(first.UpdatedAt) {
		t.Fatalf("retry: %+v %v", retry, err)
	}
	changed := input
	changed.Purpose = "Other purpose"
	if _, err := invoices.Save(ctx, action, user, key, changed); !errors.Is(err, invoicing.ErrConflict) {
		t.Fatalf("key reuse: %v", err)
	}
	if _, err := invoices.Save(ctx, action, user, uuid.New(), input); !errors.Is(err, invoicing.ErrConflict) {
		t.Fatalf("stale revision: %v", err)
	}
	changed.Revision = 1
	changed.Kind = "payment_link"
	second, err := invoices.Save(ctx, action, user, uuid.New(), changed)
	if err != nil || second.Revision != 2 {
		t.Fatalf("update: %+v %v", second, err)
	}
	if _, err := invoices.Save(ctx, action, user, key, input); !errors.Is(err, invoicing.ErrConflict) {
		t.Fatalf("old retry cannot overwrite: %v", err)
	}
	workspace, err = invoices.Read(ctx, action, user)
	if err != nil || workspace.Draft.Kind != "payment_link" || workspace.PaymentStatus != "pending" {
		t.Fatalf("persistence/status: %+v %v", workspace, err)
	}
	foreign := uuid.New()
	foreignCtx := authz.WithSessionClaims(ctx, authz.SessionClaims{UserID: foreign})
	if _, err := invoices.Read(foreignCtx, action, foreign); !errors.Is(err, errs.ErrForbidden) {
		t.Fatalf("foreign read: %v", err)
	}
	if _, err := invoices.Save(foreignCtx, action, foreign, uuid.New(), changed); !errors.Is(err, errs.ErrForbidden) {
		t.Fatalf("foreign write: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE project_members SET role='read_only_auditor' WHERE project_id=$1`, project); err != nil {
		t.Fatal(err)
	}
	if _, err := invoices.Read(ctx, action, user); err != nil {
		t.Fatalf("reader: %v", err)
	}
	readOnly, err := invoices.Read(ctx, action, user)
	if err != nil || readOnly.CanEdit {
		t.Fatalf("read-only capability: %+v %v", readOnly, err)
	}
	if _, err := invoices.Save(ctx, action, user, uuid.New(), changed); !errors.Is(err, errs.ErrForbidden) {
		t.Fatalf("read-only write: %v", err)
	}
	payments := &repository.Clients{P: pool}
	if _, err := service.NewCommands(payments).CreateManualPaymentAction(ctx, project, user, repository.CreateManualPaymentActionInput{}); !errors.Is(err, errs.ErrForbidden) {
		t.Fatalf("read-only payment creation: %v", err)
	}
	if err := service.NewCommands(payments).ConfirmPaymentAction(ctx, action, user, repository.ConfirmPaymentInput{}); !errors.Is(err, errs.ErrForbidden) {
		t.Fatalf("read-only payment confirmation: %v", err)
	}
	if err := payments.SnoozePaymentAction(ctx, action, user, repository.SnoozePaymentInput{}); !errors.Is(err, errs.ErrForbidden) {
		t.Fatalf("read-only snooze: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE project_members SET role='owner' WHERE project_id=$1`, project); err != nil {
		t.Fatal(err)
	}
	keyCtx := authz.WithSessionClaims(ctx, authz.SessionClaims{UserID: user, ActorType: authz.ActorAPIKey, Grants: []authz.Grant{{ScopeKind: authz.ScopeProject, ScopeID: project, Permissions: []string{authz.PermRead}}}})
	if _, err := invoices.Save(keyCtx, action, user, uuid.New(), changed); !errors.Is(err, errs.ErrForbidden) {
		t.Fatalf("read-only API key: %v", err)
	}
	foreignScope := authz.WithSessionClaims(ctx, authz.SessionClaims{UserID: user, ActorType: authz.ActorAPIKey, Grants: []authz.Grant{{ScopeKind: authz.ScopeProject, ScopeID: uuid.New(), Permissions: []string{authz.PermWrite}}}})
	if _, err := invoices.Read(foreignScope, action, user); !errors.Is(err, errs.ErrForbidden) {
		t.Fatalf("foreign project scope: %v", err)
	}
	changed.Revision = 2
	for _, status := range []string{"paid", "cancelled"} {
		if _, err := pool.Exec(ctx, `UPDATE payment_actions SET status=$2 WHERE payment_action_id=$1`, action, status); err != nil {
			t.Fatal(err)
		}
		if _, err := invoices.Save(ctx, action, user, uuid.New(), changed); !errors.Is(err, invoicing.ErrConflict) {
			t.Fatalf("closed %s: %v", status, err)
		}
	}
}

func TestConcurrentBankInvoiceDraftPostgres(t *testing.T) {
	ctx, pool, user, _, action := bankFixture(t)
	invoices := service.NewBankInvoices(&repository.Clients{P: pool}, invoicing.NewRegistry())
	in := invoicing.Input{Kind: "bank_invoice", PayerName: "Demo", Purpose: "Concurrent save"}
	var wg sync.WaitGroup
	results := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := invoices.Save(ctx, action, user, uuid.New(), in); results <- err }()
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		} else if !errors.Is(err, invoicing.ErrConflict) {
			t.Fatal(err)
		}
	}
	if successes != 1 {
		t.Fatalf("concurrent winners: %d", successes)
	}
}
