package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireBrowserCSRF(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		authHeader string
		cookies    []http.Cookie
		csrfHeader string
		wantStatus int
	}{
		{
			name:       "safe method with cookie is allowed",
			method:     http.MethodGet,
			cookies:    []http.Cookie{{Name: accessCookieName, Value: "access"}},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "unsafe browser cookie request without header is rejected",
			method:     http.MethodPost,
			cookies:    []http.Cookie{{Name: accessCookieName, Value: "access"}, {Name: csrfCookieName, Value: "csrf"}},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "unsafe browser cookie request with matching header is allowed",
			method:     http.MethodPatch,
			cookies:    []http.Cookie{{Name: accessCookieName, Value: "access"}, {Name: csrfCookieName, Value: "csrf"}},
			csrfHeader: "csrf",
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "refresh cookie also requires csrf",
			method:     http.MethodPost,
			cookies:    []http.Cookie{{Name: refreshCookieName, Value: "refresh"}},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "explicit bearer request is not browser csrf gated",
			method:     http.MethodDelete,
			authHeader: "Bearer service-token",
			cookies:    []http.Cookie{{Name: accessCookieName, Value: "access"}},
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/protected", nil)
			for i := range tt.cookies {
				req.AddCookie(&tt.cookies[i])
			}
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			if tt.csrfHeader != "" {
				req.Header.Set(csrfHeaderName, tt.csrfHeader)
			}

			rec := httptest.NewRecorder()
			RequireBrowserCSRF(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
