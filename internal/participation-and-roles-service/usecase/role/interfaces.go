package role

import (
	"context"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
)

type StaffRoleRepository interface {
	Create(ctx context.Context, role *entity.StaffRole) error
	GetByHackathonID(ctx context.Context, hackathonID uuid.UUID) ([]*entity.StaffRole, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.StaffRole, error)
	GetByHackathonAndUser(ctx context.Context, hackathonID, userID uuid.UUID) ([]*entity.StaffRole, error)
	Delete(ctx context.Context, hackathonID, userID uuid.UUID, role string) error
	HasRole(ctx context.Context, hackathonID, userID uuid.UUID, role string) (bool, error)
	GetRoleStrings(ctx context.Context, hackathonID, userID uuid.UUID) ([]string, error)
}

type StaffInvitationRepository interface {
	Create(ctx context.Context, invitation *entity.StaffInvitation) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.StaffInvitation, error)
	GetPendingInvitationForUser(ctx context.Context, hackathonID, userID uuid.UUID, role string) (*entity.StaffInvitation, error)
	GetByTargetUserID(ctx context.Context, userID uuid.UUID) ([]*entity.StaffInvitation, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, updatedAt time.Time) error
	GetStatusAndHackathonID(ctx context.Context, id uuid.UUID) (string, uuid.UUID, error)
	GetDetails(ctx context.Context, id uuid.UUID) (exists bool, status string, targetUserID uuid.UUID, hackathonID uuid.UUID, requestedRole string, err error)
	GetBasicInfo(ctx context.Context, id uuid.UUID) (exists bool, status string, targetUserID uuid.UUID, err error)
}

type ParticipationRepository interface {
	GetStatus(ctx context.Context, hackathonID, userID uuid.UUID) (string, error)
}

type UnitOfWork = pgxutil.UnitOfWork[*TxRepositories]

type TxRepositories struct {
	Roles       StaffRoleRepository
	Invitations StaffInvitationRepository
}
