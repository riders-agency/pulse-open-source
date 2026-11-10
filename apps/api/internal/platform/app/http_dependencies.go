package app

import (
	actiondependenciessvc "riders.agency/pulse/api/internal/modules/action_dependencies/service"
	aiproviders "riders.agency/pulse/api/internal/modules/ai/providers"
	apikeyssvc "riders.agency/pulse/api/internal/modules/apikeys/service"
	companiessvc "riders.agency/pulse/api/internal/modules/companies/service"
	projectssvc "riders.agency/pulse/api/internal/modules/projects/service"
	runssvc "riders.agency/pulse/api/internal/modules/runs/service"
	sitechatsvc "riders.agency/pulse/api/internal/modules/sitechat/service"
	webhookssvc "riders.agency/pulse/api/internal/modules/webhooks/service"
	workflowssvc "riders.agency/pulse/api/internal/modules/workflows/service"

	"github.com/jackc/pgx/v5/pgxpool"

	"riders.agency/pulse/api/internal/executor"
	deprepo "riders.agency/pulse/api/internal/modules/action_dependencies/repository"
	airepo "riders.agency/pulse/api/internal/modules/ai/repository"
	aisvc "riders.agency/pulse/api/internal/modules/ai/service"
	apirepo "riders.agency/pulse/api/internal/modules/apikeys/repository"
	authsvc "riders.agency/pulse/api/internal/modules/auth/service"
	catrepo "riders.agency/pulse/api/internal/modules/catalog/repository"
	clientrepo "riders.agency/pulse/api/internal/modules/clients/repository"
	clientsvc "riders.agency/pulse/api/internal/modules/clients/service"
	companyrepo "riders.agency/pulse/api/internal/modules/companies/repository"
	desktopmodule "riders.agency/pulse/api/internal/modules/desktop"
	docsrepo "riders.agency/pulse/api/internal/modules/integrationdocs/repository"
	projectrepo "riders.agency/pulse/api/internal/modules/projects/repository"
	runsrepo "riders.agency/pulse/api/internal/modules/runs/repository"
	sitechatmodule "riders.agency/pulse/api/internal/modules/sitechat"
	sitechatrepo "riders.agency/pulse/api/internal/modules/sitechat/repository"
	telegramrelaymodule "riders.agency/pulse/api/internal/modules/telegramrelay"
	tplrepo "riders.agency/pulse/api/internal/modules/templates/repository"
	whrepo "riders.agency/pulse/api/internal/modules/webhooks/repository"
	wfrepo "riders.agency/pulse/api/internal/modules/workflows/repository"
	workrepo "riders.agency/pulse/api/internal/modules/worktracker/repository"
	worksvc "riders.agency/pulse/api/internal/modules/worktracker/service"
	"riders.agency/pulse/api/internal/platform/httpserver"
	"riders.agency/pulse/api/internal/platform/uow"
)

func BuildHTTPDependencies(db *pgxpool.Pool, auth *authsvc.Auth, engine *executor.Engine, masterKey string) httpserver.Dependencies {
	defaultHTTPClient := engine.RuntimeHTTPClient()
	clients := clientsvc.NewCommands(&clientrepo.Clients{P: db, MasterHEX: masterKey})
	workflows := workflowssvc.NewCommands(&wfrepo.Workflows{P: db})
	workTrackerRepo := &workrepo.WorkTracker{P: db, TaskEvents: worksvc.TaskEvents{}}
	workTrackerCommands := worksvc.NewCommands(workTrackerRepo)
	siteChat := sitechatsvc.NewCommands(&sitechatrepo.SiteChat{P: db})
	return httpserver.Dependencies{
		Pool:             db,
		Auth:             auth,
		Engine:           engine,
		MasterKey:        masterKey,
		Companies:        companiessvc.NewCommands(&companyrepo.Companies{P: db}),
		APIKeys:          apikeyssvc.NewCommands(&apirepo.APIKeys{P: db}),
		Projects:         projectssvc.NewCommands(&projectrepo.Projects{P: db}),
		WebhookEvents:    webhookssvc.NewCommands(&whrepo.WebhookEvents{P: db, E: engine}),
		Workflows:        workflows,
		Runs:             runssvc.NewCommands(&runsrepo.Runs{P: db, E: engine}),
		Webhooks:         &whrepo.Webhooks{P: db, MasterHEX: masterKey},
		Catalog:          &catrepo.Catalog{P: db},
		Clients:          clients,
		ClientsDaData:    clientsvc.NewDaData(actiondependenciessvc.New(db, masterKey), defaultHTTPClient, uow.New(db)),
		TelegramRelay:    &telegramrelaymodule.Relay{P: db, E: engine, HTTPClient: defaultHTTPClient, MasterHEX: masterKey},
		ActionDeps:       actiondependenciessvc.NewCommands(&deprepo.Deps{P: db, MasterHEX: masterKey}),
		IntegrationDocs:  &docsrepo.Docs{P: db},
		EmailTemplates:   &tplrepo.EmailTemplates{P: db},
		MessageTemplates: &tplrepo.MessageTemplates{P: db},
		Desktop:          &desktopmodule.Handlers{P: db},
		SiteChat:         siteChat,
		SiteChatHub:      sitechatmodule.NewHub(siteChat),
		WorkTracker:      workTrackerCommands,
		AI:               aisvc.NewCommands(&airepo.AI{P: db, MasterHEX: masterKey, ModelClient: &aiproviders.ModelClient{HTTP: defaultHTTPClient}, Clients: clients, Workflows: workflows, WorkTracker: workTrackerCommands}),
	}
}
