package users

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/google/uuid"
)

type enrichmentOptions struct {
	IncludeSkills   bool
	IncludeContacts bool
}

func (s *Service) enrichUsersWithData(
	ctx context.Context,
	users []*entity.User,
	opts enrichmentOptions,
) (map[uuid.UUID]*GetUserOut, error) {
	if len(users) == 0 {
		return map[uuid.UUID]*GetUserOut{}, nil
	}

	userMap := make(map[uuid.UUID]*GetUserOut)
	userIDs := make([]uuid.UUID, 0, len(users))

	for _, user := range users {
		userMap[user.ID] = &GetUserOut{
			User: user,
			Skills: GetUserSkills{
				Catalog: []*entity.CatalogSkill{},
				Custom:  []*entity.CustomSkill{},
			},
			Contacts: []*entity.Contact{},
		}
		userIDs = append(userIDs, user.ID)
	}

	visibilityMap, err := s.visibilityRepo.GetByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get visibility: %w", err)
	}

	if opts.IncludeSkills {
		if err := s.enrichSkills(ctx, userIDs, visibilityMap, userMap); err != nil {
			return nil, err
		}
	}

	if opts.IncludeContacts {
		if err := s.enrichContacts(ctx, userIDs, visibilityMap, userMap); err != nil {
			return nil, err
		}
	}

	return userMap, nil
}

func (s *Service) enrichSkills(
	ctx context.Context,
	userIDs []uuid.UUID,
	visibilityMap map[uuid.UUID]*entity.Visibility,
	userMap map[uuid.UUID]*GetUserOut,
) error {
	userIDsWithPublicSkills := make([]uuid.UUID, 0)
	for _, userID := range userIDs {
		if vis, ok := visibilityMap[userID]; ok && vis.SkillsVisibility == domain.VisibilityLevelPublic {
			userIDsWithPublicSkills = append(userIDsWithPublicSkills, userID)
		}
	}

	if len(userIDsWithPublicSkills) == 0 {
		return nil
	}

	catalogSkillsMap, err := s.skillRepo.GetUsersCatalogSkills(ctx, userIDsWithPublicSkills)
	if err != nil {
		return fmt.Errorf("failed to get catalog skills: %w", err)
	}

	customSkillsMap, err := s.skillRepo.GetUsersCustomSkills(ctx, userIDsWithPublicSkills)
	if err != nil {
		return fmt.Errorf("failed to get custom skills: %w", err)
	}

	for userID, catalogSkills := range catalogSkillsMap {
		if userOut, ok := userMap[userID]; ok {
			userOut.Skills.Catalog = catalogSkills
		}
	}

	for userID, customSkills := range customSkillsMap {
		if userOut, ok := userMap[userID]; ok {
			userOut.Skills.Custom = customSkills
		}
	}

	return nil
}

func (s *Service) enrichContacts(
	ctx context.Context,
	userIDs []uuid.UUID,
	visibilityMap map[uuid.UUID]*entity.Visibility,
	userMap map[uuid.UUID]*GetUserOut,
) error {
	userIDsWithPublicContacts := make([]uuid.UUID, 0)
	for _, userID := range userIDs {
		if vis, ok := visibilityMap[userID]; ok && vis.ContactsVisibility == domain.VisibilityLevelPublic {
			userIDsWithPublicContacts = append(userIDsWithPublicContacts, userID)
		}
	}

	if len(userIDsWithPublicContacts) == 0 {
		return nil
	}

	contactsMap, err := s.contactRepo.GetByUserIDs(ctx, userIDsWithPublicContacts)
	if err != nil {
		return fmt.Errorf("failed to get contacts: %w", err)
	}

	for userID, contacts := range contactsMap {
		publicContacts := make([]*entity.Contact, 0)
		for _, contact := range contacts {
			if contact.Visibility == domain.VisibilityLevelPublic {
				publicContacts = append(publicContacts, contact)
			}
		}

		if userOut, ok := userMap[userID]; ok {
			userOut.Contacts = publicContacts
		}
	}

	return nil
}
