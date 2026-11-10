package registry

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"riders.agency/pulse/api/internal/runtime/contracts"
)

type ActionDescriptor struct {
	TypeKey string
}

type TriggerDescriptor struct {
	TypeKey string
}

type Registry struct {
	mu       sync.RWMutex
	actions  map[string]contracts.ActionHandler
	triggers map[string]contracts.TriggerHandler
}

func New() *Registry {
	return &Registry{
		actions:  map[string]contracts.ActionHandler{},
		triggers: map[string]contracts.TriggerHandler{},
	}
}

func (r *Registry) RegisterAction(handler contracts.ActionHandler) error {
	if r == nil {
		return fmt.Errorf("runtime registry is nil")
	}
	if handler == nil {
		return fmt.Errorf("action handler is nil")
	}
	typeKey := strings.TrimSpace(handler.TypeKey())
	if typeKey == "" {
		return fmt.Errorf("action handler type key is empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.actions == nil {
		r.actions = map[string]contracts.ActionHandler{}
	}
	if _, exists := r.actions[typeKey]; exists {
		return fmt.Errorf("action handler already registered: %s", typeKey)
	}
	r.actions[typeKey] = handler
	return nil
}

func (r *Registry) MustRegisterAction(handler contracts.ActionHandler) {
	if err := r.RegisterAction(handler); err != nil {
		panic(err)
	}
}

func (r *Registry) Action(typeKey string) (contracts.ActionHandler, bool) {
	typeKey = strings.TrimSpace(typeKey)
	if r == nil || typeKey == "" {
		return nil, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, ok := r.actions[typeKey]
	return handler, ok
}

func (r *Registry) ListActions() []ActionDescriptor {
	if r == nil {
		return []ActionDescriptor{}
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	keys := make([]string, 0, len(r.actions))
	for key := range r.actions {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]ActionDescriptor, 0, len(keys))
	for _, key := range keys {
		out = append(out, ActionDescriptor{TypeKey: key})
	}
	return out
}

func (r *Registry) RegisterTrigger(handler contracts.TriggerHandler) error {
	if r == nil {
		return fmt.Errorf("runtime registry is nil")
	}
	if handler == nil {
		return fmt.Errorf("trigger handler is nil")
	}
	typeKey := strings.TrimSpace(handler.TypeKey())
	if typeKey == "" {
		return fmt.Errorf("trigger handler type key is empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.triggers == nil {
		r.triggers = map[string]contracts.TriggerHandler{}
	}
	if _, exists := r.triggers[typeKey]; exists {
		return fmt.Errorf("trigger handler already registered: %s", typeKey)
	}
	r.triggers[typeKey] = handler
	return nil
}

func (r *Registry) MustRegisterTrigger(handler contracts.TriggerHandler) {
	if err := r.RegisterTrigger(handler); err != nil {
		panic(err)
	}
}

func (r *Registry) Trigger(typeKey string) (contracts.TriggerHandler, bool) {
	typeKey = strings.TrimSpace(typeKey)
	if r == nil || typeKey == "" {
		return nil, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, ok := r.triggers[typeKey]
	return handler, ok
}

func (r *Registry) ListTriggers() []TriggerDescriptor {
	if r == nil {
		return []TriggerDescriptor{}
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	keys := make([]string, 0, len(r.triggers))
	for key := range r.triggers {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]TriggerDescriptor, 0, len(keys))
	for _, key := range keys {
		out = append(out, TriggerDescriptor{TypeKey: key})
	}
	return out
}
