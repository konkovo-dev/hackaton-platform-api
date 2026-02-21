package participationservice

import (
	"context"
	"log/slog"

	commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"
	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/participation"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (a *API) ListHackathonParticipants(ctx context.Context, req *participationrolesv1.ListHackathonParticipantsRequest) (*participationrolesv1.ListHackathonParticipantsResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	var statuses []string
	if req.StatusFilter != nil && len(req.StatusFilter.Statuses) > 0 {
		for _, protoStatus := range req.StatusFilter.Statuses {
			domainStatus := mapProtoStatusToDomain(protoStatus)
			if domainStatus != "" {
				statuses = append(statuses, domainStatus)
			}
		}
	}

	var wishedRoleIDs []uuid.UUID
	if len(req.WishedRoleIdsFilter) > 0 {
		for _, roleIDStr := range req.WishedRoleIdsFilter {
			roleID, err := uuid.Parse(roleIDStr)
			if err != nil {
				continue
			}
			wishedRoleIDs = append(wishedRoleIDs, roleID)
		}
	}

	var pageSize uint32 = 20
	var offset int32 = 0

	if req.Query != nil && req.Query.Page != nil {
		if req.Query.Page.PageSize > 0 {
			pageSize = req.Query.Page.PageSize
		}
	}

	result, err := a.participationService.List(ctx, participation.ListIn{
		HackathonID:   hackathonID,
		Statuses:      statuses,
		WishedRoleIDs: wishedRoleIDs,
		Limit:         int32(pageSize),
		Offset:        offset,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "list_hackathon_participants")
	}

	protoParticipants := make([]*participationrolesv1.HackathonParticipation, 0, len(result.Participants))
	for _, p := range result.Participants {
		protoWishedRoles := make([]*participationrolesv1.TeamRole, 0, len(p.WishedRoles))
		for _, role := range p.WishedRoles {
			protoWishedRoles = append(protoWishedRoles, &participationrolesv1.TeamRole{
				Id:   role.ID.String(),
				Name: role.Name,
			})
		}

		participation := &participationrolesv1.HackathonParticipation{
			HackathonId: p.HackathonID.String(),
			UserId:      p.UserID.String(),
			Status:      mapDomainStatusToProto(p.Status),
			TeamId:      "",
			Profile: &participationrolesv1.ParticipationProfile{
				WishedRoles:    protoWishedRoles,
				MotivationText: p.MotivationText,
			},
			RegisteredAt: timestamppb.New(p.RegisteredAt),
			UpdatedAt:    timestamppb.New(p.UpdatedAt),
		}

		if p.TeamID != nil {
			participation.TeamId = p.TeamID.String()
		}

		protoParticipants = append(protoParticipants, participation)
	}

	resp := &participationrolesv1.ListHackathonParticipantsResponse{
		Participants: protoParticipants,
		Page: &commonv1.PageResponse{
			NextPageToken: "",
		},
	}

	a.logger.InfoContext(ctx, "list_hackathon_participants: success",
		slog.String("hackathon_id", req.HackathonId),
		slog.Int("count", len(protoParticipants)),
	)

	return resp, nil
}
