package hackathonservice

import (
	"context"
	"errors"
	"log/slog"

	hackathonv1 "github.com/belikoooova/hackaton-platform-api/api/hackathon/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/usecase/hackathon"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *HackathonService) PublishHackathon(ctx context.Context, req *hackathonv1.PublishHackathonRequest) (*hackathonv1.PublishHackathonResponse, error) {
	var idempotencyKey string
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, "publish")
		resp := &hackathonv1.PublishHackathonResponse{}
		found, err := s.idempotency.CheckAndGet(ctx, idempotencyKey, "publish_hackathon", requestHash, resp)
		if err != nil {
			var conflictErr *idempotency.ConflictError
			if errors.As(err, &conflictErr) {
				s.logger.WarnContext(ctx, "idempotency key conflict", slog.String("key", idempotencyKey))
				return nil, status.Error(codes.AlreadyExists, "idempotency key already used with different request")
			}
			s.logger.ErrorContext(ctx, "failed to check idempotency", slog.String("error", err.Error()))
			return nil, status.Error(codes.Internal, "failed to check idempotency")
		}
		if found {
			s.logger.InfoContext(ctx, "returning cached response", slog.String("key", idempotencyKey))
			return resp, nil
		}
	}

	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	_, err = s.hackathonService.PublishHackathon(ctx, hackathon.PublishHackathonIn{
		HackathonID: hackathonID,
	})
	if err != nil {
		return nil, s.handleError(ctx, err, "publish_hackathon")
	}

	resp := &hackathonv1.PublishHackathonResponse{}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, "publish")
		if err := s.idempotency.Save(ctx, idempotencyKey, "publish_hackathon", requestHash, resp); err != nil {
			s.logger.ErrorContext(ctx, "failed to save idempotency key", slog.String("error", err.Error()))
		}
	}

	s.logger.InfoContext(ctx, "hackathon published", slog.String("hackathon_id", req.HackathonId))
	return resp, nil
}
