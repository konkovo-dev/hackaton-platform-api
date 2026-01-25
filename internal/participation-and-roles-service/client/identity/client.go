package identity

import (
	"context"
	"fmt"

	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Client struct {
	conn         *grpc.ClientConn
	usersService identityv1.UsersServiceClient
	serviceToken string
}

func NewClient(cfg *Config) (*Client, error) {
	conn, err := grpc.NewClient(cfg.IdentityServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial identity service: %w", err)
	}

	usersService := identityv1.NewUsersServiceClient(conn)

	return &Client{
		conn:         conn,
		usersService: usersService,
		serviceToken: cfg.ServiceToken,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) UserExists(ctx context.Context, userID string) (bool, error) {
	md := metadata.New(map[string]string{
		"x-service-token": c.serviceToken,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	req := &identityv1.GetUserRequest{
		UserId: userID,
	}

	_, err := c.usersService.GetUser(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return true, nil
}
