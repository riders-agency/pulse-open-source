package service

import repo "riders.agency/pulse/api/internal/modules/companies/repository"

type Commands struct{ *repo.Companies }

func NewCommands(r *repo.Companies) *Commands {
	if r == nil {
		return nil
	}
	return &Commands{Companies: r}
}
