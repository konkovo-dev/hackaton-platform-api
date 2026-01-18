package hackathon

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
)

type HackathonRepository interface {
	Create(ctx context.Context, hackathon *entity.Hackathon) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Hackathon, error)
	Update(ctx context.Context, hackathon *entity.Hackathon) error
}

type HackathonLinkRepository interface {
	Create(ctx context.Context, link *entity.HackathonLink) error
	GetByHackathonID(ctx context.Context, hackathonID uuid.UUID) ([]*entity.HackathonLink, error)
	DeleteByHackathonID(ctx context.Context, hackathonID uuid.UUID) error
}

type OutboxRepository interface {
	Create(ctx context.Context, event *outbox.Event) error
}

type UnitOfWork = pgxutil.UnitOfWork[*TxRepositories]

type TxRepositories struct {
	Hackathons HackathonRepository
	Links      HackathonLinkRepository
	Outbox     OutboxRepository
}
