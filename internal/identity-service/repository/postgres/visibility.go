package postgres

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type VisibilityRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewVisibilityRepository(db queries.DBTX) *VisibilityRepository {
	return &VisibilityRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *VisibilityRepository) Create(ctx context.Context, visibility *entity.Visibility) error {
	err := r.Queries().VisibilityCreate(ctx, queries.VisibilityCreateParams{
		UserID:             pgxutil.UUIDToPgtype(visibility.UserID),
		SkillsVisibility:   string(visibility.SkillsVisibility),
		ContactsVisibility: string(visibility.ContactsVisibility),
	})

	if err != nil {
		return fmt.Errorf("failed to create user visibility: %w", pgxutil.MapDBError(err))
	}

	return nil
}

func (r *VisibilityRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Visibility, error) {
	row, err := r.Queries().VisibilityGetByUserID(ctx, pgxutil.UUIDToPgtype(userID))
	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return nil, fmt.Errorf("visibility not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user visibility: %w", err)
	}

	return &entity.Visibility{
		UserID:             pgxutil.PgtypeToUUID(row.UserID),
		SkillsVisibility:   domain.VisibilityLevel(row.SkillsVisibility),
		ContactsVisibility: domain.VisibilityLevel(row.ContactsVisibility),
	}, nil
}

func (r *VisibilityRepository) Upsert(ctx context.Context, visibility *entity.Visibility) error {
	err := r.Queries().VisibilityUpsert(ctx, queries.VisibilityUpsertParams{
		UserID:             pgxutil.UUIDToPgtype(visibility.UserID),
		SkillsVisibility:   string(visibility.SkillsVisibility),
		ContactsVisibility: string(visibility.ContactsVisibility),
	})

	if err != nil {
		return fmt.Errorf("failed to upsert user visibility: %w", pgxutil.MapDBError(err))
	}

	return nil
}

func (r *VisibilityRepository) GetByUserIDs(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]*entity.Visibility, error) {
	if len(userIDs) == 0 {
		return map[uuid.UUID]*entity.Visibility{}, nil
	}

	pgIDs := make([]pgtype.UUID, len(userIDs))
	for i, id := range userIDs {
		pgIDs[i] = pgxutil.UUIDToPgtype(id)
	}

	rows, err := r.Queries().VisibilityGetByUserIDs(ctx, pgIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get users visibility: %w", pgxutil.MapDBError(err))
	}

	result := make(map[uuid.UUID]*entity.Visibility)
	for _, row := range rows {
		userID := pgxutil.PgtypeToUUID(row.UserID)
		result[userID] = &entity.Visibility{
			UserID:             userID,
			SkillsVisibility:   domain.VisibilityLevel(row.SkillsVisibility),
			ContactsVisibility: domain.VisibilityLevel(row.ContactsVisibility),
		}
	}

	return result, nil
}
