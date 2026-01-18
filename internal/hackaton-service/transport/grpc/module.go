package grpc

import (
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/transport/grpc/hackathonservice"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/transport/grpc/pingservice"
	commongrpc "github.com/belikoooova/hackaton-platform-api/pkg/grpc"
	"go.uber.org/fx"
)

var Module = fx.Module("grpc",
	fx.Provide(
		MustNewConfig,
		commongrpc.NewListener,
		pingservice.New,
		hackathonservice.NewHackathonService,
		NewGRPCServer,
	),
	fx.Invoke(commongrpc.RunServer),
)
