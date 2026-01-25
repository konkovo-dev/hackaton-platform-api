package identity

import "go.uber.org/fx"

var Module = fx.Module("identity-client",
	fx.Provide(
		MustNewConfig,
		NewClient,
	),
)
