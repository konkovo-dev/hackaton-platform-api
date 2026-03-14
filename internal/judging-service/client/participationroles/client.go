package participationroles

import (
	"context"
	"fmt"

	participationandrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn                 *grpc.ClientConn
	staffService         participationandrolesv1.StaffServiceClient
	participationService participationandrolesv1.ParticipationServiceClient
	serviceToken         string
}

func NewClient(cfg *Config) (*Client, error) {
	conn, err := grpc.NewClient(
		cfg.ParticipationRolesServiceURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial participation-roles service: %w", err)
	}

	staffService := participationandrolesv1.NewStaffServiceClient(conn)
	participationService := participationandrolesv1.NewParticipationServiceClient(conn)

	return &Client{
		conn:                 conn,
		staffService:         staffService,
		participationService: participationService,
		serviceToken:         cfg.ServiceToken,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) GetHackathonContext(ctx context.Context, hackathonID string) (userID, participationStatus string, roles []string, teamID string, err error) {
	req := &participationandrolesv1.GetHackathonContextRequest{
		HackathonId: hackathonID,
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "x-service-token", c.serviceToken)
	if incomingMD, ok := metadata.FromIncomingContext(ctx); ok {
		if auth := incomingMD.Get("authorization"); len(auth) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", auth[0])
		}
	}

	resp, err := c.staffService.GetHackathonContext(ctx, req)
	if err != nil {
		return "", "", nil, "", fmt.Errorf("failed to get hackathon context: %w", err)
	}

	userID = resp.UserId

	rolesStr := make([]string, 0, len(resp.Roles))
	for _, protoRole := range resp.Roles {
		var roleStr string
		switch protoRole {
		case participationandrolesv1.HackathonRole_HX_ROLE_OWNER:
			roleStr = "owner"
		case participationandrolesv1.HackathonRole_HX_ROLE_ORGANIZER:
			roleStr = "organizer"
		case participationandrolesv1.HackathonRole_HX_ROLE_MENTOR:
			roleStr = "mentor"
		case participationandrolesv1.HackathonRole_HX_ROLE_JUDGE:
			roleStr = "judge"
		}
		if roleStr != "" {
			rolesStr = append(rolesStr, roleStr)
		}
	}

	switch resp.ParticipationStatus {
	case participationandrolesv1.ParticipationStatus_PART_INDIVIDUAL,
		participationandrolesv1.ParticipationStatus_PART_LOOKING_FOR_TEAM,
		participationandrolesv1.ParticipationStatus_PART_TEAM_MEMBER,
		participationandrolesv1.ParticipationStatus_PART_TEAM_CAPTAIN:
		participationStatus = "active"
	default:
		participationStatus = "none"
	}

	if resp.ParticipationStatus == participationandrolesv1.ParticipationStatus_PART_TEAM_MEMBER ||
		resp.ParticipationStatus == participationandrolesv1.ParticipationStatus_PART_TEAM_CAPTAIN {

		partReq := &participationandrolesv1.GetUserParticipationRequest{
			HackathonId: hackathonID,
			UserId:      userID,
		}

		ctx = metadata.AppendToOutgoingContext(ctx, "x-service-token", c.serviceToken)

		partResp, err := c.participationService.GetUserParticipation(ctx, partReq)
		if err == nil && partResp.Participation != nil {
			teamID = partResp.Participation.TeamId
		}
	}

	return userID, participationStatus, rolesStr, teamID, nil
}

func (c *Client) ListJudges(ctx context.Context, hackathonID string) ([]string, error) {
	req := &participationandrolesv1.ListHackathonStaffRequest{
		HackathonId: hackathonID,
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "x-service-token", c.serviceToken)
	if incomingMD, ok := metadata.FromIncomingContext(ctx); ok {
		if auth := incomingMD.Get("authorization"); len(auth) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", auth[0])
		}
	}

	resp, err := c.staffService.ListHackathonStaff(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list hackathon staff: %w", err)
	}

	// Filter only judges from all staff members
	judgeIDs := make([]string, 0)
	for _, staff := range resp.Staff {
		for _, role := range staff.Roles {
			if role == participationandrolesv1.HackathonRole_HX_ROLE_JUDGE {
				judgeIDs = append(judgeIDs, staff.UserId)
				break
			}
		}
	}

	return judgeIDs, nil
}
