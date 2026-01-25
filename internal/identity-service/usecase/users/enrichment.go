package users

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
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

	var actorUserID uuid.UUID
	if actorID, ok := auth.GetUserID(ctx); ok {
		if parsed, err := uuid.Parse(actorID); err == nil {
			actorUserID = parsed
		}
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
		if err := s.enrichSkills(ctx, actorUserID, userIDs, visibilityMap, userMap); err != nil {
			return nil, err
		}
	}

	if opts.IncludeContacts {
		if err := s.enrichContacts(ctx, actorUserID, userIDs, visibilityMap, userMap); err != nil {
			return nil, err
		}
	}

	return userMap, nil
}

func (s *Service) enrichSkills(
	ctx context.Context,
	actorUserID uuid.UUID,
	userIDs []uuid.UUID,
	visibilityMap map[uuid.UUID]*entity.Visibility,
	userMap map[uuid.UUID]*GetUserOut,
) error {
	userIDsToLoad := make([]uuid.UUID, 0)
	for _, userID := range userIDs {
		isMe := actorUserID != uuid.Nil && actorUserID == userID
		if isMe {
			userIDsToLoad = append(userIDsToLoad, userID)
			continue
		}

		if vis, ok := visibilityMap[userID]; ok && vis.SkillsVisibility == domain.VisibilityLevelPublic {
			userIDsToLoad = append(userIDsToLoad, userID)
		}
	}

	if len(userIDsToLoad) == 0 {
		return nil
	}

	catalogSkillsMap, err := s.skillRepo.GetUsersCatalogSkills(ctx, userIDsToLoad)
	if err != nil {
		return fmt.Errorf("failed to get catalog skills: %w", err)
	}

	customSkillsMap, err := s.skillRepo.GetUsersCustomSkills(ctx, userIDsToLoad)
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
	actorUserID uuid.UUID,
	userIDs []uuid.UUID,
	visibilityMap map[uuid.UUID]*entity.Visibility,
	userMap map[uuid.UUID]*GetUserOut,
) error {
	userIDsToLoad := make([]uuid.UUID, 0)
	for _, userID := range userIDs {
		isMe := actorUserID != uuid.Nil && actorUserID == userID
		if isMe {
			userIDsToLoad = append(userIDsToLoad, userID)
			continue
		}

		if vis, ok := visibilityMap[userID]; ok && vis.ContactsVisibility == domain.VisibilityLevelPublic {
			userIDsToLoad = append(userIDsToLoad, userID)
		}
	}

	if len(userIDsToLoad) == 0 {
		return nil
	}

	contactsMap, err := s.contactRepo.GetByUserIDs(ctx, userIDsToLoad)
	if err != nil {
		return fmt.Errorf("failed to get contacts: %w", err)
	}

	for userID, contacts := range contactsMap {
		isMe := actorUserID != uuid.Nil && actorUserID == userID

		filteredContacts := make([]*entity.Contact, 0)
		for _, contact := range contacts {
			if isMe || contact.Visibility == domain.VisibilityLevelPublic {
				filteredContacts = append(filteredContacts, contact)
			}
		}

		if userOut, ok := userMap[userID]; ok {
			userOut.Contacts = filteredContacts
		}
	}

	return nil
}
