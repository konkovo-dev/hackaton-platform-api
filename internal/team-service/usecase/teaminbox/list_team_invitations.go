package teaminbox

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	pkgpolicy "github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

const (
	DefaultPageSize = 50
	MaxPageSize     = 100
)

type ListTeamInvitationsIn struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
	PageSize    uint32
	PageToken   string
}

type ListTeamInvitationsOut struct {
	Invitations   []*entity.TeamInvitation
	NextPageToken string
}

func (s *Service) ListTeamInvitations(ctx context.Context, in ListTeamInvitationsIn) (*ListTeamInvitationsOut, error) {
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

	team, err := s.teamRepo.GetByIDAndHackathonID(ctx, in.TeamID, in.HackathonID)
	if err != nil {
		return nil, ErrNotFound
	}

	isCaptain, err := s.membershipRepo.CheckIsCaptain(ctx, in.TeamID, userUUID)
	if err != nil {
		s.logger.Error("failed to check captain status", "error", err)
		return nil, fmt.Errorf("failed to check captain status: %w", err)
	}

	listPolicy := policy.NewListTeamInvitationsPolicy()
	pctx, err := listPolicy.LoadContext(ctx, policy.ListTeamInvitationsParams{
		HackathonID: in.HackathonID,
		TeamID:      in.TeamID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetHackathonStage(stage)
	pctx.SetIsCaptain(isCaptain)

	decision := listPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	pageSize := in.PageSize
	if pageSize == 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	offset, err := parsePageToken(in.PageToken)
	if err != nil {
		return nil, fmt.Errorf("invalid page token: %w", err)
	}

	invitations, err := s.invitationRepo.ListByTeamID(ctx, team.ID, int32(pageSize+1), int32(offset))
	if err != nil {
		s.logger.Error("failed to list team invitations", "error", err)
		return nil, fmt.Errorf("failed to list team invitations: %w", err)
	}

	var nextPageToken string
	if len(invitations) > int(pageSize) {
		invitations = invitations[:pageSize]
		nextPageToken = encodePageToken(offset + int(pageSize))
	}

	return &ListTeamInvitationsOut{
		Invitations:   invitations,
		NextPageToken: nextPageToken,
	}, nil
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
