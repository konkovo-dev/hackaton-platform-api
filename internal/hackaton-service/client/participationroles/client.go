package participationroles

import (
	"context"
	"fmt"

	commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"
	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn         *grpc.ClientConn
	client       participationrolesv1.ParticipationAndRolesServiceClient
	serviceToken string
}

func NewClient(cfg *Config) (*Client, error) {
	conn, err := grpc.NewClient(
		cfg.ParticipationRolesServiceURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial participation-roles service: %w", err)
	}

	client := participationrolesv1.NewParticipationAndRolesServiceClient(conn)

	return &Client{
		conn:         conn,
		client:       client,
		serviceToken: cfg.ServiceToken,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) CheckAccess(ctx context.Context, hackathonID string, policy participationrolesv1.AccessPolicy) (bool, error) {
	req := &participationrolesv1.CheckAccessRequest{
		HackathonId: hackathonID,
		Policy:      policy,
	}

	resp, err := c.client.CheckAccess(ctx, req)
	if err != nil {
		return false, fmt.Errorf("failed to check access: %w", err)
	}

	return resp.Allowed, nil
}

func (c *Client) AssignHackathonRole(ctx context.Context, hackathonID, userID string, role participationrolesv1.HackathonRole) error {
	ctx = c.addServiceToken(ctx)

	idempotencyKeyValue := idempotency.ComputeHash(hackathonID, userID, role.String())

	req := &participationrolesv1.AssignHackathonRoleRequest{
		HackathonId: hackathonID,
		UserId:      userID,
		Role:        role,
		IdempotencyKey: &commonv1.IdempotencyKey{
			Key: idempotencyKeyValue,
		},
	}

	_, err := c.client.AssignHackathonRole(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to assign hackathon role: %w", err)
	}

	return nil
}

func (c *Client) addServiceToken(ctx context.Context) context.Context {
	md := metadata.New(map[string]string{
		"x-service-token": c.serviceToken,
	})
	return metadata.NewOutgoingContext(ctx, md)
}
