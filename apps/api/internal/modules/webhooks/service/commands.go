package service

import repo "riders.agency/pulse/api/internal/modules/webhooks/repository"

type Commands struct{ *repo.WebhookEvents }

func NewCommands(r *repo.WebhookEvents) *Commands {
	if r == nil {
		return nil
	}
	return &Commands{WebhookEvents: r}
}
