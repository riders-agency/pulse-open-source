package app

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	authsvc "riders.agency/pulse/api/internal/modules/auth/service"
	"riders.agency/pulse/api/internal/platform/httpserver"
	"riders.agency/pulse/api/internal/platform/outbox"
	"riders.agency/pulse/api/internal/runtime/registry"
)

type RuntimeNotifier interface {
	Notify()
}

type OutboxWriter interface {
	Enqueue(ctx context.Context, req outbox.EnqueueRequest) (outbox.Event, error)
}

type TemplateRenderer interface {
	Render(template string, input json.RawMessage, steps map[string]any, workflowRunID string) string
}

type Deps struct {
	DB              *pgxpool.Pool
	Auth            *authsvc.Auth
	RuntimeNotifier RuntimeNotifier
	RuntimeRegistry *registry.Registry
	OutboxRegistry  *outbox.Registry
	MasterKey       string
	HTTPClient      *http.Client
	MaxHTTPClient   *http.Client
}

type Module interface {
	Name() string
	RegisterHTTP(r chi.Router, deps httpserver.Dependencies)
	RegisterAPIV1(scope httpserver.APIV1Scope, deps httpserver.Dependencies)
	RegisterRuntime(reg *registry.Registry, deps Deps)
	RegisterOutboxHandlers(reg *outbox.Registry, deps Deps)
}

type Modules []Module

func (mods Modules) Names() []string {
	names := make([]string, 0, len(mods))
	for _, mod := range mods {
		if mod == nil {
			continue
		}
		names = append(names, mod.Name())
	}
	return names
}

func (mods Modules) RegisterHTTP(r chi.Router, deps httpserver.Dependencies) {
	for _, mod := range mods {
		if mod == nil {
			continue
		}
		mod.RegisterHTTP(r, deps)
	}
	httpserver.MountAPIV1(r, deps, mods)
}

func (mods Modules) RegisterAPIV1(scope httpserver.APIV1Scope, deps httpserver.Dependencies) {
	for _, mod := range mods {
		if mod == nil {
			continue
		}
		mod.RegisterAPIV1(scope, deps)
	}
}

func (mods Modules) RegisterRuntime(reg *registry.Registry, deps Deps) {
	for _, mod := range mods {
		if mod == nil {
			continue
		}
		mod.RegisterRuntime(reg, deps)
	}
}

func (mods Modules) RegisterOutboxHandlers(reg *outbox.Registry, deps Deps) {
	for _, mod := range mods {
		if mod == nil {
			continue
		}
		mod.RegisterOutboxHandlers(reg, deps)
	}
}
