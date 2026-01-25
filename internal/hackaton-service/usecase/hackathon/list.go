package hackathon

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain/entity"
	"github.com/google/uuid"
)

const (
	DefaultPageSize = 50
	MaxPageSize     = 100
)

type ListHackathonsIn struct {
	PageSize  uint32
	PageToken string

	IncludeDescription bool
	IncludeLinks       bool
	IncludeLimits      bool
}

type ListHackathonsOut struct {
	Hackathons    []*entity.Hackathon
	Links         map[string][]*entity.HackathonLink
	NextPageToken string
}

func (s *Service) ListHackathons(ctx context.Context, in ListHackathonsIn) (*ListHackathonsOut, error) {
	pageSize := in.PageSize
	if pageSize == 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	offset := uint32(0)
	if in.PageToken != "" {
		parsedOffset, err := parsePageToken(in.PageToken)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid page token", ErrInvalidInput)
		}
		offset = parsedOffset
	}

	hackathons, err := s.hackathonRepo.List(ctx, int32(pageSize), int32(offset))
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
