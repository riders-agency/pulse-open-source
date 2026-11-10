package service

import repo "riders.agency/pulse/api/internal/modules/workflows/repository"

type Commands struct{ *repo.Workflows }

func NewCommands(r *repo.Workflows) *Commands {
	if r == nil {
		return nil
	}
	return &Commands{Workflows: r}
}
