package kernel

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"

	"riders.agency/pulse/api/internal/runtime/contracts"
	"riders.agency/pulse/api/internal/runtime/registry"
)

func RegisterDatabaseTriggers(reg *registry.Registry, pool *pgxpool.Pool, notifier triggerNotifier) {
	if reg == nil {
		return
	}
	reg.MustRegisterTrigger(&ScheduleTrigger{P: pool, Notifier: notifier})
	for _, handler := range []contracts.TriggerHandler{
		eventDrivenTriggerHandler{typeKey: "telegram.relay.trigger"},
		eventDrivenTriggerHandler{typeKey: "module.event.trigger"},
	} {
		reg.MustRegisterTrigger(handler)
	}
}

type eventDrivenTriggerHandler struct {
	typeKey string
}

func (h eventDrivenTriggerHandler) TypeKey() string { return h.typeKey }

func (h eventDrivenTriggerHandler) Sync(context.Context) error { return nil }

func (h eventDrivenTriggerHandler) ClaimDue(context.Context, int) ([]contracts.TriggerEvent, error) {
	return nil, nil
}

func (h eventDrivenTriggerHandler) MarkEnqueued(context.Context, contracts.TriggerEvent, string) error {
	return nil
}

func (h eventDrivenTriggerHandler) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"type_key": h.typeKey, "mode": "event_driven"})
}
