package demomode

import (
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"riders.agency/pulse/api/internal/platform/problem"
)

func Enabled() bool { return os.Getenv("PULSE_PUBLIC_DEMO") == "true" }

// Middleware bounds requests in a process-wide window, including spoofed forwarding headers.
func Middleware(next http.Handler) http.Handler {
	var mu sync.Mutex
	var minute, hour time.Time
	var writes, registrations int
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !Enabled() {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("X-Pulse-Demo", "true")
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if !allowedWrite(r.URL.Path) {
			problem.Write(w, http.StatusForbidden, "Действие недоступно в демо", "Можно работать с задачами, клиентами и ИИ. Внешние отправки и настройки подключений отключены.")
			return
		}
		mu.Lock()
		now := time.Now()
		if now.Sub(minute) >= time.Minute {
			minute = now
			writes = 0
		}
		if now.Sub(hour) >= time.Hour {
			hour = now
			registrations = 0
		}
		register := r.URL.Path == "/api/v1/auth/register"
		limited := writes >= 180 || (register && registrations >= 30)
		if !limited {
			writes++
			if register {
				registrations++
			}
		}
		mu.Unlock()
		if limited {
			w.Header().Set("Retry-After", "60")
			problem.Write(w, 429, "Демо временно занято", "Повторите попытку позже.")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
		next.ServeHTTP(w, r)
	})
}

func allowedWrite(path string) bool {
	for _, part := range []string{"/attachments", "/upload", "/runs", "/run", "/replay", "/publish", "/execute", "/action-dependencies", "/integrations", "/secrets", "/api-keys", "/webhook", "/site-chat", "/models"} {
		if strings.Contains(path, part) {
			return false
		}
	}

	for _, allowed := range []string{"/api/v1/auth/register", "/api/v1/auth/login", "/api/v1/auth/logout", "/api/v1/auth/refresh", "/api/v1/auth/workspace-context", "/api/v1/companies"} {
		if path == allowed {
			return true
		}
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 5 && parts[2] == "companies" && (parts[4] == "projects" || parts[4] == "switch") {
		return true
	}
	if len(parts) >= 5 && parts[2] == "projects" {
		if parts[4] == "ai" {
			return true
		}
		if len(parts) == 5 && (parts[4] == "clients" || parts[4] == "workflows") {
			return true
		}
	}
	for _, prefix := range []string{"/api/v1/work/", "/api/v1/clients/", "/api/v1/client-services/", "/api/v1/payment-actions/", "/api/v1/workflows/"} {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
