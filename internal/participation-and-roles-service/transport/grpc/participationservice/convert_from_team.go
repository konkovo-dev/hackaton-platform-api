package participationservice

import (
	"context"
	"errors"
	"log/slog"

	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/participation"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *API) ConvertFromTeamParticipation(ctx context.Context, req *participationrolesv1.ConvertFromTeamParticipationRequest) (*participationrolesv1.ConvertFromTeamParticipationResponse, error) {
	var idempotencyKey string
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, req.UserId)
		resp := &participationrolesv1.ConvertFromTeamParticipationResponse{}
		found, err := a.idempotencyHelper.CheckAndGet(ctx, idempotencyKey, "convert_from_team_participation", requestHash, resp)
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

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	_, err = a.participationService.ConvertFromTeam(ctx, participation.ConvertFromTeamIn{
		HackathonID: hackathonID,
		UserID:      userID,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "convert_from_team_participation")
	}

	resp := &participationrolesv1.ConvertFromTeamParticipationResponse{
		Participation: &participationrolesv1.HackathonParticipation{
			HackathonId: hackathonID.String(),
			UserId:      userID.String(),
			Status:      participationrolesv1.ParticipationStatus_PART_LOOKING_FOR_TEAM,
			TeamId:      "",
		},
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, req.UserId)
		if err := a.idempotencyHelper.Save(ctx, idempotencyKey, "convert_from_team_participation", requestHash, resp); err != nil {
			a.logger.ErrorContext(ctx, "failed to save idempotency key", slog.String("error", err.Error()))
		}
	}

	a.logger.InfoContext(ctx, "convert_from_team_participation: success",
		slog.String("hackathon_id", req.HackathonId),
		slog.String("user_id", req.UserId),
	)

	return resp, nil
}
