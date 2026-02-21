package teaminbox

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type ListMyTeamInvitationsIn struct {
	PageSize  uint32
	PageToken string
}

type ListMyTeamInvitationsOut struct {
	Invitations   []*entity.TeamInvitation
	NextPageToken string
}

func (s *Service) ListMyTeamInvitations(ctx context.Context, in ListMyTeamInvitationsIn) (*ListMyTeamInvitationsOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	listPolicy := policy.NewListMyTeamInvitationsPolicy()
	pctx, err := listPolicy.LoadContext(ctx, policy.ListMyTeamInvitationsParams{})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)

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

	invitations, err := s.invitationRepo.ListByTargetUserID(ctx, userUUID, int32(pageSize+1), int32(offset))
	if err != nil {
		s.logger.Error("failed to list my team invitations", "error", err)
		return nil, fmt.Errorf("failed to list my team invitations: %w", err)
	}

	var nextPageToken string
	if len(invitations) > int(pageSize) {
		invitations = invitations[:pageSize]
		nextPageToken = encodePageToken(offset + int(pageSize))
	}

	return &ListMyTeamInvitationsOut{
		Invitations:   invitations,
		NextPageToken: nextPageToken,
	}, nil
}
