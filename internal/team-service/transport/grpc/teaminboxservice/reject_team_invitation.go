package teaminboxservice

import (
	"context"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/teaminbox"
	"github.com/google/uuid"
)

func (a *API) RejectTeamInvitation(ctx context.Context, req *teamv1.RejectTeamInvitationRequest) (*teamv1.RejectTeamInvitationResponse, error) {
	invitationID, err := uuid.Parse(req.InvitationId)
	if err != nil {
		return nil, a.handleError(ctx, err, "RejectTeamInvitation")
	}

	_, err = a.teamInboxService.RejectTeamInvitation(ctx, teaminbox.RejectTeamInvitationIn{
		InvitationID: invitationID,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "RejectTeamInvitation")
	}

	return &teamv1.RejectTeamInvitationResponse{}, nil
}
