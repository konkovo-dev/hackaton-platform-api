package txrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type TicketRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewTicketRepository(tx pgx.Tx) *TicketRepository {
	return &TicketRepository{
		BaseRepository: pgxutil.NewBaseRepository[*queries.Queries, queries.DBTX](tx, queries.New),
	}
}

func (r *TicketRepository) Create(ctx context.Context, ticket *entity.Ticket) error {
	var assignedMentorUserID pgtype.UUID
	if ticket.AssignedMentorUserID != nil {
		assignedMentorUserID = pgtype.UUID{
			Bytes: *ticket.AssignedMentorUserID,
			Valid: true,
		}
	}

	row, err := r.Queries().CreateTicket(ctx, queries.CreateTicketParams{
		ID:                   ticket.ID,
		HackathonID:          ticket.HackathonID,
		OwnerKind:            ticket.OwnerKind,
		OwnerID:              ticket.OwnerID,
		Status:               ticket.Status,
		AssignedMentorUserID: assignedMentorUserID,
	})
	if err != nil {
		return fmt.Errorf("failed to create ticket: %w", pgxutil.MapDBError(err))
	}

	ticket.CreatedAt = row.CreatedAt
	ticket.UpdatedAt = row.UpdatedAt

	return nil
}

func (r *TicketRepository) UpdateStatus(ctx context.Context, ticketID uuid.UUID, status string, closedAt *time.Time) error {
	var closedAtPg pgtype.Timestamptz
	if closedAt != nil {
		closedAtPg = pgtype.Timestamptz{
			Time:  *closedAt,
			Valid: true,
		}
	}

	_, err := r.Queries().UpdateTicketStatus(ctx, queries.UpdateTicketStatusParams{
		ID:       ticketID,
		Status:   status,
		ClosedAt: closedAtPg,
	})
	if err != nil {
		return fmt.Errorf("failed to update ticket status: %w", pgxutil.MapDBError(err))
	}

	return nil
}

func (r *TicketRepository) ClaimTicket(ctx context.Context, ticketID uuid.UUID, mentorUserID uuid.UUID) (int64, error) {
	mentorUUID := pgtype.UUID{
		Bytes: mentorUserID,
		Valid: true,
	}

	rowsAffected, err := r.Queries().ClaimTicket(ctx, queries.ClaimTicketParams{
		ID:                   ticketID,
		AssignedMentorUserID: mentorUUID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to claim ticket: %w", pgxutil.MapDBError(err))
	}

	return rowsAffected, nil
}

func (r *TicketRepository) AssignTicket(ctx context.Context, ticketID uuid.UUID, mentorUserID uuid.UUID) (int64, error) {
	mentorUUID := pgtype.UUID{
		Bytes: mentorUserID,
		Valid: true,
	}

	rowsAffected, err := r.Queries().AssignTicket(ctx, queries.AssignTicketParams{
		ID:                   ticketID,
		AssignedMentorUserID: mentorUUID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to assign ticket: %w", pgxutil.MapDBError(err))
	}

	return rowsAffected, nil
}

func (r *TicketRepository) FindOpenTicketByOwner(ctx context.Context, hackathonID uuid.UUID, ownerKind string, ownerID uuid.UUID) (*entity.Ticket, error) {
	row, err := r.Queries().FindOpenTicketByOwner(ctx, queries.FindOpenTicketByOwnerParams{
		HackathonID: hackathonID,
		OwnerKind:   ownerKind,
		OwnerID:     ownerID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find open ticket: %w", pgxutil.MapDBError(err))
	}

	var mentorPtr *uuid.UUID
	if row.AssignedMentorUserID.Valid {
		mentorID := uuid.UUID(row.AssignedMentorUserID.Bytes)
		mentorPtr = &mentorID
	}

	var closedAtPtr *time.Time
	if row.ClosedAt.Valid {
		closedAtPtr = &row.ClosedAt.Time
	}

	return &entity.Ticket{
		ID:                   row.ID,
		HackathonID:          row.HackathonID,
		OwnerKind:            row.OwnerKind,
		OwnerID:              row.OwnerID,
		Status:               row.Status,
		AssignedMentorUserID: mentorPtr,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
		ClosedAt:             closedAtPtr,
	}, nil
}

type MessageRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewMessageRepository(tx pgx.Tx) *MessageRepository {
	return &MessageRepository{
		BaseRepository: pgxutil.NewBaseRepository[*queries.Queries, queries.DBTX](tx, queries.New),
	}
}

func (r *MessageRepository) Create(ctx context.Context, message *entity.Message) error {
	var clientMessageID pgtype.Text
	if message.ClientMessageID != "" {
		clientMessageID = pgtype.Text{
			String: message.ClientMessageID,
			Valid:  true,
		}
	}

	row, err := r.Queries().CreateMessage(ctx, queries.CreateMessageParams{
		ID:              message.ID,
		TicketID:        message.TicketID,
		AuthorUserID:    message.AuthorUserID,
		AuthorRole:      message.AuthorRole,
		Text:            message.Text,
		ClientMessageID: clientMessageID,
	})
	if err != nil {
		return fmt.Errorf("failed to create message: %w", pgxutil.MapDBError(err))
	}

	message.CreatedAt = row.CreatedAt

	return nil
}

type OutboxRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewOutboxRepository(tx pgx.Tx) *OutboxRepository {
	return &OutboxRepository{
		BaseRepository: pgxutil.NewBaseRepository[*queries.Queries, queries.DBTX](tx, queries.New),
	}
}

func (r *OutboxRepository) Create(ctx context.Context, event *outbox.Event) error {
	err := r.Queries().CreateOutboxEvent(ctx, queries.CreateOutboxEventParams{
		ID:            event.ID,
		AggregateID:   event.AggregateID,
		AggregateType: event.AggregateType,
		EventType:     event.EventType,
		Payload:       event.Payload,
		Status:        string(event.Status),
		AttemptCount:  int32(event.AttemptCount),
		LastError:     event.LastError,
	})
	if err != nil {
		return fmt.Errorf("failed to create outbox event: %w", pgxutil.MapDBError(err))
	}
	return nil
}
