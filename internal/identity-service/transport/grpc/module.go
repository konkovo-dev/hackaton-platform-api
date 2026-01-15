package grpc

import (
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/transport/grpc/meservice"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/transport/grpc/pingservice"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/transport/grpc/skillsservice"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/transport/grpc/usersservice"
	commongrpc "github.com/belikoooova/hackaton-platform-api/pkg/grpc"
	"go.uber.org/fx"
)

var Module = fx.Module("grpc",
	fx.Provide(
		MustNewConfig,
		commongrpc.NewListener,
		pingservice.New,
		meservice.NewMeService,
		usersservice.NewUsersService,
		skillsservice.NewSkillsService,
		NewGRPCServer,
	),
	fx.Invoke(commongrpc.RunServer),
)
