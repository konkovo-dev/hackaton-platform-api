package postgres

import (
	"context"
	"log/slog"

	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/usecase/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

var Module = fx.Module("postgres",
	fx.Provide(
		MustNewConfig,
		NewTxManager,
		NewUserRepository,
		NewCredentialsRepository,
		NewRefreshTokenRepository,
		NewIdempotencyRepository,
		NewOutboxRepository,
		func(r *OutboxRepository) outbox.EventRepository { return r },
		func(r *OutboxRepository) auth.OutboxRepository { return r },
	),
	fx.Provide(
		func(lc fx.Lifecycle, cfg *Config, logger *slog.Logger) (*pgxpool.Pool, error) {
			pool, err := NewPool(context.Background(), cfg)
			if err != nil {
				return nil, err
			}

			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					logger.Info("database connection pool initialized")
					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.Info("closing database connection pool")
					pool.Close()
					return nil
				},
			})

			return pool, nil
		},
	),
)
