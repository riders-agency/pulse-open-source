package service

import workrepo "riders.agency/pulse/api/internal/modules/worktracker/repository"

type Commands struct {
	*workrepo.WorkTracker
}

func NewCommands(repo *workrepo.WorkTracker) *Commands {
	if repo == nil {
		return nil
	}
	return &Commands{WorkTracker: repo}
}
