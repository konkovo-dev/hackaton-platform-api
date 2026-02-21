package team

import "go.uber.org/fx"

var Module = fx.Module("team",
	fx.Provide(
		NewService,
	),
)
