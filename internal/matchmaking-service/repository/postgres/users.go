package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
	db queries.DBTX // kept for raw queries that bypass sqlc type system
}

func NewUserRepository(db queries.DBTX) *UserRepository {
	return &UserRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
		db:             db,
	}
}

// UpdateSkills updates only catalog_skill_ids and custom_skill_names for an existing user.
// Uses raw query with text[]::uuid[] cast to bypass pgx encoding issues with []uuid.UUID.
// sqlc is NOT used here to avoid generated-code type conflicts in CI (sqlc overrides manual .sql.go edits).
func (r *UserRepository) UpdateSkills(ctx context.Context, userID uuid.UUID, catalogSkillIDs []uuid.UUID, customSkillNames []string, updatedAt time.Time) error {
	skillStrs := make([]string, len(catalogSkillIDs))
	for i, id := range catalogSkillIDs {
		skillStrs[i] = id.String()
	}
	if customSkillNames == nil {
		customSkillNames = []string{}
	}
	_, err := r.db.Exec(ctx, `
		UPDATE matchmaking.users
		SET
			catalog_skill_ids  = $2::text[]::uuid[],
			custom_skill_names = $3,
			updated_at         = $4
		WHERE user_id = $1
	`, userID, skillStrs, customSkillNames, updatedAt)
	if err != nil {
		return fmt.Errorf("failed to update user skills: %w", pgxutil.MapDBError(err))
	}
	return nil
}

// InsertStubIfNotExists inserts a minimal user record only if the user doesn't exist yet.
// Unlike Upsert, it never overwrites existing data (e.g. skills synced from identity-service).

func (r *UserRepository) InsertStubIfNotExists(ctx context.Context, user *entity.User) error {
	catalogSkillIDs := user.CatalogSkillIDs
	if catalogSkillIDs == nil {
		catalogSkillIDs = []uuid.UUID{}
	}
	customSkillNames := user.CustomSkillNames
	if customSkillNames == nil {
		customSkillNames = []string{}
	}
	err := r.Queries().InsertUserStubIfNotExists(ctx, queries.InsertUserStubIfNotExistsParams{
		UserID:           user.UserID,
		Username:         user.Username,
		AvatarUrl:        pgtype.Text{String: user.AvatarURL, Valid: user.AvatarURL != ""},
		CatalogSkillIds:  catalogSkillIDs,
		CustomSkillNames: customSkillNames,
		UpdatedAt:        user.UpdatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to insert user stub: %w", pgxutil.MapDBError(err))
	}
	return nil
}

func (r *UserRepository) Upsert(ctx context.Context, user *entity.User) error {
	catalogSkillIDs := user.CatalogSkillIDs
	if catalogSkillIDs == nil {
		catalogSkillIDs = []uuid.UUID{}
	}
	customSkillNames := user.CustomSkillNames
	if customSkillNames == nil {
		customSkillNames = []string{}
	}
	err := r.Queries().UpsertUser(ctx, queries.UpsertUserParams{
		UserID:           user.UserID,
		Username:         user.Username,
		AvatarUrl:        pgtype.Text{String: user.AvatarURL, Valid: user.AvatarURL != ""},
		CatalogSkillIds:  catalogSkillIDs,
		CustomSkillNames: customSkillNames,
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
