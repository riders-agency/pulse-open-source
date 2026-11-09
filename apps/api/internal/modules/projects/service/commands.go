package service

import repo "riders.agency/pulse/api/internal/modules/projects/repository"

type Commands struct{ *repo.Projects }

func NewCommands(r *repo.Projects) *Commands {
	if r == nil {
		return nil
	}
	return &Commands{Projects: r}
}
