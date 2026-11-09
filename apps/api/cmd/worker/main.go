package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"riders.agency/pulse/api/internal/platform/app"
	"riders.agency/pulse/api/internal/platform/config"
	"riders.agency/pulse/api/internal/platform/outbox"
)

func main() {
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(filepath.Join("..", "..", ".env.local"))

	cfg := config.Load()
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	if err := config.ValidateStartup(config.RoleWorker, cfg); err != nil {
		slog.Error("config validation", slog.Any("err", err))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	workerCfg := cfg
	workerCfg.SkipMigrations = true
	application, err := app.Build(ctx, workerCfg)
	if err != nil {
		slog.Error("app build", slog.Any("err", err))
		os.Exit(1)
	}
	defer application.Close()

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	worker := &outbox.Worker{
		Repo:          &outbox.Repository{P: application.Deps.DB},
		Registry:      application.OutboxRegistry,
		Logger:        slog.Default(),
		BatchSize:     intFromEnv("PULSE_OUTBOX_BATCH_SIZE", 10),
		LeaseDuration: durationFromEnv("PULSE_OUTBOX_LEASE_SECONDS", 60*time.Second),
		PollInterval:  durationFromEnv("PULSE_OUTBOX_POLL_SECONDS", 2*time.Second),
	}

	errCh := make(chan error, 3)
	var wg sync.WaitGroup
	start := func(name string, fn func(context.Context) error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			slog.Info("worker component starting", slog.String("component", name))
			if err := fn(runCtx); err != nil && runCtx.Err() == nil {
				errCh <- err
				cancel()
			}
			slog.Info("worker component stopped", slog.String("component", name))
		}()
	}
	if application.RuntimeKernel != nil {
		start("runtime_kernel", func(ctx context.Context) error {
			application.RuntimeKernel.Run(ctx)
			return nil
		})
	}
	if application.TelegramPoller != nil {
		start("telegram_poller", func(ctx context.Context) error {
			application.TelegramPoller.Run(ctx)
			return nil
		})
	}
	start("outbox", worker.Run)

	var runErr error
	select {
	case <-ctx.Done():
		cancel()
	case runErr = <-errCh:
		cancel()
	}
	wg.Wait()
	if runErr != nil {
		slog.Error("worker stopped with error", slog.Any("err", runErr))
		os.Exit(1)
	}
	slog.Info("worker stopped")
}

func intFromEnv(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func durationFromEnv(key string, fallback time.Duration) time.Duration {
	seconds := intFromEnv(key, int(fallback/time.Second))
	return time.Duration(seconds) * time.Second
}
