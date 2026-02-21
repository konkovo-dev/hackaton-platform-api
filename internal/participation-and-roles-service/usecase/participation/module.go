package participation

import "go.uber.org/fx"

var Module = fx.Module("participation",
	fx.Provide(NewService),
)
