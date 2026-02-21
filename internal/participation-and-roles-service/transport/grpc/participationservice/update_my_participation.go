package participationservice

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/participation"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (a *API) UpdateMyParticipation(ctx context.Context, req *participationrolesv1.UpdateMyParticipationRequest) (*participationrolesv1.UpdateMyParticipationResponse, error) {
	var idempotencyKey string
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, strings.Join(req.WishedRoleIds, ","), req.MotivationText)
		resp := &participationrolesv1.UpdateMyParticipationResponse{}
		found, err := a.idempotencyHelper.CheckAndGet(ctx, idempotencyKey, "update_my_participation", requestHash, resp)
		if err != nil {
			var conflictErr *idempotency.ConflictError
			if errors.As(err, &conflictErr) {
				a.logger.WarnContext(ctx, "idempotency key conflict", slog.String("key", idempotencyKey))
				return nil, status.Error(codes.AlreadyExists, "idempotency key already used with different request")
			}
			a.logger.ErrorContext(ctx, "failed to check idempotency", slog.String("error", err.Error()))
			return nil, status.Error(codes.Internal, "failed to check idempotency")
		}
		if found {
			a.logger.InfoContext(ctx, "returning cached response", slog.String("key", idempotencyKey))
			return resp, nil
		}
	}

	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	wishedRoleIDs := make([]uuid.UUID, 0, len(req.WishedRoleIds))
	for _, roleIDStr := range req.WishedRoleIds {
		roleID, err := uuid.Parse(roleIDStr)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid wished_role_id")
		}
		wishedRoleIDs = append(wishedRoleIDs, roleID)
	}

	result, err := a.participationService.UpdateMy(ctx, participation.UpdateMyIn{
		HackathonID:    hackathonID,
		WishedRoleIDs:  wishedRoleIDs,
		MotivationText: req.MotivationText,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "update_my_participation")
	}

	protoWishedRoles := make([]*participationrolesv1.TeamRole, 0, len(result.Participation.WishedRoles))
	for _, role := range result.Participation.WishedRoles {
		protoWishedRoles = append(protoWishedRoles, &participationrolesv1.TeamRole{
			Id:   role.ID.String(),
			Name: role.Name,
		})
	}

	resp := &participationrolesv1.UpdateMyParticipationResponse{
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

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, strings.Join(req.WishedRoleIds, ","), req.MotivationText)
		if err := a.idempotencyHelper.Save(ctx, idempotencyKey, "update_my_participation", requestHash, resp); err != nil {
			a.logger.ErrorContext(ctx, "failed to save idempotency key", slog.String("error", err.Error()))
		}
	}

	a.logger.InfoContext(ctx, "update_my_participation: success",
		slog.String("hackathon_id", req.HackathonId),
	)

	return resp, nil
}
