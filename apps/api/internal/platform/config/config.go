package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultSessionTTL = 15 * time.Minute
	defaultRefreshTTL = 3650 * 24 * time.Hour
)

type Config struct {
	HTTPAddr              string
	ToolGRPCAddr          string
	ModelGRPCAddr         string
	DatabaseURL           string
	MigrationsPath        string
	Environment           string
	RuntimeTopology       string
	SkipMigrations        bool
	SessionTTL            time.Duration
	RefreshTTL            time.Duration
	HTTPReadHeaderTimeout time.Duration
	HTTPShutdownTimeout   time.Duration
	ExecutorWorkers       int
}

func Load() Config {
	return Config{
		HTTPAddr:        firstNonEmpty(os.Getenv("HTTP_ADDR"), ":8080"),
		ToolGRPCAddr:    firstNonEmpty(os.Getenv("PULSE_TOOL_GRPC_ADDR"), "127.0.0.1:9098"),
		ModelGRPCAddr:   firstNonEmpty(os.Getenv("PULSE_MODEL_GRPC_ADDR"), "127.0.0.1:9099"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		MigrationsPath:  firstNonEmpty(os.Getenv("MIGRATIONS_PATH"), "db/migrations"),
		Environment:     firstNonEmpty(os.Getenv("RP_ENV"), "development"),
		RuntimeTopology: firstNonEmpty(os.Getenv("PULSE_RUNTIME_TOPOLOGY"), "embedded"),
		SkipMigrations:  strings.EqualFold(os.Getenv("SKIP_MIGRATIONS"), "true"),
		SessionTTL:      durationSecondsEnv("PULSE_SESSION_TTL_SECONDS", defaultSessionTTL),
		RefreshTTL:      durationDaysEnv("PULSE_REFRESH_TTL_DAYS", defaultRefreshTTL),
		HTTPReadHeaderTimeout: durationSecondsEnv(
			"PULSE_HTTP_READ_HEADER_TIMEOUT_SECONDS",
			10*time.Second,
		),
		HTTPShutdownTimeout: durationSecondsEnv(
			"PULSE_HTTP_SHUTDOWN_TIMEOUT_SECONDS",
			15*time.Second,
		),
		ExecutorWorkers: 2,
	}
}

func (c Config) EmbeddedRuntimeEnabled() bool {
	return !strings.EqualFold(c.RuntimeTopology, "split-worker")
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func durationSecondsEnv(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	seconds, err := time.ParseDuration(value + "s")
	if err != nil || seconds <= 0 {
		return fallback
	}
	return seconds
}

func durationDaysEnv(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	days, err := strconv.Atoi(value)
	if err != nil || days <= 0 {
		return fallback
	}
	return time.Duration(days) * 24 * time.Hour
}
