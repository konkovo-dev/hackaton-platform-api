package grpc

import (
	"log/slog"

	submissionv1 "github.com/belikoooova/hackaton-platform-api/api/submission/v1"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/client"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/interceptor"
	"github.com/belikoooova/hackaton-platform-api/pkg/env"
	commongrpc "github.com/belikoooova/hackaton-platform-api/pkg/grpc"
	"google.golang.org/grpc"
)

func NewGRPCServer(
	submissionService *SubmissionServiceServer,
	filesService *SubmissionFilesServiceServer,
	authClient client.AuthClient,
	logger *slog.Logger,
) *grpc.Server {
	publicMethods := []string{}

	internalMethods := []string{}

	hybridMethods := []string{
		"/submission.v1.SubmissionService/CreateSubmission",
		"/submission.v1.SubmissionService/UpdateSubmission",
		"/submission.v1.SubmissionService/ListSubmissions",
		"/submission.v1.SubmissionService/SelectFinalSubmission",
		"/submission.v1.SubmissionService/GetSubmission",
		"/submission.v1.SubmissionService/GetFinalSubmission",
		"/submission.v1.SubmissionFilesService/CreateSubmissionUpload",
		"/submission.v1.SubmissionFilesService/CompleteSubmissionUpload",
		"/submission.v1.SubmissionFilesService/GetSubmissionFileDownloadURL",
	}

	serviceToken := env.GetEnv("SERVICE_AUTH_TOKEN", "")

	authInterceptor := interceptor.NewUnaryInterceptorWithHybrid(authClient, publicMethods, nil, internalMethods, hybridMethods, serviceToken, logger)

	grpcServer := commongrpc.NewServer(commongrpc.ServerOptions{
		UnaryInterceptors: []grpc.UnaryServerInterceptor{authInterceptor},
	})

	submissionv1.RegisterSubmissionServiceServer(grpcServer, submissionService)
	submissionv1.RegisterSubmissionFilesServiceServer(grpcServer, filesService)

	return grpcServer
}
