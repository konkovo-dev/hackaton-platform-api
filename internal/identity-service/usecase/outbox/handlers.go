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
	EventTypeUserSkillsUpdated = "user.skills.updated"
)

type UserSkillsUpdatedPayload struct {
	UserID           string    `json:"user_id"`
	CatalogSkillIDs  []string  `json:"catalog_skill_ids"`
	CustomSkillNames []string  `json:"custom_skill_names"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type UserSkillsUpdatedHandler struct {
	natsClient *natsclient.Client
	logger     *slog.Logger
}

func NewUserSkillsUpdatedHandler(
	natsClient *natsclient.Client,
	logger *slog.Logger,
) *UserSkillsUpdatedHandler {
	return &UserSkillsUpdatedHandler{
		natsClient: natsClient,
		logger:     logger,
	}
}

func (h *UserSkillsUpdatedHandler) EventType() string {
	return EventTypeUserSkillsUpdated
}

func (h *UserSkillsUpdatedHandler) Handle(ctx context.Context, event *outbox.Event) error {
	var payload UserSkillsUpdatedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	natsPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal NATS payload: %w", err)
	}

	subject := fmt.Sprintf("identity.user.skills.updated.%s", payload.UserID)
	if err := h.natsClient.Publish(ctx, subject, natsPayload); err != nil {
		return fmt.Errorf("failed to publish to NATS: %w", err)
	}

	h.logger.InfoContext(ctx, "published user.skills.updated event",
		"event_id", event.ID.String(),
		"user_id", payload.UserID,
	)

	return nil
}
