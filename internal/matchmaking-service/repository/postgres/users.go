package postgres

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewUserRepository(db queries.DBTX) *UserRepository {
	return &UserRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *UserRepository) Upsert(ctx context.Context, user *entity.User) error {
	err := r.Queries().UpsertUser(ctx, queries.UpsertUserParams{
		UserID:           user.UserID,
		Username:         user.Username,
		AvatarUrl:        pgtype.Text{String: user.AvatarURL, Valid: user.AvatarURL != ""},
		CatalogSkillIds:  user.CatalogSkillIDs,
		CustomSkillNames: user.CustomSkillNames,
		UpdatedAt:        user.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("failed to upsert user: %w", pgxutil.MapDBError(err))
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	row, err := r.Queries().GetUserByID(ctx, userID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &entity.User{
		UserID:           row.UserID,
		Username:         row.Username,
		AvatarURL:        row.AvatarUrl.String,
		CatalogSkillIDs:  row.CatalogSkillIds,
		CustomSkillNames: row.CustomSkillNames,
		UpdatedAt:        row.UpdatedAt,
	}, nil
}

func (r *UserRepository) GetByIDs(ctx context.Context, userIDs []uuid.UUID) ([]*entity.User, error) {
	rows, err := r.Queries().GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", pgxutil.MapDBError(err))
	}

	users := make([]*entity.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, &entity.User{
			UserID:           row.UserID,
			Username:         row.Username,
			AvatarURL:        row.AvatarUrl.String,
			CatalogSkillIDs:  row.CatalogSkillIds,
			CustomSkillNames: row.CustomSkillNames,
			UpdatedAt:        row.UpdatedAt,
		})
	}

	return users, nil
}
