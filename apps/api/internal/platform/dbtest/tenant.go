// Package dbtest builds isolated tenants for integration tests against a disposable database.
package dbtest

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"riders.agency/pulse/api/internal/platform/authz"
)

type Tenant struct {
	Pool                         *pgxpool.Pool
	UserID, CompanyID, ProjectID uuid.UUID
	Context                      context.Context
}

func NewTenant(t *testing.T) Tenant {
	t.Helper()
	dsn := os.Getenv("PULSE_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PULSE_TEST_DATABASE_URL is not set")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	f := Tenant{Pool: pool, UserID: uuid.New(), CompanyID: uuid.New(), ProjectID: uuid.New()}
	f.Context = authz.WithSessionClaims(context.Background(), authz.SessionClaims{UserID: f.UserID, ActorType: authz.ActorSession})
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM companies WHERE company_id=$1`, f.CompanyID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE user_id=$1`, f.UserID)
		pool.Close()
	})
	f.Exec(t, `INSERT INTO users(user_id,email,password_hash,first_name,display_name,is_activated) VALUES($1,$2,'test-only','Test','Test',true)`, f.UserID, f.UserID.String()+"@example.test")
	f.Exec(t, `INSERT INTO companies(company_id,owner_user_id,slug,name,created_by_user_id) VALUES($1,$2,$3,'Test',$2)`, f.CompanyID, f.UserID, f.CompanyID.String())
	f.Exec(t, `INSERT INTO company_memberships(company_id,user_id,role) VALUES($1,$2,'owner')`, f.CompanyID, f.UserID)
	f.Exec(t, `INSERT INTO projects(project_id,company_id,name,slug,hook_key,hook_secret_hash,created_by_user_id) VALUES($1,$2,'Test',$3,$3,'test-only',$4)`, f.ProjectID, f.CompanyID, f.ProjectID.String(), f.UserID)
	f.Exec(t, `INSERT INTO project_members(project_id,user_id,role,status) VALUES($1,$2,'owner','active')`, f.ProjectID, f.UserID)
	return f
}

func (f Tenant) Exec(t *testing.T, sql string, args ...any) {
	t.Helper()
	if _, err := f.Pool.Exec(f.Context, sql, args...); err != nil {
		t.Fatal(err)
	}
}
