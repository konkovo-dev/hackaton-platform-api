package grpc

import (
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/transport/grpc/pingservice"
	"go.uber.org/fx"
)

var Module = fx.Module("grpc",
	fx.Provide(
		MustNewConfig,
		NewListener,
		pingservice.New,
		NewGRPCServer,
	),
	fx.Invoke(Run),
)
