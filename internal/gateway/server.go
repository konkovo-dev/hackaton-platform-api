package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"

	authv1 "github.com/belikoooova/hackaton-platform-api/api/auth/v1"
	hackathonv1 "github.com/belikoooova/hackaton-platform-api/api/hackathon/v1"
	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	participationandrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func NewListener(cfg *Config) (net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GatewayHTTPPort))
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	return listener, nil
}

func NewMux() *runtime.ServeMux {
	return runtime.NewServeMux(
		runtime.WithMetadata(func(ctx context.Context, req *http.Request) metadata.MD {
			md := metadata.MD{}

			// Forward Authorization header
			if auth := req.Header.Get("Authorization"); auth != "" {
				md.Set("authorization", auth)
			}

			// Forward X-Service-Token header (for internal service calls)
			if token := req.Header.Get("X-Service-Token"); token != "" {
				md.Set("x-service-token", token)
			}

			return md
		}),
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			switch strings.ToLower(key) {
			case "authorization", "x-service-token":
				return key, true
			default:
				return runtime.DefaultHeaderMatcher(key)
			}
		}),
	)
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
				return fmt.Errorf("failed to register identity ping gateway handlers: %v", err)
			}

			if err := identityv1.RegisterUsersServiceHandlerFromEndpoint(bgCtx, mux, cfg.IdentityGRPCEndpoint, opts); err != nil {
				return fmt.Errorf("failed to register identity users service gateway handlers: %v", err)
			}

			if err := identityv1.RegisterMeServiceHandlerFromEndpoint(bgCtx, mux, cfg.IdentityGRPCEndpoint, opts); err != nil {
				return fmt.Errorf("failed to register identity me service gateway handlers (MUST be after UsersService): %v", err)
			}

			if err := identityv1.RegisterSkillsServiceHandlerFromEndpoint(bgCtx, mux, cfg.IdentityGRPCEndpoint, opts); err != nil {
				return fmt.Errorf("failed to register identity skills service gateway handlers: %v", err)
			}

			if err := authv1.RegisterPingServiceHandlerFromEndpoint(bgCtx, mux, cfg.AuthGRPCEndpoint, opts); err != nil {
				return fmt.Errorf("failed to register auth ping gateway handlers: %v", err)
			}

			if err := authv1.RegisterAuthServiceHandlerFromEndpoint(bgCtx, mux, cfg.AuthGRPCEndpoint, opts); err != nil {
				return fmt.Errorf("failed to register auth gateway handlers: %v", err)
			}

			if err := hackathonv1.RegisterHackathonServiceHandlerFromEndpoint(bgCtx, mux, cfg.HackatonGRPCEndpoint, opts); err != nil {
				return fmt.Errorf("failed to register hackathon service gateway handlers: %v", err)
			}

			if err := participationandrolesv1.RegisterParticipationAndRolesServiceHandlerFromEndpoint(bgCtx, mux, cfg.ParticipationAndRolesGRPCEndpoint, opts); err != nil {
				return fmt.Errorf("failed to register participation and roles service gateway handlers: %v", err)
			}

			logger.Info("starting http gateway",
				slog.String("addr", lis.Addr().String()),
				slog.String("identity_grpc_endpoint", cfg.IdentityGRPCEndpoint),
				slog.String("auth_grpc_endpoint", cfg.AuthGRPCEndpoint),
				slog.String("hackaton_grpc_endpoint", cfg.HackatonGRPCEndpoint),
				slog.String("participation_roles_grpc_endpoint", cfg.ParticipationAndRolesGRPCEndpoint),
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
