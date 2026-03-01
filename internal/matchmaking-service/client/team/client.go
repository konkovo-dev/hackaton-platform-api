package team

import (
	"context"
	"fmt"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn               *grpc.ClientConn
	teamMembersService teamv1.TeamMembersServiceClient
	serviceToken       string
}

func NewClient(cfg *Config) (*Client, error) {
	conn, err := grpc.NewClient(
		cfg.TeamServiceURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial team service: %w", err)
	}

	teamMembersService := teamv1.NewTeamMembersServiceClient(conn)

	return &Client{
		conn:               conn,
		teamMembersService: teamMembersService,
		serviceToken:       cfg.ServiceToken,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) GetTeam(ctx context.Context, hackathonID, teamID string) (captainUserID string, err error) {
	req := &teamv1.ListTeamMembersRequest{
		HackathonId: hackathonID,
		TeamId:      teamID,
	}

	// Create fresh outgoing metadata
	md := metadata.New(map[string]string{
		"x-service-token": c.serviceToken,
	})
	
	// Add authorization if present
	if incomingMD, ok := metadata.FromIncomingContext(ctx); ok {
		if auth := incomingMD.Get("authorization"); len(auth) > 0 {
			md.Set("authorization", auth[0])
		}
	}
	
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := c.teamMembersService.ListTeamMembers(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to list team members: %w", err)
	}

	// Find captain
	for _, member := range resp.Members {
		if member.IsCaptain {
			return member.UserId, nil
		}
	}

	return "", fmt.Errorf("team has no captain")
}
