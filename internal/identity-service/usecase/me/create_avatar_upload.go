package me

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type CreateAvatarUploadIn struct {
	Filename    string
	SizeBytes   int64
	ContentType string
}

type CreateAvatarUploadOut struct {
	UploadID  uuid.UUID
	UploadURL string
	ExpiresAt time.Time
}

const (
	maxAvatarSizeBytes = 5 * 1024 * 1024 // 5 MB
	avatarBucketPrefix = "avatars"
	uploadExpiryDuration = 15 * time.Minute
)

var allowedImageExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
var allowedImageContentTypes = []string{"image/jpeg", "image/png", "image/gif", "image/webp"}

func (s *Service) CreateAvatarUpload(ctx context.Context, in CreateAvatarUploadIn) (*CreateAvatarUploadOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	// Validate filename
	if in.Filename == "" {
		return nil, fmt.Errorf("%w: filename is required", ErrInvalidInput)
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(in.Filename))
	if !contains(allowedImageExtensions, ext) {
		return nil, fmt.Errorf("%w: only image files are allowed (%s)", ErrInvalidInput, strings.Join(allowedImageExtensions, ", "))
	}

	// Validate content type
	if !contains(allowedImageContentTypes, in.ContentType) {
		return nil, fmt.Errorf("%w: only image content types are allowed (%s)", ErrInvalidInput, strings.Join(allowedImageContentTypes, ", "))
	}

	// Validate size
	if in.SizeBytes <= 0 {
		return nil, fmt.Errorf("%w: size_bytes must be positive", ErrInvalidInput)
	}

	if in.SizeBytes > maxAvatarSizeBytes {
		return nil, fmt.Errorf("%w: max avatar size is %d bytes (5 MB)", ErrInvalidInput, maxAvatarSizeBytes)
	}

	// Check if user exists
	_, err = s.userRepo.GetByID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	uploadID := uuid.New()
	
	// Generate storage key: avatars/{user_id}/{upload_id}/{filename}
	storageKey := fmt.Sprintf("%s/%s/%s/%s", avatarBucketPrefix, userUUID.String(), uploadID.String(), in.Filename)

	// Generate presigned PUT URL
	uploadURL, err := s.s3Client.GeneratePresignedPutURL(ctx, s.s3Client.Config().BucketName, storageKey, uploadExpiryDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	// Save upload record
	err = s.avatarUploadRepo.CreateAvatarUpload(ctx, uploadID, userUUID, in.Filename, in.SizeBytes, in.ContentType, storageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create avatar upload record: %w", err)
	}

	expiresAt := time.Now().Add(uploadExpiryDuration)

	return &CreateAvatarUploadOut{
		UploadID:  uploadID,
		UploadURL: uploadURL,
		ExpiresAt: expiresAt,
	}, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
