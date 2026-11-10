package http

import "github.com/go-chi/chi/v5"

func MountRelayRoutes(r chi.Router, relay *Relay) {
	r.Route("/telegram/relay", func(tr chi.Router) {
		tr.Post("/{relay_key}", Post(relay))
	})
}

func MountAPIRoutes(r chi.Router, relay *Relay) {
	r.Post("/telegram/relay/{relay_key}", Post(relay))
	r.Post("/telegram/bot-api/{relay_key}/{method}", BotAPIProxy(relay))
}
