package app

import (
	"github.com/go-chi/chi/v5"

	depsmodule "riders.agency/pulse/api/internal/modules/action_dependencies"
	aimodule "riders.agency/pulse/api/internal/modules/ai"
	apikeysmodule "riders.agency/pulse/api/internal/modules/apikeys"
	authmodule "riders.agency/pulse/api/internal/modules/auth"
	catalogmodule "riders.agency/pulse/api/internal/modules/catalog"
	clientsmodule "riders.agency/pulse/api/internal/modules/clients"
	companiesmodule "riders.agency/pulse/api/internal/modules/companies"
	desktopmodule "riders.agency/pulse/api/internal/modules/desktop"
	ingestmodule "riders.agency/pulse/api/internal/modules/ingest"
	docsmodule "riders.agency/pulse/api/internal/modules/integrationdocs"
	projectsmodule "riders.agency/pulse/api/internal/modules/projects"
	providersmodule "riders.agency/pulse/api/internal/modules/providers"
	runsmodule "riders.agency/pulse/api/internal/modules/runs"
	sitechatmodule "riders.agency/pulse/api/internal/modules/sitechat"
	telegramrelaymodule "riders.agency/pulse/api/internal/modules/telegramrelay"
	templatesmodule "riders.agency/pulse/api/internal/modules/templates"
	webhooksmodule "riders.agency/pulse/api/internal/modules/webhooks"
	workflowsmodule "riders.agency/pulse/api/internal/modules/workflows"
	worktrackermodule "riders.agency/pulse/api/internal/modules/worktracker"
	"riders.agency/pulse/api/internal/platform/httpserver"
	"riders.agency/pulse/api/internal/platform/httptransport"
	"riders.agency/pulse/api/internal/platform/outbox"
	"riders.agency/pulse/api/internal/runtime/registry"
)

type moduleAdapter struct {
	name    string
	http    func(chi.Router, httpserver.Dependencies)
	apiV1   func(httpserver.APIV1Scope, httpserver.Dependencies)
	runtime func(*registry.Registry, Deps)
	outbox  func(*outbox.Registry, Deps)
}

func (m moduleAdapter) Name() string {
	return m.name
}

func (m moduleAdapter) RegisterHTTP(r chi.Router, deps httpserver.Dependencies) {
	if m.http != nil {
		m.http(r, deps)
	}
}

func (m moduleAdapter) RegisterAPIV1(scope httpserver.APIV1Scope, deps httpserver.Dependencies) {
	if m.apiV1 != nil {
		m.apiV1(scope, deps)
	}
}

func (m moduleAdapter) RegisterRuntime(reg *registry.Registry, deps Deps) {
	if m.runtime != nil {
		m.runtime(reg, deps)
	}
}

func (m moduleAdapter) RegisterOutboxHandlers(reg *outbox.Registry, deps Deps) {
	if m.outbox != nil {
		m.outbox(reg, deps)
	}
}

func BuildModules() Modules {
	return Modules{
		moduleAdapter{name: "api.docs", apiV1: registerPulseDocs},
		moduleAdapter{name: "sitechat.http", http: sitechatmodule.RegisterHTTP, apiV1: sitechatmodule.RegisterAPIV1},
		moduleAdapter{name: "ingest.http", http: ingestmodule.RegisterHTTP},
		moduleAdapter{name: "telegramrelay.http", http: telegramrelaymodule.RegisterHTTP, apiV1: telegramrelaymodule.RegisterAPIV1},
		moduleAdapter{name: "auth.http", apiV1: authmodule.RegisterAPIV1},
		moduleAdapter{name: "ai.http", apiV1: aimodule.RegisterAPIV1},
		moduleAdapter{name: "catalog.http", apiV1: catalogmodule.RegisterAPIV1},
		moduleAdapter{name: "apikeys.http", apiV1: apikeysmodule.RegisterAPIV1},
		moduleAdapter{name: "worktracker.http", apiV1: worktrackermodule.RegisterAPIV1},
		moduleAdapter{name: "companies.http", apiV1: companiesmodule.RegisterAPIV1},
		moduleAdapter{name: "clients.http", apiV1: clientsmodule.RegisterAPIV1},
		moduleAdapter{name: "projects.http", apiV1: projectsmodule.RegisterAPIV1},
		moduleAdapter{name: "integrationdocs.http", apiV1: docsmodule.RegisterAPIV1},
		moduleAdapter{name: "action_dependencies.http", apiV1: depsmodule.RegisterAPIV1},
		moduleAdapter{name: "templates.http", apiV1: templatesmodule.RegisterAPIV1},
		moduleAdapter{name: "workflows.http", apiV1: workflowsmodule.RegisterAPIV1},
		moduleAdapter{name: "webhooks.http", apiV1: webhooksmodule.RegisterAPIV1},
		moduleAdapter{name: "runs.http", apiV1: runsmodule.RegisterAPIV1},
		moduleAdapter{name: "desktop.http", apiV1: desktopmodule.RegisterAPIV1},
		moduleAdapter{name: "providers.runtime", runtime: registerProvidersRuntime},
		moduleAdapter{name: "clients.runtime", runtime: registerClientsRuntime},
		moduleAdapter{name: "workflows.outbox", outbox: registerWorkflowOutbox},
		moduleAdapter{name: "ai.outbox", outbox: registerAIOutbox},
	}
}

func registerPulseDocs(scope httpserver.APIV1Scope, _ httpserver.Dependencies) {
	if scope.Kind == httpserver.APIV1Public {
		httpserver.MountPulseDocs(scope.Router)
	}
}

func registerProvidersRuntime(reg *registry.Registry, deps Deps) {
	client := deps.HTTPClient
	if client == nil {
		client = httptransport.NewDefaultClient(0)
	}
	maxClient := deps.MaxHTTPClient
	if maxClient == nil {
		maxClient = client
	}
	if err := providersmodule.RegisterRuntimeHandlers(reg, deps.DB, deps.MasterKey, client, maxClient); err != nil {
		panic(err)
	}
}

func registerClientsRuntime(reg *registry.Registry, deps Deps) {
	if err := clientsmodule.RegisterRuntimeHandlers(reg, deps.DB, deps.RuntimeNotifier); err != nil {
		panic(err)
	}
}

func registerWorkflowOutbox(reg *outbox.Registry, deps Deps) {
	var notify workflowsmodule.NotifyFunc
	if deps.RuntimeNotifier != nil {
		notify = deps.RuntimeNotifier.Notify
	}
	if err := workflowsmodule.RegisterOutboxHandlers(reg, deps.DB, notify); err != nil {
		panic(err)
	}
}

func registerAIOutbox(reg *outbox.Registry, _ Deps) {
	if err := aimodule.RegisterOutboxHandlers(reg); err != nil {
		panic(err)
	}
}
