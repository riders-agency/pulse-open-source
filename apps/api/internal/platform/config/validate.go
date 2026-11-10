package config

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

const SchemaVersion = "pulse.config.v1"

type StartupRole string

const (
	RoleAPI    StartupRole = "api"
	RoleWorker StartupRole = "worker"
)

type ValidationProblem struct {
	Key     string
	Message string
}

type ValidationError struct {
	Schema   string
	Role     StartupRole
	Problems []ValidationProblem
}

func (e *ValidationError) Error() string {
	parts := make([]string, 0, len(e.Problems))
	for _, problem := range e.Problems {
		parts = append(parts, problem.Key+": "+problem.Message)
	}
	return fmt.Sprintf("production config invalid schema=%s role=%s: %s", e.Schema, e.Role, strings.Join(parts, "; "))
}

func ValidateStartup(role StartupRole, cfg Config) error {
	return ValidateStartupEnv(role, cfg, os.Getenv)
}

func ValidateStartupEnv(role StartupRole, cfg Config, getenv func(string) string) error {
	if !strings.EqualFold(cfg.Environment, "production") {
		return nil
	}
	v := validator{getenv: getenv}
	v.requireStrongHexKey("RIDERS_PULSE_MASTER_KEY")
	v.requireNonPlaceholder("RIDERS_PULSE_MASTER_KEY_ID", 3)
	v.validateDatabaseTLS(cfg.DatabaseURL)
	v.forbidTrue("AI_AGENT_INSECURE_DEV")
	v.forbidTrue("AI_AGENT_INSECURE_BIND_ALL")
	v.forbidTrue("PULSE_TOOL_GRPC_INSECURE_DEV")
	v.forbidTrue("PULSE_TOOL_GRPC_INSECURE_ALLOW_NONLOCAL")
	v.requireNonPlaceholder("PULSE_INTERNAL_SERVICE_ID", 6)
	v.requireStrongSecret("AI_AGENT_SERVICE_TOKEN", 32)
	v.requireHTTPSOrigin("PULSE_PUBLIC_ORIGIN")
	v.requireTrue("PULSE_COOKIE_SECURE")
	v.requireTrustedProxies("PULSE_TRUSTED_PROXIES")
	v.validateTopology(role, cfg.RuntimeTopology)
	v.requirePositiveSeconds("PULSE_SESSION_TTL_SECONDS")
	v.requirePositiveInt("PULSE_REFRESH_TTL_DAYS")
	v.requirePositiveSeconds("PULSE_HTTP_READ_HEADER_TIMEOUT_SECONDS")
	v.requirePositiveSeconds("PULSE_HTTP_SHUTDOWN_TIMEOUT_SECONDS")
	if role == RoleWorker {
		v.requirePositiveInt("PULSE_OUTBOX_BATCH_SIZE")
		v.requirePositiveSeconds("PULSE_OUTBOX_LEASE_SECONDS")
		v.requirePositiveSeconds("PULSE_OUTBOX_POLL_SECONDS")
	}
	if len(v.problems) > 0 {
		return &ValidationError{Schema: SchemaVersion, Role: role, Problems: v.problems}
	}
	return nil
}

type validator struct {
	getenv   func(string) string
	problems []ValidationProblem
}

func (v *validator) problem(key string, message string) {
	v.problems = append(v.problems, ValidationProblem{Key: key, Message: message})
}

func (v *validator) requireNonPlaceholder(key string, minLen int) string {
	value := strings.TrimSpace(v.getenv(key))
	if len(value) < minLen || isPlaceholder(value) {
		v.problem(key, "required and must not be a placeholder")
	}
	return value
}

func (v *validator) requireStrongSecret(key string, minLen int) {
	value := v.requireNonPlaceholder(key, minLen)
	if len(value) > 0 && len(value) < minLen {
		v.problem(key, fmt.Sprintf("must be at least %d characters", minLen))
	}
}

func (v *validator) requireStrongHexKey(key string) {
	value := strings.TrimSpace(v.getenv(key))
	raw, err := hex.DecodeString(value)
	if err != nil || len(raw) != 32 || allZero(raw) {
		v.problem(key, "must be a non-zero 32-byte hex key")
	}
}
