package txrepo

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/repository/postgres/queries"
	"github.com/jackc/pgx/v5"
)

type AssignmentRepository struct {
	queries *queries.Queries
}

func NewAssignmentRepository(tx pgx.Tx) *AssignmentRepository {
	return &AssignmentRepository{
		queries: queries.New(tx),
	}
}

func (r *AssignmentRepository) Create(ctx context.Context, assignment *entity.Assignment) error {
	row, err := r.queries.CreateAssignment(ctx, queries.CreateAssignmentParams{
		ID:           assignment.ID,
		HackathonID:  assignment.HackathonID,
		SubmissionID: assignment.SubmissionID,
		JudgeUserID:  assignment.JudgeUserID,
	})
	if err != nil {
		return fmt.Errorf("failed to create assignment: %w", err)
	}

	assignment.AssignedAt = row.AssignedAt

	return nil
}
