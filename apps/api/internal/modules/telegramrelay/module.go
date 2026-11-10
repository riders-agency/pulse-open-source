package telegramrelay

import (
	"github.com/go-chi/chi/v5"

	relayruntime "riders.agency/pulse/api/internal/modules/telegramrelay/runtime"
	tgrhttp "riders.agency/pulse/api/internal/modules/telegramrelay/transport/http"
	"riders.agency/pulse/api/internal/platform/httpserver"
)

type Poller = relayruntime.Poller
type Relay = tgrhttp.Relay

func RegisterHTTP(r chi.Router, deps httpserver.Dependencies) {
	MountPublicHTTP(r, deps.TelegramRelay)
}

func RegisterAPIV1(scope httpserver.APIV1Scope, deps httpserver.Dependencies) {
	if scope.Kind == httpserver.APIV1Public {
		MountAPIHTTP(scope.Router, deps.TelegramRelay)
	}
}

func MountPublicHTTP(r chi.Router, relay *Relay) {
	tgrhttp.MountRelayRoutes(r, relay)
}

func MountAPIHTTP(r chi.Router, relay *Relay) {
	tgrhttp.MountAPIRoutes(r, relay)
}
