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

func (a *API) GetMyHackathonIDs(ctx context.Context, req *participationrolesv1.GetMyHackathonIDsRequest) (*participationrolesv1.GetMyHackathonIDsResponse, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		a.logger.WarnContext(ctx, "get_my_hackathon_ids: no user_id in context")
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		a.logger.WarnContext(ctx, "get_my_hackathon_ids: invalid user_id", slog.String("user_id", userID))
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	in := role.GetMyHackathonIDsIn{
		UserID: userUUID,
	}

	if req.RoleFilter != nil {
		domainRole := domain.MapProtoRoleToDomain(*req.RoleFilter)
		if domainRole == "" {
			return nil, status.Error(codes.InvalidArgument, "invalid role_filter")
		}
		roleStr := string(domainRole)
		in.RoleFilter = &roleStr
	}

	if req.ParticipationFilter != nil {
		in.ParticipationFilter = req.ParticipationFilter
	}

	if req.ParticipationStatusFilter != nil {
		domainStatus := domain.MapProtoParticipationToDomain(*req.ParticipationStatusFilter)
		if domainStatus == "" {
			return nil, status.Error(codes.InvalidArgument, "invalid participation_status_filter")
		}
		statusStr := string(domainStatus)
		in.ParticipationStatusFilter = &statusStr
	}

	result, err := a.roleService.GetMyHackathonIDs(ctx, in)
	if err != nil {
		return nil, a.handleError(ctx, err, "get_my_hackathon_ids")
	}

	hackathonIDs := make([]string, 0, len(result.HackathonIDs))
	for _, id := range result.HackathonIDs {
		hackathonIDs = append(hackathonIDs, id.String())
	}

	a.logger.InfoContext(ctx, "get_my_hackathon_ids: success",
		slog.String("user_id", userID),
		slog.Int("hackathon_count", len(hackathonIDs)),
	)

	return &participationrolesv1.GetMyHackathonIDsResponse{
		HackathonIds: hackathonIDs,
	}, nil
}
