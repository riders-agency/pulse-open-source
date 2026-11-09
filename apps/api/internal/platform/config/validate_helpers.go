package config

import (
	"net"
	"net/url"
	"strconv"
	"strings"
)

func (v *validator) validateDatabaseTLS(databaseURL string) {
	policy := strings.TrimSpace(v.getenv("PULSE_DB_TLS_POLICY"))
	if policy == "" {
		v.problem("PULSE_DB_TLS_POLICY", "required: local-only, verify-ca or verify-full")
		return
	}
	u, err := url.Parse(databaseURL)
	if err != nil || u.Scheme == "" {
		v.problem("DATABASE_URL", "must be a PostgreSQL URL with explicit TLS policy")
		return
	}
	sslmode := u.Query().Get("sslmode")
	switch policy {
	case "local-only":
		if !isLocalDBHost(u.Hostname()) || sslmode != "disable" {
			v.problem("DATABASE_URL", "local-only requires loopback host and sslmode=disable")
		}
	case "verify-ca", "verify-full":
		if sslmode != policy {
			v.problem("DATABASE_URL", "TLS policy requires matching sslmode")
		}
	default:
		v.problem("PULSE_DB_TLS_POLICY", "unsupported value")
	}
}

func (v *validator) forbidTrue(key string) {
	if boolEnv(v.getenv(key)) {
		v.problem(key, "must be false in production")
	}
}

func (v *validator) requireTrue(key string) {
	if !boolEnv(v.getenv(key)) {
		v.problem(key, "must be true in production")
	}
}

func (v *validator) requireHTTPSOrigin(key string) {
	value := v.requireNonPlaceholder(key, len("https://a.io"))
	u, err := url.Parse(value)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.RawQuery != "" || u.Fragment != "" {
		v.problem(key, "must be an https origin without query or fragment")
	}
}

func (v *validator) requireTrustedProxies(key string) {
	value := v.requireNonPlaceholder(key, 3)
	if strings.Contains(value, "*") || strings.EqualFold(value, "0.0.0.0/0") {
		v.problem(key, "must not be wildcard")
	}
}

func (v *validator) validateTopology(role StartupRole, topology string) {
	switch topology {
	case "embedded", "split-worker":
	case "":
		v.problem("PULSE_RUNTIME_TOPOLOGY", "required: embedded or split-worker")
	default:
		v.problem("PULSE_RUNTIME_TOPOLOGY", "unsupported value")
	}
	if role == RoleWorker && topology != "split-worker" {
		v.problem("PULSE_RUNTIME_TOPOLOGY", "worker requires split-worker topology")
	}
	if role == RoleAPI && topology != "split-worker" {
		v.problem("PULSE_RUNTIME_TOPOLOGY", "production api requires split-worker topology")
	}
}

func (v *validator) requirePositiveInt(key string) {
	value := strings.TrimSpace(v.getenv(key))
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		v.problem(key, "must be a positive integer")
	}
}

func (v *validator) requirePositiveSeconds(key string) {
	v.requirePositiveInt(key)
}

func isPlaceholder(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if strings.HasPrefix(normalized, "replace-with") {
		return true
	}
	switch normalized {
	case "", "change-me", "changeme", "replace-me", "password", "secret", "demo", "dev":
		return true
	default:
		return false
	}
}

func allZero(raw []byte) bool {
	for _, b := range raw {
		if b != 0 {
			return false
		}
	}
	return true
}

func boolEnv(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func isLocalDBHost(host string) bool {
	if host == "" || strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
