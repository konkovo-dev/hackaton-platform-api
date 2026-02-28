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

type ReplyInTicketIn struct {
	HackathonID    string
	TicketID       string
	Text           string
	IdempotencyKey string
}

type ReplyInTicketOut struct {
	MessageID string
}

func (s *Service) ReplyInTicket(ctx context.Context, in ReplyInTicketIn) (*ReplyInTicketOut, error) {
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

	if in.Text == "" {
		return nil, fmt.Errorf("%w: text is required", ErrInvalidInput)
	}

	if in.IdempotencyKey != "" {
		scope := fmt.Sprintf("mentors:reply_in_ticket:%s:%s:%s", hackathonID.String(), ticketID.String(), userUUID.String())
		requestHash := in.Text

		stored, err := s.idempotencyRepo.Get(ctx, in.IdempotencyKey, scope)
		if err == nil {
			if stored.RequestHash != requestHash {
				return nil, fmt.Errorf("%w: idempotency key already used with different request", ErrConflict)
			}
			var out ReplyInTicketOut
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

	replyInTicketPolicy := mentorspolicy.NewReplyInTicketPolicy()
	pctx, err := replyInTicketPolicy.LoadContext(ctx, mentorspolicy.ReplyInTicketParams{
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

	decision := replyInTicketPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	// Determine author role (only mentor or organizer can reach here)
	authorRole := domain.AuthorRoleMentor
	for _, role := range roles {
		if role == domain.RoleOrganizer || role == domain.RoleOwner {
			authorRole = domain.AuthorRoleOrganizer
			break
		}
	}

	message := &entity.Message{
		ID:           uuid.New(),
		TicketID:     ticketID,
		AuthorUserID: pgtype.UUID{Bytes: userUUID, Valid: true},
		AuthorRole:   authorRole,
		Text:         in.Text,
	}

	var messageID uuid.UUID

	err = s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		messageRepoTx := txrepo.NewMessageRepository(tx)
		if err := messageRepoTx.Create(ctx, message); err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}

		messageID = message.ID

		recipients, err := s.getTicketRecipients(ctx, ticketID, ticket.OwnerKind, ticket.OwnerID, hackathonID)
		if err != nil {
			return fmt.Errorf("failed to get recipients: %w", err)
		}

		eventPayload := outboxusecase.MessageCreatedPayload{
			MessageID:      messageID.String(),
			TicketID:       ticketID.String(),
			HackathonID:    hackathonID.String(),
			AuthorUserID:   userUUID.String(),
			AuthorRole:     authorRole,
			Text:           in.Text,
			RecipientUsers: recipients,
		}

		payloadBytes, err := json.Marshal(eventPayload)
		if err != nil {
			return fmt.Errorf("failed to marshal event payload: %w", err)
		}

		outboxRepoTx := txrepo.NewOutboxRepository(tx)
		outboxEvent := &outbox.Event{
			ID:            uuid.New(),
			AggregateID:   messageID.String(),
			AggregateType: "message",
			EventType:     outboxusecase.EventTypeMessageCreated,
			Payload:       payloadBytes,
			Status:        outbox.EventStatusPending,
			AttemptCount:  0,
			LastError:     "",
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}

		if err := outboxRepoTx.Create(ctx, outboxEvent); err != nil {
			return fmt.Errorf("failed to create outbox event: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	out := &ReplyInTicketOut{
		MessageID: messageID.String(),
	}

	if in.IdempotencyKey != "" {
		scope := fmt.Sprintf("mentors:reply_in_ticket:%s:%s:%s", hackathonID.String(), ticketID.String(), userUUID.String())
		requestHash := in.Text
		responseBlob, _ := json.Marshal(out)
		expiresAt := time.Now().Add(24 * time.Hour)
		_ = s.idempotencyRepo.Set(ctx, in.IdempotencyKey, scope, requestHash, responseBlob, expiresAt)
	}

	return out, nil
}
