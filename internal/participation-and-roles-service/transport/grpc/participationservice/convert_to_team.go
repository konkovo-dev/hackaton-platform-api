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

func (a *API) ConvertToTeamParticipation(ctx context.Context, req *participationrolesv1.ConvertToTeamParticipationRequest) (*participationrolesv1.ConvertToTeamParticipationResponse, error) {
	var idempotencyKey string
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	if idempotencyKey != "" {
		isCaptainStr := "false"
		if req.IsCaptain {
			isCaptainStr = "true"
		}
		requestHash := idempotency.ComputeHash(req.HackathonId, req.UserId, req.TeamId, isCaptainStr)
		resp := &participationrolesv1.ConvertToTeamParticipationResponse{}
		found, err := a.idempotencyHelper.CheckAndGet(ctx, idempotencyKey, "convert_to_team_participation", requestHash, resp)
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

	teamID, err := uuid.Parse(req.TeamId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid team_id")
	}

	_, err = a.participationService.ConvertToTeam(ctx, participation.ConvertToTeamIn{
		HackathonID: hackathonID,
		UserID:      userID,
		TeamID:      teamID,
		IsCaptain:   req.IsCaptain,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "convert_to_team_participation")
	}

	resp := &participationrolesv1.ConvertToTeamParticipationResponse{
		Participation: &participationrolesv1.HackathonParticipation{
			HackathonId: hackathonID.String(),
			UserId:      userID.String(),
			Status:      participationrolesv1.ParticipationStatus_PART_TEAM_MEMBER,
			TeamId:      teamID.String(),
		},
	}

	if req.IsCaptain {
		resp.Participation.Status = participationrolesv1.ParticipationStatus_PART_TEAM_CAPTAIN
	}

	if idempotencyKey != "" {
		isCaptainStr := "false"
		if req.IsCaptain {
			isCaptainStr = "true"
		}
		requestHash := idempotency.ComputeHash(req.HackathonId, req.UserId, req.TeamId, isCaptainStr)
		if err := a.idempotencyHelper.Save(ctx, idempotencyKey, "convert_to_team_participation", requestHash, resp); err != nil {
			a.logger.ErrorContext(ctx, "failed to save idempotency key", slog.String("error", err.Error()))
		}
	}

	a.logger.InfoContext(ctx, "convert_to_team_participation: success",
		slog.String("hackathon_id", req.HackathonId),
		slog.String("user_id", req.UserId),
		slog.String("team_id", req.TeamId),
		slog.Bool("is_captain", req.IsCaptain),
	)

	return resp, nil
}
