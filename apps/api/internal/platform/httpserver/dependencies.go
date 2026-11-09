package httpserver

import (
	"github.com/jackc/pgx/v5/pgxpool"
	actiondependenciessvc "riders.agency/pulse/api/internal/modules/action_dependencies/service"
	apikeyssvc "riders.agency/pulse/api/internal/modules/apikeys/service"
	companiessvc "riders.agency/pulse/api/internal/modules/companies/service"
	projectssvc "riders.agency/pulse/api/internal/modules/projects/service"
	runssvc "riders.agency/pulse/api/internal/modules/runs/service"
	sitechatsvc "riders.agency/pulse/api/internal/modules/sitechat/service"
	webhookssvc "riders.agency/pulse/api/internal/modules/webhooks/service"
	workflowssvc "riders.agency/pulse/api/internal/modules/workflows/service"

	"riders.agency/pulse/api/internal/executor"

	aisvc "riders.agency/pulse/api/internal/modules/ai/service"

	authsvc "riders.agency/pulse/api/internal/modules/auth/service"
	catrepo "riders.agency/pulse/api/internal/modules/catalog/repository"

	clientsvc "riders.agency/pulse/api/internal/modules/clients/service"

	desktophttp "riders.agency/pulse/api/internal/modules/desktop/transport/http"
	docsrepo "riders.agency/pulse/api/internal/modules/integrationdocs/repository"

	sitechathttp "riders.agency/pulse/api/internal/modules/sitechat/transport/http"
	telegramrelayhttp "riders.agency/pulse/api/internal/modules/telegramrelay/transport/http"
	tplrepo "riders.agency/pulse/api/internal/modules/templates/repository"
	whrepo "riders.agency/pulse/api/internal/modules/webhooks/repository"

	worksvc "riders.agency/pulse/api/internal/modules/worktracker/service"
)

type Dependencies struct {
	Pool      *pgxpool.Pool
	Auth      *authsvc.Auth
	Engine    *executor.Engine
	MasterKey string

	Companies        *companiessvc.Commands
	APIKeys          *apikeyssvc.Commands
	Projects         *projectssvc.Commands
	WebhookEvents    *webhookssvc.Commands
	Workflows        *workflowssvc.Commands
	Runs             *runssvc.Commands
	Webhooks         *whrepo.Webhooks
	Catalog          *catrepo.Catalog
	Clients          *clientsvc.Commands
	ClientsDaData    *clientsvc.DaData
	TelegramRelay    *telegramrelayhttp.Relay
	ActionDeps       *actiondependenciessvc.Commands
	IntegrationDocs  *docsrepo.Docs
	EmailTemplates   *tplrepo.EmailTemplates
	MessageTemplates *tplrepo.MessageTemplates
	Desktop          *desktophttp.Handlers
	SiteChat         *sitechatsvc.Commands
	SiteChatHub      *sitechathttp.Hub
	WorkTracker      *worksvc.Commands
	AI               *aisvc.Commands
}
