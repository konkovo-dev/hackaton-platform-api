package gateway

import (
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	IdentityGRPCEndpoint              string
	AuthGRPCEndpoint                  string
	HackatonGRPCEndpoint              string
	TeamGRPCEndpoint                  string
	ParticipationAndRolesGRPCEndpoint string
	MentorsGRPCEndpoint               string
	MatchmakingGRPCEndpoint           string
	GatewayHTTPPort                   int
}

func NewConfig() *Config {
	identityGRPCEndpoint := env.GetEnv("IDENTITY_GRPC_ENDPOINT", "localhost:50051")
	authGRPCEndpoint := env.GetEnv("AUTH_GRPC_ENDPOINT", "localhost:50057")
	hackatonGRPCEndpoint := env.GetEnv("HACKATON_GRPC_ENDPOINT", "localhost:50052")
	teamGRPCEndpoint := env.GetEnv("TEAM_GRPC_ENDPOINT", "localhost:50053")
	participationAndRolesGRPCEndpoint := env.GetEnv("PARTICIPATION_ROLES_GRPC_ENDPOINT", "localhost:50055")
	mentorsGRPCEndpoint := env.GetEnv("MENTORS_GRPC_ENDPOINT", "localhost:50056")
	matchmakingGRPCEndpoint := env.GetEnv("MATCHMAKING_GRPC_ENDPOINT", "localhost:50059")

	gatewayHTTPPort, err := env.GetEnvInt("GATEWAY_HTTP_PORT", 8080)
	if err != nil {
		panic(fmt.Errorf("invalid GATEWAY_HTTP_PORT: %w", err))
	}

	return &Config{
		IdentityGRPCEndpoint:              identityGRPCEndpoint,
		AuthGRPCEndpoint:                  authGRPCEndpoint,
		HackatonGRPCEndpoint:              hackatonGRPCEndpoint,
		TeamGRPCEndpoint:                  teamGRPCEndpoint,
		ParticipationAndRolesGRPCEndpoint: participationAndRolesGRPCEndpoint,
		MentorsGRPCEndpoint:               mentorsGRPCEndpoint,
		MatchmakingGRPCEndpoint:           matchmakingGRPCEndpoint,
		GatewayHTTPPort:                   gatewayHTTPPort,
	}
}
