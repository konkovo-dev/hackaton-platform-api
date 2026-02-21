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
	conn, err := grpc.NewClient(
		cfg.HackathonServiceURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
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

func (c *Client) GetHackathon(ctx context.Context, hackathonID string) (stage string, allowTeam bool, teamSizeMax int32, err error) {
	req := &hackathonv1.GetHackathonRequest{
		HackathonId:   hackathonID,
		IncludeLimits: true,
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "x-service-token", c.serviceToken)
	if incomingMD, ok := metadata.FromIncomingContext(ctx); ok {
		if auth := incomingMD.Get("authorization"); len(auth) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", auth[0])
		}
	}

	resp, err := c.hackathonService.GetHackathon(ctx, req)
	if err != nil {
		return "", false, 0, fmt.Errorf("failed to get hackathon: %w", err)
	}

	var stageStr string
	switch resp.Hackathon.Stage {
	case hackathonv1.HackathonStage_HACKATHON_STAGE_DRAFT:
		stageStr = "draft"
	case hackathonv1.HackathonStage_HACKATHON_STAGE_UPCOMING:
		stageStr = "upcoming"
	case hackathonv1.HackathonStage_HACKATHON_STAGE_REGISTRATION:
		stageStr = "registration"
	case hackathonv1.HackathonStage_HACKATHON_STAGE_PRE_START:
		stageStr = "prestart"
	case hackathonv1.HackathonStage_HACKATHON_STAGE_RUNNING:
		stageStr = "running"
	case hackathonv1.HackathonStage_HACKATHON_STAGE_JUDGING:
		stageStr = "judging"
	case hackathonv1.HackathonStage_HACKATHON_STAGE_FINISHED:
		stageStr = "finished"
	}

	var allowTeamVal bool
	var teamSizeMaxVal int32
	if resp.Hackathon.RegistrationPolicy != nil {
		allowTeamVal = resp.Hackathon.RegistrationPolicy.AllowTeam
	}
	if resp.Hackathon.Limits != nil {
		teamSizeMaxVal = int32(resp.Hackathon.Limits.TeamSizeMax)
	}

	return stageStr, allowTeamVal, teamSizeMaxVal, nil
}
