package http

import (
	"github.com/go-chi/chi/v5"
	workflowssvc "riders.agency/pulse/api/internal/modules/workflows/service"
)

func MountProjectRoutes(r chi.Router, workflows *workflowssvc.Commands) {
	r.Route("/workflows", func(w chi.Router) {
		w.Get("/", ListByProject(workflows))
		w.Post("/", Create(workflows))
	})
}

func MountRoutes(r chi.Router, workflows *workflowssvc.Commands) {
	r.Put("/workflows/{workflow_id}/draft", PutDraft(workflows))
	r.Post("/workflows/{workflow_id}/validate", Validate(workflows))
	r.Post("/workflows/{workflow_id}/publish", Publish(workflows))
	r.Get("/workflows/{workflow_id}/versions", ListVersions(workflows))
	r.Post("/workflows/{workflow_id}/versions/{workflow_version_id}/restore", RestoreVersion(workflows))
	r.Get("/workflow-versions/{workflow_version_id}", GetVersion(workflows))
	r.Get("/workflows/{workflow_id}", Get(workflows))
}
