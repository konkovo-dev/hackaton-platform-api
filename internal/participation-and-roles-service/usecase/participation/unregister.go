package participation

import (
	"context"
	"fmt"

	participationpolicy "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type UnregisterIn struct {
	HackathonID uuid.UUID
}

type UnregisterOut struct{}

func (s *Service) Unregister(ctx context.Context, in UnregisterIn) (*UnregisterOut, error) {
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

	unregisterPolicy := participationpolicy.NewUnregisterFromHackathonPolicy(adapter)
	pctx, err := unregisterPolicy.LoadContext(ctx, participationpolicy.UnregisterFromHackathonParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	decision := unregisterPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	err = s.participRepo.Delete(ctx, in.HackathonID, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete participation: %w", err)
	}

	return &UnregisterOut{}, nil
}
