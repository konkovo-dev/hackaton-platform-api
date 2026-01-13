package main

import (
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/repository/postgres"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/transport/grpc"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/me"
	authclient "github.com/belikoooova/hackaton-platform-api/pkg/auth/client"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/belikoooova/hackaton-platform-api/pkg/logger"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		logger.Module,
		authclient.Module,
		postgres.Module,
		idempotency.Module,
		me.Module,
		grpc.Module,
	)

	app.Run()
}
