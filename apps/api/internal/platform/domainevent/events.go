// Package domainevent records business facts without choosing a delivery mechanism.
package domainevent

import (
	"context"
	"encoding/json"
	"fmt"
)

type Event struct {
	TypeKey        string
	AggregateType  string
	AggregateID    string
	IdempotencyKey string
	PayloadJSON    json.RawMessage
}
type Collector struct{ Events []Event }
type collectorKey struct{}

func WithCollector(ctx context.Context, c *Collector) context.Context {
	return context.WithValue(ctx, collectorKey{}, c)
}
func Record(ctx context.Context, event Event) error {
	collector, ok := ctx.Value(collectorKey{}).(*Collector)
	if !ok {
		return fmt.Errorf("domain event requires an application transaction")
	}
	collector.Events = append(collector.Events, event)
	return nil
}

func Merge(ctx context.Context, events []Event) error {
	collector, ok := ctx.Value(collectorKey{}).(*Collector)
	if !ok {
		return fmt.Errorf("domain event requires an application transaction")
	}
	collector.Events = append(collector.Events, events...)
	return nil
}
