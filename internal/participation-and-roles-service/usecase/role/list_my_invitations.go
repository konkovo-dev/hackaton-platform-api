package role

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	rolepolicy "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/policy"
)

type ListMyInvitationsIn struct{}

type ListMyInvitationsOut struct {
	Invitations []*entity.StaffInvitation
}

func (s *Service) ListMyInvitations(ctx context.Context, in ListMyInvitationsIn) (*ListMyInvitationsOut, error) {
	listPolicy := rolepolicy.NewListMyInvitationsPolicy()
	pctx, err := listPolicy.LoadContext(ctx, rolepolicy.ListMyInvitationsParams{})
	if err != nil {
		return nil, err
	}

	decision := listPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, s.mapPolicyError(decision)
	}

	listCtx := pctx.(*rolepolicy.ListMyInvitationsContext)
	userID := listCtx.ActorUserID()

	invitations, err := s.invitationRepo.GetByTargetUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &ListMyInvitationsOut{
		Invitations: invitations,
	}, nil
}
