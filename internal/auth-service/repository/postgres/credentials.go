package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/usecase/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CredentialsRepository struct {
	queries *queries.Queries
}

func NewCredentialsRepository(pool *pgxpool.Pool) *CredentialsRepository {
	return &CredentialsRepository{
		queries: queries.New(pool),
	}
}

func (r *CredentialsRepository) WithTx(tx pgx.Tx) auth.CredentialsRepository {
	return &CredentialsRepository{
		queries: queries.New(tx),
	}
}

func (r *CredentialsRepository) Create(ctx context.Context, creds *entity.Credentials) error {
	now := time.Now().UTC()
	creds.CreatedAt = now
	creds.UpdatedAt = now

	err := r.queries.CreateCredentials(ctx, queries.CreateCredentialsParams{
		UserID:       uuidToPgtype(creds.UserID),
		PasswordHash: creds.PasswordHash,
		CreatedAt:    creds.CreatedAt,
		UpdatedAt:    creds.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("failed to create credentials: %w", err)
	}

	return nil
}

func (r *CredentialsRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Credentials, error) {
	row, err := r.queries.GetCredentialsByUserID(ctx, uuidToPgtype(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("credentials not found")
		}
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	return &entity.Credentials{
		UserID:       pgtypeToUUID(row.UserID),
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}, nil
}

func (r *CredentialsRepository) Update(ctx context.Context, creds *entity.Credentials) error {
	creds.UpdatedAt = time.Now().UTC()

	err := r.queries.UpdateCredentials(ctx, queries.UpdateCredentialsParams{
		UserID:       uuidToPgtype(creds.UserID),
		PasswordHash: creds.PasswordHash,
		UpdatedAt:    creds.UpdatedAt,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("credentials not found")
		}
		return fmt.Errorf("failed to update credentials: %w", err)
	}

	return nil
}
