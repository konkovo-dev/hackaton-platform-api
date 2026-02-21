package participationservice

import (
	"context"
	"log/slog"

	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/participation"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (a *API) GetMyParticipation(ctx context.Context, req *participationrolesv1.GetMyParticipationRequest) (*participationrolesv1.GetMyParticipationResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	result, err := a.participationService.GetMy(ctx, participation.GetMyIn{
		HackathonID: hackathonID,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "get_my_participation")
	}

	protoWishedRoles := make([]*participationrolesv1.TeamRole, 0, len(result.Participation.WishedRoles))
	for _, role := range result.Participation.WishedRoles {
		protoWishedRoles = append(protoWishedRoles, &participationrolesv1.TeamRole{
			Id:   role.ID.String(),
			Name: role.Name,
		})
	}

	resp := &participationrolesv1.GetMyParticipationResponse{
		Participation: &participationrolesv1.HackathonParticipation{
			HackathonId: result.Participation.HackathonID.String(),
			UserId:      result.Participation.UserID.String(),
			Status:      mapDomainStatusToProto(result.Participation.Status),
			TeamId:      "",
			Profile: &participationrolesv1.ParticipationProfile{
				WishedRoles:    protoWishedRoles,
				MotivationText: result.Participation.MotivationText,
			},
			RegisteredAt: timestamppb.New(result.Participation.RegisteredAt),
			UpdatedAt:    timestamppb.New(result.Participation.UpdatedAt),
		},
	}

	if result.Participation.TeamID != nil {
		resp.Participation.TeamId = result.Participation.TeamID.String()
	}

	a.logger.InfoContext(ctx, "get_my_participation: success",
		slog.String("hackathon_id", req.HackathonId),
	)

	return resp, nil
}
