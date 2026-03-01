package matchmaking

import "go.uber.org/fx"

var Module = fx.Module("matchmaking",
	fx.Provide(
		NewService,
	),
)
