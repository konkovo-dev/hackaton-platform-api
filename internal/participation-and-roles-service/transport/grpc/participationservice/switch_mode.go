package participationservice

import (
	"context"
	"errors"
	"log/slog"

	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/participation"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (a *API) SwitchParticipationMode(ctx context.Context, req *participationrolesv1.SwitchParticipationModeRequest) (*participationrolesv1.SwitchParticipationModeResponse, error) {
	var idempotencyKey string
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, req.NewStatus.String())
		resp := &participationrolesv1.SwitchParticipationModeResponse{}
		found, err := a.idempotencyHelper.CheckAndGet(ctx, idempotencyKey, "switch_participation_mode", requestHash, resp)
		if err != nil {
			var conflictErr *idempotency.ConflictError
			if errors.As(err, &conflictErr) {
				a.logger.WarnContext(ctx, "idempotency key conflict", slog.String("key", idempotencyKey))
				return nil, status.Error(codes.FailedPrecondition, "idempotency key conflict")
			}
			a.logger.ErrorContext(ctx, "failed to check idempotency key", slog.String("error", err.Error()))
		}
		if found {
			return resp, nil
		}
	}

	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	statusStr := ""
	switch req.NewStatus {
	case participationrolesv1.ParticipationStatus_PART_INDIVIDUAL:
		statusStr = string(domain.ParticipationIndividual)
	case participationrolesv1.ParticipationStatus_PART_LOOKING_FOR_TEAM:
		statusStr = string(domain.ParticipationLookingForTeam)
	default:
		return nil, status.Error(codes.InvalidArgument, "new_status must be INDIVIDUAL_ACTIVE or LOOKING_FOR_TEAM")
	}

	out, err := a.participationService.SwitchMode(ctx, participation.SwitchModeIn{
		HackathonID: hackathonID,
		NewStatus:   statusStr,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "switch_participation_mode")
	}

	protoWishedRoles := make([]*participationrolesv1.TeamRole, 0, len(out.Participation.WishedRoles))
	for _, role := range out.Participation.WishedRoles {
		protoWishedRoles = append(protoWishedRoles, &participationrolesv1.TeamRole{
			Id:   role.ID.String(),
			Name: role.Name,
		})
	}

	resp := &participationrolesv1.SwitchParticipationModeResponse{
		Participation: &participationrolesv1.HackathonParticipation{
			HackathonId: out.Participation.HackathonID.String(),
			UserId:      out.Participation.UserID.String(),
			Status:      mapDomainStatusToProto(out.Participation.Status),
			TeamId:      "",
			Profile: &participationrolesv1.ParticipationProfile{
				WishedRoles:    protoWishedRoles,
				MotivationText: out.Participation.MotivationText,
			},
			RegisteredAt: timestamppb.New(out.Participation.RegisteredAt),
			UpdatedAt:    timestamppb.New(out.Participation.UpdatedAt),
		},
	}

	if out.Participation.TeamID != nil {
		resp.Participation.TeamId = out.Participation.TeamID.String()
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, req.NewStatus.String())
		if err := a.idempotencyHelper.Save(ctx, idempotencyKey, "switch_participation_mode", requestHash, resp); err != nil {
			a.logger.ErrorContext(ctx, "failed to save idempotency key", slog.String("error", err.Error()))
		}
	}

	a.logger.InfoContext(ctx, "switch_participation_mode: success",
		slog.String("hackathon_id", req.HackathonId),
		slog.String("new_status", statusStr),
	)

	return resp, nil
}
