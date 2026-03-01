package main

import (
	"log/slog"
	"os"

	submissionservice "github.com/belikoooova/hackaton-platform-api/internal/submission-service"
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/client/hackathon"
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/client/participationroles"
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/client/team"
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/repository/postgres"
	grpctransport "github.com/belikoooova/hackaton-platform-api/internal/submission-service/transport/grpc"
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/usecase/cleanup"
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/usecase/submission"
	authclient "github.com/belikoooova/hackaton-platform-api/pkg/auth/client"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/belikoooova/hackaton-platform-api/pkg/s3"
	"go.uber.org/fx"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	app := fx.New(
		fx.Supply(logger),
		fx.Provide(submissionservice.MustNewConfig),
		authclient.Module,
		postgres.Module,
		s3.Module,
		hackathon.Module,
		participationroles.Module,
		team.Module,
		idempotency.Module,
		submission.Module,
		cleanup.Module,
		grpctransport.Module,
	)

	app.Run()
}
