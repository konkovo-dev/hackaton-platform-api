package users

import (
	"context"
	"fmt"

	identitypolicy "github.com/belikoooova/hackaton-platform-api/internal/identity-service/policy"
	"github.com/google/uuid"
)

const MaxBatchSize = 100

type BatchGetUsersIn struct {
	UserIDs         []uuid.UUID
	IncludeSkills   bool
	IncludeContacts bool
}

type BatchGetUsersOut struct {
	Users []*GetUserOut
}

func (s *Service) BatchGetUsers(ctx context.Context, in BatchGetUsersIn) (*BatchGetUsersOut, error) {
	policy := identitypolicy.NewRequireAuthPolicy(identitypolicy.ActionBatchGetUsers)
	pctx, err := policy.LoadContext(ctx, identitypolicy.NoParams{})
	if err != nil {
		return nil, err
	}

	decision := policy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, s.mapPolicyError(decision)
	}

	if err := s.validateBatchGetUsersIn(in); err != nil {
		return nil, err
	}

	if len(in.UserIDs) == 0 {
		return &BatchGetUsersOut{Users: []*GetUserOut{}}, nil
	}

	users, err := s.userRepo.GetByIDs(ctx, in.UserIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	enrichedMap, err := s.enrichUsersWithData(ctx, users, enrichmentOptions{
		IncludeSkills:   in.IncludeSkills,
		IncludeContacts: in.IncludeContacts,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*GetUserOut, 0, len(in.UserIDs))
	for _, userID := range in.UserIDs {
		if userOut, ok := enrichedMap[userID]; ok {
			result = append(result, userOut)
		}
	}

	return &BatchGetUsersOut{Users: result}, nil
}

func (s *Service) validateBatchGetUsersIn(in BatchGetUsersIn) error {
	if len(in.UserIDs) > MaxBatchSize {
		return fmt.Errorf("%w: maximum %d users allowed, got %d", ErrTooManyUsers, MaxBatchSize, len(in.UserIDs))
	}

	for _, userID := range in.UserIDs {
		if userID == uuid.Nil {
			return fmt.Errorf("%w: invalid user_id in list", ErrInvalidInput)
		}
	}

	return nil
}
