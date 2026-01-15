package me

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/google/uuid"
)

type UpdateMyContactsIn struct {
	UserID             uuid.UUID
	Contacts           []*ContactInput
	ContactsVisibility domain.VisibilityLevel
}

type ContactInput struct {
	Type       string
	Value      string
	Visibility domain.VisibilityLevel
}

type UpdateMyContactsOut struct {
	Contacts   []*entity.Contact
	Visibility *entity.Visibility
}

func (s *Service) UpdateMyContacts(ctx context.Context, in UpdateMyContactsIn) (*UpdateMyContactsOut, error) {
	if err := s.validateUpdateMyContactsIn(in); err != nil {
		return nil, err
	}

	_, err := s.userRepo.GetByID(ctx, in.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	contacts := make([]*entity.Contact, 0, len(in.Contacts))
	for _, c := range in.Contacts {
		contacts = append(contacts, &entity.Contact{
			ID:         uuid.New(),
			UserID:     in.UserID,
			Type:       c.Type,
			Value:      c.Value,
			Visibility: c.Visibility,
		})
	}

	var visibility *entity.Visibility

	err = s.uow.Do(ctx, func(ctx context.Context, txRepos *TxRepositories) error {
		if err := txRepos.Contacts.Update(ctx, in.UserID, contacts); err != nil {
			return fmt.Errorf("failed to update user contacts: %w", err)
		}

		visibility, err = txRepos.Visibility.GetByUserID(ctx, in.UserID)
		if err != nil {
			return fmt.Errorf("failed to get visibility: %w", err)
		}

		visibility.ContactsVisibility = in.ContactsVisibility

		if err := txRepos.Visibility.Upsert(ctx, visibility); err != nil {
			return fmt.Errorf("failed to update visibility: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &UpdateMyContactsOut{
		Contacts:   contacts,
		Visibility: visibility,
	}, nil
}

func (s *Service) validateUpdateMyContactsIn(in UpdateMyContactsIn) error {
	if in.UserID == uuid.Nil {
		return fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}

	if in.ContactsVisibility == "" {
		return fmt.Errorf("%w: contacts_visibility is required", ErrInvalidInput)
	}

	if in.ContactsVisibility != domain.VisibilityLevelPublic && in.ContactsVisibility != domain.VisibilityLevelPrivate {
		return fmt.Errorf("%w: contacts_visibility must be 'public' or 'private'", ErrInvalidInput)
	}

	for i, contact := range in.Contacts {
		if contact.Type == "" {
			return fmt.Errorf("%w: contact[%d].type is required", ErrInvalidInput, i)
		}

		if contact.Value == "" {
			return fmt.Errorf("%w: contact[%d].value is required", ErrInvalidInput, i)
		}

		if contact.Visibility == "" {
			return fmt.Errorf("%w: contact[%d].visibility is required", ErrInvalidInput, i)
		}

		if contact.Visibility != domain.VisibilityLevelPublic && contact.Visibility != domain.VisibilityLevelPrivate {
			return fmt.Errorf("%w: contact[%d].visibility must be 'public' or 'private'", ErrInvalidInput, i)
		}
	}

	return nil
}
