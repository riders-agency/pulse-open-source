package clients

import (
	"github.com/go-chi/chi/v5"

	"riders.agency/pulse/api/internal/modules/clients/invoicing"

	"riders.agency/pulse/api/internal/modules/clients/service"
	clienthttp "riders.agency/pulse/api/internal/modules/clients/transport/http"
	"riders.agency/pulse/api/internal/platform/httpserver"
)

type RunNotifier = clienthttp.RunNotifier
type DaDataFinder = clienthttp.DaDataFinder
type DaDataIntegrationSaver = clienthttp.DaDataIntegrationSaver
type DaDataService = clienthttp.DaDataService

func RegisterAPIV1(scope httpserver.APIV1Scope, deps httpserver.Dependencies) {
	switch scope.Kind {
	case httpserver.APIV1Protected:
		MountCompanyHTTP(scope.Router, deps.Clients, deps.ClientsDaData)
		MountGlobalHTTP(scope.Router, deps.Clients, deps.Engine)
		MountClientHTTP(scope.Router, deps.Clients, deps.Engine)
	case httpserver.APIV1Project:
		MountProjectHTTP(scope.Router, deps.Clients, deps.ClientsDaData, deps.Engine)
	}
}

func MountCompanyHTTP(r chi.Router, clients *service.Commands, dadata DaDataService) {
	clienthttp.MountCompanyRoutes(r, clients, dadata)
}

func MountGlobalHTTP(r chi.Router, clients *service.Commands, notifier RunNotifier) {
	clienthttp.MountGlobalRoutes(r, clients, notifier)
	clienthttp.MountBankInvoiceRoutes(r, service.NewBankInvoices(clients, invoicing.NewRegistry()))
}

func MountClientHTTP(r chi.Router, clients *service.Commands, notifier RunNotifier) {
	clienthttp.MountClientRoutes(r, clients, notifier)
}

func MountProjectHTTP(r chi.Router, clients *service.Commands, dadata DaDataIntegrationSaver, notifier RunNotifier) {
	clienthttp.MountProjectRoutes(r, clients, dadata, notifier)
}
