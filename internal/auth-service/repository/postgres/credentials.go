package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
)

type CredentialsRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewCredentialsRepository(db queries.DBTX) *CredentialsRepository {
	return &CredentialsRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *CredentialsRepository) Create(ctx context.Context, creds *entity.Credentials) error {
	now := time.Now().UTC()
	creds.CreatedAt = now
	creds.UpdatedAt = now

	err := r.Queries().CreateCredentials(ctx, queries.CreateCredentialsParams{
		UserID:       pgxutil.UUIDToPgtype(creds.UserID),
		PasswordHash: creds.PasswordHash,
		CreatedAt:    creds.CreatedAt,
		UpdatedAt:    creds.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("failed to create credentials: %w", pgxutil.MapDBError(err))
	}

	return nil
}

func (r *CredentialsRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Credentials, error) {
	row, err := r.Queries().GetCredentialsByUserID(ctx, pgxutil.UUIDToPgtype(userID))
	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return nil, fmt.Errorf("credentials not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	return &entity.Credentials{
		UserID:       pgxutil.PgtypeToUUID(row.UserID),
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}, nil
}

func (r *CredentialsRepository) Update(ctx context.Context, creds *entity.Credentials) error {
	creds.UpdatedAt = time.Now().UTC()

	err := r.Queries().UpdateCredentials(ctx, queries.UpdateCredentialsParams{
		UserID:       pgxutil.UUIDToPgtype(creds.UserID),
		PasswordHash: creds.PasswordHash,
		UpdatedAt:    creds.UpdatedAt,
	})

	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return fmt.Errorf("credentials not found: %w", err)
		}
		return fmt.Errorf("failed to update credentials: %w", err)
	}

	return nil
}
