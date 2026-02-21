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

func (a *API) UnregisterFromHackathon(ctx context.Context, req *participationrolesv1.UnregisterFromHackathonRequest) (*participationrolesv1.UnregisterFromHackathonResponse, error) {
	var idempotencyKey string
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId)
		resp := &participationrolesv1.UnregisterFromHackathonResponse{}
		found, err := a.idempotencyHelper.CheckAndGet(ctx, idempotencyKey, "unregister_from_hackathon", requestHash, resp)
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

	_, err = a.participationService.Unregister(ctx, participation.UnregisterIn{
		HackathonID: hackathonID,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "unregister_from_hackathon")
	}

	resp := &participationrolesv1.UnregisterFromHackathonResponse{}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId)
		if err := a.idempotencyHelper.Save(ctx, idempotencyKey, "unregister_from_hackathon", requestHash, resp); err != nil {
			a.logger.ErrorContext(ctx, "failed to save idempotency key", slog.String("error", err.Error()))
		}
	}

	a.logger.InfoContext(ctx, "unregister_from_hackathon: success",
		slog.String("hackathon_id", req.HackathonId),
	)

	return resp, nil
}
