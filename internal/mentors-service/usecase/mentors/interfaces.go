package mentors

import (
	"context"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain/entity"
	"github.com/google/uuid"
)

type TicketRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Ticket, error)
	GetByIDAndHackathonID(ctx context.Context, id, hackathonID uuid.UUID) (*entity.Ticket, error)
	ListByOwner(ctx context.Context, hackathonID uuid.UUID, ownerKind string, ownerID uuid.UUID, limit, offset int32) ([]*entity.Ticket, error)
	FindOpenTicketByOwner(ctx context.Context, hackathonID uuid.UUID, ownerKind string, ownerID uuid.UUID) (*entity.Ticket, error)
	CreateOrGetOpenTicket(ctx context.Context, hackathonID uuid.UUID, ownerKind string, ownerID uuid.UUID) (*entity.Ticket, error)
	ListByMentor(ctx context.Context, hackathonID, mentorUserID uuid.UUID, limit, offset int32) ([]*entity.Ticket, error)
	ListAll(ctx context.Context, hackathonID uuid.UUID, limit, offset int32) ([]*entity.Ticket, error)
	Create(ctx context.Context, ticket *entity.Ticket) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, closedAt *time.Time) error
	CountOpenTicketsByMentor(ctx context.Context, hackathonID, mentorUserID uuid.UUID) (int64, error)
}

type MessageRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Message, error)
	ListByTicket(ctx context.Context, ticketID uuid.UUID, limit, offset int32) ([]*entity.Message, error)
	ListByOwner(ctx context.Context, hackathonID uuid.UUID, ownerKind string, ownerID uuid.UUID, limit, offset int32) ([]*entity.Message, error)
	Create(ctx context.Context, message *entity.Message) error
	FindByClientMessageID(ctx context.Context, clientMessageID string) (*entity.Message, error)
}

type HackathonClient interface {
	GetHackathon(ctx context.Context, hackathonID string) (stage string, err error)
}

type ParticipationRolesClient interface {
	GetParticipationAndRoles(ctx context.Context, userID, hackathonID string) (roles []string, teamID *string, err error)
}

type TeamClient interface {
	ListTeamMembers(ctx context.Context, hackathonID, teamID string) ([]string, error)
}
