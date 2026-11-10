package demomode

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/authz"
)

type tenantReader struct {
	demo bool
	err  error
}

func (r tenantReader) QueryRow(context.Context, string, ...any) pgx.Row { return r }
func (r tenantReader) Scan(dest ...any) error {
	if r.err == nil {
		*dest[0].(*bool) = r.demo
	}
	return r.err
}

func TestTenantPolicyOnMixedProduction(t *testing.T) {
	t.Setenv("PULSE_PUBLIC_DEMO", "false")
	for _, tc := range []struct {
		name string
		demo bool
		err  error
		path string
		want int
	}{
		{"real integration unaffected", false, nil, "/api/v1/projects/p/action-dependencies", 204},
		{"demo integration denied", true, nil, "/api/v1/projects/p/action-dependencies", 403},
		{"demo API key denied", true, nil, "/api/v1/api-keys", 403},
		{"demo profile immutable", true, nil, "/api/v1/auth/profile", 403},
		{"demo office available", true, nil, "/api/v1/projects/p/office-demo", 204},
		{"lookup fails closed", false, errors.New("database down"), "/api/v1/work/tasks", 503},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := TenantMiddleware(tenantReader{tc.demo, tc.err})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
			req := httptest.NewRequest("POST", tc.path, nil)
			req = req.WithContext(authz.WithSessionClaims(req.Context(), authz.SessionClaims{UserID: uuid.New()}))
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			if w.Code != tc.want {
				t.Fatalf("got %d, want %d", w.Code, tc.want)
			}
		})
	}
}

func TestReservedGuestRegistrationIsBoundedButRealRegistrationRemainsAvailable(t *testing.T) {
	t.Setenv("PULSE_PUBLIC_DEMO", "false")
	h := TenantMiddleware(tenantReader{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	for i := 0; i < 31; i++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(`{"email":"guest-x@DEMO.example.test"}`)))
		want := 204
		if i == 30 {
			want = 429
		}
		if w.Code != want {
			t.Fatalf("registration %d: %d", i, w.Code)
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(`{"email":"person@example.test"}`)))
	if w.Code != 204 {
		t.Fatal("real registration affected")
	}
}
