package identity

import (
	"context"
	"fmt"

	commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"
	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn      *grpc.ClientConn
	meService identityv1.MeServiceClient
}

func NewClient(cfg *Config) (*Client, error) {
	conn, err := grpc.NewClient(cfg.IdentityServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial identity service: %w", err)
	}

	meService := identityv1.NewMeServiceClient(conn)

	return &Client{
		conn:      conn,
		meService: meService,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) CreateUser(ctx context.Context, userID, username, firstName, lastName, timezone string) error {
	idempotencyKeyValue := idempotency.ComputeHash(userID, "identity.create_me")

	req := &identityv1.CreateMeRequest{
		UserId:    userID,
		Username:  username,
		FirstName: firstName,
		LastName:  lastName,
		Timezone:  timezone,
		IdempotencyKey: &commonv1.IdempotencyKey{
			Key: idempotencyKeyValue,
		},
	}

	_, err := c.meService.CreateMe(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create user in identity service: %w", err)
	}

	return nil
}
