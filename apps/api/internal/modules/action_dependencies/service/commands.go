package service

import (
	repo "riders.agency/pulse/api/internal/modules/action_dependencies/repository"
	dbpkg "riders.agency/pulse/api/internal/platform/db"
)

type Commands struct{ *repo.Deps }

func NewCommands(r *repo.Deps) *Commands {
	if r == nil {
		return nil
	}
	return &Commands{Deps: r}
}

func New(database dbpkg.DB, masterKey string) *Commands {
	return NewCommands(&repo.Deps{P: database, MasterHEX: masterKey})
}
