package participationroles

import (
	"context"
	"fmt"

	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn         *grpc.ClientConn
	staffService participationrolesv1.StaffServiceClient
	participationService participationrolesv1.ParticipationServiceClient
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

	staffService := participationrolesv1.NewStaffServiceClient(conn)
	participationService := participationrolesv1.NewParticipationServiceClient(conn)

	return &Client{
		conn:         conn,
		staffService: staffService,
		participationService: participationService,
		serviceToken: cfg.ServiceToken,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) GetHackathonContext(ctx context.Context, hackathonID string) (userID, participationStatus string, roles []string, err error) {
	req := &participationrolesv1.GetHackathonContextRequest{
		HackathonId: hackathonID,
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if auth := md.Get("authorization"); len(auth) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", auth[0])
		}
	}

	resp, err := c.staffService.GetHackathonContext(ctx, req)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to get hackathon context: %w", err)
	}

	rolesStr := make([]string, 0, len(resp.Roles))
	for _, protoRole := range resp.Roles {
		var roleStr string
		switch protoRole {
		case participationrolesv1.HackathonRole_HX_ROLE_OWNER:
			roleStr = "owner"
		case participationrolesv1.HackathonRole_HX_ROLE_ORGANIZER:
			roleStr = "organizer"
		case participationrolesv1.HackathonRole_HX_ROLE_MENTOR:
			roleStr = "mentor"
		case participationrolesv1.HackathonRole_HX_ROLE_JUDGE:
			roleStr = "judge"
		}
		if roleStr != "" {
			rolesStr = append(rolesStr, roleStr)
		}
	}

	var participationStatusStr string
	switch resp.ParticipationStatus {
	case participationrolesv1.ParticipationStatus_PART_NONE:
		participationStatusStr = "none"
	case participationrolesv1.ParticipationStatus_PART_INDIVIDUAL:
		participationStatusStr = "individual"
	case participationrolesv1.ParticipationStatus_PART_LOOKING_FOR_TEAM:
		participationStatusStr = "looking_for_team"
	case participationrolesv1.ParticipationStatus_PART_TEAM_MEMBER:
		participationStatusStr = "team_member"
	case participationrolesv1.ParticipationStatus_PART_TEAM_CAPTAIN:
		participationStatusStr = "team_captain"
	}

	return resp.UserId, participationStatusStr, rolesStr, nil
}

func (c *Client) ConvertToTeamParticipation(ctx context.Context, hackathonID, userID, teamID string, isCaptain bool) error {
	req := &participationrolesv1.ConvertToTeamParticipationRequest{
		HackathonId: hackathonID,
		UserId:      userID,
		TeamId:      teamID,
		IsCaptain:   isCaptain,
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if auth := md.Get("authorization"); len(auth) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", auth[0])
		}
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "x-service-token", c.serviceToken)

	_, err := c.participationService.ConvertToTeamParticipation(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to convert to team participation: %w", err)
	}

	return nil
}

func (c *Client) ConvertFromTeamParticipation(ctx context.Context, hackathonID, userID string) error {
	req := &participationrolesv1.ConvertFromTeamParticipationRequest{
		HackathonId: hackathonID,
		UserId:      userID,
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if auth := md.Get("authorization"); len(auth) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", auth[0])
		}
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "x-service-token", c.serviceToken)

	_, err := c.participationService.ConvertFromTeamParticipation(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to convert from team participation: %w", err)
	}

	return nil
}

func (c *Client) GetUserParticipation(ctx context.Context, hackathonID, userID string) (participationStatus string, err error) {
	req := &participationrolesv1.GetUserParticipationRequest{
		HackathonId: hackathonID,
		UserId:      userID,
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if auth := md.Get("authorization"); len(auth) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", auth[0])
		}
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "x-service-token", c.serviceToken)

	resp, err := c.participationService.GetUserParticipation(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to get user participation: %w", err)
	}

	var participationStatusStr string
	switch resp.Participation.Status {
	case participationrolesv1.ParticipationStatus_PART_NONE:
		participationStatusStr = "none"
	case participationrolesv1.ParticipationStatus_PART_INDIVIDUAL:
		participationStatusStr = "individual"
	case participationrolesv1.ParticipationStatus_PART_LOOKING_FOR_TEAM:
		participationStatusStr = "looking_for_team"
	case participationrolesv1.ParticipationStatus_PART_TEAM_MEMBER:
		participationStatusStr = "team_member"
	case participationrolesv1.ParticipationStatus_PART_TEAM_CAPTAIN:
		participationStatusStr = "team_captain"
	}

	return participationStatusStr, nil
}
