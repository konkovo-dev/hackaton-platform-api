package mentors

import (
	"context"
	"encoding/json"
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

type SendMessageIn struct {
	HackathonID     string
	Text            string
	ClientMessageID string
	IdempotencyKey  string
}

type SendMessageOut struct {
	MessageID string
	TicketID  string
}

func (s *Service) SendMessage(ctx context.Context, in SendMessageIn) (*SendMessageOut, error) {
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

	if in.Text == "" {
		return nil, fmt.Errorf("%w: text is required", ErrInvalidInput)
	}

	if in.IdempotencyKey != "" {
		scope := fmt.Sprintf("mentors:send_message:%s:%s", hackathonID.String(), userUUID.String())
		requestHash := fmt.Sprintf("%s:%s", in.Text, in.ClientMessageID)

		stored, err := s.idempotencyRepo.Get(ctx, in.IdempotencyKey, scope)
		if err == nil {
			if stored.RequestHash != requestHash {
				return nil, fmt.Errorf("%w: idempotency key already used with different request", ErrConflict)
			}
			var out SendMessageOut
			if err := json.Unmarshal(stored.ResponseBlob, &out); err != nil {
				return nil, fmt.Errorf("failed to unmarshal idempotent response: %w", err)
			}
			return &out, nil
		}
	}

	sendMessagePolicy := mentorspolicy.NewSendMessagePolicy()
	pctx, err := sendMessagePolicy.LoadContext(ctx, mentorspolicy.SendMessageParams{
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

	decision := sendMessagePolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	_, teamIDPtr, err := s.prClient.GetParticipationAndRoles(ctx, userID, in.HackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participation and roles: %w", err)
	}

	var ownerKind string
	var ownerID uuid.UUID

	if teamIDPtr != nil && *teamIDPtr != "" {
		teamID, err := uuid.Parse(*teamIDPtr)
		if err != nil {
			return nil, fmt.Errorf("invalid team_id from participation service: %w", err)
		}
		ownerKind = domain.OwnerKindTeam
		ownerID = teamID
	} else {
		ownerKind = domain.OwnerKindUser
		ownerID = userUUID
	}

	baseRecipients := make([]string, 0)
	switch ownerKind {
	case domain.OwnerKindTeam:
		if ownerID == uuid.Nil {
			return nil, fmt.Errorf("owner_id is nil for TEAM owner (teamIDPtr=%v)", teamIDPtr)
		}
		teamMembers, err := s.teamClient.ListTeamMembers(ctx, hackathonID.String(), ownerID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to list team members (ownerID=%s): %w", ownerID.String(), err)
		}
		baseRecipients = append(baseRecipients, teamMembers...)
	case domain.OwnerKindUser:
		baseRecipients = append(baseRecipients, ownerID.String())
	}

	if in.ClientMessageID != "" {
		existingMsg, err := s.messageRepo.FindByClientMessageID(ctx, in.ClientMessageID)
		if err == nil && existingMsg != nil {
			return &SendMessageOut{
				MessageID: existingMsg.ID.String(),
				TicketID:  existingMsg.TicketID.String(),
			}, nil
		}
	}

	var messageID uuid.UUID
	var ticketID uuid.UUID

	err = s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		ticketRepoTx := txrepo.NewTicketRepository(tx)

		ticket, err := ticketRepoTx.CreateOrGetOpenTicket(ctx, hackathonID, ownerKind, ownerID)
		if err != nil {
			return fmt.Errorf("failed to create or get ticket: %w", err)
		}

		ticketID = ticket.ID

		authorRole := domain.AuthorRoleParticipant
		for _, role := range roles {
			if role == domain.RoleOrganizer || role == domain.RoleOwner {
				authorRole = domain.AuthorRoleOrganizer
				break
			}
		}

		message := &entity.Message{
			ID:              uuid.New(),
			TicketID:        ticketID,
			AuthorUserID:    pgtype.UUID{Bytes: userUUID, Valid: true},
			AuthorRole:      authorRole,
			Text:            in.Text,
			ClientMessageID: in.ClientMessageID,
		}

		messageRepoTx := txrepo.NewMessageRepository(tx)
		if err := messageRepoTx.Create(ctx, message); err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}

		messageID = message.ID

		recipients := make([]string, len(baseRecipients))
		copy(recipients, baseRecipients)
		currentTicket, err := s.ticketRepo.GetByID(ctx, ticketID)
		if err == nil && currentTicket.AssignedMentorUserID != nil {
			mentorID := currentTicket.AssignedMentorUserID.String()
			alreadyIncluded := false
			for _, r := range recipients {
				if r == mentorID {
					alreadyIncluded = true
					break
				}
			}
			if !alreadyIncluded {
				recipients = append(recipients, mentorID)
			}
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
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	out := &SendMessageOut{
		MessageID: messageID.String(),
		TicketID:  ticketID.String(),
	}

	if in.IdempotencyKey != "" {
		scope := fmt.Sprintf("mentors:send_message:%s:%s", hackathonID.String(), userUUID.String())
		requestHash := fmt.Sprintf("%s:%s", in.Text, in.ClientMessageID)
		responseBlob, _ := json.Marshal(out)
		expiresAt := time.Now().Add(24 * time.Hour)
		_ = s.idempotencyRepo.Set(ctx, in.IdempotencyKey, scope, requestHash, responseBlob, expiresAt)
	}

	return out, nil
}
