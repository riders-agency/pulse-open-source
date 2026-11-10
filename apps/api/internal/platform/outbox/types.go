package outbox

import (
	"encoding/json"
	"time"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusProcessed  Status = "processed"
	StatusFailed     Status = "failed"
)

type Event struct {
	ID             string
	TypeKey        string
	AggregateType  string
	AggregateID    string
	IdempotencyKey string
	Status         Status
	PayloadJSON    json.RawMessage
	Attempts       int
	MaxAttempts    int
	AvailableAt    time.Time
	LockedUntil    *time.Time
	LastError      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type EnqueueRequest struct {
	TypeKey        string
	SchemaVersion  string
	AggregateType  string
	AggregateID    string
	IdempotencyKey string
	PayloadJSON    json.RawMessage
	AvailableAt    time.Time
	MaxAttempts    int
}
