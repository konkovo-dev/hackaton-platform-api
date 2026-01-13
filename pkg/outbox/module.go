package outbox

import "go.uber.org/fx"

var Module = fx.Module("outbox",
	fx.Provide(
		MustNewConfig,
		NewPublisher,
		NewProcessor,
	),
)
