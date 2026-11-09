package service

import repo "riders.agency/pulse/api/internal/modules/apikeys/repository"

type Commands struct{ *repo.APIKeys }

func NewCommands(r *repo.APIKeys) *Commands {
	if r == nil {
		return nil
	}
	return &Commands{APIKeys: r}
}
