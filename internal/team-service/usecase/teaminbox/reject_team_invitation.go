package teaminbox

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type RejectTeamInvitationIn struct {
	InvitationID uuid.UUID
}

type RejectTeamInvitationOut struct{}

func (s *Service) RejectTeamInvitation(ctx context.Context, in RejectTeamInvitationIn) (*RejectTeamInvitationOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	invitation, err := s.invitationRepo.GetByID(ctx, in.InvitationID)
	if err != nil {
		return nil, ErrNotFound
	}

	stage, allowTeam, _, err := s.hackathonClient.GetHackathon(ctx, invitation.HackathonID.String())
	if err != nil {
		s.logger.Error("failed to get hackathon", "error", err)
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	rejectPolicy := policy.NewRejectTeamInvitationPolicy()
	pctx, err := rejectPolicy.LoadContext(ctx, policy.RejectTeamInvitationParams{
		InvitationID: in.InvitationID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetHackathonStage(stage)
	pctx.SetAllowTeam(allowTeam)
	pctx.SetInvitationStatus(invitation.Status)
	pctx.SetInvitationTarget(invitation.TargetUserID)

	decision := rejectPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	err = s.invitationRepo.UpdateStatus(ctx, in.InvitationID, "declined")
	if err != nil {
		s.logger.Error("failed to reject invitation", "error", err)
		return nil, fmt.Errorf("failed to reject invitation: %w", err)
	}

	return &RejectTeamInvitationOut{}, nil
}
