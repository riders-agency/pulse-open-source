package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"riders.agency/pulse/api/internal/executor"
	aitools "riders.agency/pulse/api/internal/modules/ai/tools"
	aihttp "riders.agency/pulse/api/internal/modules/ai/transport/http"
	authrepo "riders.agency/pulse/api/internal/modules/auth/repository"
	authsvc "riders.agency/pulse/api/internal/modules/auth/service"
	maxprovider "riders.agency/pulse/api/internal/modules/providers/max"
	telegramrelaymodule "riders.agency/pulse/api/internal/modules/telegramrelay"
	"riders.agency/pulse/api/internal/platform/config"
	dbpkg "riders.agency/pulse/api/internal/platform/db"
	"riders.agency/pulse/api/internal/platform/httpserver"
	"riders.agency/pulse/api/internal/platform/migrate"
	"riders.agency/pulse/api/internal/platform/outbox"
	runtimekernel "riders.agency/pulse/api/internal/runtime/kernel"
	runtimeregistry "riders.agency/pulse/api/internal/runtime/registry"
)

type Application struct {
	Config          config.Config
	Deps            Deps
	Handler         http.Handler
	TelegramPoller  *telegramrelaymodule.Poller
	RuntimeKernel   *runtimekernel.Kernel
	RuntimeRegistry *runtimeregistry.Registry
	OutboxRegistry  *outbox.Registry
	ToolGateway     aitools.ToolGateway
}

func Build(ctx context.Context, cfg config.Config) (*Application, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL required")
	}
	if !cfg.SkipMigrations {
		root, _ := os.Getwd()
		mp := cfg.MigrationsPath
		if !filepath.IsAbs(mp) {
			mp = filepath.Join(root, mp)
		}
		if err := migrate.Up(cfg.DatabaseURL, mp); err != nil {
			return nil, fmt.Errorf("migrate: %w", err)
		}
	}

	pool, err := dbpkg.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	if isProductionEnvironment(cfg.Environment) {
		if err := guardNoProductionDemoTenant(ctx, pool); err != nil {
			pool.Close()
			return nil, err
		}
	}

	engine := executor.New(pool)
	authRepo := &authrepo.Pool{P: pool}
	auth := authsvc.New(authRepo, cfg.SessionTTL, cfg.RefreshTTL)
	runtimeRegistry := engine.Registry
	if runtimeRegistry == nil {
		runtimeRegistry = runtimeregistry.New()
		engine.Registry = runtimeRegistry
	}
	outboxRegistry := outbox.NewRegistry()
	masterKey := os.Getenv("RIDERS_PULSE_MASTER_KEY")
	maxHTTPClient, err := maxprovider.NewHTTPClientFromEnv(0)
	if err != nil {
		return nil, fmt.Errorf("max http transport: %w", err)
	}

	deps := Deps{
		DB:              pool,
		Auth:            auth,
		RuntimeNotifier: engine,
		RuntimeRegistry: runtimeRegistry,
		OutboxRegistry:  outboxRegistry,
		MasterKey:       masterKey,
		HTTPClient:      engine.RuntimeHTTPClient(),
		MaxHTTPClient:   maxHTTPClient,
	}
	modules := BuildModules()
	modules.RegisterRuntime(runtimeRegistry, deps)
	modules.RegisterOutboxHandlers(outboxRegistry, deps)
	runtimeKernel := runtimekernel.New(
		pool,
		engine,
		runtimekernel.WithTriggerRegistry(runtimeRegistry),
		runtimekernel.WithRunClaimAllowlist(os.Getenv("PULSE_WORKFLOW_RUN_ALLOWLIST")),
		runtimekernel.WithWorkerCount(runtimekernel.RuntimeWorkerCountFromEnv(os.Getenv("PULSE_RUNTIME_WORKERS"))),
	)
	engine.SetNotifier(runtimeKernel.Notify)

	httpDeps := BuildHTTPDependencies(pool, auth, engine, masterKey)
	handler := httpserver.NewRouter(httpDeps, modules)
	toolGateway := aihttp.NewLangGraphToolGateway(httpDeps.AI)

	return &Application{
		Config:          cfg,
		Deps:            deps,
		Handler:         handler,
		TelegramPoller:  &telegramrelaymodule.Poller{P: pool, E: engine, MasterHEX: masterKey},
		RuntimeKernel:   runtimeKernel,
		RuntimeRegistry: runtimeRegistry,
		OutboxRegistry:  outboxRegistry,
		ToolGateway:     toolGateway,
	}, nil
}
