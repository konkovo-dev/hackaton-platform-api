package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type SubmissionFileRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewSubmissionFileRepository(db queries.DBTX) *SubmissionFileRepository {
	return &SubmissionFileRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *SubmissionFileRepository) Create(ctx context.Context, file *entity.SubmissionFile) error {
	row, err := r.Queries().CreateSubmissionFile(ctx, queries.CreateSubmissionFileParams{
		ID:           file.ID,
		SubmissionID: file.SubmissionID,
		Filename:     file.Filename,
		SizeBytes:    file.SizeBytes,
		ContentType:  file.ContentType,
		StorageKey:   file.StorageKey,
		UploadStatus: file.UploadStatus,
	})
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	file.CreatedAt = row.CreatedAt
	file.CompletedAt = toTimePtr(row.CompletedAt)

	return nil
}

func (r *SubmissionFileRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.SubmissionFile, error) {
	row, err := r.Queries().GetFileByID(ctx, id)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get file by id: %w", err)
	}

	return &entity.SubmissionFile{
		ID:           row.ID,
		SubmissionID: row.SubmissionID,
		Filename:     row.Filename,
		SizeBytes:    row.SizeBytes,
		ContentType:  row.ContentType,
		StorageKey:   row.StorageKey,
		UploadStatus: row.UploadStatus,
		CreatedAt:    row.CreatedAt,
		CompletedAt:  toTimePtr(row.CompletedAt),
	}, nil
}

func (r *SubmissionFileRepository) GetByIDAndSubmissionID(ctx context.Context, id, submissionID uuid.UUID) (*entity.SubmissionFile, error) {
	row, err := r.Queries().GetFileByIDAndSubmissionID(ctx, queries.GetFileByIDAndSubmissionIDParams{
		ID:           id,
		SubmissionID: submissionID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return &entity.SubmissionFile{
		ID:           row.ID,
		SubmissionID: row.SubmissionID,
		Filename:     row.Filename,
		SizeBytes:    row.SizeBytes,
		ContentType:  row.ContentType,
		StorageKey:   row.StorageKey,
		UploadStatus: row.UploadStatus,
		CreatedAt:    row.CreatedAt,
		CompletedAt:  toTimePtr(row.CompletedAt),
	}, nil
}

func (r *SubmissionFileRepository) ListBySubmission(ctx context.Context, submissionID uuid.UUID) ([]*entity.SubmissionFile, error) {
	rows, err := r.Queries().ListFilesBySubmission(ctx, submissionID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to list files by submission: %w", err)
	}

	files := make([]*entity.SubmissionFile, 0, len(rows))
	for _, row := range rows {
		files = append(files, &entity.SubmissionFile{
			ID:           row.ID,
			SubmissionID: row.SubmissionID,
			Filename:     row.Filename,
			SizeBytes:    row.SizeBytes,
			ContentType:  row.ContentType,
			StorageKey:   row.StorageKey,
			UploadStatus: row.UploadStatus,
			CreatedAt:    row.CreatedAt,
			CompletedAt:  toTimePtr(row.CompletedAt),
		})
	}

	return files, nil
}

func (r *SubmissionFileRepository) CountBySubmission(ctx context.Context, submissionID uuid.UUID) (int64, error) {
	count, err := r.Queries().CountFilesBySubmission(ctx, submissionID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return 0, fmt.Errorf("failed to count files by submission: %w", err)
	}

	return count, nil
}

func (r *SubmissionFileRepository) CountCompletedBySubmission(ctx context.Context, submissionID uuid.UUID) (int64, error) {
	count, err := r.Queries().CountCompletedFilesBySubmission(ctx, submissionID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return 0, fmt.Errorf("failed to count completed files by submission: %w", err)
	}

	return count, nil
}

func (r *SubmissionFileRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, completedAt *time.Time) (*entity.SubmissionFile, error) {
	row, err := r.Queries().UpdateFileStatus(ctx, queries.UpdateFileStatusParams{
		ID:           id,
		UploadStatus: status,
		CompletedAt:  toNullTime(completedAt),
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to update file status: %w", err)
	}

	return &entity.SubmissionFile{
		ID:           row.ID,
		SubmissionID: row.SubmissionID,
		Filename:     row.Filename,
		SizeBytes:    row.SizeBytes,
		ContentType:  row.ContentType,
		StorageKey:   row.StorageKey,
		UploadStatus: row.UploadStatus,
		CreatedAt:    row.CreatedAt,
		CompletedAt:  toTimePtr(row.CompletedAt),
	}, nil
}

func (r *SubmissionFileRepository) MarkExpiredFilesAsFailed(ctx context.Context, expiredBefore time.Duration) (int64, error) {
	interval := pgtype.Interval{
		Microseconds: expiredBefore.Microseconds(),
		Valid:        true,
	}
	
	affected, err := r.Queries().MarkExpiredFilesAsFailed(ctx, interval)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return 0, fmt.Errorf("failed to mark expired files as failed: %w", err)
	}

	return affected, nil
}

func toTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func toNullTime(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}
