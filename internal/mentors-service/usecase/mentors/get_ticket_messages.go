package mentors

import (
	"context"
	"errors"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain/entity"
	mentorspolicy "github.com/belikoooova/hackaton-platform-api/internal/mentors-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type GetTicketMessagesIn struct {
	HackathonID string
	TicketID    string
	Limit       int32
	Offset      int32
}

type GetTicketMessagesOut struct {
	Messages []*entity.Message
	HasMore  bool
}

func (s *Service) GetTicketMessages(ctx context.Context, in GetTicketMessagesIn) (*GetTicketMessagesOut, error) {
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

	ticketID, err := uuid.Parse(in.TicketID)
	if err != nil {
		return nil, fmt.Errorf("invalid ticket_id: %w", err)
	}

	ticket, err := s.ticketRepo.GetByIDAndHackathonID(ctx, ticketID, hackathonID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	getTicketMessagesPolicy := mentorspolicy.NewGetTicketMessagesPolicy()
	pctx, err := getTicketMessagesPolicy.LoadContext(ctx, mentorspolicy.GetTicketMessagesParams{
		HackathonID: hackathonID,
		TicketID:    ticketID,
		Ticket:      ticket,
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

	decision := getTicketMessagesPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	limit := in.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	messages, err := s.messageRepo.ListByTicket(ctx, ticketID, limit+1, in.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	hasMore := false
	if len(messages) > int(limit) {
		hasMore = true
		messages = messages[:limit]
	}

	return &GetTicketMessagesOut{
		Messages: messages,
		HasMore:  hasMore,
	}, nil
}
