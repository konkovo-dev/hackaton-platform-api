package idempotency

import "go.uber.org/fx"

var Module = fx.Module(
	"idempotency",
	fx.Provide(
		MustNewConfig,
		NewHelper,
	),
)
