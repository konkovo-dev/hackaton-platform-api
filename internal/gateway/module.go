package gateway

import "go.uber.org/fx"

var Module = fx.Module("gateway",
	fx.Provide(
		NewConfig,
		NewListener,
		NewMux,
		NewHTTPServer,
	),
	fx.Invoke(Run),
)
