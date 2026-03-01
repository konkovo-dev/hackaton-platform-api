package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	natsclient "github.com/belikoooova/hackaton-platform-api/pkg/nats"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
)

const (
	EventTypeParticipationRegistered   = "participation.registered"
	EventTypeParticipationUpdated      = "participation.updated"
	EventTypeParticipationStatusChanged = "participation.status_changed"
	EventTypeParticipationTeamAssigned = "participation.team_assigned"
	EventTypeParticipationTeamRemoved  = "participation.team_removed"
)

type ParticipationRegisteredPayload struct {
	HackathonID    string    `json:"hackathon_id"`
	UserID         string    `json:"user_id"`
	Status         string    `json:"status"`
	WishedRoleIDs  []string  `json:"wished_role_ids"`
	MotivationText string    `json:"motivation_text"`
	RegisteredAt   time.Time `json:"registered_at"`
}

type ParticipationUpdatedPayload struct {
	HackathonID    string    `json:"hackathon_id"`
	UserID         string    `json:"user_id"`
	WishedRoleIDs  []string  `json:"wished_role_ids"`
	MotivationText string    `json:"motivation_text"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ParticipationStatusChangedPayload struct {
	HackathonID string    `json:"hackathon_id"`
	UserID      string    `json:"user_id"`
	OldStatus   string    `json:"old_status"`
	NewStatus   string    `json:"new_status"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ParticipationTeamAssignedPayload struct {
	HackathonID string    `json:"hackathon_id"`
	UserID      string    `json:"user_id"`
	TeamID      string    `json:"team_id"`
	IsCaptain   bool      `json:"is_captain"`
	AssignedAt  time.Time `json:"assigned_at"`
}

type ParticipationTeamRemovedPayload struct {
	HackathonID string    `json:"hackathon_id"`
	UserID      string    `json:"user_id"`
	TeamID      string    `json:"team_id"`
	RemovedAt   time.Time `json:"removed_at"`
}

type ParticipationRegisteredHandler struct {
	natsClient *natsclient.Client
	logger     *slog.Logger
}

func NewParticipationRegisteredHandler(
	natsClient *natsclient.Client,
	logger *slog.Logger,
) *ParticipationRegisteredHandler {
	return &ParticipationRegisteredHandler{
		natsClient: natsClient,
		logger:     logger,
	}
}

func (h *ParticipationRegisteredHandler) EventType() string {
	return EventTypeParticipationRegistered
}

func (h *ParticipationRegisteredHandler) Handle(ctx context.Context, event *outbox.Event) error {
	var payload ParticipationRegisteredPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	natsPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal NATS payload: %w", err)
	}

	subject := fmt.Sprintf("participation.registered.%s", payload.HackathonID)
	if err := h.natsClient.Publish(ctx, subject, natsPayload); err != nil {
		return fmt.Errorf("failed to publish to NATS: %w", err)
	}

	h.logger.InfoContext(ctx, "published participation.registered event",
		"event_id", event.ID.String(),
		"user_id", payload.UserID,
		"hackathon_id", payload.HackathonID,
	)

	return nil
}

type ParticipationUpdatedHandler struct {
	natsClient *natsclient.Client
	logger     *slog.Logger
}

func NewParticipationUpdatedHandler(
	natsClient *natsclient.Client,
	logger *slog.Logger,
) *ParticipationUpdatedHandler {
	return &ParticipationUpdatedHandler{
		natsClient: natsClient,
		logger:     logger,
	}
}

func (h *ParticipationUpdatedHandler) EventType() string {
	return EventTypeParticipationUpdated
}

func (h *ParticipationUpdatedHandler) Handle(ctx context.Context, event *outbox.Event) error {
	var payload ParticipationUpdatedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	natsPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal NATS payload: %w", err)
	}

	subject := fmt.Sprintf("participation.updated.%s", payload.HackathonID)
	if err := h.natsClient.Publish(ctx, subject, natsPayload); err != nil {
		return fmt.Errorf("failed to publish to NATS: %w", err)
	}

	h.logger.InfoContext(ctx, "published participation.updated event",
		"event_id", event.ID.String(),
		"user_id", payload.UserID,
		"hackathon_id", payload.HackathonID,
	)

	return nil
}

type ParticipationStatusChangedHandler struct {
	natsClient *natsclient.Client
	logger     *slog.Logger
}

func NewParticipationStatusChangedHandler(
	natsClient *natsclient.Client,
	logger *slog.Logger,
) *ParticipationStatusChangedHandler {
	return &ParticipationStatusChangedHandler{
		natsClient: natsClient,
		logger:     logger,
	}
}

func (h *ParticipationStatusChangedHandler) EventType() string {
	return EventTypeParticipationStatusChanged
}

func (h *ParticipationStatusChangedHandler) Handle(ctx context.Context, event *outbox.Event) error {
	var payload ParticipationStatusChangedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	natsPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal NATS payload: %w", err)
	}

	subject := fmt.Sprintf("participation.status_changed.%s", payload.HackathonID)
	if err := h.natsClient.Publish(ctx, subject, natsPayload); err != nil {
		return fmt.Errorf("failed to publish to NATS: %w", err)
	}

	h.logger.InfoContext(ctx, "published participation.status_changed event",
		"event_id", event.ID.String(),
		"user_id", payload.UserID,
		"hackathon_id", payload.HackathonID,
		"old_status", payload.OldStatus,
		"new_status", payload.NewStatus,
	)

	return nil
}

type ParticipationTeamAssignedHandler struct {
	natsClient *natsclient.Client
	logger     *slog.Logger
}

func NewParticipationTeamAssignedHandler(
	natsClient *natsclient.Client,
	logger *slog.Logger,
) *ParticipationTeamAssignedHandler {
	return &ParticipationTeamAssignedHandler{
		natsClient: natsClient,
		logger:     logger,
	}
}

func (h *ParticipationTeamAssignedHandler) EventType() string {
	return EventTypeParticipationTeamAssigned
}

func (h *ParticipationTeamAssignedHandler) Handle(ctx context.Context, event *outbox.Event) error {
	var payload ParticipationTeamAssignedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	natsPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal NATS payload: %w", err)
	}

	subject := fmt.Sprintf("participation.team_assigned.%s", payload.HackathonID)
	if err := h.natsClient.Publish(ctx, subject, natsPayload); err != nil {
		return fmt.Errorf("failed to publish to NATS: %w", err)
	}

	h.logger.InfoContext(ctx, "published participation.team_assigned event",
		"event_id", event.ID.String(),
		"user_id", payload.UserID,
		"hackathon_id", payload.HackathonID,
		"team_id", payload.TeamID,
	)

	return nil
}

type ParticipationTeamRemovedHandler struct {
	natsClient *natsclient.Client
	logger     *slog.Logger
}

func NewParticipationTeamRemovedHandler(
	natsClient *natsclient.Client,
	logger *slog.Logger,
) *ParticipationTeamRemovedHandler {
	return &ParticipationTeamRemovedHandler{
		natsClient: natsClient,
		logger:     logger,
	}
}

func (h *ParticipationTeamRemovedHandler) EventType() string {
	return EventTypeParticipationTeamRemoved
}

func (h *ParticipationTeamRemovedHandler) Handle(ctx context.Context, event *outbox.Event) error {
	var payload ParticipationTeamRemovedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	natsPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal NATS payload: %w", err)
	}

	subject := fmt.Sprintf("participation.team_removed.%s", payload.HackathonID)
	if err := h.natsClient.Publish(ctx, subject, natsPayload); err != nil {
		return fmt.Errorf("failed to publish to NATS: %w", err)
	}

	h.logger.InfoContext(ctx, "published participation.team_removed event",
		"event_id", event.ID.String(),
		"user_id", payload.UserID,
		"hackathon_id", payload.HackathonID,
		"team_id", payload.TeamID,
	)

	return nil
}
