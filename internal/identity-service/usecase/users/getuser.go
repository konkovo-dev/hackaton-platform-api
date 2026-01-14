package users

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
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
	if err := s.validateGetUserIn(in); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, in.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUserNotFound, err)
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
