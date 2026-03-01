package postgres

import (
	"context"
	"log/slog"

	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

var Module = fx.Module("postgres",
	fx.Provide(
		MustNewConfig,
		func(pool *pgxpool.Pool) *IdempotencyRepository {
			return NewIdempotencyRepository(pool)
		},
		func(r *IdempotencyRepository) idempotency.Repository {
			return r
		},
		func(pool *pgxpool.Pool) queries.DBTX { return pool },
		NewUserRepository,
		NewParticipationRepository,
		NewTeamRepository,
		NewVacancyRepository,
		func(pool *pgxpool.Pool) *OutboxRepository {
			return NewOutboxRepository(pool)
		},
		func(r *OutboxRepository) outbox.EventRepository {
			return r
		},
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
					logger.Info("matchmaking-service database connection pool initialized")
					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.Info("closing matchmaking-service database connection pool")
					pool.Close()
					return nil
				},
			})

			return pool, nil
		},
	),
)
