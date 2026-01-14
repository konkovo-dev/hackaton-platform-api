package users

import (
	"context"
	"fmt"
	"strings"

	commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/queryutil"
	"github.com/belikoooova/hackaton-platform-api/pkg/queryutil/sqlbuilder"
	"github.com/google/uuid"
)

type ListUsersIn struct {
	Query           *commonv1.Query
	IncludeSkills   bool
	IncludeContacts bool
}

type ListUsersOut struct {
	Users         []*GetUserOut
	NextPageToken string
}

func (s *Service) ListUsers(ctx context.Context, in ListUsersIn) (*ListUsersOut, error) {
	if err := s.validateListUsersIn(in); err != nil {
		return nil, err
	}

	var page *commonv1.PageRequest
	if in.Query != nil {
		page = in.Query.Page
	}
	pagination := queryutil.ParsePagination(page, 0, 0)

	var cursor *ListUsersCursor
	if pagination.PageToken != "" {
		cursor = &ListUsersCursor{}
		if err := queryutil.DecodeCursor(pagination.PageToken, cursor); err != nil {
			return nil, fmt.Errorf("%w: invalid page_token", ErrInvalidInput)
		}
	}

	filters, err := s.parseListUsersFilters(in.Query)
	if err != nil {
		return nil, err
	}

	sort := s.parseListUsersSort(in.Query)

	searchQuery := ""
	if in.Query != nil && in.Query.Q != "" {
		searchQuery = in.Query.Q
	}

	userIDs, _, err := s.userRepo.ListUsers(ctx, ListUsersRepoParams{
		SearchQuery:  searchQuery,
		Filters:      filters,
		Sort:         sort,
		Cursor:       cursor,
		Limit:        queryutil.ShouldFetchMore(pagination.PageSize),
		FieldMapping: GetFieldMapping(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	hasMoreInDB := queryutil.HasMoreResults(len(userIDs), pagination.PageSize)

	filteredUsers := userIDs
	if filters.HasSkillsFilter {
		filtered, err := s.filterUsersBySkills(ctx, userIDs, filters)
		if err != nil {
			return nil, fmt.Errorf("failed to filter by skills: %w", err)
		}
		filteredUsers = filtered
	}

	nextPageToken := ""
	if hasMoreInDB && len(userIDs) > 0 {
		lastLoadedUser := userIDs[len(userIDs)-1]
		nextPageToken, err = queryutil.EncodeCursor(&ListUsersCursor{
			Username: lastLoadedUser.Username,
			UserID:   lastLoadedUser.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to encode cursor: %w", err)
		}
	}

	userIDsToFetch := extractUserIDs(filteredUsers)

	batchResult, err := s.BatchGetUsers(ctx, BatchGetUsersIn{
		UserIDs:         userIDsToFetch,
		IncludeSkills:   in.IncludeSkills,
		IncludeContacts: in.IncludeContacts,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to batch get users: %w", err)
	}

	return &ListUsersOut{
		Users:         batchResult.Users,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Service) filterUsersBySkills(ctx context.Context, users []*UserListResult, filters *ListUsersFilters) ([]*UserListResult, error) {
	if len(users) == 0 {
		return users, nil
	}

	userIDs := extractUserIDs(users)

	catalogSkillsMap, err := s.skillRepo.GetUsersCatalogSkills(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog skills: %w", err)
	}

	customSkillsMap, err := s.skillRepo.GetUsersCustomSkills(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom skills: %w", err)
	}

	result := []*UserListResult{}

	for _, user := range users {
		if s.userMatchesSkillsFilter(user.ID, catalogSkillsMap, customSkillsMap, filters) {
			result = append(result, user)
		}
	}

	return result, nil
}

func (s *Service) userMatchesSkillsFilter(
	userID uuid.UUID,
	catalogSkillsMap map[uuid.UUID][]*entity.CatalogSkill,
	customSkillsMap map[uuid.UUID][]*entity.CustomSkill,
	filters *ListUsersFilters,
) bool {
	for _, group := range filters.FilterGroups {
		groupMatches := true

		for _, filter := range group.Filters {
			if filter.Field != FieldSkills {
				continue
			}

			filterValue, ok := filter.Value.(string)
			if !ok {
				groupMatches = false
				break
			}

			skillMatches := false

			for _, skill := range catalogSkillsMap[userID] {
				if s.matchesSkillFilter(skill.Name, filterValue, filter.Operation) {
					skillMatches = true
					break
				}
			}

			if !skillMatches {
				for _, skill := range customSkillsMap[userID] {
					if s.matchesSkillFilter(skill.Name, filterValue, filter.Operation) {
						skillMatches = true
						break
					}
				}
			}

			if !skillMatches {
				groupMatches = false
				break
			}
		}

		if groupMatches {
			return true
		}
	}

	return false
}

func (s *Service) matchesSkillFilter(skillName, filterValue string, operation commonv1.FilterOperation) bool {
	skillNameLower := strings.ToLower(skillName)
	filterValueLower := strings.ToLower(filterValue)

	switch operation {
	case commonv1.FilterOperation_FILTER_OPERATION_EQUAL:
		return skillNameLower == filterValueLower
	case commonv1.FilterOperation_FILTER_OPERATION_CONTAINS:
		return strings.Contains(skillNameLower, filterValueLower)
	case commonv1.FilterOperation_FILTER_OPERATION_PREFIX:
		return strings.HasPrefix(skillNameLower, filterValueLower)
	default:
		return false
	}
}

func (s *Service) validateListUsersIn(in ListUsersIn) error {
	return nil
}

func (s *Service) parseListUsersFilters(query *commonv1.Query) (*ListUsersFilters, error) {
	result := &ListUsersFilters{
		FilterGroups:    []*sqlbuilder.FilterGroup{},
		HasSkillsFilter: false,
	}

	if query == nil || len(query.FilterGroups) == 0 {
		return result, nil
	}

	allowedFields := map[string]bool{
		FieldUsername:  true,
		FieldUserID:    true,
		FieldFirstName: true,
		FieldLastName:  true,
		FieldSkills:    true,
	}

	for _, group := range query.FilterGroups {
		fg := &sqlbuilder.FilterGroup{Filters: []*sqlbuilder.Filter{}}

		for _, f := range group.Filters {
			if !allowedFields[f.Field] {
				return nil, fmt.Errorf("%w: unsupported filter field: %s", ErrInvalidInput, f.Field)
			}

			switch f.Field {
			case FieldUsername:
				if !queryutil.IsValidStringOperation(f.Operation) {
					return nil, fmt.Errorf("%w: invalid operation for username", ErrInvalidInput)
				}
				stringVal := queryutil.GetStringValue(f)
				if stringVal == "" {
					return nil, fmt.Errorf("%w: username filter requires string value", ErrInvalidInput)
				}
				fg.Filters = append(fg.Filters, &sqlbuilder.Filter{
					Field:     FieldUsername,
					Operation: f.Operation,
					Value:     stringVal,
				})

			case FieldUserID:
				if !queryutil.IsValidInOperation(f.Operation) {
					return nil, fmt.Errorf("%w: user_id only supports IN operation", ErrInvalidInput)
				}
				userIDs := queryutil.GetStringListValue(f)
				if len(userIDs) == 0 {
					return nil, fmt.Errorf("%w: user_id IN requires non-empty list", ErrInvalidInput)
				}
				parsedIDs := make([]interface{}, 0, len(userIDs))
				for _, idStr := range userIDs {
					id, err := uuid.Parse(idStr)
					if err != nil {
						return nil, fmt.Errorf("%w: invalid user_id: %s", ErrInvalidInput, idStr)
					}
					parsedIDs = append(parsedIDs, id)
				}
				fg.Filters = append(fg.Filters, &sqlbuilder.Filter{
					Field:     FieldUserID,
					Operation: f.Operation,
					Value:     parsedIDs,
				})

			case FieldFirstName, FieldLastName:
				if !queryutil.IsValidStringOperation(f.Operation) {
					return nil, fmt.Errorf("%w: invalid operation for %s", ErrInvalidInput, f.Field)
				}
				stringVal := queryutil.GetStringValue(f)
				if stringVal == "" {
					return nil, fmt.Errorf("%w: %s filter requires string value", ErrInvalidInput, f.Field)
				}
				fg.Filters = append(fg.Filters, &sqlbuilder.Filter{
					Field:     f.Field,
					Operation: f.Operation,
					Value:     stringVal,
				})

			case FieldSkills:
				if !queryutil.IsValidStringOperation(f.Operation) {
					return nil, fmt.Errorf("%w: invalid operation for skill", ErrInvalidInput)
				}
				stringVal := queryutil.GetStringValue(f)
				if stringVal == "" {
					return nil, fmt.Errorf("%w: skill filter requires string value", ErrInvalidInput)
				}
				fg.Filters = append(fg.Filters, &sqlbuilder.Filter{
					Field:     FieldSkills,
					Operation: f.Operation,
					Value:     stringVal,
				})
				result.HasSkillsFilter = true
			}
		}

		if len(fg.Filters) > 0 {
			result.FilterGroups = append(result.FilterGroups, fg)
		}
	}

	return result, nil
}

func (s *Service) parseListUsersSort(query *commonv1.Query) []sqlbuilder.SortField {
	allowedSortFields := map[string]bool{
		FieldUsername:  true,
		FieldFirstName: true,
		FieldLastName:  true,
	}

	if query == nil || len(query.Sort) == 0 {
		return []sqlbuilder.SortField{
			{Field: FieldUsername, Direction: "ASC"},
			{Field: FieldUserID, Direction: "ASC"},
		}
	}

	result := []sqlbuilder.SortField{}
	addedFields := make(map[string]bool)

	for _, sort := range query.Sort {
		if !allowedSortFields[sort.Field] {
			continue
		}

		if addedFields[sort.Field] {
			continue
		}

		direction := queryutil.GetSortDirection(sort.Direction)
		result = append(result, sqlbuilder.SortField{Field: sort.Field, Direction: direction})
		addedFields[sort.Field] = true
	}

	if len(result) == 0 {
		result = append(result, sqlbuilder.SortField{Field: FieldUsername, Direction: "ASC"})
	}

	result = append(result, sqlbuilder.SortField{Field: FieldUserID, Direction: "ASC"})

	return result
}

func extractUserIDs(users []*UserListResult) []uuid.UUID {
	ids := make([]uuid.UUID, len(users))
	for i, u := range users {
		ids[i] = u.ID
	}
	return ids
}
