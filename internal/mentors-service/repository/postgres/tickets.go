package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type TicketRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewTicketRepository(db queries.DBTX) *TicketRepository {
	return &TicketRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *TicketRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Ticket, error) {
	row, err := r.Queries().GetTicketByID(ctx, id)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get ticket by id: %w", err)
	}

	return mapTicketToEntity(row), nil
}

func (r *TicketRepository) GetByIDAndHackathonID(ctx context.Context, id, hackathonID uuid.UUID) (*entity.Ticket, error) {
	row, err := r.Queries().GetTicketByIDAndHackathonID(ctx, queries.GetTicketByIDAndHackathonIDParams{
		ID:          id,
		HackathonID: hackathonID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	return mapTicketToEntity(row), nil
}

func (r *TicketRepository) ListByOwner(ctx context.Context, hackathonID uuid.UUID, ownerKind string, ownerID uuid.UUID, limit, offset int32) ([]*entity.Ticket, error) {
	rows, err := r.Queries().ListTicketsByOwner(ctx, queries.ListTicketsByOwnerParams{
		HackathonID: hackathonID,
		OwnerKind:   ownerKind,
		OwnerID:     ownerID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to list tickets by owner: %w", err)
	}

	tickets := make([]*entity.Ticket, 0, len(rows))
	for _, row := range rows {
		tickets = append(tickets, mapTicketToEntity(row))
	}

	return tickets, nil
}

func (r *TicketRepository) FindOpenTicketByOwner(ctx context.Context, hackathonID uuid.UUID, ownerKind string, ownerID uuid.UUID) (*entity.Ticket, error) {
	row, err := r.Queries().FindOpenTicketByOwner(ctx, queries.FindOpenTicketByOwnerParams{
		HackathonID: hackathonID,
		OwnerKind:   ownerKind,
		OwnerID:     ownerID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return mapTicketToEntity(row), nil
}

func (r *TicketRepository) ListByMentor(ctx context.Context, hackathonID, mentorUserID uuid.UUID, limit, offset int32) ([]*entity.Ticket, error) {
	rows, err := r.Queries().ListTicketsByMentor(ctx, queries.ListTicketsByMentorParams{
		HackathonID:          hackathonID,
		AssignedMentorUserID: pgtype.UUID{Bytes: mentorUserID, Valid: true},
		Limit:                limit,
		Offset:               offset,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to list tickets by mentor: %w", err)
	}

	tickets := make([]*entity.Ticket, 0, len(rows))
	for _, row := range rows {
		tickets = append(tickets, mapTicketToEntity(row))
	}

	return tickets, nil
}

func (r *TicketRepository) ListAll(ctx context.Context, hackathonID uuid.UUID, limit, offset int32) ([]*entity.Ticket, error) {
	rows, err := r.Queries().ListAllTickets(ctx, queries.ListAllTicketsParams{
		HackathonID: hackathonID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to list all tickets: %w", err)
	}

	tickets := make([]*entity.Ticket, 0, len(rows))
	for _, row := range rows {
		tickets = append(tickets, mapTicketToEntity(row))
	}

	return tickets, nil
}

func (r *TicketRepository) CreateOrGetOpenTicket(ctx context.Context, hackathonID uuid.UUID, ownerKind string, ownerID uuid.UUID) (*entity.Ticket, error) {
	row, err := r.Queries().CreateOrGetOpenTicket(ctx, queries.CreateOrGetOpenTicketParams{
		ID:          uuid.New(),
		HackathonID: hackathonID,
		OwnerKind:   ownerKind,
		OwnerID:     ownerID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to create or get open ticket: %w", err)
	}

	return mapTicketToEntity(row), nil
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
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to create ticket: %w", err)
	}

	ticket.CreatedAt = row.CreatedAt
	ticket.UpdatedAt = row.UpdatedAt

	return nil
}

func (r *TicketRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, closedAt *time.Time) error {
	var closedAtPg pgtype.Timestamptz
	if closedAt != nil {
		closedAtPg = pgtype.Timestamptz{
			Time:  *closedAt,
			Valid: true,
		}
	}

	_, err := r.Queries().UpdateTicketStatus(ctx, queries.UpdateTicketStatusParams{
		ID:       id,
		Status:   status,
		ClosedAt: closedAtPg,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to update ticket status: %w", err)
	}

	return nil
}

func (r *TicketRepository) CountOpenTicketsByMentor(ctx context.Context, hackathonID, mentorUserID uuid.UUID) (int64, error) {
	count, err := r.Queries().CountOpenTicketsByMentor(ctx, queries.CountOpenTicketsByMentorParams{
		HackathonID:          hackathonID,
		AssignedMentorUserID: pgtype.UUID{Bytes: mentorUserID, Valid: true},
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return 0, fmt.Errorf("failed to count open tickets: %w", err)
	}

	return count, nil
}

func mapTicketToEntity(t queries.MentorsTicket) *entity.Ticket {
	var mentorPtr *uuid.UUID
	if t.AssignedMentorUserID.Valid {
		mentorID := uuid.UUID(t.AssignedMentorUserID.Bytes)
		mentorPtr = &mentorID
	}

	var closedAtPtr *time.Time
	if t.ClosedAt.Valid {
		closedAtPtr = &t.ClosedAt.Time
	}

	return &entity.Ticket{
		ID:                   t.ID,
		HackathonID:          t.HackathonID,
		OwnerKind:            t.OwnerKind,
		OwnerID:              t.OwnerID,
		Status:               t.Status,
		AssignedMentorUserID: mentorPtr,
		CreatedAt:            t.CreatedAt,
		UpdatedAt:            t.UpdatedAt,
		ClosedAt:             closedAtPtr,
	}
}
