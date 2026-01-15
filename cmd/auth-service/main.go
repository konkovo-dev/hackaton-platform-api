package main

import (
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/client/identity"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/repository/postgres"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/transport/grpc"
	authUsecase "github.com/belikoooova/hackaton-platform-api/internal/auth-service/usecase/auth"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/usecase/jwt"
	outboxHandlers "github.com/belikoooova/hackaton-platform-api/internal/auth-service/usecase/outbox"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/usecase/password"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/belikoooova/hackaton-platform-api/pkg/logger"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		logger.Module,
		postgres.Module,
		password.Module,
		jwt.Module,
		authUsecase.Module,
		idempotency.Module,
		identity.Module,
		outboxHandlers.Module,
		outbox.Module,
		grpc.Module,
		fx.Provide(
			fx.Annotate(password.NewService, fx.As(new(authUsecase.PasswordService))),
			fx.Annotate(jwt.NewService, fx.As(new(authUsecase.JWTService))),
		),
		// Wire repositories to auth usecase interfaces
		fx.Provide(
			func(repo *postgres.UserRepository) authUsecase.UserRepository { return repo },
			func(repo *postgres.CredentialsRepository) authUsecase.CredentialsRepository { return repo },
			func(repo *postgres.RefreshTokenRepository) authUsecase.RefreshTokenRepository { return repo },
			func(pool *pgxpool.Pool) authUsecase.UnitOfWork {
				return pgxutil.NewUnitOfWork(pool, func(tx pgx.Tx) *authUsecase.TxRepositories {
					return &authUsecase.TxRepositories{
						Users:       postgres.NewUserRepository(tx),
						Credentials: postgres.NewCredentialsRepository(tx),
						Outbox:      postgres.NewOutboxRepository(tx),
					}
				})
			},
		),
	)

	app.Run()
}
