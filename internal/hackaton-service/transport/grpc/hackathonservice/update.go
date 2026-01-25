package hackathonservice

import (
	"context"
	"errors"
	"log/slog"

	hackathonv1 "github.com/belikoooova/hackaton-platform-api/api/hackathon/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/transport/grpc/mappers"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *HackathonService) UpdateHackathon(ctx context.Context, req *hackathonv1.UpdateHackathonRequest) (*hackathonv1.UpdateHackathonResponse, error) {
	var idempotencyKey string
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.Name, req.ShortDescription)
		resp := &hackathonv1.UpdateHackathonResponse{}
		found, err := s.idempotency.CheckAndGet(ctx, idempotencyKey, "update_hackathon", requestHash, resp)
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

	in := mappers.ProtoToUpdateHackathonIn(req)

	result, err := s.hackathonService.UpdateHackathon(ctx, in)
	if err != nil {
		return nil, s.handleError(ctx, err, "update_hackathon")
	}

	validationErrors := make([]*hackathonv1.ValidationError, 0, len(result.ValidationErrors))
	for _, ve := range result.ValidationErrors {
		validationErrors = append(validationErrors, &hackathonv1.ValidationError{
			Code:    ve.Code,
			Field:   ve.Field,
			Message: ve.Message,
			Meta:    ve.Meta,
		})
	}

	resp := &hackathonv1.UpdateHackathonResponse{
		ValidationErrors: validationErrors,
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.Name, req.ShortDescription)
		if err := s.idempotency.Save(ctx, idempotencyKey, "update_hackathon", requestHash, resp); err != nil {
			s.logger.ErrorContext(ctx, "failed to save idempotency key", slog.String("error", err.Error()))
		}
	}

	s.logger.InfoContext(ctx, "hackathon updated", slog.String("hackathon_id", req.HackathonId))
	return resp, nil
}
