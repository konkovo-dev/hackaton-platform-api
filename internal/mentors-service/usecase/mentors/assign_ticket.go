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
	"github.com/jackc/pgx/v5/pgtype"
)

type AssignTicketIn struct {
	HackathonID    string
	TicketID       string
	MentorUserID   string
	IdempotencyKey string
}

type AssignTicketOut struct {
	TicketID             string
	AssignedMentorUserID string
	AssignedAt           time.Time
}

func (s *Service) AssignTicket(ctx context.Context, in AssignTicketIn) (*AssignTicketOut, error) {
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

	mentorUserID, err := uuid.Parse(in.MentorUserID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid mentor_user_id", ErrInvalidInput)
	}

	if in.IdempotencyKey != "" {
		scope := fmt.Sprintf("mentors:assign_ticket:%s:%s", hackathonID.String(), ticketID.String())
		requestHash := mentorUserID.String()

		stored, err := s.idempotencyRepo.Get(ctx, in.IdempotencyKey, scope)
		if err == nil {
			if stored.RequestHash != requestHash {
				return nil, fmt.Errorf("%w: idempotency key already used with different request", ErrConflict)
			}
			var out AssignTicketOut
			if err := json.Unmarshal(stored.ResponseBlob, &out); err != nil {
				return nil, fmt.Errorf("failed to unmarshal idempotent response: %w", err)
			}
			return &out, nil
		}
	}

	ticket, err := s.ticketRepo.GetByIDAndHackathonID(ctx, ticketID, hackathonID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	assignTicketPolicy := mentorspolicy.NewAssignTicketPolicy()
	pctx, err := assignTicketPolicy.LoadContext(ctx, mentorspolicy.AssignTicketParams{
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

	decision := assignTicketPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	baseRecipients := make([]string, 0)
	switch ticket.OwnerKind {
	case domain.OwnerKindTeam:
		teamMembers, err := s.teamClient.ListTeamMembers(ctx, ticket.HackathonID.String(), ticket.OwnerID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to list team members: %w", err)
		}
		baseRecipients = append(baseRecipients, teamMembers...)
	case domain.OwnerKindUser:
		baseRecipients = append(baseRecipients, ticket.OwnerID.String())
	}

	assignedAt := time.Now().UTC()

	err = s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		ticketRepoTx := txrepo.NewTicketRepository(tx)
		rowsAffected, err := ticketRepoTx.AssignTicket(ctx, ticketID, mentorUserID)
		if err != nil {
			return fmt.Errorf("failed to assign ticket: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("%w: ticket not found or not open", ErrConflict)
		}

		recipients := make([]string, len(baseRecipients))
		copy(recipients, baseRecipients)
		recipients = append(recipients, mentorUserID.String())

		// Create system message about ticket assignment
		messageRepoTx := txrepo.NewMessageRepository(tx)
		systemMessage := &entity.Message{
			ID:              uuid.New(),
			TicketID:        ticketID,
			AuthorUserID:    pgtype.UUID{Valid: false},
			AuthorRole:      domain.AuthorRoleSystem,
			Text:            "Mentor joined the chat",
			ClientMessageID: "",
		}

		if err := messageRepoTx.Create(ctx, systemMessage); err != nil {
			return fmt.Errorf("failed to create system message: %w", err)
		}

		eventPayload := outboxusecase.TicketAssignedPayload{
			TicketID:             ticketID.String(),
			HackathonID:          hackathonID.String(),
			AssignedMentorUserID: mentorUserID.String(),
			AssignedAt:           assignedAt,
			RecipientUsers:       recipients,
		}

		payloadBytes, err := json.Marshal(eventPayload)
		if err != nil {
			return fmt.Errorf("failed to marshal event payload: %w", err)
		}

		outboxRepoTx := txrepo.NewOutboxRepository(tx)
		ticketAssignedEvent := &outbox.Event{
			ID:            uuid.New(),
			AggregateID:   ticketID.String(),
			AggregateType: "ticket",
			EventType:     outboxusecase.EventTypeTicketAssigned,
			Payload:       payloadBytes,
			Status:        outbox.EventStatusPending,
			AttemptCount:  0,
			LastError:     "",
			CreatedAt:     assignedAt,
			UpdatedAt:     assignedAt,
		}

		if err := outboxRepoTx.Create(ctx, ticketAssignedEvent); err != nil {
			return fmt.Errorf("failed to create ticket.assigned event: %w", err)
		}

		// Publish message.created event for system message
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
			CreatedAt:     assignedAt,
			UpdatedAt:     assignedAt,
		}

		if err := outboxRepoTx.Create(ctx, messageCreatedEvent); err != nil {
			return fmt.Errorf("failed to create message.created event: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	out := &AssignTicketOut{
		TicketID:             ticketID.String(),
		AssignedMentorUserID: mentorUserID.String(),
		AssignedAt:           assignedAt,
	}

	if in.IdempotencyKey != "" {
		scope := fmt.Sprintf("mentors:assign_ticket:%s:%s", hackathonID.String(), ticketID.String())
		requestHash := mentorUserID.String()
		responseBlob, _ := json.Marshal(out)
		expiresAt := time.Now().Add(24 * time.Hour)
		_ = s.idempotencyRepo.Set(ctx, in.IdempotencyKey, scope, requestHash, responseBlob, expiresAt)
	}

	return out, nil
}
