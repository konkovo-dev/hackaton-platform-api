package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/belikoooova/hackaton-platform-api/pkg/centrifugo"
	natsclient "github.com/belikoooova/hackaton-platform-api/pkg/nats"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
)

const (
	EventTypeMessageCreated = "message.created"
	EventTypeTicketClosed   = "ticket.closed"
	EventTypeTicketAssigned = "ticket.assigned"
)

type MessageCreatedPayload struct {
	MessageID      string   `json:"message_id"`
	TicketID       string   `json:"ticket_id"`
	HackathonID    string   `json:"hackathon_id"`
	AuthorUserID   string   `json:"author_user_id"`
	AuthorRole     string   `json:"author_role"`
	Text           string   `json:"text"`
	RecipientUsers []string `json:"recipient_users"`
}

type TicketClosedPayload struct {
	TicketID       string   `json:"ticket_id"`
	HackathonID    string   `json:"hackathon_id"`
	RecipientUsers []string `json:"recipient_users"`
}

type TicketAssignedPayload struct {
	TicketID             string    `json:"ticket_id"`
	HackathonID          string    `json:"hackathon_id"`
	AssignedMentorUserID string    `json:"assigned_mentor_user_id"`
	AssignedAt           time.Time `json:"assigned_at"`
	RecipientUsers       []string  `json:"recipient_users"`
}

type MessageCreatedHandler struct {
	natsClient       *natsclient.Client
	centrifugoClient *centrifugo.Client
	logger           *slog.Logger
}

func NewMessageCreatedHandler(
	natsClient *natsclient.Client,
	centrifugoClient *centrifugo.Client,
	logger *slog.Logger,
) *MessageCreatedHandler {
	return &MessageCreatedHandler{
		natsClient:       natsClient,
		centrifugoClient: centrifugoClient,
		logger:           logger,
	}
}

func (h *MessageCreatedHandler) EventType() string {
	return EventTypeMessageCreated
}

func (h *MessageCreatedHandler) Handle(ctx context.Context, event *outbox.Event) error {
	var payload MessageCreatedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	var publishErrors []error

	// Publish to NATS for audit/logging
	natsPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal NATS payload: %w", err)
	}

	subject := fmt.Sprintf("mentors.message.created.%s", payload.HackathonID)
	if err := h.natsClient.Publish(ctx, subject, natsPayload); err != nil {
		publishErrors = append(publishErrors, fmt.Errorf("NATS: %w", err))
	}

	// Publish to Centrifugo for real-time delivery
	centrifugoData := map[string]interface{}{
		"type":           "message.created",
		"message_id":     payload.MessageID,
		"ticket_id":      payload.TicketID,
		"hackathon_id":   payload.HackathonID,
		"author_user_id": payload.AuthorUserID,
		"author_role":    payload.AuthorRole,
		"text":           payload.Text,
	}

	if err := h.centrifugoClient.PublishToUsers(ctx, payload.RecipientUsers, "support:feed", centrifugoData); err != nil {
		publishErrors = append(publishErrors, fmt.Errorf("Centrifugo: %w", err))
	}

	if len(publishErrors) > 0 {
		return fmt.Errorf("failed to publish: %v", publishErrors)
	}

	h.logger.InfoContext(ctx, "published message.created event",
		"event_id", event.ID.String(),
		"message_id", payload.MessageID,
		"ticket_id", payload.TicketID,
		"recipients_count", len(payload.RecipientUsers),
	)

	return nil
}

type TicketClosedHandler struct {
	natsClient       *natsclient.Client
	centrifugoClient *centrifugo.Client
	logger           *slog.Logger
}

func NewTicketClosedHandler(
	natsClient *natsclient.Client,
	centrifugoClient *centrifugo.Client,
	logger *slog.Logger,
) *TicketClosedHandler {
	return &TicketClosedHandler{
		natsClient:       natsClient,
		centrifugoClient: centrifugoClient,
		logger:           logger,
	}
}

func (h *TicketClosedHandler) EventType() string {
	return EventTypeTicketClosed
}

func (h *TicketClosedHandler) Handle(ctx context.Context, event *outbox.Event) error {
	var payload TicketClosedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	var publishErrors []error

	// Publish to NATS for audit/logging
	natsPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal NATS payload: %w", err)
	}

	subject := fmt.Sprintf("mentors.ticket.closed.%s", payload.HackathonID)
	if err := h.natsClient.Publish(ctx, subject, natsPayload); err != nil {
		publishErrors = append(publishErrors, fmt.Errorf("NATS: %w", err))
	}

	// Publish to Centrifugo for real-time delivery
	centrifugoData := map[string]interface{}{
		"type":         "ticket.closed",
		"ticket_id":    payload.TicketID,
		"hackathon_id": payload.HackathonID,
	}

	if err := h.centrifugoClient.PublishToUsers(ctx, payload.RecipientUsers, "support:feed", centrifugoData); err != nil {
		publishErrors = append(publishErrors, fmt.Errorf("Centrifugo: %w", err))
	}

	if len(publishErrors) > 0 {
		return fmt.Errorf("failed to publish: %v", publishErrors)
	}

	h.logger.InfoContext(ctx, "published ticket.closed event",
		"event_id", event.ID.String(),
		"ticket_id", payload.TicketID,
		"recipients_count", len(payload.RecipientUsers),
	)

	return nil
}

type TicketAssignedHandler struct {
	natsClient       *natsclient.Client
	centrifugoClient *centrifugo.Client
	logger           *slog.Logger
}

func NewTicketAssignedHandler(
	natsClient *natsclient.Client,
	centrifugoClient *centrifugo.Client,
	logger *slog.Logger,
) *TicketAssignedHandler {
	return &TicketAssignedHandler{
		natsClient:       natsClient,
		centrifugoClient: centrifugoClient,
		logger:           logger,
	}
}

func (h *TicketAssignedHandler) EventType() string {
	return EventTypeTicketAssigned
}

func (h *TicketAssignedHandler) Handle(ctx context.Context, event *outbox.Event) error {
	var payload TicketAssignedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	var publishErrors []error

	natsPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal NATS payload: %w", err)
	}

	subject := fmt.Sprintf("mentors.ticket.assigned.%s", payload.HackathonID)
	if err := h.natsClient.Publish(ctx, subject, natsPayload); err != nil {
		publishErrors = append(publishErrors, fmt.Errorf("NATS: %w", err))
	}

	centrifugoData := map[string]interface{}{
		"type":                    "ticket.assigned",
		"ticket_id":               payload.TicketID,
		"hackathon_id":            payload.HackathonID,
		"assigned_mentor_user_id": payload.AssignedMentorUserID,
		"assigned_at":             payload.AssignedAt,
	}

	if err := h.centrifugoClient.PublishToUsers(ctx, payload.RecipientUsers, "support:feed", centrifugoData); err != nil {
		publishErrors = append(publishErrors, fmt.Errorf("Centrifugo: %w", err))
	}

	if len(publishErrors) > 0 {
		return fmt.Errorf("failed to publish: %v", publishErrors)
	}

	h.logger.InfoContext(ctx, "published ticket.assigned event",
		"event_id", event.ID.String(),
		"ticket_id", payload.TicketID,
		"mentor_id", payload.AssignedMentorUserID,
		"recipients_count", len(payload.RecipientUsers),
	)

	return nil
}
