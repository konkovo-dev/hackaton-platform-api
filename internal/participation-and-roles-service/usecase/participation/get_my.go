package participation

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	participationpolicy "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type GetMyIn struct {
	HackathonID uuid.UUID
}

type GetMyOut struct {
	Participation *entity.Participation
}

func (s *Service) GetMy(ctx context.Context, in GetMyIn) (*GetMyOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	adapter := &policyRepositoryAdapter{
		roleRepo:     s.roleRepo,
		participRepo: s.participRepo,
	}

	getMyPolicy := participationpolicy.NewGetMyParticipationPolicy(adapter)
	pctx, err := getMyPolicy.LoadContext(ctx, participationpolicy.GetMyParticipationParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	decision := getMyPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	participation, err := s.participRepo.Get(ctx, in.HackathonID, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participation: %w", err)
	}

	if participation == nil {
		return nil, ErrNotFound
	}

	wishedRoles, err := s.teamRoleRepo.GetByParticipation(ctx, in.HackathonID, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wished roles: %w", err)
	}

	participation.WishedRoles = wishedRoles

	return &GetMyOut{
		Participation: participation,
	}, nil
}
