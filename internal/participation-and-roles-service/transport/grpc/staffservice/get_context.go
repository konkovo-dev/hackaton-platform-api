package staffservice

import (
	"context"
	"log/slog"

	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/role"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *API) GetHackathonContext(ctx context.Context, req *participationrolesv1.GetHackathonContextRequest) (*participationrolesv1.GetHackathonContextResponse, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		a.logger.WarnContext(ctx, "get_hackathon_context: no user_id in context")
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		a.logger.WarnContext(ctx, "get_hackathon_context: invalid user_id", slog.String("user_id", userID))
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	result, err := a.roleService.GetHackathonContext(ctx, role.GetHackathonContextIn{
		UserID:      userUUID,
		HackathonID: hackathonID,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "get_hackathon_context")
	}

	protoRoles := make([]participationrolesv1.HackathonRole, 0, len(result.Roles))
	for _, roleStr := range result.Roles {
		protoRole := domain.MapDomainRoleToProto(domain.HackathonRole(roleStr))
		if protoRole != participationrolesv1.HackathonRole_HACKATHON_ROLE_UNSPECIFIED {
			protoRoles = append(protoRoles, protoRole)
		}
	}

	protoParticipationStatus := domain.MapDomainParticipationToProto(domain.ParticipationStatus(result.ParticipationStatus))

	a.logger.InfoContext(ctx, "get_hackathon_context: success",
		slog.String("hackathon_id", req.HackathonId),
		slog.String("user_id", userID),
		slog.Int("roles_count", len(protoRoles)),
	)

	return &participationrolesv1.GetHackathonContextResponse{
		UserId:              result.UserID.String(),
		ParticipationStatus: protoParticipationStatus,
		Roles:               protoRoles,
	}, nil
}
