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
	conn         *grpc.ClientConn
	teamService  teamv1.TeamMembersServiceClient
	serviceToken string
}

func NewClient(cfg *Config) (*Client, error) {
	conn, err := grpc.NewClient(
		cfg.TeamServiceURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial team service: %w", err)
	}

	teamService := teamv1.NewTeamMembersServiceClient(conn)

	return &Client{
		conn:         conn,
		teamService:  teamService,
		serviceToken: cfg.ServiceToken,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) GetTeamCaptain(ctx context.Context, hackathonID, teamID string) (captainUserID string, err error) {
	req := &teamv1.ListTeamMembersRequest{
		HackathonId: hackathonID,
		TeamId:      teamID,
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "x-service-token", c.serviceToken)
	if incomingMD, ok := metadata.FromIncomingContext(ctx); ok {
		if auth := incomingMD.Get("authorization"); len(auth) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", auth[0])
		}
	}

	resp, err := c.teamService.ListTeamMembers(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to list team members: %w", err)
	}

	for _, member := range resp.Members {
		if member.IsCaptain {
			return member.UserId, nil
		}
	}

	return "", fmt.Errorf("team has no captain")
}

func (c *Client) ListTeamMembers(ctx context.Context, hackathonID, teamID string) ([]string, error) {
	req := &teamv1.ListTeamMembersRequest{
		HackathonId: hackathonID,
		TeamId:      teamID,
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "x-service-token", c.serviceToken)
	if incomingMD, ok := metadata.FromIncomingContext(ctx); ok {
		if auth := incomingMD.Get("authorization"); len(auth) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", auth[0])
		}
	}

	resp, err := c.teamService.ListTeamMembers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list team members: %w", err)
	}

	userIDs := make([]string, 0, len(resp.Members))
	for _, member := range resp.Members {
		userIDs = append(userIDs, member.UserId)
	}

	return userIDs, nil
}
