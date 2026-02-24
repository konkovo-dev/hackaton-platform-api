package mentors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain/entity"
	mentorspolicy "github.com/belikoooova/hackaton-platform-api/internal/mentors-service/policy"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/txrepo"
	outboxusecase "github.com/belikoooova/hackaton-platform-api/internal/mentors-service/usecase/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CloseTicketIn struct {
	HackathonID    string
	TicketID       string
	IdempotencyKey string
}

type CloseTicketOut struct{}

func (s *Service) CloseTicket(ctx context.Context, in CloseTicketIn) (*CloseTicketOut, error) {
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
		return nil, fmt.Errorf("%w: invalid hackathon_id", ErrInvalidInput)
	}

	ticketID, err := uuid.Parse(in.TicketID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid ticket_id", ErrInvalidInput)
	}

	ticket, err := s.ticketRepo.GetByIDAndHackathonID(ctx, ticketID, hackathonID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	if ticket.Status == domain.TicketStatusClosed {
		return &CloseTicketOut{}, nil
	}

	if in.IdempotencyKey != "" {
		scope := fmt.Sprintf("mentors:close_ticket:%s:%s", hackathonID.String(), ticketID.String())
		requestHash := userUUID.String()

		stored, err := s.idempotencyRepo.Get(ctx, in.IdempotencyKey, scope)
		if err == nil {
			if stored.RequestHash != requestHash {
				return nil, fmt.Errorf("%w: idempotency key already used with different request", ErrConflict)
			}
			var out CloseTicketOut
			if err := json.Unmarshal(stored.ResponseBlob, &out); err != nil {
				return nil, fmt.Errorf("failed to unmarshal idempotent response: %w", err)
			}
			return &out, nil
		}
	}

	closeTicketPolicy := mentorspolicy.NewCloseTicketPolicy()
	pctx, err := closeTicketPolicy.LoadContext(ctx, mentorspolicy.CloseTicketParams{
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

	decision := closeTicketPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	baseRecipients := make([]string, 0)
	switch ticket.OwnerKind {
	case domain.OwnerKindTeam:
		teamMembers, err := s.teamClient.ListTeamMembers(ctx, ticket.OwnerID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to list team members: %w", err)
		}
		baseRecipients = append(baseRecipients, teamMembers...)
	case domain.OwnerKindUser:
		baseRecipients = append(baseRecipients, ticket.OwnerID.String())
	}

	if ticket.AssignedMentorUserID != nil {
		mentorID := ticket.AssignedMentorUserID.String()
		alreadyIncluded := false
		for _, r := range baseRecipients {
			if r == mentorID {
				alreadyIncluded = true
				break
			}
		}
		if !alreadyIncluded {
			baseRecipients = append(baseRecipients, mentorID)
		}
	}

	now := time.Now().UTC()

	err = s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		ticketRepoTx := txrepo.NewTicketRepository(tx)
		if err := ticketRepoTx.UpdateStatus(ctx, ticketID, domain.TicketStatusClosed, &now); err != nil {
			return fmt.Errorf("failed to close ticket: %w", err)
		}

		messageRepoTx := txrepo.NewMessageRepository(tx)
		systemMessage := &entity.Message{
			ID:              uuid.New(),
			TicketID:        ticketID,
			AuthorUserID:    uuid.Nil,
			AuthorRole:      domain.AuthorRoleSystem,
			Text:            "Ticket closed",
			ClientMessageID: "",
			CreatedAt:       now,
		}

		if err := messageRepoTx.Create(ctx, systemMessage); err != nil {
			return fmt.Errorf("failed to create system message: %w", err)
		}

		recipients := make([]string, len(baseRecipients))
		copy(recipients, baseRecipients)

		ticketClosedPayload := outboxusecase.TicketClosedPayload{
			TicketID:       ticketID.String(),
			HackathonID:    hackathonID.String(),
			RecipientUsers: recipients,
		}

		ticketClosedBytes, err := json.Marshal(ticketClosedPayload)
		if err != nil {
			return fmt.Errorf("failed to marshal ticket.closed payload: %w", err)
		}

		outboxRepoTx := txrepo.NewOutboxRepository(tx)
		ticketClosedEvent := &outbox.Event{
			ID:            uuid.New(),
			AggregateID:   ticketID.String(),
			AggregateType: "ticket",
			EventType:     outboxusecase.EventTypeTicketClosed,
			Payload:       ticketClosedBytes,
			Status:        outbox.EventStatusPending,
			AttemptCount:  0,
			LastError:     "",
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if err := outboxRepoTx.Create(ctx, ticketClosedEvent); err != nil {
			return fmt.Errorf("failed to create ticket.closed event: %w", err)
		}

		messageCreatedPayload := outboxusecase.MessageCreatedPayload{
			MessageID:      systemMessage.ID.String(),
			TicketID:       ticketID.String(),
			HackathonID:    hackathonID.String(),
			AuthorUserID:   "",
			AuthorRole:     domain.AuthorRoleSystem,
			Text:           systemMessage.Text,
			RecipientUsers: recipients,
		}

		messageCreatedBytes, err := json.Marshal(messageCreatedPayload)
		if err != nil {
			return fmt.Errorf("failed to marshal message.created payload: %w", err)
		}

		messageCreatedEvent := &outbox.Event{
			ID:            uuid.New(),
			AggregateID:   systemMessage.ID.String(),
			AggregateType: "message",
			EventType:     outboxusecase.EventTypeMessageCreated,
			Payload:       messageCreatedBytes,
			Status:        outbox.EventStatusPending,
			AttemptCount:  0,
			LastError:     "",
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if err := outboxRepoTx.Create(ctx, messageCreatedEvent); err != nil {
			return fmt.Errorf("failed to create message.created event: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	out := &CloseTicketOut{}

	if in.IdempotencyKey != "" {
		scope := fmt.Sprintf("mentors:close_ticket:%s:%s", hackathonID.String(), ticketID.String())
		requestHash := userUUID.String()
		responseBlob, _ := json.Marshal(out)
		expiresAt := time.Now().Add(24 * time.Hour)
		_ = s.idempotencyRepo.Set(ctx, in.IdempotencyKey, scope, requestHash, responseBlob, expiresAt)
	}

	return out, nil
}
