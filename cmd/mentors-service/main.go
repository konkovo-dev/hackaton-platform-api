package main

import (
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/client/hackathon"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/client/participationroles"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/client/team"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/repository/postgres"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/transport/grpc"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/usecase/mentors"
	outboxusecase "github.com/belikoooova/hackaton-platform-api/internal/mentors-service/usecase/outbox"
	authclient "github.com/belikoooova/hackaton-platform-api/pkg/auth/client"
	"github.com/belikoooova/hackaton-platform-api/pkg/centrifugo"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/belikoooova/hackaton-platform-api/pkg/logger"
	natsclient "github.com/belikoooova/hackaton-platform-api/pkg/nats"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		logger.Module,
		authclient.Module,
		idempotency.Module,
		natsclient.Module,
		centrifugo.Module,
		postgres.Module,
		hackathon.Module,
		participationroles.Module,
		team.Module,
		mentors.Module,
		outboxusecase.Module,
		outbox.Module,
		grpc.Module,
	)

	app.Run()
}
