package grpc

import "go.uber.org/fx"

var Module = fx.Module("grpc",
	fx.Provide(
		MustNewConfig,
		NewListener,
		NewGRPCServer,
	),
	fx.Invoke(Run),
)
