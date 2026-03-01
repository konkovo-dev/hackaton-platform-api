package vacancy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/google/uuid"
)

type VacancyRepository interface {
	GetByID(ctx context.Context, vacancyID uuid.UUID) (*entity.Vacancy, error)
	Create(ctx context.Context, vacancy *entity.Vacancy) error
	Update(ctx context.Context, vacancy *entity.Vacancy) error
	CountOccupiedSlots(ctx context.Context, vacancyID uuid.UUID) (int64, error)
	CountTotalOpenSlots(ctx context.Context, teamID uuid.UUID) (int64, error)
}

type TeamRepository interface {
	GetByIDAndHackathonID(ctx context.Context, teamID, hackathonID uuid.UUID) (*entity.Team, error)
}

type MembershipRepository interface {
	CheckIsCaptain(ctx context.Context, teamID, userID uuid.UUID) (bool, error)
	CountMembers(ctx context.Context, teamID uuid.UUID) (int64, error)
}

type HackathonClient interface {
	GetHackathon(ctx context.Context, hackathonID string) (stage string, allowTeam bool, teamSizeMax int32, err error)
}

type OutboxRepository interface {
	Create(ctx context.Context, event *outbox.Event) error
}
