package teaminbox

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type CancelTeamInvitationIn struct {
	HackathonID  uuid.UUID
	TeamID       uuid.UUID
	InvitationID uuid.UUID
}

type CancelTeamInvitationOut struct{}

func (s *Service) CancelTeamInvitation(ctx context.Context, in CancelTeamInvitationIn) (*CancelTeamInvitationOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	stage, allowTeam, _, err := s.hackathonClient.GetHackathon(ctx, in.HackathonID.String())
	if err != nil {
		s.logger.Error("failed to get hackathon", "error", err)
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	team, err := s.teamRepo.GetByIDAndHackathonID(ctx, in.TeamID, in.HackathonID)
	if err != nil {
		return nil, ErrNotFound
	}

	isCaptain, err := s.membershipRepo.CheckIsCaptain(ctx, in.TeamID, userUUID)
	if err != nil {
		s.logger.Error("failed to check captain status", "error", err)
		return nil, fmt.Errorf("failed to check captain status: %w", err)
	}

	invitation, err := s.invitationRepo.GetByID(ctx, in.InvitationID)
	if err != nil {
		return nil, ErrNotFound
	}

	invitationBelongs := invitation.TeamID == team.ID
	invitationStatus := invitation.Status

	cancelPolicy := policy.NewCancelTeamInvitationPolicy()
	pctx, err := cancelPolicy.LoadContext(ctx, policy.CancelTeamInvitationParams{
		HackathonID:  in.HackathonID,
		TeamID:       in.TeamID,
		InvitationID: in.InvitationID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetHackathonStage(stage)
	pctx.SetAllowTeam(allowTeam)
	pctx.SetIsCaptain(isCaptain)
	pctx.SetInvitationBelongs(invitationBelongs)
	pctx.SetInvitationStatus(invitationStatus)

	decision := cancelPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	err = s.invitationRepo.UpdateStatus(ctx, in.InvitationID, "canceled")
	if err != nil {
		s.logger.Error("failed to cancel invitation", "error", err)
		return nil, fmt.Errorf("failed to cancel invitation: %w", err)
	}

	return &CancelTeamInvitationOut{}, nil
}
