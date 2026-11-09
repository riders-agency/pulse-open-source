package outbox

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

type Handler interface {
	TypeKey() string
	Handle(ctx context.Context, ev Event) error
}

type HandlerDescriptor struct {
	TypeKey string
}

type Registry struct {
	mu       sync.RWMutex
	handlers map[string]Handler
}

func NewRegistry() *Registry {
	return &Registry{handlers: map[string]Handler{}}
}

func (r *Registry) Register(handler Handler) error {
	if r == nil {
		return fmt.Errorf("outbox registry is nil")
	}
	if handler == nil {
		return fmt.Errorf("outbox handler is nil")
	}
	typeKey := strings.TrimSpace(handler.TypeKey())
	if typeKey == "" {
		return fmt.Errorf("outbox handler type key is empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handlers == nil {
		r.handlers = map[string]Handler{}
	}
	if _, exists := r.handlers[typeKey]; exists {
		return fmt.Errorf("outbox handler already registered: %s", typeKey)
	}
	r.handlers[typeKey] = handler
	return nil
}

func (r *Registry) Handler(typeKey string) (Handler, bool) {
	typeKey = strings.TrimSpace(typeKey)
	if r == nil || typeKey == "" {
		return nil, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, ok := r.handlers[typeKey]
	return handler, ok
}

func (r *Registry) List() []HandlerDescriptor {
	if r == nil {
		return []HandlerDescriptor{}
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	keys := make([]string, 0, len(r.handlers))
	for key := range r.handlers {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]HandlerDescriptor, 0, len(keys))
	for _, key := range keys {
		out = append(out, HandlerDescriptor{TypeKey: key})
	}
	return out
}
