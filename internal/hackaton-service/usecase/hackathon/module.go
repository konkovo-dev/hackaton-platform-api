package hackathon

import "go.uber.org/fx"

var Module = fx.Module("hackathon",
	fx.Provide(
		NewService,
	),
)
