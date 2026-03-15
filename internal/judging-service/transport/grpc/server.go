package grpc

import (
	"log/slog"

	judgingv1 "github.com/belikoooova/hackaton-platform-api/api/judging/v1"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/client"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/interceptor"
	"github.com/belikoooova/hackaton-platform-api/pkg/env"
	commongrpc "github.com/belikoooova/hackaton-platform-api/pkg/grpc"
	"google.golang.org/grpc"
)

func NewGRPCServer(
	judgingService *JudgingServiceServer,
	authClient client.AuthClient,
	logger *slog.Logger,
) *grpc.Server {
	publicMethods := []string{}

	internalMethods := []string{}

	hybridMethods := []string{
		"/judging.v1.JudgingService/AssignSubmissionsToJudges",
		"/judging.v1.JudgingService/GetMyAssignments",
		"/judging.v1.JudgingService/SubmitEvaluation",
		"/judging.v1.JudgingService/GetMyEvaluations",
		"/judging.v1.JudgingService/GetSubmissionEvaluations",
		"/judging.v1.JudgingService/GetLeaderboard",
		"/judging.v1.JudgingService/GetMyEvaluationResult",
	}

	serviceToken := env.GetEnv("SERVICE_AUTH_TOKEN", "")

	authInterceptor := interceptor.NewUnaryInterceptorWithHybrid(authClient, publicMethods, nil, internalMethods, hybridMethods, serviceToken, logger)

	grpcServer := commongrpc.NewServer(commongrpc.ServerOptions{
		UnaryInterceptors: []grpc.UnaryServerInterceptor{authInterceptor},
	})

	judgingv1.RegisterJudgingServiceServer(grpcServer, judgingService)

	return grpcServer
}
