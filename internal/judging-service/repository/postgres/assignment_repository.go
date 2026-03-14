package postgres

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
)

type AssignmentRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewAssignmentRepository(db queries.DBTX) *AssignmentRepository {
	return &AssignmentRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *AssignmentRepository) Create(ctx context.Context, assignment *entity.Assignment) error {
	row, err := r.Queries().CreateAssignment(ctx, queries.CreateAssignmentParams{
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

func (r *AssignmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Assignment, error) {
	row, err := r.Queries().GetAssignmentByID(ctx, id)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get assignment by id: %w", err)
	}

	return &entity.Assignment{
		ID:           row.ID,
		HackathonID:  row.HackathonID,
		SubmissionID: row.SubmissionID,
		JudgeUserID:  row.JudgeUserID,
		AssignedAt:   row.AssignedAt,
	}, nil
}

func (r *AssignmentRepository) ListByJudge(ctx context.Context, hackathonID, judgeUserID uuid.UUID, limit, offset int32) ([]*entity.Assignment, []bool, error) {
	rows, err := r.Queries().ListAssignmentsByJudge(ctx, queries.ListAssignmentsByJudgeParams{
		HackathonID: hackathonID,
		JudgeUserID: judgeUserID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, nil, fmt.Errorf("failed to list assignments by judge: %w", err)
	}

	assignments := make([]*entity.Assignment, 0, len(rows))
	isEvaluated := make([]bool, 0, len(rows))
	for _, row := range rows {
		assignments = append(assignments, &entity.Assignment{
			ID:           row.ID,
			HackathonID:  row.HackathonID,
			SubmissionID: row.SubmissionID,
			JudgeUserID:  row.JudgeUserID,
			AssignedAt:   row.AssignedAt,
		})
		isEvaluated = append(isEvaluated, row.IsEvaluated)
	}

	return assignments, isEvaluated, nil
}

func (r *AssignmentRepository) ListByJudgeFiltered(ctx context.Context, hackathonID, judgeUserID uuid.UUID, evaluated bool, limit, offset int32) ([]*entity.Assignment, []bool, error) {
	rows, err := r.Queries().ListAssignmentsByJudgeFiltered(ctx, queries.ListAssignmentsByJudgeFilteredParams{
		HackathonID: hackathonID,
		JudgeUserID: judgeUserID,
		Limit:       limit,
		Offset:      offset,
		Evaluated:   evaluated,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, nil, fmt.Errorf("failed to list assignments by judge filtered: %w", err)
	}

	assignments := make([]*entity.Assignment, 0, len(rows))
	isEvaluated := make([]bool, 0, len(rows))
	for _, row := range rows {
		assignments = append(assignments, &entity.Assignment{
			ID:           row.ID,
			HackathonID:  row.HackathonID,
			SubmissionID: row.SubmissionID,
			JudgeUserID:  row.JudgeUserID,
			AssignedAt:   row.AssignedAt,
		})
		isEvaluated = append(isEvaluated, row.IsEvaluated)
	}

	return assignments, isEvaluated, nil
}

func (r *AssignmentRepository) CountByJudge(ctx context.Context, hackathonID, judgeUserID uuid.UUID) (int64, error) {
	count, err := r.Queries().CountAssignmentsByJudge(ctx, queries.CountAssignmentsByJudgeParams{
		HackathonID: hackathonID,
		JudgeUserID: judgeUserID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return 0, fmt.Errorf("failed to count assignments by judge: %w", err)
	}

	return count, nil
}

func (r *AssignmentRepository) CountByJudgeFiltered(ctx context.Context, hackathonID, judgeUserID uuid.UUID, evaluated bool) (int64, error) {
	count, err := r.Queries().CountAssignmentsByJudgeFiltered(ctx, queries.CountAssignmentsByJudgeFilteredParams{
		HackathonID: hackathonID,
		JudgeUserID: judgeUserID,
		Evaluated:   evaluated,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return 0, fmt.Errorf("failed to count assignments by judge filtered: %w", err)
	}

	return count, nil
}

func (r *AssignmentRepository) CheckExists(ctx context.Context, hackathonID, submissionID, judgeUserID uuid.UUID) (bool, error) {
	exists, err := r.Queries().CheckAssignmentExists(ctx, queries.CheckAssignmentExistsParams{
		HackathonID:  hackathonID,
		SubmissionID: submissionID,
		JudgeUserID:  judgeUserID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return false, fmt.Errorf("failed to check assignment exists: %w", err)
	}

	return exists, nil
}

func (r *AssignmentRepository) CountByHackathon(ctx context.Context, hackathonID uuid.UUID) (int64, error) {
	count, err := r.Queries().CountAssignmentsByHackathon(ctx, hackathonID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return 0, fmt.Errorf("failed to count assignments by hackathon: %w", err)
	}

	return count, nil
}
