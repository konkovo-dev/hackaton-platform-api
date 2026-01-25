package users

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	identitypolicy "github.com/belikoooova/hackaton-platform-api/internal/identity-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type GetUserIn struct {
	UserID          uuid.UUID
	IncludeSkills   bool
	IncludeContacts bool
}

type GetUserOut struct {
	User     *entity.User
	Skills   GetUserSkills
	Contacts []*entity.Contact
}

type GetUserSkills struct {
	Catalog []*entity.CatalogSkill
	Custom  []*entity.CustomSkill
}

func (s *Service) GetUser(ctx context.Context, in GetUserIn) (*GetUserOut, error) {
	policy := identitypolicy.NewRequireAuthPolicy(identitypolicy.ActionReadUser)
	pctx, err := policy.LoadContext(ctx, identitypolicy.NoParams{})
	if err != nil {
		return nil, err
	}

	decision := policy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, s.mapPolicyError(decision)
	}

	if err := s.validateGetUserIn(in); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, in.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	enrichedMap, err := s.enrichUsersWithData(ctx, []*entity.User{user}, enrichmentOptions{
		IncludeSkills:   in.IncludeSkills,
		IncludeContacts: in.IncludeContacts,
	})
	if err != nil {
		return nil, err
	}

	result, ok := enrichedMap[user.ID]
	if !ok {
		return nil, fmt.Errorf("enrichment failed: user not found in result")
	}

	return result, nil
}

func (s *Service) validateGetUserIn(in GetUserIn) error {
	if in.UserID == uuid.Nil {
		return fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}

	return nil
}

func (s *Service) mapPolicyError(decision *policy.Decision) error {
	if len(decision.Violations) == 0 {
		return fmt.Errorf("access denied")
	}

	v := decision.Violations[0]
	if v.Code == policy.ViolationCodeForbidden {
		return fmt.Errorf("access denied: %s", v.Message)
	}

	return fmt.Errorf("%s", v.Message)
}
