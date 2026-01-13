package client

import "go.uber.org/fx"

var Module = fx.Module("auth-client",
	fx.Provide(
		MustNewConfig,
		NewAuthClient,
	),
)
