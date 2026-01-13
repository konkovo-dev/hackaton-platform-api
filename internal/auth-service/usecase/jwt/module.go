package jwt

import "go.uber.org/fx"

var Module = fx.Module("jwt",
	fx.Provide(
		MustNewConfig,
		NewService,
	),
)
