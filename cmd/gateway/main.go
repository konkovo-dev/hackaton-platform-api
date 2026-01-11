package main

import (
	"github.com/belikoooova/hackaton-platform-api/internal/gateway"
	"github.com/belikoooova/hackaton-platform-api/pkg/logger"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		logger.Module,
		gateway.Module,
	)

	app.Run()
}
