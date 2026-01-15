package skills

import (
	"context"
	"fmt"

	commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/repository/postgres"
	"github.com/belikoooova/hackaton-platform-api/pkg/queryutil"
	"github.com/google/uuid"
)

const (
	FieldName    = "name"
	DefaultLimit = 50
	MaxLimit     = 100
)

type ListSkillCatalogIn struct {
	Query *commonv1.Query
}

type ListSkillCatalogOut struct {
	Skills        []*SkillCatalogResult
	NextPageToken string
}

type ListSkillCatalogCursor struct {
	Name string    `json:"name"`
	ID   uuid.UUID `json:"id"`
}

func (s *Service) ListSkillCatalog(ctx context.Context, in ListSkillCatalogIn) (*ListSkillCatalogOut, error) {
	if err := s.validateListSkillCatalogIn(in); err != nil {
		return nil, err
	}

	var page *commonv1.PageRequest
	if in.Query != nil {
		page = in.Query.Page
	}
	pagination := queryutil.ParsePagination(page, DefaultLimit, MaxLimit)

	var cursor *ListSkillCatalogCursor
	if pagination.PageToken != "" {
		cursor = &ListSkillCatalogCursor{}
		if err := queryutil.DecodeCursor(pagination.PageToken, cursor); err != nil {
			return nil, fmt.Errorf("%w: invalid page_token", ErrInvalidInput)
		}
	}

	filters, err := s.parseListSkillCatalogFilters(in.Query)
	if err != nil {
		return nil, err
	}

	sort := s.parseListSkillCatalogSort(in.Query)

	searchQuery := ""
	if in.Query != nil && in.Query.Q != "" {
		searchQuery = in.Query.Q
	}

	var repoCursor *postgres.ListSkillCatalogCursor
	if cursor != nil {
		repoCursor = &postgres.ListSkillCatalogCursor{
			Name: cursor.Name,
			ID:   cursor.ID,
		}
	}

	results, hasMore, err := s.skillRepo.ListSkillCatalog(ctx, postgres.ListSkillCatalogParams{
		SearchQuery: searchQuery,
		Filters:     filters,
		SortField:   sort.Field,
		SortDesc:    sort.Desc,
		Cursor:      repoCursor,
		Limit:       queryutil.ShouldFetchMore(pagination.PageSize),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list skill catalog: %w", err)
	}

	results = queryutil.TrimResults(results, pagination.PageSize)

	skills := make([]*SkillCatalogResult, len(results))
	for i, result := range results {
		skills[i] = &SkillCatalogResult{
			ID:   result.ID,
			Name: result.Name,
		}
	}

	nextPageToken := ""
	if hasMore && len(results) > 0 {
		last := results[len(results)-1]
		nextCursor := &ListSkillCatalogCursor{
			Name: last.Name,
			ID:   last.ID,
		}
		nextPageToken, err = queryutil.EncodeCursor(nextCursor)
		if err != nil {
			return nil, fmt.Errorf("failed to encode cursor: %w", err)
		}
	}

	return &ListSkillCatalogOut{
		Skills:        skills,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Service) validateListSkillCatalogIn(in ListSkillCatalogIn) error {
	return nil
}

type ListSkillCatalogSort struct {
	Field string
	Desc  bool
}

func (s *Service) parseListSkillCatalogSort(query *commonv1.Query) ListSkillCatalogSort {
	result := ListSkillCatalogSort{
		Field: FieldName,
		Desc:  false,
	}

	if query == nil || len(query.Sort) == 0 {
		return result
	}

	allowedFields := map[string]bool{
		FieldName: true,
	}

	for _, sortItem := range query.Sort {
		if allowedFields[sortItem.Field] {
			result.Field = sortItem.Field
			result.Desc = sortItem.Direction == commonv1.SortDirection_SORT_DIRECTION_DESC
			break
		}
	}

	return result
}

func (s *Service) parseListSkillCatalogFilters(query *commonv1.Query) ([]postgres.ListSkillCatalogFilter, error) {
	if query == nil || len(query.FilterGroups) == 0 {
		return []postgres.ListSkillCatalogFilter{}, nil
	}

	allowedFields := map[string]bool{
		FieldName: true,
	}

	var filters []postgres.ListSkillCatalogFilter

	for _, group := range query.FilterGroups {
		for _, filter := range group.Filters {
			if !allowedFields[filter.Field] {
				continue
			}

			switch filter.Field {
			case FieldName:
				if !queryutil.IsValidStringOperation(filter.Operation) {
					return nil, fmt.Errorf("%w: invalid operation %s for field %s", ErrInvalidInput, filter.Operation.String(), filter.Field)
				}

				value := queryutil.GetStringValue(filter)
				if value == "" {
					return nil, fmt.Errorf("%w: empty value for field %s", ErrInvalidInput, filter.Field)
				}

				var operation string
				switch filter.Operation {
				case commonv1.FilterOperation_FILTER_OPERATION_EQUAL,
					commonv1.FilterOperation_FILTER_OPERATION_CONTAINS:
					operation = "CONTAINS"
				case commonv1.FilterOperation_FILTER_OPERATION_PREFIX:
					operation = "PREFIX"
				default:
					continue
				}

				filters = append(filters, postgres.ListSkillCatalogFilter{
					Field:     FieldName,
					Operation: operation,
					Value:     value,
				})
			}
		}
	}

	return filters, nil
}
