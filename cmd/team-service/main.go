package main

import (
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/transport/grpc"
	"github.com/belikoooova/hackaton-platform-api/pkg/logger"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		logger.Module,
		grpc.Module,
	)

	app.Run()
}

