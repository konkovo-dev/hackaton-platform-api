package participationandrolesservice

import (
	"context"
	"log/slog"

	commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"
	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/role"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (a *API) ListMyStaffInvitations(ctx context.Context, req *participationrolesv1.ListMyStaffInvitationsRequest) (*participationrolesv1.ListMyStaffInvitationsResponse, error) {
	result, err := a.roleService.ListMyInvitations(ctx, role.ListMyInvitationsIn{})
	if err != nil {
		return nil, a.handleError(ctx, err, "list_my_staff_invitations")
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

	a.logger.InfoContext(ctx, "list_my_staff_invitations: success",
		slog.Int("invitations_count", len(invitations)),
	)

	return &participationrolesv1.ListMyStaffInvitationsResponse{
		Invitations: invitations,
		Page:        &commonv1.PageResponse{},
	}, nil
}
