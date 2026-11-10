package http

import (
	"github.com/go-chi/chi/v5"
	runssvc "riders.agency/pulse/api/internal/modules/runs/service"
)

func MountWorkflowRoutes(r chi.Router, runs *runssvc.Commands) {
	r.Post("/workflows/{workflow_id}/runs/manual", Manual(runs))
	r.Post("/workflows/{workflow_id}/test-runs", TestRun(runs))
	r.Get("/workflows/{workflow_id}/runs", ListByWorkflow(runs))
}

func MountRunRoutes(r chi.Router, runs *runssvc.Commands) {
	r.Get("/runs/{workflow_run_id}", Get(runs))
	r.Get("/runs/{workflow_run_id}/graph", Graph(runs))
	r.Get("/runs/{workflow_run_id}/actions", Actions(runs))
	r.Get("/runs/{workflow_run_id}/events", Events(runs))
	r.Get("/runs/{workflow_run_id}/events/stream", Stream(runs))
	r.Post("/runs/{workflow_run_id}/cancel", Cancel(runs))
	r.Post("/runs/{workflow_run_id}/replay", Replay(runs))
	r.Post("/runs/{workflow_run_id}/retry-action", ReplayAction(runs))
}
