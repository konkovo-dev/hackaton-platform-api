package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	authv1 "github.com/belikoooova/hackaton-platform-api/api/auth/v1"
	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewListener(cfg *Config) (net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GatewayHTTPPort))
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	return listener, nil
}

func NewMux() *runtime.ServeMux {
	return runtime.NewServeMux()
}

func NewHTTPServer(mux *runtime.ServeMux, cfg *Config) *http.Server {
	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.GatewayHTTPPort),
		Handler: mux,
	}

	return httpSrv
}

func Run(lc fx.Lifecycle, s *http.Server, lis net.Listener, mux *runtime.ServeMux, cfg *Config, logger *slog.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

			bgCtx := context.Background()

			if err := identityv1.RegisterPingServiceHandlerFromEndpoint(bgCtx, mux, cfg.IdentityGRPCEndpoint, opts); err != nil {
				return fmt.Errorf("failed to register identity gateway handlers: %v", err)
			}

			if err := authv1.RegisterAuthServiceHandlerFromEndpoint(bgCtx, mux, cfg.AuthGRPCEndpoint, opts); err != nil {
				return fmt.Errorf("failed to register auth gateway handlers: %v", err)
			}

			if err := authv1.RegisterPingServiceHandlerFromEndpoint(bgCtx, mux, cfg.AuthGRPCEndpoint, opts); err != nil {
				return fmt.Errorf("failed to register auth ping gateway handlers: %v", err)
			}

			logger.Info("starting http gateway",
				slog.String("addr", lis.Addr().String()),
				slog.String("identity_grpc_endpoint", cfg.IdentityGRPCEndpoint),
				slog.String("auth_grpc_endpoint", cfg.AuthGRPCEndpoint),
			)

			go func() {
				if err := s.Serve(lis); err != nil {
					logger.Error("failed to start http gateway", slog.String("error", err.Error()))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping http gateway")
			return s.Shutdown(ctx)
		},
	})
}
