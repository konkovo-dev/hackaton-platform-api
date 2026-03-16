package postgres

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/me"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
)

type AvatarUploadRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewAvatarUploadRepository(db queries.DBTX) *AvatarUploadRepository {
	return &AvatarUploadRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *AvatarUploadRepository) CreateAvatarUpload(ctx context.Context, uploadID, userID uuid.UUID, filename string, sizeBytes int64, contentType, storageKey string) error {
	_, err := r.Queries().CreateAvatarUpload(ctx, queries.CreateAvatarUploadParams{
		UploadID:    pgxutil.UUIDToPgtype(uploadID),
		UserID:      pgxutil.UUIDToPgtype(userID),
		Filename:    filename,
		SizeBytes:   sizeBytes,
		ContentType: contentType,
		StorageKey:  storageKey,
	})

	if err != nil {
		return fmt.Errorf("failed to create avatar upload: %w", pgxutil.MapDBError(err))
	}

	return nil
}

func (r *AvatarUploadRepository) GetAvatarUploadByID(ctx context.Context, uploadID uuid.UUID) (*me.AvatarUpload, error) {
	row, err := r.Queries().GetAvatarUploadByID(ctx, pgxutil.UUIDToPgtype(uploadID))
	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return nil, fmt.Errorf("avatar upload not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get avatar upload: %w", err)
	}

	return &me.AvatarUpload{
		UploadID:    pgxutil.PgtypeToUUID(row.UploadID),
		UserID:      pgxutil.PgtypeToUUID(row.UserID),
		Filename:    row.Filename,
		SizeBytes:   row.SizeBytes,
		ContentType: row.ContentType,
		StorageKey:  row.StorageKey,
		Status:      row.Status,
	}, nil
}

func (r *AvatarUploadRepository) CompleteAvatarUpload(ctx context.Context, uploadID uuid.UUID) error {
	_, err := r.Queries().CompleteAvatarUpload(ctx, pgxutil.UUIDToPgtype(uploadID))
	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return fmt.Errorf("avatar upload not found: %w", err)
		}
		return fmt.Errorf("failed to complete avatar upload: %w", err)
	}

	return nil
}
