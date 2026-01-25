package hackathonservice

import (
	"context"

	commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"
	hackathonv1 "github.com/belikoooova/hackaton-platform-api/api/hackathon/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/transport/grpc/mappers"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/usecase/hackathon"
)

func (s *HackathonService) ListHackathons(ctx context.Context, req *hackathonv1.ListHackathonsRequest) (*hackathonv1.ListHackathonsResponse, error) {
	var pageSize uint32
	var pageToken string

	if req.Query != nil && req.Query.Page != nil {
		pageSize = req.Query.Page.PageSize
		pageToken = req.Query.Page.PageToken
	}

	result, err := s.hackathonService.ListHackathons(ctx, hackathon.ListHackathonsIn{
		PageSize:           pageSize,
		PageToken:          pageToken,
		IncludeDescription: req.IncludeDescription,
		IncludeLinks:       req.IncludeLinks,
		IncludeLimits:      req.IncludeLimits,
	})
	if err != nil {
		return nil, s.handleError(ctx, err, "list_hackathons")
	}

	hackathons := make([]*hackathonv1.Hackathon, 0, len(result.Hackathons))
	for _, h := range result.Hackathons {
		protoHackathon := mappers.HackathonToProto(h, mappers.HackathonConversionOptions{
			IncludeDescription: req.IncludeDescription,
			IncludeLimits:      req.IncludeLimits,
		})

		if req.IncludeLinks {
			if links, ok := result.Links[h.ID.String()]; ok {
				for _, link := range links {
					protoHackathon.Links = append(protoHackathon.Links, &hackathonv1.HackathonLink{
						Title: link.Title,
						Url:   link.URL,
					})
				}
			}
		}

		hackathons = append(hackathons, protoHackathon)
	}

	return &hackathonv1.ListHackathonsResponse{
		Hackathons: hackathons,
		Page: &commonv1.PageResponse{
			NextPageToken: result.NextPageToken,
		},
	}, nil
}
