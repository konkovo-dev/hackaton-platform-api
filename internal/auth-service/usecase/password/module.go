package password

import "go.uber.org/fx"

var Module = fx.Module("password",
	fx.Provide(
		NewService,
	),
)

