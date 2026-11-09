package http

import "github.com/go-chi/chi/v5"

func MountRoutes(r chi.Router, desktop *Handlers) {
	r.Post("/desktop/auth/start", desktop.AuthStart())
	r.Post("/desktop/auth/poll", desktop.Poll())
	r.Get("/desktop/messages", desktop.ListMessages())
	r.Get("/desktop/messages/stream", desktop.MessagesStream)
	r.Post("/desktop/messages/{desktop_message_id}/ack", desktop.AckMessage())
}
