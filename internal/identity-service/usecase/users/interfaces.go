package users

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/queryutil/sqlbuilder"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.User, error)
	ListUsers(ctx context.Context, params ListUsersRepoParams) ([]*UserListResult, bool, error)
}

type UserListResult struct {
	ID       uuid.UUID
	Username string
}

type ListUsersRepoParams struct {
	SearchQuery  string
	Filters      *ListUsersFilters
	Sort         []sqlbuilder.SortField
	Cursor       *ListUsersCursor
	Limit        int
	FieldMapping sqlbuilder.FieldMapping
}

type ListUsersFilters struct {
	FilterGroups    []*sqlbuilder.FilterGroup
	HasSkillsFilter bool
}

type ListUsersCursor struct {
	Username string    `json:"username"`
	UserID   uuid.UUID `json:"user_id"`
}

type SkillRepository interface {
	GetUserCatalogSkills(ctx context.Context, userID uuid.UUID) ([]*entity.CatalogSkill, error)
	GetUserCustomSkills(ctx context.Context, userID uuid.UUID) ([]*entity.CustomSkill, error)
	GetUsersCatalogSkills(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID][]*entity.CatalogSkill, error)
	GetUsersCustomSkills(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID][]*entity.CustomSkill, error)
}

type ContactRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Contact, error)
	GetByUserIDs(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID][]*entity.Contact, error)
}

type VisibilityRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Visibility, error)
	GetByUserIDs(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]*entity.Visibility, error)
}
