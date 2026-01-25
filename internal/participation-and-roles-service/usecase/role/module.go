package role

import (
	"go.uber.org/fx"
)

var Module = fx.Module("role",
	fx.Provide(
		NewService,
	),
)
