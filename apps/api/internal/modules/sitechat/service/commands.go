package service

import repo "riders.agency/pulse/api/internal/modules/sitechat/repository"

type Commands struct{ *repo.SiteChat }

func NewCommands(r *repo.SiteChat) *Commands {
	if r == nil {
		return nil
	}
	return &Commands{SiteChat: r}
}
