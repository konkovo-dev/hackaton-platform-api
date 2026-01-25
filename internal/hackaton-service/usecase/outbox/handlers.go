package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/client/participationroles"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
)

const (
	EventTypeOwnerAssigned = "hackathon.owner_assigned"
)

type OwnerAssignedPayload struct {
	HackathonID string `json:"hackathon_id"`
	UserID      string `json:"user_id"`
	Role        string `json:"role"`
}

type OwnerAssignedHandler struct {
	participationRolesClient *participationroles.Client
	logger                   *slog.Logger
}

func NewOwnerAssignedHandler(participationRolesClient *participationroles.Client, logger *slog.Logger) *OwnerAssignedHandler {
	return &OwnerAssignedHandler{
		participationRolesClient: participationRolesClient,
		logger:                   logger,
	}
}

func (h *OwnerAssignedHandler) EventType() string {
	return EventTypeOwnerAssigned
}

func (h *OwnerAssignedHandler) Handle(ctx context.Context, event *outbox.Event) error {
	var payload OwnerAssignedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	h.logger.InfoContext(ctx, "processing hackathon.owner_assigned event",
		slog.String("event_id", event.ID.String()),
		slog.String("hackathon_id", payload.HackathonID),
		slog.String("user_id", payload.UserID),
		slog.String("role", payload.Role),
	)

	protoRole := mapDomainRoleToProto(domain.HackathonRole(payload.Role))

	err := h.participationRolesClient.AssignHackathonRole(
		ctx,
		payload.HackathonID,
		payload.UserID,
		protoRole,
	)
	if err != nil {
		return fmt.Errorf("failed to assign hackathon role: %w", err)
	}

	h.logger.InfoContext(ctx, "hackathon owner role assigned",
		slog.String("hackathon_id", payload.HackathonID),
		slog.String("user_id", payload.UserID),
	)

	return nil
}

func mapDomainRoleToProto(role domain.HackathonRole) participationrolesv1.HackathonRole {
	switch role {
	case domain.RoleOwner:
		return participationrolesv1.HackathonRole_HX_ROLE_OWNER
	case domain.RoleOrganizer:
		return participationrolesv1.HackathonRole_HX_ROLE_ORGANIZER
	case domain.RoleMentor:
		return participationrolesv1.HackathonRole_HX_ROLE_MENTOR
	case domain.RoleJury:
		return participationrolesv1.HackathonRole_HX_ROLE_JUDGE
	default:
		return participationrolesv1.HackathonRole_HACKATHON_ROLE_UNSPECIFIED
	}
}
