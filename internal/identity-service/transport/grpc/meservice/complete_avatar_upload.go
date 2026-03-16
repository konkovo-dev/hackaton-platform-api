package meservice

import (
	"context"
	"errors"
	"log/slog"

	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/me"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *MeService) CompleteAvatarUpload(ctx context.Context, req *identityv1.CompleteAvatarUploadRequest) (*identityv1.CompleteAvatarUploadResponse, error) {
	uploadID, err := uuid.Parse(req.UploadId)
	if err != nil {
		return nil, s.handleError(ctx, me.ErrInvalidInput, "complete_avatar_upload")
	}

	var idempotencyKey string
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.UploadId)
		resp := &identityv1.CompleteAvatarUploadResponse{}
		found, err := s.idempotency.CheckAndGet(ctx, idempotencyKey, "complete_avatar_upload", requestHash, resp)
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

	result, err := s.meService.CompleteAvatarUpload(ctx, me.CompleteAvatarUploadIn{
		UploadID: uploadID,
	})

	if err != nil {
		return nil, s.handleError(ctx, err, "complete_avatar_upload")
	}

	resp := &identityv1.CompleteAvatarUploadResponse{
		AvatarUrl: result.AvatarURL,
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.UploadId)
		if err := s.idempotency.Save(ctx, idempotencyKey, "complete_avatar_upload", requestHash, resp); err != nil {
			s.logger.ErrorContext(ctx, "failed to save idempotency key", slog.String("error", err.Error()))
		}
	}

	return resp, nil
}
