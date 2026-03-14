package judging

import (
	"context"
	"time"

	submissionv1 "github.com/belikoooova/hackaton-platform-api/api/submission/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/repository/postgres"
	"github.com/google/uuid"
)

type AssignmentRepository interface {
	Create(ctx context.Context, assignment *entity.Assignment) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Assignment, error)
	ListByJudge(ctx context.Context, hackathonID, judgeUserID uuid.UUID, limit, offset int32) ([]*entity.Assignment, []bool, error)
	ListByJudgeFiltered(ctx context.Context, hackathonID, judgeUserID uuid.UUID, evaluated bool, limit, offset int32) ([]*entity.Assignment, []bool, error)
	CountByJudge(ctx context.Context, hackathonID, judgeUserID uuid.UUID) (int64, error)
	CountByJudgeFiltered(ctx context.Context, hackathonID, judgeUserID uuid.UUID, evaluated bool) (int64, error)
	CheckExists(ctx context.Context, hackathonID, submissionID, judgeUserID uuid.UUID) (bool, error)
	CountByHackathon(ctx context.Context, hackathonID uuid.UUID) (int64, error)
}

type EvaluationRepository interface {
	Create(ctx context.Context, evaluation *entity.Evaluation) error
	Update(ctx context.Context, id uuid.UUID, score int32, comment string) (*entity.Evaluation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Evaluation, error)
	GetBySubmissionAndJudge(ctx context.Context, submissionID, judgeUserID uuid.UUID) (*entity.Evaluation, error)
	ListByJudge(ctx context.Context, hackathonID, judgeUserID uuid.UUID, limit, offset int32) ([]*entity.Evaluation, error)
	CountByJudge(ctx context.Context, hackathonID, judgeUserID uuid.UUID) (int64, error)
	ListBySubmission(ctx context.Context, submissionID uuid.UUID) ([]*entity.Evaluation, error)
}

type LeaderboardRepository interface {
	GetLeaderboardScores(ctx context.Context, hackathonID uuid.UUID, limit, offset int32) ([]*postgres.LeaderboardScoreEntry, error)
	CountEvaluatedSubmissions(ctx context.Context, hackathonID uuid.UUID) (int64, error)
	GetSubmissionAverageScore(ctx context.Context, submissionID uuid.UUID) (averageScore float64, evaluationCount int32, err error)
	GetEvaluationsByOwner(ctx context.Context, hackathonID, submissionID uuid.UUID) ([]*entity.Evaluation, error)
}

type HackathonClient interface {
	GetHackathon(ctx context.Context, hackathonID string) (stage string, resultPublishedAt *time.Time, err error)
}

type ParticipationRolesClient interface {
	GetHackathonContext(ctx context.Context, hackathonID string) (userID, participationStatus string, roles []string, teamID string, err error)
	ListJudges(ctx context.Context, hackathonID string) ([]string, error)
}

type SubmissionClient interface {
	ListFinalSubmissions(ctx context.Context, hackathonID string) ([]*submissionv1.Submission, error)
	GetSubmission(ctx context.Context, hackathonID, submissionID string) (*submissionv1.Submission, error)
}
