package config

import (
	"strings"
	"testing"
)

func TestValidateStartupSkipsDevelopment(t *testing.T) {
	err := ValidateStartupEnv(RoleAPI, Config{Environment: "development"}, func(string) string { return "" })
	if err != nil {
		t.Fatalf("development config should not be fail-closed: %v", err)
	}
}

func TestValidateStartupProductionNegativeFixtures(t *testing.T) {
	tests := []struct {
		name    string
		role    StartupRole
		mutate  func(map[string]string)
		wantKey string
	}{
		{
			name: "missing master key",
			mutate: func(env map[string]string) {
				delete(env, "RIDERS_PULSE_MASTER_KEY")
			},
			wantKey: "RIDERS_PULSE_MASTER_KEY",
		},
		{
			name: "weak master key id",
			mutate: func(env map[string]string) {
				env["RIDERS_PULSE_MASTER_KEY_ID"] = "dev"
			},
			wantKey: "RIDERS_PULSE_MASTER_KEY_ID",
		},
		{
			name: "database tls disabled remotely",
			mutate: func(env map[string]string) {
				env["DATABASE_URL"] = "postgres://pulse@10.0.0.2:5432/pulse?sslmode=disable"
				env["PULSE_DB_TLS_POLICY"] = "verify-full"
			},
			wantKey: "DATABASE_URL",
		},
		{
			name: "ai insecure dev",
			mutate: func(env map[string]string) {
				env["AI_AGENT_INSECURE_DEV"] = "true"
			},
			wantKey: "AI_AGENT_INSECURE_DEV",
		},
		{
			name: "missing service identity",
			mutate: func(env map[string]string) {
				delete(env, "PULSE_INTERNAL_SERVICE_ID")
			},
			wantKey: "PULSE_INTERNAL_SERVICE_ID",
		},
		{
			name: "placeholder service token",
			mutate: func(env map[string]string) {
				env["AI_AGENT_SERVICE_TOKEN"] = "change-me"
			},
			wantKey: "AI_AGENT_SERVICE_TOKEN",
		},
		{
			name: "public origin must be https",
			mutate: func(env map[string]string) {
				env["PULSE_PUBLIC_ORIGIN"] = "http://pulse.example.test"
			},
			wantKey: "PULSE_PUBLIC_ORIGIN",
		},
		{
			name: "cookies secure",
			mutate: func(env map[string]string) {
				env["PULSE_COOKIE_SECURE"] = "false"
			},
			wantKey: "PULSE_COOKIE_SECURE",
		},
		{
			name: "wildcard proxies",
			mutate: func(env map[string]string) {
				env["PULSE_TRUSTED_PROXIES"] = "0.0.0.0/0"
			},
			wantKey: "PULSE_TRUSTED_PROXIES",
		},
		{
			name: "worker requires split worker",
			role: RoleWorker,
			mutate: func(env map[string]string) {
				env["PULSE_RUNTIME_TOPOLOGY"] = "embedded"
			},
			wantKey: "PULSE_RUNTIME_TOPOLOGY",
		},
		{
			name: "api production rejects embedded worker",
			role: RoleAPI,
			mutate: func(env map[string]string) {
				env["PULSE_RUNTIME_TOPOLOGY"] = "embedded"
			},
			wantKey: "PULSE_RUNTIME_TOPOLOGY",
		},
		{
			name: "missing explicit http timeout",
			mutate: func(env map[string]string) {
				delete(env, "PULSE_HTTP_READ_HEADER_TIMEOUT_SECONDS")
			},
			wantKey: "PULSE_HTTP_READ_HEADER_TIMEOUT_SECONDS",
		},
		{
			name: "missing explicit session ttl",
			mutate: func(env map[string]string) {
				delete(env, "PULSE_SESSION_TTL_SECONDS")
			},
			wantKey: "PULSE_SESSION_TTL_SECONDS",
		},
		{
			name: "missing explicit refresh ttl",
			mutate: func(env map[string]string) {
				delete(env, "PULSE_REFRESH_TTL_DAYS")
			},
			wantKey: "PULSE_REFRESH_TTL_DAYS",
		},
		{
			name: "worker missing outbox limit",
			role: RoleWorker,
			mutate: func(env map[string]string) {
				delete(env, "PULSE_OUTBOX_BATCH_SIZE")
			},
			wantKey: "PULSE_OUTBOX_BATCH_SIZE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := validProductionEnv()
			tt.mutate(env)
			role := tt.role
			if role == "" {
				role = RoleAPI
			}
			cfg := configFromEnv(env)
			err := ValidateStartupEnv(role, cfg, func(key string) string { return env[key] })
			if err == nil {
				t.Fatalf("expected production config to fail")
			}
			message := err.Error()
			if !strings.Contains(message, tt.wantKey) {
				t.Fatalf("expected %s in error, got %q", tt.wantKey, message)
			}
			if strings.Contains(message, env["AI_AGENT_SERVICE_TOKEN"]) {
				t.Fatalf("error leaked secret value: %q", message)
			}
		})
	}
}

func TestValidateStartupProductionPositiveFixtures(t *testing.T) {
	env := validProductionEnv()
	if err := ValidateStartupEnv(RoleAPI, configFromEnv(env), func(key string) string { return env[key] }); err != nil {
		t.Fatalf("valid api config failed: %v", err)
	}
	if err := ValidateStartupEnv(RoleWorker, configFromEnv(env), func(key string) string { return env[key] }); err != nil {
		t.Fatalf("valid worker config failed: %v", err)
	}
}

func validProductionEnv() map[string]string {
	return map[string]string{
		"AI_AGENT_SERVICE_TOKEN":                 strings.Repeat("a", 40),
		"DATABASE_URL":                           "postgres://pulse@127.0.0.1:5432/pulse?sslmode=disable",
		"PULSE_COOKIE_SECURE":                    "true",
		"PULSE_DB_TLS_POLICY":                    "local-only",
		"PULSE_HTTP_READ_HEADER_TIMEOUT_SECONDS": "10",
		"PULSE_HTTP_SHUTDOWN_TIMEOUT_SECONDS":    "15",
		"PULSE_INTERNAL_SERVICE_ID":              "pulse-api",
		"PULSE_OUTBOX_BATCH_SIZE":                "10",
		"PULSE_OUTBOX_LEASE_SECONDS":             "60",
		"PULSE_OUTBOX_POLL_SECONDS":              "2",
		"PULSE_PUBLIC_ORIGIN":                    "https://pulse.example.test",
		"PULSE_RUNTIME_TOPOLOGY":                 "split-worker",
		"PULSE_SESSION_TTL_SECONDS":              "900",
		"PULSE_REFRESH_TTL_DAYS":                 "3650",
		"PULSE_TRUSTED_PROXIES":                  "127.0.0.1/32,::1/128",
		"RIDERS_PULSE_MASTER_KEY":                strings.Repeat("a", 64),
		"RIDERS_PULSE_MASTER_KEY_ID":             "prod-2026-06",
	}
}

func configFromEnv(env map[string]string) Config {
	return Config{
		DatabaseURL:           env["DATABASE_URL"],
		Environment:           "production",
		RuntimeTopology:       env["PULSE_RUNTIME_TOPOLOGY"],
		HTTPReadHeaderTimeout: 10,
		HTTPShutdownTimeout:   15,
	}
}
