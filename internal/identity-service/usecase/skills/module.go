package skills

import "go.uber.org/fx"

var Module = fx.Module(
	"usecase.skills",
	fx.Provide(
		NewService,
	),
)
