package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IdempotencyRepository struct {
	queries *queries.Queries
}

func NewIdempotencyRepository(pool *pgxpool.Pool) *IdempotencyRepository {
	return &IdempotencyRepository{
		queries: queries.New(pool),
	}
}

func (r *IdempotencyRepository) Get(ctx context.Context, key, scope string) (*idempotency.StoredResponse, error) {
	row, err := r.queries.GetIdempotencyKey(ctx, queries.GetIdempotencyKeyParams{
		Key:   key,
		Scope: scope,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, idempotency.ErrNotFound
		}
		return nil, err
	}

	return &idempotency.StoredResponse{
		Key:          row.Key,
		Scope:        row.Scope,
		RequestHash:  row.RequestHash,
		ResponseBlob: row.ResponseBlob,
		CreatedAt:    row.CreatedAt,
		ExpiresAt:    row.ExpiresAt,
	}, nil
}

func (r *IdempotencyRepository) Set(ctx context.Context, key, scope, requestHash string, responseBlob []byte, expiresAt time.Time) error {
	_, err := r.queries.SetIdempotencyKey(ctx, queries.SetIdempotencyKeyParams{
		Key:          key,
		Scope:        scope,
		RequestHash:  requestHash,
		ResponseBlob: responseBlob,
		ExpiresAt:    expiresAt,
	})
	return err
}
