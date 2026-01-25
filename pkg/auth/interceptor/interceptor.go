package interceptor

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type unaryInterceptor struct {
	authClient      client.AuthClient
	publicMethods   map[string]bool
	optionalMethods map[string]bool
	internalMethods map[string]bool
	hybridMethods   map[string]bool
	serviceToken    string
	logger          *slog.Logger
}

func NewUnaryInterceptor(
	authClient client.AuthClient,
	publicMethods []string,
	optionalMethods []string,
	internalMethods []string,
	serviceToken string,
	logger *slog.Logger,
) grpc.UnaryServerInterceptor {
	return NewUnaryInterceptorWithHybrid(authClient, publicMethods, optionalMethods, internalMethods, nil, serviceToken, logger)
}

func NewUnaryInterceptorWithHybrid(
	authClient client.AuthClient,
	publicMethods []string,
	optionalMethods []string,
	internalMethods []string,
	hybridMethods []string,
	serviceToken string,
	logger *slog.Logger,
) grpc.UnaryServerInterceptor {
	publicMap := make(map[string]bool, len(publicMethods))
	for _, method := range publicMethods {
		publicMap[method] = true
	}

	optionalMap := make(map[string]bool, len(optionalMethods))
	for _, method := range optionalMethods {
		optionalMap[method] = true
	}

	internalMap := make(map[string]bool, len(internalMethods))
	for _, method := range internalMethods {
		internalMap[method] = true
	}

	hybridMap := make(map[string]bool, len(hybridMethods))
	for _, method := range hybridMethods {
		hybridMap[method] = true
	}

	i := &unaryInterceptor{
		authClient:      authClient,
		publicMethods:   publicMap,
		optionalMethods: optionalMap,
		internalMethods: internalMap,
		hybridMethods:   hybridMap,
		serviceToken:    serviceToken,
		logger:          logger,
	}

	return i.intercept
}

func (i *unaryInterceptor) intercept(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	method := info.FullMethod

	if i.internalMethods[method] {
		if err := i.verifyServiceToken(ctx, method); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}

	if i.hybridMethods[method] {
		// Try service token first
		if err := i.verifyServiceToken(ctx, method); err == nil {
			// Set service-to-service flag in context
			ctx = auth.WithServiceCall(ctx)
			return handler(ctx, req)
		}

		// If service token fails, try user token
		token, err := i.extractToken(ctx, method)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "authentication required (either service token or user token)")
		}

		claims, err := i.authenticate(ctx, token, method)
		if err != nil {
			return nil, err
		}

		ctx = auth.WithClaims(ctx, claims)
		return handler(ctx, req)
	}

	if i.publicMethods[method] {
		return handler(ctx, req)
	}

	if i.optionalMethods[method] {
		token, err := i.extractToken(ctx, method)
		if err != nil {
			return handler(ctx, req)
		}

		claims, err := i.authenticate(ctx, token, method)
		if err != nil {
			return handler(ctx, req)
		}

		ctx = auth.WithClaims(ctx, claims)
		return handler(ctx, req)
	}

	token, err := i.extractToken(ctx, method)
	if err != nil {
		return nil, err
	}

	claims, err := i.authenticate(ctx, token, method)
	if err != nil {
		return nil, err
	}

	ctx = auth.WithClaims(ctx, claims)

	return handler(ctx, req)
}

func (i *unaryInterceptor) extractToken(ctx context.Context, method string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		i.logAuthFailure(method, "missing_metadata", nil)
		return "", status.Error(codes.Unauthenticated, "missing authentication metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		i.logAuthFailure(method, "missing_token", nil)
		return "", status.Error(codes.Unauthenticated, "missing authorization token")
	}

	token := values[0]
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimSpace(token)

	if token == "" {
		i.logAuthFailure(method, "empty_token", nil)
		return "", status.Error(codes.Unauthenticated, "empty authorization token")
	}

	return token, nil
}

func (i *unaryInterceptor) authenticate(ctx context.Context, token, method string) (*auth.Claims, error) {
	claims, err := i.authClient.IntrospectToken(ctx, token)
	if err != nil {
		if errors.Is(err, client.ErrInvalidToken) {
			i.logAuthFailure(method, "invalid_token", err)
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		i.logAuthFailure(method, "auth_service_error", err)
		return nil, status.Error(codes.Internal, "authentication service error")
	}

	return claims, nil
}

func (i *unaryInterceptor) logAuthFailure(method, reason string, err error) {
	if i.logger == nil {
		return
	}

	if err != nil {
		i.logger.Warn("authentication failed",
			"method", method,
			"reason", reason,
			"error", err,
		)
	} else {
		i.logger.Warn("authentication failed",
			"method", method,
			"reason", reason,
		)
	}
}

func (i *unaryInterceptor) extractServiceToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("x-service-token")
	if len(values) == 0 {
		return "", status.Error(codes.PermissionDenied, "service authentication required")
	}

	token := strings.TrimSpace(values[0])
	if token == "" {
		return "", status.Error(codes.PermissionDenied, "empty service token")
	}

	return token, nil
}

func (i *unaryInterceptor) verifyServiceToken(ctx context.Context, method string) error {
	token, err := i.extractServiceToken(ctx)
	if err != nil {
		i.logAuthFailure(method, "missing_service_token", err)
		return err
	}

	if token != i.serviceToken {
		i.logAuthFailure(method, "invalid_service_token", nil)
		return status.Error(codes.PermissionDenied, "invalid service token")
	}

	return nil
}
