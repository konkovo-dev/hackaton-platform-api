package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/usecase/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	pgxadapter "github.com/belikoooova/hackaton-platform-api/pkg/pgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxRepository struct {
	queries *queries.Queries
}

func NewOutboxRepository(pool *pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{
		queries: queries.New(pool),
	}
}

func (r *OutboxRepository) WithTx(tx pgx.Tx) auth.OutboxRepository {
	return &OutboxRepository{
		queries: queries.New(tx),
	}
}

func (r *OutboxRepository) Create(ctx context.Context, event *outbox.Event) error {
	err := r.queries.CreateOutboxEvent(ctx, queries.CreateOutboxEventParams{
		ID:            pgxadapter.UUIDToPgtype(event.ID),
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
	rows, err := r.queries.GetPendingOutboxEvents(ctx, int32(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to get pending outbox events: %w", err)
	}

	events := make([]*outbox.Event, 0, len(rows))
	for _, row := range rows {
		events = append(events, &outbox.Event{
			ID:            pgxadapter.PgtypeToUUID(row.ID),
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

	err := r.queries.UpdateOutboxEvent(ctx, queries.UpdateOutboxEventParams{
		ID:           pgxadapter.UUIDToPgtype(event.ID),
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
