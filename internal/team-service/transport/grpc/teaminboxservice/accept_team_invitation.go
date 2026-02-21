package teaminboxservice

import (
	"context"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/teaminbox"
	"github.com/google/uuid"
)

func (a *API) AcceptTeamInvitation(ctx context.Context, req *teamv1.AcceptTeamInvitationRequest) (*teamv1.AcceptTeamInvitationResponse, error) {
	invitationID, err := uuid.Parse(req.InvitationId)
	if err != nil {
		return nil, a.handleError(ctx, err, "AcceptTeamInvitation")
	}

	_, err = a.teamInboxService.AcceptTeamInvitation(ctx, teaminbox.AcceptTeamInvitationIn{
		InvitationID: invitationID,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "AcceptTeamInvitation")
	}

	return &teamv1.AcceptTeamInvitationResponse{}, nil
}
