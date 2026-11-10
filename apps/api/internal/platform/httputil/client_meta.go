package httputil

import (
	"net"
	"net/http"
	"strings"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// ClientIP returns a host IP string suitable for sessions.ip_address (no port).
func ClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	if ip := chimiddleware.GetClientIP(r.Context()); ip != "" {
		return ip
	}
	addr := strings.TrimSpace(r.RemoteAddr)
	if addr == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

// ClientUserAgent returns the User-Agent header trimmed.
func ClientUserAgent(r *http.Request) string {
	if r == nil {
		return ""
	}
	return strings.TrimSpace(r.UserAgent())
}
