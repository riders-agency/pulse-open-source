package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/joho/godotenv"

	"riders.agency/pulse/api/internal/platform/app"
	"riders.agency/pulse/api/internal/platform/config"
)

func main() {
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(filepath.Join("..", "..", ".env.local"))

	cfg := config.Load()
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	if err := config.ValidateStartup(config.RoleAPI, cfg); err != nil {
		slog.Error("config validation", slog.Any("err", err))
		os.Exit(1)
	}

	ctx := context.Background()
	application, err := app.Build(ctx, cfg)
	if err != nil {
		slog.Error("app build", slog.Any("err", err))
		os.Exit(1)
	}
	defer application.Close()

	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()
	stopToolGateway, err := application.StartToolGatewayGRPC(appCtx)
	if err != nil {
		slog.Error("tool gateway grpc", slog.Any("err", err))
		os.Exit(1)
	}
	defer stopToolGateway()
	stopModelGateway, err := application.StartModelGatewayGRPC(appCtx)
	if err != nil {
		slog.Error("model gateway grpc", slog.Any("err", err))
		os.Exit(1)
	}
	defer stopModelGateway()
	if cfg.EmbeddedRuntimeEnabled() {
		application.StartBackground(appCtx)
	}

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           application.Handler,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
	}

	go func() {
		slog.Info("listen", slog.String("addr", cfg.HTTPAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	sig := <-ch
	slog.Info("shutdown signal", slog.String("signal", sig.String()))
	appCancel()
	shCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTPShutdownTimeout)
	defer cancel()
	_ = srv.Shutdown(shCtx)
}
