package middleware

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/httputil"
	problem "riders.agency/pulse/api/internal/platform/problem"
)

const (
	accessCookieName  = "rp_access"
	refreshCookieName = "rp_refresh"
	csrfCookieName    = "rp_csrf"
	csrfHeaderName    = "X-CSRF-Token"
)

func OptionalAuth(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := httputil.Bearer(r.Header.Get("Authorization"))
			ctx := r.Context()
			if strings.HasPrefix(token, "rp_live_") {
				ctx = attachAPIKey(ctx, pool, token)
			} else {
				if token == "" {
					if cookie, err := r.Cookie(accessCookieName); err == nil {
						token = strings.TrimSpace(cookie.Value)
					}
				}
				if token == "" && r.URL.Path == "/api/v1/site-chat/ws/operator" {
					token = strings.TrimSpace(r.URL.Query().Get("access_token"))
				}
				ctx = attachSession(ctx, pool, token)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAuth(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := SessionClaimsFrom(r.Context())
			if !ok {
				problem.Write(w, http.StatusUnauthorized, "Unauthorized", "")
				return
			}
			if claims.ActorType == authz.ActorAPIKey {
				if err := authorizeAPIKeyRequest(r, pool); err != nil {
					problem.Write(w, http.StatusForbidden, "forbidden", "")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireBrowserCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !BrowserCSRFValid(r) {
			problem.Write(w, http.StatusForbidden, "csrf required", "")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func BrowserCSRFValid(r *http.Request) bool {
	if !RequiresBrowserCSRF(r) {
		return true
	}
	csrfCookie, err := r.Cookie(csrfCookieName)
	if err != nil {
		return false
	}
	header := strings.TrimSpace(r.Header.Get(csrfHeaderName))
	cookieValue := strings.TrimSpace(csrfCookie.Value)
	if header == "" || cookieValue == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(header), []byte(cookieValue)) == 1
}

func RequiresBrowserCSRF(r *http.Request) bool {
	if isSafeMethod(r.Method) || strings.TrimSpace(r.Header.Get("Authorization")) != "" {
		return false
	}
	return hasCookie(r, accessCookieName) || hasCookie(r, refreshCookieName)
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

func hasCookie(r *http.Request, name string) bool {
	cookie, err := r.Cookie(name)
	return err == nil && strings.TrimSpace(cookie.Value) != ""
}

func attachSession(ctx context.Context, pool *pgxpool.Pool, wire string) context.Context {
	if wire == "" || pool == nil {
		return ctx
	}
	dec, err := base64.RawURLEncoding.DecodeString(wire)
	if err != nil {
		return ctx
	}
	parts := strings.SplitN(string(dec), "|", 3)
	if len(parts) != 3 || parts[0] != "pulse" {
		return ctx
	}
	sid, err := uuid.Parse(parts[1])
	if err != nil {
		return ctx
	}
	secBytes, err := hex.DecodeString(parts[2])
	if err != nil {
		return ctx
	}
	row := pool.QueryRow(ctx, `
SELECT user_id::text, session_token_hash::text
  FROM sessions
 WHERE session_id = $1
   AND revoked_at IS NULL
   AND (access_expires_at IS NULL OR access_expires_at > now())
   AND expires_at > now()`, sid.String())
	var uidStr, hash string
	if err := row.Scan(&uidStr, &hash); err != nil {
		return ctx
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), secBytes) != nil {
		return ctx
	}
	u, err := uuid.Parse(uidStr)
	if err != nil {
		slog.Warn("bad user uuid session", slog.String("session_id", sid.String()))
		return ctx
	}
	return WithSessionClaims(ctx, SessionClaims{
		SessionID: sid,
		UserID:    u,
		Plaintext: hex.EncodeToString(secBytes),
		ActorType: authz.ActorSession,
	})
}

func attachAPIKey(ctx context.Context, pool *pgxpool.Pool, token string) context.Context {
	if pool == nil || token == "" {
		return ctx
	}
	prefix, ok := apiKeyPrefix(token)
	if !ok {
		return ctx
	}
	row := pool.QueryRow(ctx, `
SELECT api_key_id::text, user_id::text, key_hash::text
  FROM api_keys
 WHERE key_prefix = $1
   AND status = 'active'
   AND revoked_at IS NULL
   AND (expires_at IS NULL OR expires_at > now())`, prefix)
	var keyIDStr, userIDStr, hash string
	if err := row.Scan(&keyIDStr, &userIDStr, &hash); err != nil {
		return ctx
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(token)) != nil {
		return ctx
	}
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		return ctx
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return ctx
	}
	grants, err := loadAPIKeyGrants(ctx, pool, keyID)
	if err != nil {
		return ctx
	}
	_, _ = pool.Exec(ctx, `UPDATE api_keys SET last_used_at = now() WHERE api_key_id = $1`, keyID.String())
	return WithSessionClaims(ctx, SessionClaims{
		UserID:       userID,
		ActorType:    authz.ActorAPIKey,
		APIKeyID:     keyID,
		APIKeyPrefix: prefix,
		Grants:       grants,
	})
}

func apiKeyPrefix(token string) (string, bool) {
	parts := strings.Split(token, "_")
	if len(parts) != 4 || parts[0] != "rp" || parts[1] != "live" || parts[2] == "" || parts[3] == "" {
		return "", false
	}
	return parts[2], true
}

func loadAPIKeyGrants(ctx context.Context, pool *pgxpool.Pool, keyID uuid.UUID) ([]authz.Grant, error) {
	rows, err := pool.Query(ctx, `
SELECT scope_kind::text, COALESCE(scope_id::text, ''), permissions
  FROM api_key_grants
 WHERE api_key_id = $1`, keyID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	grants := make([]authz.Grant, 0)
	for rows.Next() {
		var grant authz.Grant
		var scopeID string
		if err := rows.Scan(&grant.ScopeKind, &scopeID, &grant.Permissions); err != nil {
			return nil, err
		}
		if scopeID != "" {
			id, err := uuid.Parse(scopeID)
			if err != nil {
				return nil, err
			}
			grant.ScopeID = id
		}
		grants = append(grants, grant)
	}
	return grants, rows.Err()
}

func authorizeAPIKeyRequest(r *http.Request, pool *pgxpool.Pool) error {
	permission := authz.PermRead
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		permission = authz.PermRead
	default:
		permission = authz.PermWrite
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 || parts[0] != "api" || parts[1] != "v1" {
		return nil
	}
	if len(parts) >= 4 {
		id, err := uuid.Parse(parts[3])
		if err == nil {
			switch parts[2] {
			case "companies":
				return authz.RequireCompany(r.Context(), pool, id, permission)
			case "projects":
				return authz.RequireProject(r.Context(), pool, id, permission)
			case "workflows":
				return authz.RequireWorkflow(r.Context(), pool, id, permission)
			case "clients":
				return authz.RequireClient(r.Context(), pool, id, permission)
			case "client-services":
				return authz.RequireClientService(r.Context(), pool, id, permission)
			case "payment-actions":
				return authz.RequirePaymentAction(r.Context(), pool, id, permission)
			case "runs":
				return authz.RequireRun(r.Context(), pool, id, permission)
			case "webhook-endpoints":
				return authz.RequireWebhookEndpoint(r.Context(), pool, id, permission)
			case "webhook-events":
				return authz.RequireWebhookEvent(r.Context(), pool, id, permission)
			}
		}
	}
	return nil
}
