package outbox

import (
	"context"
)

type EventRepository interface {
	Create(ctx context.Context, event *Event) error
	GetPending(ctx context.Context, limit int) ([]*Event, error)
	Update(ctx context.Context, event *Event) error
}

type Handler interface {
	EventType() string
	Handle(ctx context.Context, event *Event) error
}

type Publisher interface {
	Publish(ctx context.Context, aggregateID, aggregateType, eventType string, payload []byte) error
}
