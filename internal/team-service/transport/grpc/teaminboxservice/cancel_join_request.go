package teaminboxservice

import (
	"context"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/teaminbox"
	"github.com/google/uuid"
)

func (a *API) CancelJoinRequest(ctx context.Context, req *teamv1.CancelJoinRequestRequest) (*teamv1.CancelJoinRequestResponse, error) {
	requestID, err := uuid.Parse(req.RequestId)
	if err != nil {
		return nil, a.handleError(ctx, err, "CancelJoinRequest")
	}

	_, err = a.teamInboxService.CancelJoinRequest(ctx, teaminbox.CancelJoinRequestIn{
		RequestID: requestID,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "CancelJoinRequest")
	}

	return &teamv1.CancelJoinRequestResponse{}, nil
}
