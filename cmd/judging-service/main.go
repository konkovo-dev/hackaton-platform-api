package main

import (
	"log/slog"
	"os"

	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/client/hackathon"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/client/participationroles"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/client/submission"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/repository/postgres"
	grpctransport "github.com/belikoooova/hackaton-platform-api/internal/judging-service/transport/grpc"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/usecase/judging"
	authclient "github.com/belikoooova/hackaton-platform-api/pkg/auth/client"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"go.uber.org/fx"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	app := fx.New(
		fx.Supply(logger),
		authclient.Module,
		postgres.Module,
		hackathon.Module,
		participationroles.Module,
		submission.Module,
		idempotency.Module,
		judging.Module,
		grpctransport.Module,
	)

	app.Run()
}
