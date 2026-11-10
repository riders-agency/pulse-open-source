package http

import (
	"github.com/go-chi/chi/v5"

	docsrepo "riders.agency/pulse/api/internal/modules/integrationdocs/repository"
)

func MountPublicRoutes(r chi.Router, docs *docsrepo.Docs) {
	r.Get("/docs/integrations/search", PublicSearch(docs))
	r.Get("/docs/integrations/{public_slug}.md", RawMarkdown(docs))
}

func MountProjectRoutes(r chi.Router, docs *docsrepo.Docs) {
	r.Get("/integration-docs/search", ProjectSearch(docs))
}
