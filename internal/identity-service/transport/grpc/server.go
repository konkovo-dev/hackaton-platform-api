package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/transport/grpc/meservice"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/transport/grpc/pingservice"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/client"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/interceptor"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func NewListener(cfg *Config) (net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	return listener, nil
}

func NewGRPCServer(
	pingService *pingservice.PingService,
	meService *meservice.MeService,
	authClient client.AuthClient,
	logger *slog.Logger,
) *grpc.Server {
	publicMethods := []string{
		"/identity.v1.PingService/Ping",
		"/identity.v1.MeService/CreateMe",
	}

	authInterceptor := interceptor.NewUnaryInterceptor(authClient, publicMethods, logger)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor),
	)

	identityv1.RegisterPingServiceServer(grpcServer, pingService)
	identityv1.RegisterMeServiceServer(grpcServer, meService)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(grpcServer)

	return grpcServer
}

func Run(lc fx.Lifecycle, s *grpc.Server, lis net.Listener, logger *slog.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting grpc server", slog.String("addr", lis.Addr().String()))
			go func() {
				if err := s.Serve(lis); err != nil {
					logger.Error("grpc serve error", slog.String("error", err.Error()))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping grpc server")
			_ = lis.Close()

			stopped := make(chan struct{})
			go func() {
				s.GracefulStop()
				close(stopped)
			}()

			select {
			case <-ctx.Done():
				s.Stop()
				logger.Warn("grpc server stopped forcefully")
			case <-stopped:
				logger.Info("grpc server stopped gracefully")
			}
			return nil
		},
	})
}
