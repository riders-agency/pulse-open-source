package demomode

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/problem"
)

type reader interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}
type tenantKey struct{}

// The reserved, non-deliverable email namespace is immutable through the profile API.
// It only restricts privileges; client flags and company names never grant demo access.
func IsUser(ctx context.Context, p reader, uid uuid.UUID) (bool, error) {
	var demo bool
	err := p.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pulse_read.users_v1
 WHERE user_id=$1 AND lower(email::text) LIKE '%@demo.example.test')`, uid).Scan(&demo)
	return demo, err
}

// Demo workspaces can execute only runs tied to their persisted incoming letters.
// This guard also covers internal tool calls that do not pass through HTTP middleware.
func RunAllowed(ctx context.Context, p reader, runID, projectID string) (bool, error) {
	var allowed bool
	err := p.QueryRow(ctx, `SELECT NOT EXISTS(
 SELECT 1 FROM pulse_read.projects_v1 p JOIN pulse_read.company_memberships_v1 cm USING(company_id)
 JOIN pulse_read.users_v1 u ON u.user_id=cm.user_id WHERE p.project_id=$2 AND cm.role='owner'
 AND lower(u.email::text) LIKE '%@demo.example.test') OR EXISTS(
 SELECT 1 FROM workflow_office_messages m JOIN workflow_office_workspaces o USING(project_id)
 JOIN pulse_read.workflow_runs_v1 r ON r.workflow_run_id=m.workflow_run_id AND r.workflow_id=o.workflow_id
 WHERE m.workflow_run_id=$1 AND m.project_id=$2)`, runID, projectID).Scan(&allowed)
	return allowed, err
}

// TenantMiddleware runs after session authentication, including on public auth routes.
// Dedicated demo installations keep the outer process-wide policy.
func TenantMiddleware(p reader) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		restricted := Middleware(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if Enabled() {
				next.ServeHTTP(w, r)
				return
			}
			demo := false
			if claims, ok := authz.SessionClaimsFrom(r.Context()); ok {
				var err error
				demo, err = IsUser(r.Context(), p, claims.UserID)
				if err != nil {
					problem.Write(w, 503, "Не удалось проверить доступ", "")
					return
				}
			}
			if r.Method == http.MethodPost && r.URL.Path == "/api/v1/auth/register" {
				raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 64<<10))
				if err != nil {
					problem.Write(w, 413, "Слишком большой запрос", "")
					return
				}
				r.Body = io.NopCloser(bytes.NewReader(raw))
				var body struct {
					Email string `json:"email"`
				}
				if json.Unmarshal(raw, &body) == nil && strings.HasSuffix(strings.ToLower(strings.TrimSpace(body.Email)), "@demo.example.test") {
					demo = true
				}
			}
			restricted.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), tenantKey{}, demo)))
		})
	}
}
