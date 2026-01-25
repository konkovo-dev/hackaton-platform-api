package hackathon

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain/entity"
	hackathonpolicy "github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/policy"
	"github.com/google/uuid"
)

type GetHackathonIn struct {
	HackathonID uuid.UUID

	IncludeDescription bool
	IncludeLinks       bool
	IncludeLimits      bool
}

type GetHackathonOut struct {
	Hackathon *entity.Hackathon
	Links     []*entity.HackathonLink
}

func (s *Service) GetHackathon(ctx context.Context, in GetHackathonIn) (*GetHackathonOut, error) {
	getPolicy := hackathonpolicy.NewGetHackathonPolicy(s.hackathonRepo, s.parClient)
	pctx, err := getPolicy.LoadContext(ctx, hackathonpolicy.GetHackathonParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	decision := getPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, s.mapPolicyError(decision)
	}

	hackathon, err := s.hackathonRepo.GetByID(ctx, in.HackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	if hackathon == nil {
		return nil, ErrHackathonNotFound
	}

	out := &GetHackathonOut{
		Hackathon: hackathon,
	}

	if !in.IncludeDescription {
		hackathon.Description = ""
	}

	if !in.IncludeLimits {
		hackathon.TeamSizeMax = 0
	}

	if in.IncludeLinks {
		links, err := s.linkRepo.GetByHackathonID(ctx, in.HackathonID)
		if err != nil {
			return nil, fmt.Errorf("failed to get hackathon links: %w", err)
		}
		out.Links = links
	}

	return out, nil
}
