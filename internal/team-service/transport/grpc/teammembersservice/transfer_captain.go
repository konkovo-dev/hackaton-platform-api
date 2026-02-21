package teammembersservice

import (
	"context"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/teammember"
	"github.com/google/uuid"
)

func (a *API) TransferCaptain(ctx context.Context, req *teamv1.TransferCaptainRequest) (*teamv1.TransferCaptainResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, a.handleError(ctx, err, "TransferCaptain")
	}

	teamID, err := uuid.Parse(req.TeamId)
	if err != nil {
		return nil, a.handleError(ctx, err, "TransferCaptain")
	}

	newCaptainID, err := uuid.Parse(req.TargetUserId)
	if err != nil {
		return nil, a.handleError(ctx, err, "TransferCaptain")
	}

	_, err = a.teamMemberService.TransferCaptain(ctx, teammember.TransferCaptainIn{
		HackathonID:  hackathonID,
		TeamID:       teamID,
		NewCaptainID: newCaptainID,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "TransferCaptain")
	}

	return &teamv1.TransferCaptainResponse{}, nil
}
