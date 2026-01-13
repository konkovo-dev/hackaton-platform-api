package outbox

import "go.uber.org/fx"

var Module = fx.Module("outbox-handlers",
	fx.Provide(
		NewUserRegisteredHandler,
	),
)
