package postgres

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
)

type LeaderboardScoreEntry struct {
	SubmissionID    uuid.UUID
	AverageScore    float64
	EvaluationCount int32
}

type LeaderboardRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewLeaderboardRepository(db queries.DBTX) *LeaderboardRepository {
	return &LeaderboardRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *LeaderboardRepository) GetLeaderboardScores(ctx context.Context, hackathonID uuid.UUID, limit, offset int32) ([]*LeaderboardScoreEntry, error) {
	rows, err := r.Queries().GetLeaderboardScores(ctx, queries.GetLeaderboardScoresParams{
		HackathonID: hackathonID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get leaderboard scores: %w", err)
	}

	entries := make([]*LeaderboardScoreEntry, 0, len(rows))
	for _, row := range rows {
		avgScore := 0.0
		if score, ok := row.AverageScore.(float64); ok {
			avgScore = score
		}
		entries = append(entries, &LeaderboardScoreEntry{
			SubmissionID:    row.SubmissionID,
			AverageScore:    avgScore,
			EvaluationCount: row.EvaluationCount,
		})
	}

	return entries, nil
}

func (r *LeaderboardRepository) CountEvaluatedSubmissions(ctx context.Context, hackathonID uuid.UUID) (int64, error) {
	count, err := r.Queries().CountEvaluatedSubmissions(ctx, hackathonID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return 0, fmt.Errorf("failed to count evaluated submissions: %w", err)
	}

	return count, nil
}

func (r *LeaderboardRepository) GetSubmissionAverageScore(ctx context.Context, submissionID uuid.UUID) (float64, int32, error) {
	row, err := r.Queries().GetSubmissionAverageScore(ctx, submissionID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return 0, 0, nil
		}
		return 0, 0, fmt.Errorf("failed to get submission average score: %w", err)
	}

	avgScore := 0.0
	if score, ok := row.AverageScore.(float64); ok {
		avgScore = score
	}

	return avgScore, row.EvaluationCount, nil
}

func (r *LeaderboardRepository) GetEvaluationsByOwner(ctx context.Context, hackathonID, submissionID uuid.UUID) ([]*entity.Evaluation, error) {
	rows, err := r.Queries().GetEvaluationsByOwner(ctx, queries.GetEvaluationsByOwnerParams{
		HackathonID:  hackathonID,
		SubmissionID: submissionID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get evaluations by owner: %w", err)
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
