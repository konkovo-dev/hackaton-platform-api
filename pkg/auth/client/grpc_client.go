package client

import (
	"context"
	"fmt"

	authv1 "github.com/belikoooova/hackaton-platform-api/api/auth/v1"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type authClient struct {
	conn   *grpc.ClientConn
	client authv1.AuthServiceClient
}

func newAuthClient(cfg *Config) (*authClient, error) {
	conn, err := grpc.NewClient(cfg.AuthServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial auth service: %w", err)
	}

	client := authv1.NewAuthServiceClient(conn)

	return &authClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *authClient) IntrospectToken(ctx context.Context, token string) (*auth.Claims, error) {
	req := &authv1.IntrospectTokenRequest{
		AccessToken: token,
	}

	resp, err := c.client.IntrospectToken(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to introspect token: %w", err)
	}

	if !resp.Active {
		return nil, ErrInvalidToken
	}

	return &auth.Claims{
		UserID:    resp.UserId,
		ExpiresAt: resp.ExpiresAt.AsTime(),
	}, nil
}
