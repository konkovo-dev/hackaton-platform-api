package grpc

import (
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/transport/grpc/participationandrolesservice"
	commongrpc "github.com/belikoooova/hackaton-platform-api/pkg/grpc"
	"go.uber.org/fx"
)

var Module = fx.Module("grpc",
	fx.Provide(
		MustNewConfig,
		commongrpc.NewListener,
		participationandrolesservice.New,
		NewGRPCServer,
	),
	fx.Invoke(commongrpc.RunServer),
)
