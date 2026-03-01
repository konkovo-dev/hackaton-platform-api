package me

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
}

type SkillRepository interface {
	ListCatalogSkillsByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.CatalogSkill, error)
	GetUserCatalogSkills(ctx context.Context, userID uuid.UUID) ([]*entity.CatalogSkill, error)
	GetUserCustomSkills(ctx context.Context, userID uuid.UUID) ([]*entity.CustomSkill, error)
	Update(ctx context.Context, userID uuid.UUID, catalogSkillIDs []uuid.UUID, customSkillNames []string) error
}

type ContactRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Contact, error)
	Update(ctx context.Context, userID uuid.UUID, contacts []*entity.Contact) error
}

type VisibilityRepository interface {
	Create(ctx context.Context, visibility *entity.Visibility) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Visibility, error)
	Upsert(ctx context.Context, visibility *entity.Visibility) error
}

type UnitOfWork = pgxutil.UnitOfWork[*TxRepositories]

type OutboxRepository interface {
	Create(ctx context.Context, event *outbox.Event) error
}

type TxRepositories struct {
	Users      UserRepository
	Skills     SkillRepository
	Contacts   ContactRepository
	Visibility VisibilityRepository
	Outbox     OutboxRepository
}
