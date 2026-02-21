package teaminboxservice

import (
	"context"

	commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"
	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/teaminbox"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (a *API) ListTeamInvitations(ctx context.Context, req *teamv1.ListTeamInvitationsRequest) (*teamv1.ListTeamInvitationsResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, a.handleError(ctx, err, "ListTeamInvitations")
	}

	teamID, err := uuid.Parse(req.TeamId)
	if err != nil {
		return nil, a.handleError(ctx, err, "ListTeamInvitations")
	}

	var pageSize uint32
	var pageToken string
	if req.Query != nil && req.Query.Page != nil {
		pageSize = req.Query.Page.PageSize
		pageToken = req.Query.Page.PageToken
	}

	out, err := a.teamInboxService.ListTeamInvitations(ctx, teaminbox.ListTeamInvitationsIn{
		HackathonID: hackathonID,
		TeamID:      teamID,
		PageSize:    pageSize,
		PageToken:   pageToken,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "ListTeamInvitations")
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

	return &teamv1.ListTeamInvitationsResponse{
		Invitations: protoInvitations,
		Page: &commonv1.PageResponse{
			NextPageToken: out.NextPageToken,
		},
	}, nil
}

func mapStatusToProto(status string) teamv1.TeamInboxStatus {
	switch status {
	case "pending":
		return teamv1.TeamInboxStatus_TEAM_INBOX_PENDING
	case "accepted":
		return teamv1.TeamInboxStatus_TEAM_INBOX_ACCEPTED
	case "declined":
		return teamv1.TeamInboxStatus_TEAM_INBOX_DECLINED
	case "canceled":
		return teamv1.TeamInboxStatus_TEAM_INBOX_CANCELED
	case "expired":
		return teamv1.TeamInboxStatus_TEAM_INBOX_EXPIRED
	default:
		return teamv1.TeamInboxStatus_TEAM_INBOX_STATUS_UNSPECIFIED
	}
}
