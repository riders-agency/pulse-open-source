package service

import (
	"context"

	"github.com/jackc/pgx/v5"
	airepo "riders.agency/pulse/api/internal/modules/ai/repository"
	clientsvc "riders.agency/pulse/api/internal/modules/clients/service"
	workflowssvc "riders.agency/pulse/api/internal/modules/workflows/service"
	worksvc "riders.agency/pulse/api/internal/modules/worktracker/service"
)

type Commands struct{ *airepo.AI }

func NewCommands(repo *airepo.AI) *Commands {
	if repo == nil {
		return nil
	}
	return &Commands{AI: repo}
}

func (c *Commands) inTransaction(ctx context.Context, tx pgx.Tx) *airepo.AI {
	bound := *c.AI
	bound.P = tx
	if c.Clients != nil {
		bound.Clients = clientsvc.NewCommands(c.Clients.Clients.WithTx(tx))
	}
	if c.Workflows != nil {
		bound.Workflows = workflowssvc.NewCommands(c.Workflows.Workflows.WithTx(tx))
	}
	if c.WorkTracker != nil {
		work := *c.WorkTracker.WorkTracker
		work.P = tx
		bound.WorkTracker = worksvc.NewCommands(&work)
	}
	return &bound
}
