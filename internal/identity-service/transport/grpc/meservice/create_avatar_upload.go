package meservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/me"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *MeService) CreateAvatarUpload(ctx context.Context, req *identityv1.CreateAvatarUploadRequest) (*identityv1.CreateAvatarUploadResponse, error) {
	var idempotencyKey string
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.Filename, fmt.Sprint(req.SizeBytes), req.ContentType)
		resp := &identityv1.CreateAvatarUploadResponse{}
		found, err := s.idempotency.CheckAndGet(ctx, idempotencyKey, "create_avatar_upload", requestHash, resp)
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

	result, err := s.meService.CreateAvatarUpload(ctx, me.CreateAvatarUploadIn{
		Filename:    req.Filename,
		SizeBytes:   req.SizeBytes,
		ContentType: req.ContentType,
	})

	if err != nil {
		return nil, s.handleError(ctx, err, "create_avatar_upload")
	}

	resp := &identityv1.CreateAvatarUploadResponse{
		UploadId:  result.UploadID.String(),
		UploadUrl: result.UploadURL,
		ExpiresAt: timestamppb.New(result.ExpiresAt),
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.Filename, fmt.Sprint(req.SizeBytes), req.ContentType)
		if err := s.idempotency.Save(ctx, idempotencyKey, "create_avatar_upload", requestHash, resp); err != nil {
			s.logger.ErrorContext(ctx, "failed to save idempotency key", slog.String("error", err.Error()))
		}
	}

	return resp, nil
}
