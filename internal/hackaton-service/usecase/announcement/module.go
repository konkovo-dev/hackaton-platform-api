package announcement

import "go.uber.org/fx"

var Module = fx.Module("announcement",
	fx.Provide(
		NewService,
	),
)

