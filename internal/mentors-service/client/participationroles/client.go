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

func (c *Client) GetParticipationAndRoles(ctx context.Context, userID, hackathonID string) (roles []string, teamID *string, err error) {
	// Get roles from StaffService
	ctxReq := &participationandrolesv1.GetHackathonContextRequest{
		HackathonId: hackathonID,
	}

	if incomingMD, ok := metadata.FromIncomingContext(ctx); ok {
		if auth := incomingMD.Get("authorization"); len(auth) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", auth[0])
		}
	}

	ctxResp, err := c.staffService.GetHackathonContext(ctx, ctxReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get hackathon context: %w", err)
	}

	rolesStr := make([]string, 0, len(ctxResp.Roles))
	for _, protoRole := range ctxResp.Roles {
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

	// Check if user is a participant
	var participantRole string
	switch ctxResp.ParticipationStatus {
	case participationandrolesv1.ParticipationStatus_PART_INDIVIDUAL,
		participationandrolesv1.ParticipationStatus_PART_LOOKING_FOR_TEAM,
		participationandrolesv1.ParticipationStatus_PART_TEAM_MEMBER,
		participationandrolesv1.ParticipationStatus_PART_TEAM_CAPTAIN:
		participantRole = "participant"
	}
	if participantRole != "" {
		rolesStr = append(rolesStr, participantRole)
	}

	// Get team ID if user is in a team
	var teamIDPtr *string
	if ctxResp.ParticipationStatus == participationandrolesv1.ParticipationStatus_PART_TEAM_MEMBER ||
		ctxResp.ParticipationStatus == participationandrolesv1.ParticipationStatus_PART_TEAM_CAPTAIN {
		
		// Get full participation to extract team_id
		partReq := &participationandrolesv1.GetUserParticipationRequest{
			HackathonId: hackathonID,
			UserId:      userID,
		}

		ctx = metadata.AppendToOutgoingContext(ctx, "x-service-token", c.serviceToken)

		partResp, err := c.participationService.GetUserParticipation(ctx, partReq)
		if err == nil && partResp.Participation != nil {
			teamID := partResp.Participation.TeamId
			if teamID != "" {
				teamIDPtr = &teamID
			}
		}
	}

	return rolesStr, teamIDPtr, nil
}
