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
	EventTypeTeamCreated       = "team.created"
	EventTypeTeamUpdated       = "team.updated"
	EventTypeTeamDeleted       = "team.deleted"
	EventTypeVacancyCreated    = "vacancy.created"
	EventTypeVacancyUpdated    = "vacancy.updated"
	EventTypeVacancySlotsChanged = "vacancy.slots_changed"
)

type TeamCreatedPayload struct {
	TeamID      string    `json:"team_id"`
	HackathonID string    `json:"hackathon_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsJoinable  bool      `json:"is_joinable"`
	CreatedAt   time.Time `json:"created_at"`
}

type TeamUpdatedPayload struct {
	TeamID      string    `json:"team_id"`
	HackathonID string    `json:"hackathon_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsJoinable  bool      `json:"is_joinable"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TeamDeletedPayload struct {
	TeamID      string    `json:"team_id"`
	HackathonID string    `json:"hackathon_id"`
	DeletedAt   time.Time `json:"deleted_at"`
}

type VacancyCreatedPayload struct {
	VacancyID        string    `json:"vacancy_id"`
	TeamID           string    `json:"team_id"`
	HackathonID      string    `json:"hackathon_id"`
	Description      string    `json:"description"`
	DesiredRoleIDs   []string  `json:"desired_role_ids"`
	DesiredSkillIDs  []string  `json:"desired_skill_ids"`
	SlotsOpen        int32     `json:"slots_open"`
	CreatedAt        time.Time `json:"created_at"`
}

type VacancyUpdatedPayload struct {
	VacancyID        string    `json:"vacancy_id"`
	TeamID           string    `json:"team_id"`
	HackathonID      string    `json:"hackathon_id"`
	Description      string    `json:"description"`
	DesiredRoleIDs   []string  `json:"desired_role_ids"`
	DesiredSkillIDs  []string  `json:"desired_skill_ids"`
	SlotsOpen        int32     `json:"slots_open"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type VacancySlotsChangedPayload struct {
	VacancyID   string    `json:"vacancy_id"`
	TeamID      string    `json:"team_id"`
	HackathonID string    `json:"hackathon_id"`
	SlotsOpen   int32     `json:"slots_open"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TeamCreatedHandler struct {
	natsClient *natsclient.Client
	logger     *slog.Logger
}

func NewTeamCreatedHandler(
	natsClient *natsclient.Client,
	logger *slog.Logger,
) *TeamCreatedHandler {
	return &TeamCreatedHandler{
		natsClient: natsClient,
		logger:     logger,
	}
}

func (h *TeamCreatedHandler) EventType() string {
	return EventTypeTeamCreated
}

func (h *TeamCreatedHandler) Handle(ctx context.Context, event *outbox.Event) error {
	var payload TeamCreatedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	natsPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal NATS payload: %w", err)
	}

	subject := fmt.Sprintf("team.created.%s", payload.HackathonID)
	if err := h.natsClient.Publish(ctx, subject, natsPayload); err != nil {
		return fmt.Errorf("failed to publish to NATS: %w", err)
	}

	h.logger.InfoContext(ctx, "published team.created event",
		"event_id", event.ID.String(),
		"team_id", payload.TeamID,
		"hackathon_id", payload.HackathonID,
	)

	return nil
}

type TeamUpdatedHandler struct {
	natsClient *natsclient.Client
	logger     *slog.Logger
}

func NewTeamUpdatedHandler(
	natsClient *natsclient.Client,
	logger *slog.Logger,
) *TeamUpdatedHandler {
	return &TeamUpdatedHandler{
		natsClient: natsClient,
		logger:     logger,
	}
}

func (h *TeamUpdatedHandler) EventType() string {
	return EventTypeTeamUpdated
}

func (h *TeamUpdatedHandler) Handle(ctx context.Context, event *outbox.Event) error {
	var payload TeamUpdatedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	natsPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal NATS payload: %w", err)
	}

	subject := fmt.Sprintf("team.updated.%s", payload.HackathonID)
	if err := h.natsClient.Publish(ctx, subject, natsPayload); err != nil {
		return fmt.Errorf("failed to publish to NATS: %w", err)
	}

	h.logger.InfoContext(ctx, "published team.updated event",
		"event_id", event.ID.String(),
		"team_id", payload.TeamID,
		"hackathon_id", payload.HackathonID,
	)

	return nil
}

type TeamDeletedHandler struct {
	natsClient *natsclient.Client
	logger     *slog.Logger
}

func NewTeamDeletedHandler(
	natsClient *natsclient.Client,
	logger *slog.Logger,
) *TeamDeletedHandler {
	return &TeamDeletedHandler{
		natsClient: natsClient,
		logger:     logger,
	}
}

func (h *TeamDeletedHandler) EventType() string {
	return EventTypeTeamDeleted
}

func (h *TeamDeletedHandler) Handle(ctx context.Context, event *outbox.Event) error {
	var payload TeamDeletedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	natsPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal NATS payload: %w", err)
	}

	subject := fmt.Sprintf("team.deleted.%s", payload.HackathonID)
	if err := h.natsClient.Publish(ctx, subject, natsPayload); err != nil {
		return fmt.Errorf("failed to publish to NATS: %w", err)
	}

	h.logger.InfoContext(ctx, "published team.deleted event",
		"event_id", event.ID.String(),
		"team_id", payload.TeamID,
		"hackathon_id", payload.HackathonID,
	)

	return nil
}

type VacancyCreatedHandler struct {
	natsClient *natsclient.Client
	logger     *slog.Logger
}

func NewVacancyCreatedHandler(
	natsClient *natsclient.Client,
	logger *slog.Logger,
) *VacancyCreatedHandler {
	return &VacancyCreatedHandler{
		natsClient: natsClient,
		logger:     logger,
	}
}

func (h *VacancyCreatedHandler) EventType() string {
	return EventTypeVacancyCreated
}

func (h *VacancyCreatedHandler) Handle(ctx context.Context, event *outbox.Event) error {
	var payload VacancyCreatedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	natsPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal NATS payload: %w", err)
	}

	subject := fmt.Sprintf("team.vacancy.created.%s", payload.HackathonID)
	if err := h.natsClient.Publish(ctx, subject, natsPayload); err != nil {
		return fmt.Errorf("failed to publish to NATS: %w", err)
	}

	h.logger.InfoContext(ctx, "published vacancy.created event",
		"event_id", event.ID.String(),
		"vacancy_id", payload.VacancyID,
		"team_id", payload.TeamID,
		"hackathon_id", payload.HackathonID,
	)

	return nil
}

type VacancyUpdatedHandler struct {
	natsClient *natsclient.Client
	logger     *slog.Logger
}

func NewVacancyUpdatedHandler(
	natsClient *natsclient.Client,
	logger *slog.Logger,
) *VacancyUpdatedHandler {
	return &VacancyUpdatedHandler{
		natsClient: natsClient,
		logger:     logger,
	}
}

func (h *VacancyUpdatedHandler) EventType() string {
	return EventTypeVacancyUpdated
}

func (h *VacancyUpdatedHandler) Handle(ctx context.Context, event *outbox.Event) error {
	var payload VacancyUpdatedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	natsPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal NATS payload: %w", err)
	}

	subject := fmt.Sprintf("team.vacancy.updated.%s", payload.HackathonID)
	if err := h.natsClient.Publish(ctx, subject, natsPayload); err != nil {
		return fmt.Errorf("failed to publish to NATS: %w", err)
	}

	h.logger.InfoContext(ctx, "published vacancy.updated event",
		"event_id", event.ID.String(),
		"vacancy_id", payload.VacancyID,
		"team_id", payload.TeamID,
		"hackathon_id", payload.HackathonID,
	)

	return nil
}

type VacancySlotsChangedHandler struct {
	natsClient *natsclient.Client
	logger     *slog.Logger
}

func NewVacancySlotsChangedHandler(
	natsClient *natsclient.Client,
	logger *slog.Logger,
) *VacancySlotsChangedHandler {
	return &VacancySlotsChangedHandler{
		natsClient: natsClient,
		logger:     logger,
	}
}

func (h *VacancySlotsChangedHandler) EventType() string {
	return EventTypeVacancySlotsChanged
}

func (h *VacancySlotsChangedHandler) Handle(ctx context.Context, event *outbox.Event) error {
	var payload VacancySlotsChangedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	natsPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal NATS payload: %w", err)
	}

	subject := fmt.Sprintf("team.vacancy.slots_changed.%s", payload.HackathonID)
	if err := h.natsClient.Publish(ctx, subject, natsPayload); err != nil {
		return fmt.Errorf("failed to publish to NATS: %w", err)
	}

	h.logger.InfoContext(ctx, "published vacancy.slots_changed event",
		"event_id", event.ID.String(),
		"vacancy_id", payload.VacancyID,
		"team_id", payload.TeamID,
		"hackathon_id", payload.HackathonID,
		"slots_open", payload.SlotsOpen,
	)

	return nil
}
