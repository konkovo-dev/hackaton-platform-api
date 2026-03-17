package hackathon

import (
	"context"
	"fmt"
	"strings"

	commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/queryutil/sqlbuilder"
	"github.com/google/uuid"
)

const (
	DefaultPageSize = 50
	MaxPageSize     = 100
)

const (
	FieldName                   = "name"
	FieldStage                  = "stage"
	FieldState                  = "state"
	FieldLocationOnline         = "location_online"
	FieldLocationCity           = "location_city"
	FieldLocationCountry        = "location_country"
	FieldHackathonID            = "hackathon_id"
	FieldMyRole                 = "my_role"
	FieldMyParticipation        = "my_participation"
	FieldMyParticipationStatus  = "my_participation_status"
)

type ListHackathonsIn struct {
	Query *commonv1.Query

	IncludeDescription bool
	IncludeLinks       bool
	IncludeLimits      bool
}

type ListHackathonsFilters struct {
	FilterGroups     []*sqlbuilder.FilterGroup
	HackathonIDsOnly []uuid.UUID
}

type RoleFilters struct {
	RoleFilter                *string
	ParticipationFilter       *bool
	ParticipationStatusFilter *string
}

type ListHackathonsRepoParams struct {
	SearchQuery  string
	Filters      *ListHackathonsFilters
	Sort         []sqlbuilder.SortField
	Limit        int32
	Offset       int32
	FieldMapping sqlbuilder.FieldMapping
}

type ListHackathonsOut struct {
	Hackathons    []*entity.Hackathon
	Links         map[string][]*entity.HackathonLink
	NextPageToken string
}

func (s *Service) ListHackathons(ctx context.Context, in ListHackathonsIn) (*ListHackathonsOut, error) {
	var page *commonv1.PageRequest
	if in.Query != nil {
		page = in.Query.Page
	}
	
	pageSize := uint32(DefaultPageSize)
	if page != nil && page.PageSize > 0 {
		pageSize = page.PageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	offset := uint32(0)
	pageToken := ""
	if page != nil {
		pageToken = page.PageToken
	}
	if pageToken != "" {
		parsedOffset, err := parsePageToken(pageToken)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid page token", ErrInvalidInput)
		}
		offset = parsedOffset
	}

	filters, roleFilters, err := s.parseListHackathonsFilters(ctx, in.Query)
	if err != nil {
		return nil, err
	}

	// Apply state filter based on user permissions
	// If no state filter is explicitly provided, only show published hackathons to regular users
	// Owners and organizers can see drafts of their own hackathons via my_role filter
	hasStateFilter := s.hasStateFilter(in.Query)
	hasDraftStateFilter := s.hasDraftStateFilter(in.Query)
	
	if !hasStateFilter && roleFilters == nil {
		// No state filter and no role filter - show only published hackathons
		// Add to existing filter group if exists, otherwise create new one
		if len(filters.FilterGroups) > 0 {
			// Add to the first (and typically only) filter group to ensure AND logic
			filters.FilterGroups[0].Filters = append(filters.FilterGroups[0].Filters, &sqlbuilder.Filter{
				Field:     FieldState,
				Operation: commonv1.FilterOperation_FILTER_OPERATION_EQUAL,
				Value:     "published",
			})
		} else {
			// No existing filters, create new group
			filters.FilterGroups = append(filters.FilterGroups, &sqlbuilder.FilterGroup{
				Filters: []*sqlbuilder.Filter{
					{
						Field:     FieldState,
						Operation: commonv1.FilterOperation_FILTER_OPERATION_EQUAL,
						Value:     "published",
					},
				},
			})
		}
	} else if hasDraftStateFilter && roleFilters == nil {
		// User requested state=draft but has no role filter
		// This means they're trying to see all drafts without specifying which ones
		// Return empty result (drafts are only accessible via my_role filter)
		return &ListHackathonsOut{
			Hackathons:    []*entity.Hackathon{},
			Links:         map[string][]*entity.HackathonLink{},
			NextPageToken: "",
		}, nil
	}
	// If hasStateFilter && roleFilters != nil, the state filter will be applied normally
	// This allows filtering by state=draft when combined with my_role=owner

	if roleFilters != nil {
		hackathonIDs, err := s.parClient.GetMyHackathonIDs(
			ctx,
			roleFilters.RoleFilter,
			roleFilters.ParticipationStatusFilter,
			roleFilters.ParticipationFilter,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get hackathon IDs by role: %w", err)
		}

		if len(hackathonIDs) == 0 {
			return &ListHackathonsOut{
				Hackathons:    []*entity.Hackathon{},
				Links:         map[string][]*entity.HackathonLink{},
				NextPageToken: "",
			}, nil
		}

		filters.HackathonIDsOnly = hackathonIDs
	}

	sort := s.parseListHackathonsSort(in.Query)

	searchQuery := ""
	if in.Query != nil && in.Query.Q != "" {
		searchQuery = in.Query.Q
	}

	hackathons, err := s.hackathonRepo.List(ctx, ListHackathonsRepoParams{
		SearchQuery:  searchQuery,
		Filters:      filters,
		Sort:         sort,
		Limit:        int32(pageSize),
		Offset:       int32(offset),
		FieldMapping: GetFieldMapping(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list hackathons: %w", err)
	}

	nextPageToken := ""
	if len(hackathons) == int(pageSize) {
		nextPageToken = encodePageToken(offset + pageSize)
	}

	if !in.IncludeDescription {
		for _, h := range hackathons {
			h.Description = ""
		}
	}

	if !in.IncludeLimits {
		for _, h := range hackathons {
			h.TeamSizeMax = 0
		}
	}

	linksMap := make(map[string][]*entity.HackathonLink)
	if in.IncludeLinks && len(hackathons) > 0 {
		hackathonIDs := make([]uuid.UUID, len(hackathons))
		for i, h := range hackathons {
			hackathonIDs[i] = h.ID
		}

		allLinks, err := s.linkRepo.GetByHackathonIDs(ctx, hackathonIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get hackathon links: %w", err)
		}

		for _, link := range allLinks {
			linksMap[link.HackathonID.String()] = append(linksMap[link.HackathonID.String()], link)
		}
	}

	return &ListHackathonsOut{
		Hackathons:    hackathons,
		Links:         linksMap,
		NextPageToken: nextPageToken,
	}, nil
}

func GetFieldMapping() sqlbuilder.FieldMapping {
	return sqlbuilder.FieldMapping{
		FieldName:            "h.name",
		FieldStage:           "h.stage",
		FieldState:           "h.state",
		FieldLocationOnline:  "h.location_online",
		FieldLocationCity:    "h.location_city",
		FieldLocationCountry: "h.location_country",
		FieldHackathonID:     "h.id",
		"created_at":         "h.created_at",
		"starts_at":          "h.starts_at",
		"ends_at":            "h.ends_at",
		"published_at":       "h.published_at",
	}
}

func (s *Service) parseListHackathonsFilters(ctx context.Context, query *commonv1.Query) (*ListHackathonsFilters, *RoleFilters, error) {
	result := &ListHackathonsFilters{
		FilterGroups:     []*sqlbuilder.FilterGroup{},
		HackathonIDsOnly: nil,
	}

	var roleFilters *RoleFilters

	if query == nil || len(query.FilterGroups) == 0 {
		return result, nil, nil
	}

	allowedFields := map[string]bool{
		FieldName:                  true,
		FieldStage:                 true,
		FieldState:                 true,
		FieldLocationOnline:        true,
		FieldLocationCity:          true,
		FieldLocationCountry:       true,
		FieldHackathonID:           true,
		FieldMyRole:                true,
		FieldMyParticipation:       true,
		FieldMyParticipationStatus: true,
	}

	for _, group := range query.FilterGroups {
		fg := &sqlbuilder.FilterGroup{Filters: []*sqlbuilder.Filter{}}

		for _, f := range group.Filters {
			// Normalize field name: convert dot notation to underscore
			// e.g., "location.online" -> "location_online"
			normalizedField := strings.ReplaceAll(f.Field, ".", "_")
			f.Field = normalizedField
			if !allowedFields[f.Field] {
				return nil, nil, fmt.Errorf("%w: unsupported filter field: %s", ErrInvalidInput, f.Field)
			}

			switch f.Field {
			case FieldName, FieldLocationCity, FieldLocationCountry:
				stringVal := getStringValue(f)
				if stringVal == "" {
					return nil, nil, fmt.Errorf("%w: %s filter requires string value", ErrInvalidInput, f.Field)
				}
				fg.Filters = append(fg.Filters, &sqlbuilder.Filter{
					Field:     f.Field,
					Operation: f.Operation,
					Value:     stringVal,
				})

			case FieldStage, FieldState:
				stringVal := getStringValue(f)
				if stringVal == "" {
					return nil, nil, fmt.Errorf("%w: %s filter requires string value", ErrInvalidInput, f.Field)
				}
				normalizedVal := normalizeEnumValue(stringVal)
				fg.Filters = append(fg.Filters, &sqlbuilder.Filter{
					Field:     f.Field,
					Operation: f.Operation,
					Value:     normalizedVal,
				})

			case FieldLocationOnline:
				boolVal := getBoolValue(f)
				fg.Filters = append(fg.Filters, &sqlbuilder.Filter{
					Field:     f.Field,
					Operation: f.Operation,
					Value:     boolVal,
				})

			case FieldHackathonID:
				if f.Operation != commonv1.FilterOperation_FILTER_OPERATION_IN {
					return nil, nil, fmt.Errorf("%w: hackathon_id only supports IN operation", ErrInvalidInput)
				}
				hackathonIDs := getStringListValue(f)
				if len(hackathonIDs) == 0 {
					return nil, nil, fmt.Errorf("%w: hackathon_id IN requires non-empty list", ErrInvalidInput)
				}
				parsedIDs := make([]interface{}, 0, len(hackathonIDs))
				for _, idStr := range hackathonIDs {
					id, err := uuid.Parse(idStr)
					if err != nil {
						return nil, nil, fmt.Errorf("%w: invalid hackathon_id: %s", ErrInvalidInput, idStr)
					}
					parsedIDs = append(parsedIDs, id)
				}
				fg.Filters = append(fg.Filters, &sqlbuilder.Filter{
					Field:     f.Field,
					Operation: f.Operation,
					Value:     parsedIDs,
				})

			case FieldMyRole:
				if f.Operation != commonv1.FilterOperation_FILTER_OPERATION_EQUAL {
					return nil, nil, fmt.Errorf("%w: my_role only supports EQUAL operation", ErrInvalidInput)
				}
				roleVal := getStringValue(f)
				if roleVal == "" {
					return nil, nil, fmt.Errorf("%w: my_role requires string value", ErrInvalidInput)
				}
				if roleFilters == nil {
					roleFilters = &RoleFilters{}
				}
				roleFilters.RoleFilter = &roleVal

			case FieldMyParticipation:
				if f.Operation != commonv1.FilterOperation_FILTER_OPERATION_EQUAL {
					return nil, nil, fmt.Errorf("%w: my_participation only supports EQUAL operation", ErrInvalidInput)
				}
				boolVal := getBoolValue(f)
				if roleFilters == nil {
					roleFilters = &RoleFilters{}
				}
				roleFilters.ParticipationFilter = &boolVal

			case FieldMyParticipationStatus:
				if f.Operation != commonv1.FilterOperation_FILTER_OPERATION_EQUAL {
					return nil, nil, fmt.Errorf("%w: my_participation_status only supports EQUAL operation", ErrInvalidInput)
				}
				statusVal := getStringValue(f)
				if statusVal == "" {
					return nil, nil, fmt.Errorf("%w: my_participation_status requires string value", ErrInvalidInput)
				}
				// Don't normalize participation status - it already has underscores in DB
				if roleFilters == nil {
					roleFilters = &RoleFilters{}
				}
				roleFilters.ParticipationStatusFilter = &statusVal
			}
		}

		if len(fg.Filters) > 0 {
			result.FilterGroups = append(result.FilterGroups, fg)
		}
	}

	return result, roleFilters, nil
}

func (s *Service) parseListHackathonsSort(query *commonv1.Query) []sqlbuilder.SortField {
	allowedSortFields := map[string]bool{
		FieldName:     true,
		"created_at":  true,
		"starts_at":   true,
		"ends_at":     true,
		"published_at": true,
	}

	if query == nil || len(query.Sort) == 0 {
		return nil
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

		direction := getSortDirection(sort.Direction)
		result = append(result, sqlbuilder.SortField{Field: sort.Field, Direction: direction})
		addedFields[sort.Field] = true
	}

	return result
}

func getStringValue(f *commonv1.Filter) string {
	return f.GetStringValue()
}

func getBoolValue(f *commonv1.Filter) bool {
	return f.GetBoolValue()
}

func getStringListValue(f *commonv1.Filter) []string {
	if sl := f.GetStringList(); sl != nil {
		return sl.Values
	}
	return nil
}

func getSortDirection(dir commonv1.SortDirection) string {
	if dir == commonv1.SortDirection_SORT_DIRECTION_DESC {
		return "DESC"
	}
	return "ASC"
}

func normalizeEnumValue(val string) string {
	val = strings.ToLower(val)
	val = strings.TrimPrefix(val, "hackathon_stage_")
	val = strings.TrimPrefix(val, "hackathon_state_")
	val = strings.ReplaceAll(val, "_", "")
	return val
}

func (s *Service) hasStateFilter(query *commonv1.Query) bool {
	if query == nil || len(query.FilterGroups) == 0 {
		return false
	}

	for _, group := range query.FilterGroups {
		for _, f := range group.Filters {
			normalizedField := strings.ReplaceAll(f.Field, ".", "_")
			if normalizedField == FieldState {
				return true
			}
		}
	}

	return false
}

func (s *Service) hasDraftStateFilter(query *commonv1.Query) bool {
	if query == nil || len(query.FilterGroups) == 0 {
		return false
	}

	for _, group := range query.FilterGroups {
		for _, f := range group.Filters {
			normalizedField := strings.ReplaceAll(f.Field, ".", "_")
			if normalizedField == FieldState {
				stringVal := getStringValue(f)
				normalizedVal := normalizeEnumValue(stringVal)
				if normalizedVal == "draft" {
					return true
				}
			}
		}
	}

	return false
}
