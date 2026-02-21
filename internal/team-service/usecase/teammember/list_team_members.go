package teammember

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	pkgpolicy "github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type ListTeamMembersIn struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
}

type ListTeamMembersOut struct {
	Members []*entity.Membership
}

func (s *Service) ListTeamMembers(ctx context.Context, in ListTeamMembersIn) (*ListTeamMembersOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	stage, _, _, err := s.hackathonClient.GetHackathon(ctx, in.HackathonID.String())
	if err != nil {
		s.logger.Error("failed to get hackathon", "error", err)
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	_, participationStatus, roles, err := s.parClient.GetHackathonContext(ctx, in.HackathonID.String())
	if err != nil {
		s.logger.Error("failed to get hackathon context", "error", err)
		return nil, fmt.Errorf("failed to get hackathon context: %w", err)
	}

	team, err := s.teamRepo.GetByIDAndHackathonID(ctx, in.TeamID, in.HackathonID)
	if err != nil {
		return nil, ErrNotFound
	}

	listPolicy := policy.NewListTeamMembersPolicy()
	pctx, err := listPolicy.LoadContext(ctx, policy.ListTeamMembersParams{
		HackathonID: in.HackathonID,
		TeamID:      in.TeamID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetHackathonStage(stage)
	pctx.SetParticipationStatus(participationStatus)
	pctx.SetRoles(roles)

	decision := listPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	members, err := s.membershipRepo.ListByTeamID(ctx, team.ID)
	if err != nil {
		s.logger.Error("failed to list team members", "error", err)
		return nil, fmt.Errorf("failed to list team members: %w", err)
	}

	return &ListTeamMembersOut{Members: members}, nil
}

func mapPolicyError(decision *pkgpolicy.Decision) error {
	if len(decision.Violations) == 0 {
		return ErrUnauthorized
	}

	v := decision.Violations[0]
	switch v.Code {
	case pkgpolicy.ViolationCodeForbidden:
		return fmt.Errorf("%w: %s", ErrForbidden, v.Message)
	case pkgpolicy.ViolationCodeNotFound:
		return fmt.Errorf("%w: %s", ErrNotFound, v.Message)
	case pkgpolicy.ViolationCodeStageRule, pkgpolicy.ViolationCodeTimeRule, pkgpolicy.ViolationCodePolicyRule:
		return fmt.Errorf("%w: %s", ErrForbidden, v.Message)
	default:
		return ErrUnauthorized
	}
}
