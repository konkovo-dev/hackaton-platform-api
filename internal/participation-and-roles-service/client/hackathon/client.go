package hackathon

import (
	"context"
	"fmt"

	hackathonv1 "github.com/belikoooova/hackaton-platform-api/api/hackathon/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn             *grpc.ClientConn
	hackathonService hackathonv1.HackathonServiceClient
	serviceToken     string
}

func NewClient(cfg *Config) (*Client, error) {
	conn, err := grpc.NewClient(cfg.HackathonServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial hackathon service: %w", err)
	}

	hackathonService := hackathonv1.NewHackathonServiceClient(conn)

	return &Client{
		conn:             conn,
		hackathonService: hackathonService,
		serviceToken:     cfg.ServiceToken,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) GetHackathonInfo(ctx context.Context, hackathonID string) (*HackathonInfo, error) {
	md := metadata.New(map[string]string{
		"x-service-token": c.serviceToken,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	req := &hackathonv1.GetHackathonRequest{
		HackathonId:        hackathonID,
		IncludeDescription: false,
		IncludeLinks:       false,
		IncludeLimits:      true,
	}

	resp, err := c.hackathonService.GetHackathon(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	return &HackathonInfo{
		Stage:           resp.Hackathon.Stage.String(),
		AllowIndividual: resp.Hackathon.RegistrationPolicy.AllowIndividual,
		AllowTeam:       resp.Hackathon.RegistrationPolicy.AllowTeam,
	}, nil
}

type HackathonInfo struct {
	Stage           string
	AllowIndividual bool
	AllowTeam       bool
}
