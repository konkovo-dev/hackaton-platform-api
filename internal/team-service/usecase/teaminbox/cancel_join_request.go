package teaminbox

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type CancelJoinRequestIn struct {
	RequestID uuid.UUID
}

type CancelJoinRequestOut struct{}

func (s *Service) CancelJoinRequest(ctx context.Context, in CancelJoinRequestIn) (*CancelJoinRequestOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	request, err := s.joinRequestRepo.GetByID(ctx, in.RequestID)
	if err != nil {
		return nil, ErrNotFound
	}

	stage, allowTeam, _, err := s.hackathonClient.GetHackathon(ctx, request.HackathonID.String())
	if err != nil {
		s.logger.Error("failed to get hackathon", "error", err)
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	cancelPolicy := policy.NewCancelJoinRequestPolicy()
	pctx, err := cancelPolicy.LoadContext(ctx, policy.CancelJoinRequestParams{
		RequestID: in.RequestID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetHackathonStage(stage)
	pctx.SetAllowTeam(allowTeam)
	pctx.SetRequestStatus(request.Status)
	pctx.SetRequestRequester(request.RequesterUserID)

	decision := cancelPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	err = s.joinRequestRepo.UpdateStatus(ctx, in.RequestID, "canceled")
	if err != nil {
		s.logger.Error("failed to cancel join request", "error", err)
		return nil, fmt.Errorf("failed to cancel join request: %w", err)
	}

	return &CancelJoinRequestOut{}, nil
}
