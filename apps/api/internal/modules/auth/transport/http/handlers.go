package http

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	pm "riders.agency/pulse/api/internal/modules/auth/service"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	"riders.agency/pulse/api/internal/platform/httputil"
	problem "riders.agency/pulse/api/internal/platform/problem"
)

func RegisterJSON(s *pm.Auth) http.HandlerFunc {
	type inBody struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var body inBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "invalid json", "")
			return
		}
		if body.Email == "" || body.Password == "" {
			problem.Write(w, http.StatusBadRequest, "missing credentials", "")
			return
		}
		pair, err := s.Register(r.Context(), body.Email, body.Password, body.DisplayName,
			httputil.ClientUserAgent(r), httputil.ClientIP(r))
		if err != nil {
			var pe *pgconn.PgError
			if errors.As(err, &pe) && pe.Code == "23505" {
				problem.Write(w, http.StatusConflict, "email exists", "")
				return
			}
			problem.Write(w, http.StatusInternalServerError, "register failed", "")
			return
		}
		setAuthCookies(w, r, pair)
		tokenJSON(w, pair)
	}
}

func LoginJSON(s *pm.Auth) http.HandlerFunc {
	type inBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var body inBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "invalid json", "")
			return
		}
		pair, err := s.Login(r.Context(), body.Email, body.Password,
			httputil.ClientUserAgent(r), httputil.ClientIP(r))
		if err != nil {
			problem.Write(w, http.StatusUnauthorized, "invalid credentials", "")
			return
		}
		setAuthCookies(w, r, pair)
		tokenJSON(w, pair)
	}
}

func RefreshJSON(s *pm.Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !mware.BrowserCSRFValid(r) {
			problem.Write(w, http.StatusForbidden, "csrf required", "")
			return
		}
		cookie, err := r.Cookie("rp_refresh")
		if err != nil || strings.TrimSpace(cookie.Value) == "" {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		pair, err := s.Refresh(r.Context(), cookie.Value)
		if err != nil {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		setAuthCookies(w, r, pair)
		tokenJSON(w, pair)
	}
}

func RefreshCSRFJSON(s *pm.Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("rp_refresh")
		if err != nil || strings.TrimSpace(cookie.Value) == "" {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		token, err := s.RefreshCSRFToken(r.Context(), cookie.Value)
		if err != nil {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		setCSRFCookie(w, r, token, time.Now().Add(s.RefreshTTL))
		problem.JSON(w, http.StatusOK, map[string]any{"csrf_token": token})
	}
}

func CSRFJSON(s *pm.Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := mware.SessionClaimsFrom(r.Context()); !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		token, err := newCSRFToken(s.CSRFRandBytes)
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "csrf failed", "")
			return
		}
		setCSRFCookie(w, r, token, time.Now().Add(s.RefreshTTL))
		problem.JSON(w, http.StatusOK, map[string]any{"csrf_token": token})
	}
}

func LogoutJSON(s *pm.Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		_ = s.Logout(r.Context(), c.SessionID)
		clearAuthCookies(w, r)
		w.WriteHeader(http.StatusNoContent)
	}
}

func MeJSON(s *pm.Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		p, err := s.Me(r.Context(), c.UserID)
		if err != nil {
			problem.Write(w, http.StatusNotFound, "user not found", "")
			return
		}
		problem.JSON(w, http.StatusOK, map[string]any{
			"user_id":      p.UserID.String(),
			"email":        p.Email,
			"display_name": p.DisplayName,
			"first_name":   p.FirstName,
		})
	}
}

func WorkspaceContextJSON(s *pm.Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		ctx, err := s.WorkspaceContext(r.Context(), c.UserID)
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "workspace context failed", "")
			return
		}
		problem.JSON(w, http.StatusOK, ctx)
	}
}

func PutWorkspaceContextJSON(s *pm.Auth) http.HandlerFunc {
	type inBody struct {
		ActiveCompanyID *string `json:"active_company_id"`
		ActiveProjectID *string `json:"active_project_id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		var body inBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "invalid json", "")
			return
		}
		companyID, err := optionalUUID(body.ActiveCompanyID)
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "bad company id", "")
			return
		}
		projectID, err := optionalUUID(body.ActiveProjectID)
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "bad project id", "")
			return
		}
		err = s.PutWorkspaceContext(r.Context(), c.UserID, pm.WorkspaceContextUpdate{
			ActiveCompanyID: companyID,
			ActiveProjectID: projectID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			problem.Write(w, http.StatusForbidden, "workspace context forbidden", "")
			return
		}
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "workspace context update failed", "")
			return
		}
		ctx, err := s.WorkspaceContext(r.Context(), c.UserID)
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "workspace context failed", "")
			return
		}
		problem.JSON(w, http.StatusOK, ctx)
	}
}

func SessionsJSON(s *pm.Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		rows, err := s.ListSessions(r.Context(), c.UserID)
		if err != nil {
			problem.Write(w, http.StatusInternalServerError, "sessions list failed", "")
			return
		}
		out := make([]map[string]any, 0, len(rows))
		for _, row := range rows {
			entry := map[string]any{
				"session_id": row.SessionID.String(),
				"created_at": row.CreatedAt.UTC().Format(time.RFC3339Nano),
				"expires_at": row.ExpiresAt.UTC().Format(time.RFC3339Nano),
				"current":    row.SessionID == c.SessionID,
			}
			if row.UserAgent != nil {
				entry["user_agent"] = *row.UserAgent
			}
			if row.IP != nil {
				entry["ip"] = *row.IP
			}
			out = append(out, entry)
		}
		problem.JSON(w, http.StatusOK, map[string]any{"sessions": out})
	}
}

func EndSessionJSON(s *pm.Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		sidStr := chi.URLParam(r, "session_id")
		target, err := uuid.Parse(sidStr)
		if err != nil || target == uuid.Nil {
			problem.Write(w, http.StatusBadRequest, "bad session id", "")
			return
		}
		err = s.EndOtherSession(r.Context(), c.UserID, target, c.SessionID)
		switch {
		case errors.Is(err, pm.ErrCannotEndCurrentSession):
			problem.Write(w, http.StatusConflict, "use logout", "")
			return
		case errors.Is(err, pm.ErrSessionNotFound):
			problem.Write(w, http.StatusNotFound, "session not found", "")
			return
		case err != nil:
			problem.Write(w, http.StatusInternalServerError, "revoke failed", "")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func PatchProfileJSON(s *pm.Auth) http.HandlerFunc {
	type inBody struct {
		DisplayName *string `json:"display_name"`
		FirstName   *string `json:"first_name"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		var body inBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			problem.Write(w, http.StatusBadRequest, "invalid json", "")
			return
		}
		if body.DisplayName == nil && body.FirstName == nil {
			problem.Write(w, http.StatusBadRequest, "empty patch", "")
			return
		}
		if err := s.PatchProfile(r.Context(), c.UserID, pm.ProfileUpdate{
			DisplayName: body.DisplayName,
			FirstName:   body.FirstName,
		}); err != nil {
			problem.Write(w, http.StatusInternalServerError, "profile update failed", "")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func optionalUUID(value *string) (*uuid.UUID, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	id, err := uuid.Parse(strings.TrimSpace(*value))
	if err != nil || id == uuid.Nil {
		return nil, err
	}
	return &id, nil
}

func tokenJSON(w http.ResponseWriter, pair *pm.TokenPair) {
	problem.JSON(w, http.StatusOK, map[string]any{
		"access_token":       pair.AccessToken,
		"token_type":         "Bearer",
		"expires_in":         maxAge(pair.AccessExpiresAt),
		"csrf_token":         pair.CSRFToken,
		"access_expires_at":  pair.AccessExpiresAt.UTC().Format(time.RFC3339Nano),
		"refresh_expires_at": pair.RefreshExpiresAt.UTC().Format(time.RFC3339Nano),
	})
}

func setAuthCookies(w http.ResponseWriter, r *http.Request, pair *pm.TokenPair) {
	secure := isSecureRequest(r)
	http.SetCookie(w, &http.Cookie{
		Name:     "rp_access",
		Value:    pair.AccessToken,
		Path:     "/",
		MaxAge:   maxAge(pair.AccessExpiresAt),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "rp_refresh",
		Value:    pair.RefreshToken,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   maxAge(pair.RefreshExpiresAt),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "rp_csrf",
		Value:    pair.CSRFToken,
		Path:     "/",
		MaxAge:   maxAge(pair.RefreshExpiresAt),
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func setCSRFCookie(w http.ResponseWriter, r *http.Request, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "rp_csrf",
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge(expiresAt),
		HttpOnly: false,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func newCSRFToken(size int) (string, error) {
	if size < 16 {
		size = 32
	}
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func clearAuthCookies(w http.ResponseWriter, r *http.Request) {
	secure := isSecureRequest(r)
	for _, cookie := range []http.Cookie{
		{Name: "rp_access", Path: "/", MaxAge: -1, HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode},
		{Name: "rp_refresh", Path: "/api/v1/auth/refresh", MaxAge: -1, HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode},
		{Name: "rp_csrf", Path: "/", MaxAge: -1, HttpOnly: false, Secure: secure, SameSite: http.SameSiteLaxMode},
	} {
		http.SetCookie(w, &cookie)
	}
}

func maxAge(t time.Time) int {
	seconds := int(time.Until(t).Seconds())
	if seconds < 1 {
		return 1
	}
	return seconds
}

func isSecureRequest(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}
