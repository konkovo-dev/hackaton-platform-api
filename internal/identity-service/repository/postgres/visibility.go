package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type VisibilityRepository struct {
	queries *queries.Queries
}

func NewVisibilityRepository(db queries.DBTX) *VisibilityRepository {
	return &VisibilityRepository{
		queries: queries.New(db),
	}
}

func (r *VisibilityRepository) Create(ctx context.Context, visibility *entity.Visibility) error {
	err := r.queries.VisibilityCreate(ctx, queries.VisibilityCreateParams{
		UserID:             pgxutil.UUIDToPgtype(visibility.UserID),
		SkillsVisibility:   string(visibility.SkillsVisibility),
		ContactsVisibility: string(visibility.ContactsVisibility),
	})

	if err != nil {
		return fmt.Errorf("failed to create user visibility: %w", err)
	}

	return nil
}

func (r *VisibilityRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Visibility, error) {
	row, err := r.queries.VisibilityGetByUserID(ctx, pgxutil.UUIDToPgtype(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("visibility not found")
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
	err := r.queries.VisibilityUpsert(ctx, queries.VisibilityUpsertParams{
		UserID:             pgxutil.UUIDToPgtype(visibility.UserID),
		SkillsVisibility:   string(visibility.SkillsVisibility),
		ContactsVisibility: string(visibility.ContactsVisibility),
	})

	if err != nil {
		return fmt.Errorf("failed to upsert user visibility: %w", err)
	}

	return nil
}
