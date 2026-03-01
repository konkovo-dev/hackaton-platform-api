package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
)

type OutboxRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewOutboxRepository(db queries.DBTX) *OutboxRepository {
	return &OutboxRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *OutboxRepository) Create(ctx context.Context, event *outbox.Event) error {
	err := r.Queries().CreateOutboxEvent(ctx, queries.CreateOutboxEventParams{
		ID:            pgxutil.UUIDToPgtype(event.ID),
		AggregateID:   event.AggregateID,
		AggregateType: event.AggregateType,
		EventType:     event.EventType,
		Payload:       event.Payload,
		Status:        string(event.Status),
		AttemptCount:  int32(event.AttemptCount),
		LastError:     event.LastError,
		CreatedAt:     event.CreatedAt,
		UpdatedAt:     event.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("failed to create outbox event: %w", err)
	}

	return nil
}

func (r *OutboxRepository) GetPending(ctx context.Context, limit int) ([]*outbox.Event, error) {
	rows, err := r.Queries().GetPendingOutboxEvents(ctx, int32(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to get pending outbox events: %w", err)
	}

	events := make([]*outbox.Event, 0, len(rows))
	for _, row := range rows {
		events = append(events, &outbox.Event{
			ID:            pgxutil.PgtypeToUUID(row.ID),
			AggregateID:   row.AggregateID,
			AggregateType: row.AggregateType,
			EventType:     row.EventType,
			Payload:       row.Payload,
			Status:        outbox.EventStatus(row.Status),
			AttemptCount:  int(row.AttemptCount),
			LastError:     row.LastError,
			CreatedAt:     row.CreatedAt,
			UpdatedAt:     row.UpdatedAt,
		})
	}

	return events, nil
}

func (r *OutboxRepository) Update(ctx context.Context, event *outbox.Event) error {
	event.UpdatedAt = time.Now().UTC()

	err := r.Queries().UpdateOutboxEvent(ctx, queries.UpdateOutboxEventParams{
		ID:           pgxutil.UUIDToPgtype(event.ID),
		Status:       string(event.Status),
		AttemptCount: int32(event.AttemptCount),
		LastError:    event.LastError,
		UpdatedAt:    event.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("failed to update outbox event: %w", err)
	}

	return nil
}
