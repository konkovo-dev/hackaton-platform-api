package announcement

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain/entity"
	"github.com/google/uuid"
)

type HackathonRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Hackathon, error)
}

type AnnouncementRepository interface {
	Create(ctx context.Context, announcement *entity.HackathonAnnouncement) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.HackathonAnnouncement, error)
	ListByHackathonID(ctx context.Context, hackathonID uuid.UUID) ([]*entity.HackathonAnnouncement, error)
	Update(ctx context.Context, announcement *entity.HackathonAnnouncement) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

type ParticipationAndRolesClient interface {
	GetHackathonContext(ctx context.Context, hackathonID string) (userID string, participationStatus string, teamID string, roles []string, err error)
}
