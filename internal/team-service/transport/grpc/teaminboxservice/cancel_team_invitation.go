package teaminboxservice

import (
	"context"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/teaminbox"
	"github.com/google/uuid"
)

func (a *API) CancelTeamInvitation(ctx context.Context, req *teamv1.CancelTeamInvitationRequest) (*teamv1.CancelTeamInvitationResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, a.handleError(ctx, err, "CancelTeamInvitation")
	}

	teamID, err := uuid.Parse(req.TeamId)
	if err != nil {
		return nil, a.handleError(ctx, err, "CancelTeamInvitation")
	}

	invitationID, err := uuid.Parse(req.InvitationId)
	if err != nil {
		return nil, a.handleError(ctx, err, "CancelTeamInvitation")
	}

	_, err = a.teamInboxService.CancelTeamInvitation(ctx, teaminbox.CancelTeamInvitationIn{
		HackathonID:  hackathonID,
		TeamID:       teamID,
		InvitationID: invitationID,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "CancelTeamInvitation")
	}

	return &teamv1.CancelTeamInvitationResponse{}, nil
}
