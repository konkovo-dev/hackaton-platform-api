package me

import "go.uber.org/fx"

var Module = fx.Module("me-usecase",
	fx.Provide(NewService),
)
