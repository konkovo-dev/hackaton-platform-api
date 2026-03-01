package participation

import (
	"context"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/google/uuid"
)

type ListParticipationsFilter struct {
	Statuses      []string
	WishedRoleIDs []uuid.UUID
	Limit         int32
	Offset        int32
}

type ParticipationRepository interface {
	Get(ctx context.Context, hackathonID, userID uuid.UUID) (*entity.Participation, error)
	Create(ctx context.Context, participation *entity.Participation) error
	GetStatus(ctx context.Context, hackathonID, userID uuid.UUID) (string, error)
	UpdateProfile(ctx context.Context, hackathonID, userID uuid.UUID, motivationText string, updatedAt time.Time) error
	UpdateStatus(ctx context.Context, hackathonID, userID uuid.UUID, status string, updatedAt time.Time) error
	Update(ctx context.Context, hackathonID, userID uuid.UUID, status string, teamID *uuid.UUID, updatedAt time.Time) error
	Delete(ctx context.Context, hackathonID, userID uuid.UUID) error
	List(ctx context.Context, hackathonID uuid.UUID, filter ListParticipationsFilter) ([]*entity.Participation, int64, error)
}

type StaffRoleRepository interface {
	GetRoleStrings(ctx context.Context, hackathonID, userID uuid.UUID) ([]string, error)
}

type TeamRoleRepository interface {
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.TeamRole, error)
	GetByParticipation(ctx context.Context, hackathonID, userID uuid.UUID) ([]*entity.TeamRole, error)
	SetForParticipation(ctx context.Context, hackathonID, userID uuid.UUID, roleIDs []uuid.UUID) error
	ListAll(ctx context.Context) ([]*entity.TeamRole, error)
}

type OutboxRepository interface {
	Create(ctx context.Context, event *outbox.Event) error
}
