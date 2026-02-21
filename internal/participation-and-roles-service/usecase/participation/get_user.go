package participation

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	participationpolicy "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type GetUserIn struct {
	HackathonID  uuid.UUID
	TargetUserID uuid.UUID
}

type GetUserOut struct {
	Participation *entity.Participation
}

func (s *Service) GetUser(ctx context.Context, in GetUserIn) (*GetUserOut, error) {
	if !auth.IsServiceCall(ctx) {
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

		getUserPolicy := participationpolicy.NewGetUserParticipationPolicy(adapter)
		pctx, err := getUserPolicy.LoadContext(ctx, participationpolicy.GetUserParticipationParams{
			HackathonID:  in.HackathonID,
			TargetUserID: in.TargetUserID,
		})
		if err != nil {
			return nil, err
		}

		decision := getUserPolicy.Check(ctx, pctx)
		if !decision.Allowed {
			return nil, mapPolicyError(decision)
		}
	}

	participation, err := s.participRepo.Get(ctx, in.HackathonID, in.TargetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participation: %w", err)
	}

	if participation == nil {
		return nil, ErrNotFound
	}

	wishedRoles, err := s.teamRoleRepo.GetByParticipation(ctx, in.HackathonID, in.TargetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wished roles: %w", err)
	}

	participation.WishedRoles = wishedRoles

	return &GetUserOut{
		Participation: participation,
	}, nil
}
