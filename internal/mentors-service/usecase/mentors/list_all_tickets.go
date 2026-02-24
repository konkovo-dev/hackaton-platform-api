package mentors

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain/entity"
	mentorspolicy "github.com/belikoooova/hackaton-platform-api/internal/mentors-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type ListAllTicketsIn struct {
	HackathonID string
	Limit       int32
	Offset      int32
}

type ListAllTicketsOut struct {
	Tickets []*entity.Ticket
	HasMore bool
}

func (s *Service) ListAllTickets(ctx context.Context, in ListAllTicketsIn) (*ListAllTicketsOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	hackathonID, err := uuid.Parse(in.HackathonID)
	if err != nil {
		return nil, fmt.Errorf("invalid hackathon_id: %w", err)
	}

	listAllTicketsPolicy := mentorspolicy.NewListAllTicketsPolicy()
	pctx, err := listAllTicketsPolicy.LoadContext(ctx, mentorspolicy.ListAllTicketsParams{
		HackathonID: hackathonID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load policy context: %w", err)
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)

	stage, err := s.hackathonClient.GetHackathon(ctx, in.HackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}
	pctx.SetHackathonStage(stage)

	roles, _, err := s.prClient.GetParticipationAndRoles(ctx, userID, in.HackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participation and roles: %w", err)
	}
	pctx.SetActorRoles(roles)

	decision := listAllTicketsPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	limit := in.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	tickets, err := s.ticketRepo.ListAll(ctx, hackathonID, limit+1, in.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tickets: %w", err)
	}

	hasMore := false
	if len(tickets) > int(limit) {
		hasMore = true
		tickets = tickets[:limit]
	}

	return &ListAllTicketsOut{
		Tickets: tickets,
		HasMore: hasMore,
	}, nil
}
