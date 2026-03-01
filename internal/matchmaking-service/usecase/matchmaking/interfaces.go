package matchmaking

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/domain/entity"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetByID(ctx context.Context, userID uuid.UUID) (*entity.User, error)
	GetByIDs(ctx context.Context, userIDs []uuid.UUID) ([]*entity.User, error)
}

type ParticipationRepository interface {
	Get(ctx context.Context, hackathonID, userID uuid.UUID) (*entity.Participation, error)
	ListByHackathon(ctx context.Context, hackathonID uuid.UUID) ([]*entity.Participation, error)
}

type TeamRepository interface {
	GetByID(ctx context.Context, teamID uuid.UUID) (*entity.Team, error)
	ListByHackathon(ctx context.Context, hackathonID uuid.UUID) ([]*entity.Team, error)
}

type VacancyRepository interface {
	GetByID(ctx context.Context, vacancyID uuid.UUID) (*entity.Vacancy, error)
	GetByTeamID(ctx context.Context, teamID uuid.UUID) ([]*entity.Vacancy, error)
	ListByHackathon(ctx context.Context, hackathonID uuid.UUID) ([]*entity.Vacancy, error)
}

type HackathonClient interface {
	GetHackathon(ctx context.Context, hackathonID string) (stage string, err error)
}

type ParticipationRolesClient interface {
	GetParticipationAndRoles(ctx context.Context, userID, hackathonID string) (roles []string, teamID *string, err error)
}

type TeamClient interface {
	GetTeam(ctx context.Context, hackathonID, teamID string) (captainUserID string, err error)
}
