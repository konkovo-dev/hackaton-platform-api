package teaminboxservice

import (
	"context"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/teaminbox"
	"github.com/google/uuid"
)

func (a *API) CreateTeamInvitation(ctx context.Context, req *teamv1.CreateTeamInvitationRequest) (*teamv1.CreateTeamInvitationResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, a.handleError(ctx, err, "CreateTeamInvitation")
	}

	teamID, err := uuid.Parse(req.TeamId)
	if err != nil {
		return nil, a.handleError(ctx, err, "CreateTeamInvitation")
	}

	targetUserID, err := uuid.Parse(req.TargetUserId)
	if err != nil {
		return nil, a.handleError(ctx, err, "CreateTeamInvitation")
	}

	vacancyID, err := uuid.Parse(req.VacancyId)
	if err != nil {
		return nil, a.handleError(ctx, err, "CreateTeamInvitation")
	}

	out, err := a.teamInboxService.CreateTeamInvitation(ctx, teaminbox.CreateTeamInvitationIn{
		HackathonID:  hackathonID,
		TeamID:       teamID,
		TargetUserID: targetUserID,
		VacancyID:    vacancyID,
		Message:      req.Message,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "CreateTeamInvitation")
	}

	return &teamv1.CreateTeamInvitationResponse{
		InvitationId: out.InvitationID.String(),
	}, nil
}
