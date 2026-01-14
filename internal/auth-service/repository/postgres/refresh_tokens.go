package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshTokenRepository struct {
	queries *queries.Queries
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		queries: queries.New(pool),
	}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	token.CreatedAt = time.Now().UTC()

	var revokedAt pgtype.Timestamptz
	if token.RevokedAt != nil {
		revokedAt = pgtype.Timestamptz{
			Time:  *token.RevokedAt,
			Valid: true,
		}
	}

	err := r.queries.CreateRefreshToken(ctx, queries.CreateRefreshTokenParams{
		ID:        pgxutil.UUIDToPgtype(token.ID),
		UserID:    pgxutil.UUIDToPgtype(token.UserID),
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt,
		RevokedAt: revokedAt,
		CreatedAt: token.CreatedAt,
	})

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	row, err := r.queries.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("refresh token not found")
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return &entity.RefreshToken{
		ID:        pgxutil.PgtypeToUUID(row.ID),
		UserID:    pgxutil.PgtypeToUUID(row.UserID),
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt,
		RevokedAt: pgxutil.PgtypeTimestampToTimePtr(row.RevokedAt),
		CreatedAt: row.CreatedAt,
	}, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, tokenHash string, revokedAt time.Time) error {
	err := r.queries.RevokeRefreshToken(ctx, queries.RevokeRefreshTokenParams{
		TokenHash: tokenHash,
		RevokedAt: pgtype.Timestamptz{
			Time:  revokedAt,
			Valid: true,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error {
	err := r.queries.RevokeAllRefreshTokensByUserID(ctx, queries.RevokeAllRefreshTokensByUserIDParams{
		UserID: pgxutil.UUIDToPgtype(userID),
		RevokedAt: pgtype.Timestamptz{
			Time:  revokedAt,
			Valid: true,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to revoke all refresh tokens: %w", err)
	}

	return nil
}
