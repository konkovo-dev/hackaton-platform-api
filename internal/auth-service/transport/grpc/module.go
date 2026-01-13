package grpc

import (
	authv1 "github.com/belikoooova/hackaton-platform-api/api/auth/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/transport/grpc/authservice"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/transport/grpc/pingservice"
	commongrpc "github.com/belikoooova/hackaton-platform-api/pkg/grpc"
	"go.uber.org/fx"
)

var Module = fx.Module("grpc",
	fx.Provide(
		MustNewConfig,
		commongrpc.NewListener,
		pingservice.New,
		authservice.NewAuthService,
		fx.Annotate(
			authservice.NewAuthService,
			fx.As(new(authv1.AuthServiceServer)),
		),
		NewGRPCServer,
	),
	fx.Invoke(commongrpc.RunServer),
)
