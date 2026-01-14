package auth

import (
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/repository/postgres"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

var Module = fx.Module("auth",
	fx.Provide(
		MustNewConfig,
		NewService,
		NewUnitOfWork,
	),
)

func NewUnitOfWork(pool *pgxpool.Pool) UnitOfWork {
	factory := func(tx pgx.Tx) *TxRepositories {
		return &TxRepositories{
			Users:       postgres.NewUserRepository(tx),
			Credentials: postgres.NewCredentialsRepository(tx),
			Outbox:      postgres.NewOutboxRepository(tx),
		}
	}

	return pgxutil.NewUnitOfWork(pool, factory)
}
