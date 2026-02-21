package teammember

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type MembershipRepository interface {
	ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]*entity.Membership, error)
	GetByTeamAndUser(ctx context.Context, teamID, userID uuid.UUID) (*entity.Membership, error)
	Delete(ctx context.Context, teamID, userID uuid.UUID) error
	UpdateCaptainStatus(ctx context.Context, teamID, userID uuid.UUID, isCaptain bool) error
	CheckIsCaptain(ctx context.Context, teamID, userID uuid.UUID) (bool, error)
}

type VacancyRepository interface {
	IncrementSlotsOpen(ctx context.Context, vacancyID uuid.UUID) error
}

type TeamRepository interface {
	GetByIDAndHackathonID(ctx context.Context, teamID, hackathonID uuid.UUID) (*entity.Team, error)
}

type TxManager interface {
	WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type HackathonClient interface {
	GetHackathon(ctx context.Context, hackathonID string) (stage string, allowTeam bool, teamSizeMax int32, err error)
}

type ParticipationAndRolesClient interface {
	GetHackathonContext(ctx context.Context, hackathonID string) (userID, participationStatus string, roles []string, err error)
	ConvertFromTeamParticipation(ctx context.Context, hackathonID, userID string) error
	ConvertToTeamParticipation(ctx context.Context, hackathonID, userID, teamID string, isCaptain bool) error
}
