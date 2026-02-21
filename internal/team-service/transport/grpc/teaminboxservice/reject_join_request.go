package teaminboxservice

import (
	"context"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/teaminbox"
	"github.com/google/uuid"
)

func (a *API) RejectJoinRequest(ctx context.Context, req *teamv1.RejectJoinRequestRequest) (*teamv1.RejectJoinRequestResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, a.handleError(ctx, err, "RejectJoinRequest")
	}

	teamID, err := uuid.Parse(req.TeamId)
	if err != nil {
		return nil, a.handleError(ctx, err, "RejectJoinRequest")
	}

	requestID, err := uuid.Parse(req.RequestId)
	if err != nil {
		return nil, a.handleError(ctx, err, "RejectJoinRequest")
	}

	_, err = a.teamInboxService.RejectJoinRequest(ctx, teaminbox.RejectJoinRequestIn{
		HackathonID: hackathonID,
		TeamID:      teamID,
		RequestID:   requestID,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "RejectJoinRequest")
	}

	return &teamv1.RejectJoinRequestResponse{}, nil
}
