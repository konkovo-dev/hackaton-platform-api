package teaminboxservice

import (
	"context"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/teaminbox"
	"github.com/google/uuid"
)

func (a *API) CreateJoinRequest(ctx context.Context, req *teamv1.CreateJoinRequestRequest) (*teamv1.CreateJoinRequestResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, a.handleError(ctx, err, "CreateJoinRequest")
	}

	teamID, err := uuid.Parse(req.TeamId)
	if err != nil {
		return nil, a.handleError(ctx, err, "CreateJoinRequest")
	}

	vacancyID, err := uuid.Parse(req.VacancyId)
	if err != nil {
		return nil, a.handleError(ctx, err, "CreateJoinRequest")
	}

	out, err := a.teamInboxService.CreateJoinRequest(ctx, teaminbox.CreateJoinRequestIn{
		HackathonID: hackathonID,
		TeamID:      teamID,
		VacancyID:   vacancyID,
		Message:     req.Message,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "CreateJoinRequest")
	}

	return &teamv1.CreateJoinRequestResponse{
		RequestId: out.RequestID.String(),
	}, nil
}
