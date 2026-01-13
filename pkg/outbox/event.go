package outbox

import (
	"time"

	"github.com/google/uuid"
)

type EventStatus string

const (
	EventStatusPending    EventStatus = "pending"
	EventStatusProcessing EventStatus = "processing"
	EventStatusProcessed  EventStatus = "processed"
	EventStatusFailed     EventStatus = "failed"
)

type Event struct {
	ID            uuid.UUID   `json:"id"`
	AggregateID   string      `json:"aggregate_id"`
	AggregateType string      `json:"aggregate_type"`
	EventType     string      `json:"event_type"`
	Payload       []byte      `json:"payload"`
	Status        EventStatus `json:"status"`
	AttemptCount  int         `json:"attempt_count"`
	LastError     string      `json:"last_error"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

func NewEvent(aggregateID, aggregateType, eventType string, payload []byte) *Event {
	now := time.Now().UTC()
	return &Event{
		ID:            uuid.New(),
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		EventType:     eventType,
		Payload:       payload,
		Status:        EventStatusPending,
		AttemptCount:  0,
		LastError:     "",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}
