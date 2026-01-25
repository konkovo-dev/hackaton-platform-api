package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
)

type IdempotencyRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewIdempotencyRepository(db queries.DBTX) *IdempotencyRepository {
	return &IdempotencyRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *IdempotencyRepository) Get(ctx context.Context, key, scope string) (*idempotency.StoredResponse, error) {
	row, err := r.Queries().GetIdempotencyKey(ctx, queries.GetIdempotencyKeyParams{
		Key:   key,
		Scope: scope,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return nil, idempotency.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get idempotency key: %w", err)
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
	err := r.Queries().CreateIdempotencyKey(ctx, queries.CreateIdempotencyKeyParams{
		Key:          key,
		Scope:        scope,
		RequestHash:  requestHash,
		ResponseBlob: responseBlob,
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to set idempotency key: %w", err)
	}
	return nil
}

func (r *IdempotencyRepository) DeleteExpired(ctx context.Context) error {
	err := r.Queries().DeleteExpiredIdempotencyKeys(ctx)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to delete expired idempotency keys: %w", err)
	}
	return nil
}
