package participationroles

import (
	"context"
	"fmt"

	commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"
	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn         *grpc.ClientConn
	staffService participationrolesv1.StaffServiceClient
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

	return &Client{
		conn:         conn,
		staffService: staffService,
		serviceToken: cfg.ServiceToken,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) AssignHackathonRole(ctx context.Context, hackathonID, userID string, role participationrolesv1.HackathonRole) error {
	idempotencyKeyValue := idempotency.ComputeHash(hackathonID, userID, role.String())

	req := &participationrolesv1.AssignHackathonRoleRequest{
		HackathonId: hackathonID,
		UserId:      userID,
		Role:        role,
		IdempotencyKey: &commonv1.IdempotencyKey{
			Key: idempotencyKeyValue,
		},
	}

	md := metadata.New(map[string]string{
		"x-service-token": c.serviceToken,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	_, err := c.staffService.AssignHackathonRole(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to assign hackathon role: %w", err)
	}

	return nil
}

func (c *Client) GetHackathonContext(ctx context.Context, hackathonID string) (userID, participationStatus, teamID string, roles []string, err error) {
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
		return "", "", "", nil, fmt.Errorf("failed to get hackathon context: %w", err)
	}

	rolesStr := make([]string, 0, len(resp.Roles))
	for _, protoRole := range resp.Roles {
		domainRole := domain.MapProtoRoleToDomain(protoRole)
		if domainRole != "" {
			rolesStr = append(rolesStr, string(domainRole))
		}
	}

	domainParticipation := domain.MapProtoParticipationToDomain(resp.ParticipationStatus)

	return resp.UserId, string(domainParticipation), "", rolesStr, nil
}
