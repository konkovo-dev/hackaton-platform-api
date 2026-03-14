package postgres

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
)

type SubmissionRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewSubmissionRepository(db queries.DBTX) *SubmissionRepository {
	return &SubmissionRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *SubmissionRepository) Create(ctx context.Context, submission *entity.Submission) error {
	row, err := r.Queries().CreateSubmission(ctx, queries.CreateSubmissionParams{
		ID:              submission.ID,
		HackathonID:     submission.HackathonID,
		OwnerKind:       submission.OwnerKind,
		OwnerID:         submission.OwnerID,
		CreatedByUserID: submission.CreatedByUserID,
		Title:           submission.Title,
		Description:     submission.Description,
		IsFinal:         submission.IsFinal,
	})
	if err != nil {
		return fmt.Errorf("failed to create submission: %w", err)
	}

	submission.CreatedAt = row.CreatedAt
	submission.UpdatedAt = row.UpdatedAt

	return nil
}

func (r *SubmissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Submission, error) {
	row, err := r.Queries().GetSubmissionByID(ctx, id)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get submission by id: %w", err)
	}

	return &entity.Submission{
		ID:              row.ID,
		HackathonID:     row.HackathonID,
		OwnerKind:       row.OwnerKind,
		OwnerID:         row.OwnerID,
		CreatedByUserID: row.CreatedByUserID,
		Title:           row.Title,
		Description:     row.Description,
		IsFinal:         row.IsFinal,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}, nil
}

func (r *SubmissionRepository) GetByIDAndHackathonID(ctx context.Context, id, hackathonID uuid.UUID) (*entity.Submission, error) {
	row, err := r.Queries().GetSubmissionByIDAndHackathonID(ctx, queries.GetSubmissionByIDAndHackathonIDParams{
		ID:          id,
		HackathonID: hackathonID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get submission: %w", err)
	}

	return &entity.Submission{
		ID:              row.ID,
		HackathonID:     row.HackathonID,
		OwnerKind:       row.OwnerKind,
		OwnerID:         row.OwnerID,
		CreatedByUserID: row.CreatedByUserID,
		Title:           row.Title,
		Description:     row.Description,
		IsFinal:         row.IsFinal,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}, nil
}

func (r *SubmissionRepository) ListByOwner(ctx context.Context, hackathonID uuid.UUID, ownerKind string, ownerID uuid.UUID, limit, offset int32) ([]*entity.Submission, error) {
	rows, err := r.Queries().ListSubmissionsByOwner(ctx, queries.ListSubmissionsByOwnerParams{
		HackathonID: hackathonID,
		OwnerKind:   ownerKind,
		OwnerID:     ownerID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to list submissions by owner: %w", err)
	}

	submissions := make([]*entity.Submission, 0, len(rows))
	for _, row := range rows {
		submissions = append(submissions, &entity.Submission{
			ID:              row.ID,
			HackathonID:     row.HackathonID,
			OwnerKind:       row.OwnerKind,
			OwnerID:         row.OwnerID,
			CreatedByUserID: row.CreatedByUserID,
			Title:           row.Title,
			Description:     row.Description,
			IsFinal:         row.IsFinal,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}

	return submissions, nil
}

func (r *SubmissionRepository) ListByHackathon(ctx context.Context, hackathonID uuid.UUID, limit, offset int32) ([]*entity.Submission, error) {
	rows, err := r.Queries().ListSubmissionsByHackathon(ctx, queries.ListSubmissionsByHackathonParams{
		HackathonID: hackathonID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to list submissions by hackathon: %w", err)
	}

	submissions := make([]*entity.Submission, 0, len(rows))
	for _, row := range rows {
		submissions = append(submissions, &entity.Submission{
			ID:              row.ID,
			HackathonID:     row.HackathonID,
			OwnerKind:       row.OwnerKind,
			OwnerID:         row.OwnerID,
			CreatedByUserID: row.CreatedByUserID,
			Title:           row.Title,
			Description:     row.Description,
			IsFinal:         row.IsFinal,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}

	return submissions, nil
}

func (r *SubmissionRepository) CountByOwner(ctx context.Context, hackathonID uuid.UUID, ownerKind string, ownerID uuid.UUID) (int64, error) {
	count, err := r.Queries().CountSubmissionsByOwner(ctx, queries.CountSubmissionsByOwnerParams{
		HackathonID: hackathonID,
		OwnerKind:   ownerKind,
		OwnerID:     ownerID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return 0, fmt.Errorf("failed to count submissions by owner: %w", err)
	}

	return count, nil
}

func (r *SubmissionRepository) GetFinalByOwner(ctx context.Context, hackathonID uuid.UUID, ownerKind string, ownerID uuid.UUID) (*entity.Submission, error) {
	row, err := r.Queries().GetFinalSubmissionByOwner(ctx, queries.GetFinalSubmissionByOwnerParams{
		HackathonID: hackathonID,
		OwnerKind:   ownerKind,
		OwnerID:     ownerID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &entity.Submission{
		ID:              row.ID,
		HackathonID:     row.HackathonID,
		OwnerKind:       row.OwnerKind,
		OwnerID:         row.OwnerID,
		CreatedByUserID: row.CreatedByUserID,
		Title:           row.Title,
		Description:     row.Description,
		IsFinal:         row.IsFinal,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}, nil
}

func (r *SubmissionRepository) UpdateDescription(ctx context.Context, id uuid.UUID, description string) (*entity.Submission, error) {
	row, err := r.Queries().UpdateSubmissionDescription(ctx, queries.UpdateSubmissionDescriptionParams{
		ID:          id,
		Description: description,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to update submission description: %w", err)
	}

	return &entity.Submission{
		ID:              row.ID,
		HackathonID:     row.HackathonID,
		OwnerKind:       row.OwnerKind,
		OwnerID:         row.OwnerID,
		CreatedByUserID: row.CreatedByUserID,
		Title:           row.Title,
		Description:     row.Description,
		IsFinal:         row.IsFinal,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}, nil
}

func (r *SubmissionRepository) GetTotalSize(ctx context.Context, submissionID uuid.UUID) (int64, error) {
	totalSize, err := r.Queries().GetSubmissionTotalSize(ctx, submissionID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return 0, fmt.Errorf("failed to get submission total size: %w", err)
	}

	return totalSize, nil
}

func (r *SubmissionRepository) UnsetFinalForOwner(ctx context.Context, hackathonID uuid.UUID, ownerKind string, ownerID uuid.UUID) error {
	err := r.Queries().UnsetFinalSubmission(ctx, queries.UnsetFinalSubmissionParams{
		HackathonID: hackathonID,
		OwnerKind:   ownerKind,
		OwnerID:     ownerID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to unset final submission: %w", err)
	}

	return nil
}
