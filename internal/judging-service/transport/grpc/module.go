package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

var Module = fx.Module("grpc-transport",
	fx.Provide(
		NewJudgingServiceServer,
		NewGRPCServer,
	),
	fx.Invoke(func(lc fx.Lifecycle, server *grpc.Server, logger *slog.Logger) {
		port := env.GetEnv("JUDGING_SERVICE_GRPC_PORT", "50060")
		addr := fmt.Sprintf(":%s", port)

		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				lis, err := net.Listen("tcp", addr)
				if err != nil {
					return fmt.Errorf("failed to listen: %w", err)
				}

				go func() {
					logger.Info("gRPC server starting", "addr", addr)
					if err := server.Serve(lis); err != nil {
						logger.Error("gRPC server failed", "error", err)
					}
				}()

				return nil
			},
			OnStop: func(ctx context.Context) error {
				logger.Info("stopping gRPC server")
				server.GracefulStop()
				return nil
			},
		})
	}),
)
