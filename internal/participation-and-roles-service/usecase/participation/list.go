package participation

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	participationpolicy "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type ListIn struct {
	HackathonID   uuid.UUID
	Statuses      []string
	WishedRoleIDs []uuid.UUID
	Limit         int32
	Offset        int32
}

type ListOut struct {
	Participants []*entity.Participation
	Total        int64
}

func (s *Service) List(ctx context.Context, in ListIn) (*ListOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	_, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	adapter := &policyRepositoryAdapter{
		roleRepo:     s.roleRepo,
		participRepo: s.participRepo,
	}

	listPolicy := participationpolicy.NewListParticipantsPolicy(adapter)
	pctx, err := listPolicy.LoadContext(ctx, participationpolicy.ListParticipantsParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	decision := listPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	if in.Limit <= 0 {
		in.Limit = 20
	}
	if in.Limit > 100 {
		in.Limit = 100
	}

	participations, total, err := s.participRepo.List(ctx, in.HackathonID, ListParticipationsFilter{
		Statuses:      in.Statuses,
		WishedRoleIDs: in.WishedRoleIDs,
		Limit:         in.Limit,
		Offset:        in.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list participations: %w", err)
	}

	for _, p := range participations {
		wishedRoles, err := s.teamRoleRepo.GetByParticipation(ctx, p.HackathonID, p.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get wished roles: %w", err)
		}
		p.WishedRoles = wishedRoles
	}

	return &ListOut{
		Participants: participations,
		Total:        total,
	}, nil
}
