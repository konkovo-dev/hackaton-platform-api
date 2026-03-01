package txrepo

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/repository/postgres/queries"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SubmissionRepository struct {
	queries *queries.Queries
}

func NewSubmissionRepository(tx pgx.Tx) *SubmissionRepository {
	return &SubmissionRepository{
		queries: queries.New(tx),
	}
}

func (r *SubmissionRepository) UnsetFinalSubmission(ctx context.Context, hackathonID uuid.UUID, ownerKind string, ownerID uuid.UUID) error {
	err := r.queries.UnsetFinalSubmission(ctx, queries.UnsetFinalSubmissionParams{
		HackathonID: hackathonID,
		OwnerKind:   ownerKind,
		OwnerID:     ownerID,
	})
	if err != nil {
		return fmt.Errorf("failed to unset final submission: %w", err)
	}
	return nil
}

func (r *SubmissionRepository) SetFinalSubmission(ctx context.Context, submissionID uuid.UUID) error {
	err := r.queries.SetFinalSubmission(ctx, submissionID)
	if err != nil {
		return fmt.Errorf("failed to set final submission: %w", err)
	}
	return nil
}
