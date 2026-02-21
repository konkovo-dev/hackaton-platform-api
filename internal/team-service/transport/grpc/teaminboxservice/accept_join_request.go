package teaminboxservice

import (
	"context"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/teaminbox"
	"github.com/google/uuid"
)

func (a *API) AcceptJoinRequest(ctx context.Context, req *teamv1.AcceptJoinRequestRequest) (*teamv1.AcceptJoinRequestResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, a.handleError(ctx, err, "AcceptJoinRequest")
	}

	teamID, err := uuid.Parse(req.TeamId)
	if err != nil {
		return nil, a.handleError(ctx, err, "AcceptJoinRequest")
	}

	requestID, err := uuid.Parse(req.RequestId)
	if err != nil {
		return nil, a.handleError(ctx, err, "AcceptJoinRequest")
	}

	_, err = a.teamInboxService.AcceptJoinRequest(ctx, teaminbox.AcceptJoinRequestIn{
		HackathonID: hackathonID,
		TeamID:      teamID,
		RequestID:   requestID,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "AcceptJoinRequest")
	}

	return &teamv1.AcceptJoinRequestResponse{}, nil
}
