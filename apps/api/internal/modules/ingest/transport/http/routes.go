package http

import "github.com/go-chi/chi/v5"

func MountHookRoutes(r chi.Router, webhook Webhook) {
	r.Get("/hooks/{hook_key}/{endpoint_key}/runs/{workflow_run_id}/receipt", webhook.Receipt())
	r.Mount("/hooks", webhook.MountHandler())
}
