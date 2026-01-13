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
			fx.Annotate(postgres.NewUserRepository, fx.As(new(authUsecase.UserRepository))),
			fx.Annotate(postgres.NewCredentialsRepository, fx.As(new(authUsecase.CredentialsRepository))),
			fx.Annotate(postgres.NewRefreshTokenRepository, fx.As(new(authUsecase.RefreshTokenRepository))),
			fx.Annotate(password.NewService, fx.As(new(authUsecase.PasswordService))),
			fx.Annotate(jwt.NewService, fx.As(new(authUsecase.JWTService))),
			fx.Annotate(postgres.NewIdempotencyRepository, fx.As(new(idempotency.Repository))),
		),
	)

	app.Run()
}
