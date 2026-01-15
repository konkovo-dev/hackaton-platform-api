package meservice

import (
	"context"
	"errors"
	"log/slog"

	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/transport/grpc/mappers"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/me"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *MeService) CreateMe(ctx context.Context, req *identityv1.CreateMeRequest) (*identityv1.CreateMeResponse, error) {
	var idempotencyKey string
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.UserId, req.Username)
		resp := &identityv1.CreateMeResponse{}
		found, err := s.idempotency.CheckAndGet(ctx, idempotencyKey, "create_me", requestHash, resp)
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

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		s.logger.WarnContext(ctx, "create_me: invalid user_id", slog.String("user_id", req.UserId))
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	result, err := s.meService.CreateMe(ctx, me.CreateMeIn{
		UserID:    userID,
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Timezone:  req.Timezone,
	})

	if err != nil {
		return nil, s.handleError(ctx, err, "create_me")
	}

	resp := mappers.CreateMeOutToResponse(result)

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.UserId, req.Username)
		if err := s.idempotency.Save(ctx, idempotencyKey, "create_me", requestHash, resp); err != nil {
			s.logger.ErrorContext(ctx, "failed to save idempotency key", slog.String("error", err.Error()))
		}
	}

	s.logger.InfoContext(ctx, "user created", slog.String("user_id", result.User.ID.String()))
	return resp, nil
}
