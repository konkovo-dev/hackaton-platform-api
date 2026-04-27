package me

import (
	"context"
	"errors"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	pkgerrors "github.com/belikoooova/hackaton-platform-api/pkg/errors"
	"github.com/google/uuid"
)

type CompleteAvatarUploadIn struct {
	UploadID uuid.UUID
}

type CompleteAvatarUploadOut struct {
	AvatarURL string
}

func (s *Service) CompleteAvatarUpload(ctx context.Context, in CompleteAvatarUploadIn) (*CompleteAvatarUploadOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	// Get upload record
	upload, err := s.avatarUploadRepo.GetAvatarUploadByID(ctx, in.UploadID)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			return nil, fmt.Errorf("%w: avatar upload not found", ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to get avatar upload: %w", err)
	}

	// Verify ownership
	if upload.UserID != userUUID {
		return nil, ErrForbidden
	}

	// Check if already completed
	if upload.Status == "completed" {
		// Generate public URL
		avatarURL := fmt.Sprintf("%s://%s/%s/%s", s.s3Client.Config().Scheme(), s.s3Client.Config().PublicEndpoint, s.s3Client.Config().BucketName, upload.StorageKey)
		return &CompleteAvatarUploadOut{
			AvatarURL: avatarURL,
		}, nil
	}

	// Verify file exists in S3
	size, exists, err := s.s3Client.HeadObject(ctx, s.s3Client.Config().BucketName, upload.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to verify file in S3: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("%w: file not found in S3", ErrInvalidInput)
	}

	// Verify size matches
	if size != upload.SizeBytes {
		return nil, fmt.Errorf("%w: file size mismatch (expected %d, got %d)", ErrInvalidInput, upload.SizeBytes, size)
	}

	// Mark upload as completed
	err = s.avatarUploadRepo.CompleteAvatarUpload(ctx, in.UploadID)
	if err != nil {
		return nil, fmt.Errorf("failed to complete avatar upload: %w", err)
	}

	// Generate public URL
	avatarURL := fmt.Sprintf("%s://%s/%s/%s", s.s3Client.Config().Scheme(), s.s3Client.Config().PublicEndpoint, s.s3Client.Config().BucketName, upload.StorageKey)

	// Update user's avatar_url
	err = s.userRepo.UpdateAvatarURL(ctx, userUUID, avatarURL)
	if err != nil {
		return nil, fmt.Errorf("failed to update user avatar URL: %w", err)
	}

	return &CompleteAvatarUploadOut{
		AvatarURL: avatarURL,
	}, nil
}
