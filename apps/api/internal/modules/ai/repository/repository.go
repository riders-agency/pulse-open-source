package repository

import (
	"errors"

	clientssvc "riders.agency/pulse/api/internal/modules/clients/service"
	workflowssvc "riders.agency/pulse/api/internal/modules/workflows/service"

	dbpkg "riders.agency/pulse/api/internal/platform/db"

	aiproviders "riders.agency/pulse/api/internal/modules/ai/providers"

	worksvc "riders.agency/pulse/api/internal/modules/worktracker/service"
)

var ErrNotFound = errors.New("ai resource not found")

type AI struct {
	P           dbpkg.DB
	MasterHEX   string
	ModelClient aiproviders.ModelGateway
	Clients     *clientssvc.Commands
	Workflows   *workflowssvc.Commands
	WorkTracker *worksvc.Commands
}

type ModelProvider = aiproviders.ModelProvider

type ModelDescriptor = aiproviders.ModelDescriptor
