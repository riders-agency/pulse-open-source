package authz_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/dbtest"
	"riders.agency/pulse/api/internal/platform/errs"
)

func TestSessionAuthorizationEnforcesRoleStatusAndTenant(t *testing.T) {
	f := dbtest.NewTenant(t)
	if err := authz.RequireProject(f.Context, f.Pool, f.ProjectID, authz.PermWrite); err != nil {
		t.Fatal(err)
	}
	for _, status := range []string{"invited", "inactive", "removed"} {
		f.Exec(t, `UPDATE project_members SET status=$2 WHERE project_id=$1`, f.ProjectID, status)
		if err := authz.RequireProject(f.Context, f.Pool, f.ProjectID, authz.PermRead); !errors.Is(err, errs.ErrForbidden) {
			t.Fatalf("status %s: %v", status, err)
		}
	}
	f.Exec(t, `UPDATE project_members SET status='active',role='read_only_auditor' WHERE project_id=$1`, f.ProjectID)
	if err := authz.RequireProject(f.Context, f.Pool, f.ProjectID, authz.PermRead); err != nil {
		t.Fatal(err)
	}
	for _, permission := range []string{authz.PermWrite, authz.PermAdmin, "unknown"} {
		if err := authz.RequireProject(f.Context, f.Pool, f.ProjectID, permission); !errors.Is(err, errs.ErrForbidden) {
			t.Fatalf("permission %s: %v", permission, err)
		}
	}
	if err := authz.RequireProject(f.Context, f.Pool, uuid.New(), authz.PermRead); !errors.Is(err, errs.ErrForbidden) {
		t.Fatalf("foreign project: %v", err)
	}
	f.Exec(t, `UPDATE company_memberships SET role='member' WHERE company_id=$1`, f.CompanyID)
	if err := authz.RequireCompany(f.Context, f.Pool, f.CompanyID, authz.PermAdmin); !errors.Is(err, errs.ErrForbidden) {
		t.Fatalf("company admin: %v", err)
	}
}
