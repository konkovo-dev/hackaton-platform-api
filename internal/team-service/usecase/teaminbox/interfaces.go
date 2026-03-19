package teaminbox

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TeamInvitationRepository interface {
	ListByTeamID(ctx context.Context, teamID uuid.UUID, limit, offset int32) ([]*entity.TeamInvitation, error)
	CountByTeamID(ctx context.Context, teamID uuid.UUID) (int64, error)
	ListByTargetUserID(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*entity.TeamInvitation, error)
	CountByTargetUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	Create(ctx context.Context, invitation *entity.TeamInvitation) error
	GetByID(ctx context.Context, invitationID uuid.UUID) (*entity.TeamInvitation, error)
	UpdateStatus(ctx context.Context, invitationID uuid.UUID, status string) error
	CancelCompeting(ctx context.Context, userID, hackathonID uuid.UUID) error
	CheckUserInTeam(ctx context.Context, teamID, userID uuid.UUID) (bool, error)
}

type VacancyRepository interface {
	GetByID(ctx context.Context, vacancyID uuid.UUID) (*entity.Vacancy, error)
	GetByTeamID(ctx context.Context, teamID uuid.UUID) ([]*entity.Vacancy, error)
	Create(ctx context.Context, vacancy *entity.Vacancy) error
	DecrementSlotsOpen(ctx context.Context, vacancyID uuid.UUID) error
}

type TeamRepository interface {
	GetByIDAndHackathonID(ctx context.Context, teamID, hackathonID uuid.UUID) (*entity.Team, error)
}

type MembershipRepository interface {
	Create(ctx context.Context, membership *entity.Membership) error
	CheckIsCaptain(ctx context.Context, teamID, userID uuid.UUID) (bool, error)
	CountMembers(ctx context.Context, teamID uuid.UUID) (int64, error)
}

type JoinRequestRepository interface {
	ListByTeamID(ctx context.Context, teamID uuid.UUID, limit, offset int32) ([]*entity.JoinRequest, error)
	CountByTeamID(ctx context.Context, teamID uuid.UUID) (int64, error)
	ListByRequesterUserID(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*entity.JoinRequest, error)
	CountByRequesterUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	GetByID(ctx context.Context, requestID uuid.UUID) (*entity.JoinRequest, error)
	Create(ctx context.Context, request *entity.JoinRequest) error
	UpdateStatus(ctx context.Context, requestID uuid.UUID, status string) error
	CancelCompeting(ctx context.Context, userID, hackathonID uuid.UUID) error
}

type TeamRepositoryForJoinRequest interface {
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
	GetUserParticipation(ctx context.Context, hackathonID, userID string) (participationStatus string, err error)
	ConvertToTeamParticipation(ctx context.Context, hackathonID, userID, teamID string, isCaptain bool) error
}
