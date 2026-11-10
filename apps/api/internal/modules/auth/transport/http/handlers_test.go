package http

import (
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	pm "riders.agency/pulse/api/internal/modules/auth/service"
	"riders.agency/pulse/api/internal/platform/authz"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
)

func TestSetAuthCookiesIncludesHttpOnlyTokensAndReadableCSRF(t *testing.T) {
	pair := &pm.TokenPair{
		AccessToken:      "access-token",
		RefreshToken:     "refresh-token",
		CSRFToken:        "csrf-token",
		AccessExpiresAt:  time.Now().Add(15 * time.Minute),
		RefreshExpiresAt: time.Now().Add(24 * time.Hour),
	}
	req := httptest.NewRequest(stdhttp.MethodPost, "/api/v1/auth/login", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()

	setAuthCookies(rec, req, pair)

	cookies := map[string]*stdhttp.Cookie{}
	for _, cookie := range rec.Result().Cookies() {
		cookies[cookie.Name] = cookie
	}
	access := cookies["rp_access"]
	refresh := cookies["rp_refresh"]
	csrf := cookies["rp_csrf"]
	if access == nil || refresh == nil || csrf == nil {
		t.Fatalf("expected access, refresh and csrf cookies, got %#v", cookies)
	}
	if !access.HttpOnly || !refresh.HttpOnly {
		t.Fatal("access and refresh cookies must be HttpOnly")
	}
	if csrf.HttpOnly {
		t.Fatal("csrf cookie must be readable by browser code")
	}
	if access.SameSite != stdhttp.SameSiteLaxMode || refresh.SameSite != stdhttp.SameSiteLaxMode || csrf.SameSite != stdhttp.SameSiteLaxMode {
		t.Fatal("auth cookies must use SameSite=Lax")
	}
	if !access.Secure || !refresh.Secure || !csrf.Secure {
		t.Fatal("auth cookies must be Secure behind https proxy")
	}
	if refresh.Path != "/api/v1/auth/refresh" {
		t.Fatalf("refresh path = %q", refresh.Path)
	}
	if csrf.Path != "/" {
		t.Fatalf("csrf path = %q", csrf.Path)
	}
}

func TestCSRFJSONIssuesReadableCSRFCookie(t *testing.T) {
	auth := &pm.Auth{RefreshTTL: 24 * time.Hour, CSRFRandBytes: 32}
	req := httptest.NewRequest(stdhttp.MethodGet, "/api/v1/auth/csrf", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req = req.WithContext(mware.WithSessionClaims(req.Context(), mware.SessionClaims{
		SessionID: uuid.New(),
		UserID:    uuid.New(),
		ActorType: authz.ActorSession,
	}))
	rec := httptest.NewRecorder()

	CSRFJSON(auth).ServeHTTP(rec, req)

	if rec.Code != stdhttp.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, stdhttp.StatusOK)
	}
	var body struct {
		CSRFToken string `json:"csrf_token"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.CSRFToken == "" {
		t.Fatal("csrf_token is required")
	}
	cookies := map[string]*stdhttp.Cookie{}
	for _, cookie := range rec.Result().Cookies() {
		cookies[cookie.Name] = cookie
	}
	csrf := cookies["rp_csrf"]
	if csrf == nil {
		t.Fatalf("expected rp_csrf cookie, got %#v", cookies)
	}
	if csrf.Value != body.CSRFToken {
		t.Fatalf("csrf cookie value does not match response token")
	}
	if csrf.HttpOnly {
		t.Fatal("csrf cookie must be readable by browser code")
	}
	if !csrf.Secure {
		t.Fatal("csrf cookie must be Secure behind https proxy")
	}
	if csrf.Path != "/" {
		t.Fatalf("csrf path = %q", csrf.Path)
	}
}
