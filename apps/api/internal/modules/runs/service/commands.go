package service

import repo "riders.agency/pulse/api/internal/modules/runs/repository"

type Commands struct{ *repo.Runs }

func NewCommands(r *repo.Runs) *Commands {
	if r == nil {
		return nil
	}
	return &Commands{Runs: r}
}
