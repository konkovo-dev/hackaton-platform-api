package grpc

import (
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/transport/grpc/mentorsservice"
	commongrpc "github.com/belikoooova/hackaton-platform-api/pkg/grpc"
	"go.uber.org/fx"
)

var Module = fx.Module("grpc",
	fx.Provide(
		MustNewConfig,
		commongrpc.NewListener,
		mentorsservice.NewAPI,
		NewGRPCServer,
	),
	fx.Invoke(commongrpc.RunServer),
)
