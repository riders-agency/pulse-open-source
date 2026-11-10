package contracts

import (
	"time"

	"github.com/google/uuid"
)

type EventCursor struct {
	At time.Time
	ID uuid.UUID
}
type EventPage struct {
	Items []map[string]any
	Next  EventCursor
}
