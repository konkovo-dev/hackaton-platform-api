package me

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/google/uuid"
)

type GetMeIn struct {
	UserID uuid.UUID
}

type GetMeOut struct {
	User       *entity.User
	Skills     GetMeSkills
	Contacts   []*GetMeContact
	Visibility *entity.Visibility
}

type GetMeSkills struct {
	Catalog []*entity.CatalogSkill
	Custom  []*entity.CustomSkill
}

type GetMeContact struct {
	Contact    *entity.Contact
	Visibility string
}

func (s *Service) GetMe(ctx context.Context, in GetMeIn) (*GetMeOut, error) {
	if err := s.validateGetMeIn(in); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, in.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	catalogSkills, err := s.skillRepo.GetUserCatalogSkills(ctx, in.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog skills: %w", err)
	}

	customSkills, err := s.skillRepo.GetUserCustomSkills(ctx, in.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom skills: %w", err)
	}

	contacts, err := s.contactRepo.GetByUserID(ctx, in.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}

	visibility, err := s.visibilityRepo.GetByUserID(ctx, in.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get visibility: %w", err)
	}

	contactsOutput := make([]*GetMeContact, 0, len(contacts))
	for _, contact := range contacts {
		contactsOutput = append(contactsOutput, &GetMeContact{
			Contact:    contact,
			Visibility: string(contact.Visibility),
		})
	}

	return &GetMeOut{
		User: user,
		Skills: GetMeSkills{
			Catalog: catalogSkills,
			Custom:  customSkills,
		},
		Contacts:   contactsOutput,
		Visibility: visibility,
	}, nil
}

func (s *Service) validateGetMeIn(in GetMeIn) error {
	if in.UserID == uuid.Nil {
		return fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}

	return nil
}
