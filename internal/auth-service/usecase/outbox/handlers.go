package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/client/identity"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
)

const (
	EventTypeUserRegistered = "user.registered"
)

type UserRegisteredPayload struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Timezone  string `json:"timezone"`
}

type UserRegisteredHandler struct {
	identityClient *identity.Client
	logger         *slog.Logger
}

func NewUserRegisteredHandler(identityClient *identity.Client, logger *slog.Logger) *UserRegisteredHandler {
	return &UserRegisteredHandler{
		identityClient: identityClient,
		logger:         logger,
	}
}

func (h *UserRegisteredHandler) EventType() string {
	return EventTypeUserRegistered
}

func (h *UserRegisteredHandler) Handle(ctx context.Context, event *outbox.Event) error {
	var payload UserRegisteredPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	h.logger.InfoContext(ctx, "processing user.registered event",
		slog.String("event_id", event.ID.String()),
		slog.String("user_id", payload.UserID),
		slog.String("username", payload.Username),
	)

	err := h.identityClient.CreateUser(ctx, payload.UserID, payload.Username, payload.FirstName, payload.LastName, payload.Timezone)
	if err != nil {
		return fmt.Errorf("failed to create user in identity service: %w", err)
	}

	h.logger.InfoContext(ctx, "user created in identity service",
		slog.String("user_id", payload.UserID),
		slog.String("username", payload.Username),
	)

	return nil
}
