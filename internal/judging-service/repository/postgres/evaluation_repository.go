package postgres

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
)

type EvaluationRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewEvaluationRepository(db queries.DBTX) *EvaluationRepository {
	return &EvaluationRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *EvaluationRepository) Create(ctx context.Context, evaluation *entity.Evaluation) error {
	row, err := r.Queries().CreateEvaluation(ctx, queries.CreateEvaluationParams{
		ID:           evaluation.ID,
		HackathonID:  evaluation.HackathonID,
		SubmissionID: evaluation.SubmissionID,
		JudgeUserID:  evaluation.JudgeUserID,
		Score:        evaluation.Score,
		Comment:      evaluation.Comment,
	})
	if err != nil {
		return fmt.Errorf("failed to create evaluation: %w", err)
	}

	evaluation.EvaluatedAt = row.EvaluatedAt
	evaluation.UpdatedAt = row.UpdatedAt

	return nil
}

func (r *EvaluationRepository) Update(ctx context.Context, id uuid.UUID, score int32, comment string) (*entity.Evaluation, error) {
	row, err := r.Queries().UpdateEvaluation(ctx, queries.UpdateEvaluationParams{
		ID:      id,
		Score:   score,
		Comment: comment,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to update evaluation: %w", err)
	}

	return &entity.Evaluation{
		ID:           row.ID,
		HackathonID:  row.HackathonID,
		SubmissionID: row.SubmissionID,
		JudgeUserID:  row.JudgeUserID,
		Score:        row.Score,
		Comment:      row.Comment,
		EvaluatedAt:  row.EvaluatedAt,
		UpdatedAt:    row.UpdatedAt,
	}, nil
}

func (r *EvaluationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Evaluation, error) {
	row, err := r.Queries().GetEvaluationByID(ctx, id)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get evaluation by id: %w", err)
	}

	return &entity.Evaluation{
		ID:           row.ID,
		HackathonID:  row.HackathonID,
		SubmissionID: row.SubmissionID,
		JudgeUserID:  row.JudgeUserID,
		Score:        row.Score,
		Comment:      row.Comment,
		EvaluatedAt:  row.EvaluatedAt,
		UpdatedAt:    row.UpdatedAt,
	}, nil
}

func (r *EvaluationRepository) GetBySubmissionAndJudge(ctx context.Context, submissionID, judgeUserID uuid.UUID) (*entity.Evaluation, error) {
	row, err := r.Queries().GetEvaluationBySubmissionAndJudge(ctx, queries.GetEvaluationBySubmissionAndJudgeParams{
		SubmissionID: submissionID,
		JudgeUserID:  judgeUserID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get evaluation by submission and judge: %w", err)
	}

	return &entity.Evaluation{
		ID:           row.ID,
		HackathonID:  row.HackathonID,
		SubmissionID: row.SubmissionID,
		JudgeUserID:  row.JudgeUserID,
		Score:        row.Score,
		Comment:      row.Comment,
		EvaluatedAt:  row.EvaluatedAt,
		UpdatedAt:    row.UpdatedAt,
	}, nil
}

func (r *EvaluationRepository) ListByJudge(ctx context.Context, hackathonID, judgeUserID uuid.UUID, limit, offset int32) ([]*entity.Evaluation, error) {
	rows, err := r.Queries().ListEvaluationsByJudge(ctx, queries.ListEvaluationsByJudgeParams{
		HackathonID: hackathonID,
		JudgeUserID: judgeUserID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to list evaluations by judge: %w", err)
	}

	evaluations := make([]*entity.Evaluation, 0, len(rows))
	for _, row := range rows {
		evaluations = append(evaluations, &entity.Evaluation{
			ID:           row.ID,
			HackathonID:  row.HackathonID,
			SubmissionID: row.SubmissionID,
			JudgeUserID:  row.JudgeUserID,
			Score:        row.Score,
			Comment:      row.Comment,
			EvaluatedAt:  row.EvaluatedAt,
			UpdatedAt:    row.UpdatedAt,
		})
	}

	return evaluations, nil
}

func (r *EvaluationRepository) CountByJudge(ctx context.Context, hackathonID, judgeUserID uuid.UUID) (int64, error) {
	count, err := r.Queries().CountEvaluationsByJudge(ctx, queries.CountEvaluationsByJudgeParams{
		HackathonID: hackathonID,
		JudgeUserID: judgeUserID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return 0, fmt.Errorf("failed to count evaluations by judge: %w", err)
	}

	return count, nil
}

func (r *EvaluationRepository) ListBySubmission(ctx context.Context, submissionID uuid.UUID) ([]*entity.Evaluation, error) {
	rows, err := r.Queries().ListEvaluationsBySubmission(ctx, submissionID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to list evaluations by submission: %w", err)
	}

	evaluations := make([]*entity.Evaluation, 0, len(rows))
	for _, row := range rows {
		evaluations = append(evaluations, &entity.Evaluation{
			ID:           row.ID,
			HackathonID:  row.HackathonID,
			SubmissionID: row.SubmissionID,
			JudgeUserID:  row.JudgeUserID,
			Score:        row.Score,
			Comment:      row.Comment,
			EvaluatedAt:  row.EvaluatedAt,
			UpdatedAt:    row.UpdatedAt,
		})
	}

	return evaluations, nil
}
