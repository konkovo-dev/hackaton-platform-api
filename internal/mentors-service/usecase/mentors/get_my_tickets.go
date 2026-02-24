package mentors

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain/entity"
	mentorspolicy "github.com/belikoooova/hackaton-platform-api/internal/mentors-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type GetMyTicketsIn struct {
	HackathonID string
	Limit       int32
	Offset      int32
}

type GetMyTicketsOut struct {
	Tickets []*entity.Ticket
	HasMore bool
}

func (s *Service) GetMyTickets(ctx context.Context, in GetMyTicketsIn) (*GetMyTicketsOut, error) {
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

	getMyTicketsPolicy := mentorspolicy.NewGetMyTicketsPolicy()
	pctx, err := getMyTicketsPolicy.LoadContext(ctx, mentorspolicy.GetMyTicketsParams{
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

	decision := getMyTicketsPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	_, teamIDPtr, err := s.prClient.GetParticipationAndRoles(ctx, userID, in.HackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participation and roles: %w", err)
	}

	var ownerKind string
	var ownerID uuid.UUID

	if teamIDPtr != nil {
		teamID, err := uuid.Parse(*teamIDPtr)
		if err == nil {
			ownerKind = domain.OwnerKindTeam
			ownerID = teamID
		} else {
			ownerKind = domain.OwnerKindUser
			ownerID = userUUID
		}
	} else {
		ownerKind = domain.OwnerKindUser
		ownerID = userUUID
	}

	limit := in.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	tickets, err := s.ticketRepo.ListByOwner(ctx, hackathonID, ownerKind, ownerID, limit+1, in.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tickets: %w", err)
	}

	hasMore := false
	if len(tickets) > int(limit) {
		hasMore = true
		tickets = tickets[:limit]
	}

	return &GetMyTicketsOut{
		Tickets: tickets,
		HasMore: hasMore,
	}, nil
}
