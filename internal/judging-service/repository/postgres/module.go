package postgres

import (
	"context"
	"log/slog"

	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

var Module = fx.Module("postgres",
	fx.Provide(
		MustNewConfig,
		func(r *IdempotencyRepository) idempotency.Repository {
			return r
		},
		func(pool *pgxpool.Pool) queries.DBTX { return pool },
		NewAssignmentRepository,
		NewEvaluationRepository,
		NewLeaderboardRepository,
		NewIdempotencyRepository,
		func(pool *pgxpool.Pool) *pgxutil.TxManager {
			return pgxutil.NewTxManager(pool)
		},
	),
	fx.Provide(
		func(lc fx.Lifecycle, cfg *pgxutil.Config, logger *slog.Logger) (*pgxpool.Pool, error) {
			pool, err := pgxutil.NewPool(context.Background(), cfg)
			if err != nil {
				return nil, err
			}

			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					logger.Info("judging-service database connection pool initialized")
					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.Info("closing judging-service database connection pool")
					pool.Close()
					return nil
				},
			})

			return pool, nil
		},
	),
)
