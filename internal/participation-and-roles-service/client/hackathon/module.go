package hackathon

import "go.uber.org/fx"

var Module = fx.Module("hackathon-client",
	fx.Provide(
		MustNewConfig,
		NewClient,
	),
)
