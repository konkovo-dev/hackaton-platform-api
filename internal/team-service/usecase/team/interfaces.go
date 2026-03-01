package team

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/google/uuid"
)

type TeamRepository interface {
	List(ctx context.Context, hackathonID uuid.UUID, limit, offset int32) ([]*entity.Team, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Team, error)
	GetByIDAndHackathonID(ctx context.Context, id, hackathonID uuid.UUID) (*entity.Team, error)
	Create(ctx context.Context, team *entity.Team) error
	CheckNameExists(ctx context.Context, hackathonID uuid.UUID, name string) (bool, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, team *entity.Team) error
}

type VacancyRepository interface {
	GetByTeamIDs(ctx context.Context, teamIDs []uuid.UUID) ([]*entity.Vacancy, error)
	GetByTeamID(ctx context.Context, teamID uuid.UUID) ([]*entity.Vacancy, error)
}

type MembershipRepository interface {
	Create(ctx context.Context, membership *entity.Membership) error
	CheckIsCaptain(ctx context.Context, teamID, userID uuid.UUID) (bool, error)
	CountMembers(ctx context.Context, teamID uuid.UUID) (int64, error)
}

type HackathonClient interface {
	GetHackathon(ctx context.Context, hackathonID string) (stage string, allowTeam bool, teamSizeMax int32, err error)
}

type ParticipationAndRolesClient interface {
	GetHackathonContext(ctx context.Context, hackathonID string) (userID, participationStatus string, roles []string, err error)
	ConvertToTeamParticipation(ctx context.Context, hackathonID, userID, teamID string, isCaptain bool) error
	ConvertFromTeamParticipation(ctx context.Context, hackathonID, userID string) error
}

type OutboxRepository interface {
	Create(ctx context.Context, event *outbox.Event) error
}
