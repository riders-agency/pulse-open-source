package ai

import (
	"context"
	"log/slog"

	"riders.agency/pulse/api/internal/platform/outbox"
)

var aiToolLifecycleEvents = []string{
	"ai.tool.approved",
	"ai.tool.applied",
	"ai.tool.failed",
	"ai.tool.rejected",
}

func RegisterOutboxHandlers(reg *outbox.Registry) error {
	for _, eventType := range aiToolLifecycleEvents {
		if err := reg.Register(aiToolLifecycleHandler{typeKey: eventType}); err != nil {
			return err
		}
	}
	return nil
}

type aiToolLifecycleHandler struct {
	typeKey string
}

func (h aiToolLifecycleHandler) TypeKey() string {
	return h.typeKey
}

func (h aiToolLifecycleHandler) Handle(_ context.Context, ev outbox.Event) error {
	slog.Info("ai tool lifecycle outbox event processed",
		slog.String("event_id", ev.ID),
		slog.String("event_type", ev.TypeKey),
		slog.String("aggregate_type", ev.AggregateType),
		slog.String("aggregate_id", ev.AggregateID),
	)
	return nil
}
