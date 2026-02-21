package teaminbox

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type RejectJoinRequestIn struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
	RequestID   uuid.UUID
}

type RejectJoinRequestOut struct{}

func (s *Service) RejectJoinRequest(ctx context.Context, in RejectJoinRequestIn) (*RejectJoinRequestOut, error) {
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

	if request.TeamID != in.TeamID || request.HackathonID != in.HackathonID {
		return nil, ErrNotFound
	}

	stage, allowTeam, _, err := s.hackathonClient.GetHackathon(ctx, in.HackathonID.String())
	if err != nil {
		s.logger.Error("failed to get hackathon", "error", err)
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	_, err = s.teamRepo.GetByIDAndHackathonID(ctx, in.TeamID, in.HackathonID)
	if err != nil {
		return nil, ErrNotFound
	}

	isCaptain, err := s.membershipRepo.CheckIsCaptain(ctx, in.TeamID, userUUID)
	if err != nil {
		s.logger.Error("failed to check captain status", "error", err)
		return nil, fmt.Errorf("failed to check captain status: %w", err)
	}

	rejectPolicy := policy.NewRejectJoinRequestPolicy()
	pctx, err := rejectPolicy.LoadContext(ctx, policy.RejectJoinRequestParams{
		HackathonID: in.HackathonID,
		TeamID:      in.TeamID,
		RequestID:   in.RequestID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetHackathonStage(stage)
	pctx.SetAllowTeam(allowTeam)
	pctx.SetIsCaptain(isCaptain)
	pctx.SetRequestStatus(request.Status)

	decision := rejectPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	err = s.joinRequestRepo.UpdateStatus(ctx, in.RequestID, "declined")
	if err != nil {
		s.logger.Error("failed to reject join request", "error", err)
		return nil, fmt.Errorf("failed to reject join request: %w", err)
	}

	return &RejectJoinRequestOut{}, nil
}
