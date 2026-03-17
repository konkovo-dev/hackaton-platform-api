package staffservice

import (
	"context"
	"log/slog"

	commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"
	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/role"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (a *API) ListHackathonStaffInvitations(ctx context.Context, req *participationrolesv1.ListHackathonStaffInvitationsRequest) (*participationrolesv1.ListHackathonStaffInvitationsResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, a.handleError(ctx, role.ErrInvalidInput, "list_hackathon_staff_invitations")
	}

	pageSize := uint32(50)
	pageToken := ""
	if req.Query != nil && req.Query.Page != nil {
		if req.Query.Page.PageSize > 0 {
			pageSize = req.Query.Page.PageSize
		}
		pageToken = req.Query.Page.PageToken
	}

	result, err := a.roleService.ListHackathonInvitations(ctx, role.ListHackathonInvitationsIn{
		HackathonID: hackathonID,
		PageSize:    pageSize,
		PageToken:   pageToken,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "list_hackathon_staff_invitations")
	}

	invitations := make([]*participationrolesv1.StaffInvitation, 0, len(result.Invitations))
	for _, inv := range result.Invitations {
		protoRole := domain.MapDomainRoleToProto(domain.HackathonRole(inv.RequestedRole))
		protoStatus := domain.MapDomainInvitationStatusToProto(domain.InvitationStatus(inv.Status))

		protoInv := &participationrolesv1.StaffInvitation{
			InvitationId:    inv.ID.String(),
			HackathonId:     inv.HackathonID.String(),
			TargetUserId:    inv.TargetUserID.String(),
			RequestedRole:   protoRole,
			CreatedByUserId: inv.CreatedByUserID.String(),
			Message:         inv.Message,
			Status:          protoStatus,
			CreatedAt:       timestamppb.New(inv.CreatedAt),
			UpdatedAt:       timestamppb.New(inv.UpdatedAt),
		}

		if inv.ExpiresAt != nil {
			protoInv.ExpiresAt = timestamppb.New(*inv.ExpiresAt)
		}

		invitations = append(invitations, protoInv)
	}

	a.logger.InfoContext(ctx, "list_hackathon_staff_invitations: success",
		slog.String("hackathon_id", req.HackathonId),
		slog.Int("invitations_count", len(invitations)),
	)

	return &participationrolesv1.ListHackathonStaffInvitationsResponse{
		Invitations: invitations,
		Page: &commonv1.PageResponse{
			NextPageToken: result.NextPageToken,
		},
	}, nil
}
