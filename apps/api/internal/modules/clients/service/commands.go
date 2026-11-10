package service

import (
	repo "riders.agency/pulse/api/internal/modules/clients/repository"
	dbpkg "riders.agency/pulse/api/internal/platform/db"
)

type Commands struct{ *repo.Clients }

func NewCommands(r *repo.Clients) *Commands {
	if r == nil {
		return nil
	}
	return &Commands{Clients: r}
}

func New(database dbpkg.DB, masterKey string) *Commands {
	return NewCommands(&repo.Clients{P: database, MasterHEX: masterKey})
}
