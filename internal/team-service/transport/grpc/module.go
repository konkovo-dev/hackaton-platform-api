package grpc

import (
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/transport/grpc/teaminboxservice"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/transport/grpc/teammembersservice"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/transport/grpc/teamsservice"
	commongrpc "github.com/belikoooova/hackaton-platform-api/pkg/grpc"
	"go.uber.org/fx"
)

var Module = fx.Module("grpc",
	fx.Provide(
		MustNewConfig,
		commongrpc.NewListener,
		teamsservice.New,
		teammembersservice.New,
		teaminboxservice.New,
		NewGRPCServer,
	),
	fx.Invoke(commongrpc.RunServer),
)
