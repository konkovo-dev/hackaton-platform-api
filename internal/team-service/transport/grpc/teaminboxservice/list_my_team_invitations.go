package teaminboxservice

import (
	"context"

	commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"
	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/teaminbox"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (a *API) ListMyTeamInvitations(ctx context.Context, req *teamv1.ListMyTeamInvitationsRequest) (*teamv1.ListMyTeamInvitationsResponse, error) {
	var pageSize uint32
	var pageToken string
	if req.Query != nil && req.Query.Page != nil {
		pageSize = req.Query.Page.PageSize
		pageToken = req.Query.Page.PageToken
	}

	out, err := a.teamInboxService.ListMyTeamInvitations(ctx, teaminbox.ListMyTeamInvitationsIn{
		PageSize:  pageSize,
		PageToken: pageToken,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "ListMyTeamInvitations")
	}

	protoInvitations := make([]*teamv1.TeamInvitation, 0, len(out.Invitations))
	for _, inv := range out.Invitations {
		protoInv := &teamv1.TeamInvitation{
			InvitationId:     inv.ID.String(),
			HackathonId:      inv.HackathonID.String(),
			TeamId:           inv.TeamID.String(),
			TargetUserId:     inv.TargetUserID.String(),
			VacancyId:        inv.VacancyID.String(),
			CreatedByUserId:  inv.CreatedByUserID.String(),
			Message:          inv.Message,
			Status:           mapStatusToProto(inv.Status),
			CreatedAt:        timestamppb.New(inv.CreatedAt),
			UpdatedAt:        timestamppb.New(inv.UpdatedAt),
		}

		if inv.ExpiresAt != nil {
			protoInv.ExpiresAt = timestamppb.New(*inv.ExpiresAt)
		}

		protoInvitations = append(protoInvitations, protoInv)
	}

	return &teamv1.ListMyTeamInvitationsResponse{
		Invitations: protoInvitations,
		Page: &commonv1.PageResponse{
			NextPageToken: out.NextPageToken,
		},
	}, nil
}
