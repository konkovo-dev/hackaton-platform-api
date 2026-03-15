package submission

import (
	"context"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain/entity"
	"github.com/google/uuid"
)

type SubmissionRepository interface {
	Create(ctx context.Context, submission *entity.Submission) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Submission, error)
	GetByIDAndHackathonID(ctx context.Context, id, hackathonID uuid.UUID) (*entity.Submission, error)
	ListByOwner(ctx context.Context, hackathonID uuid.UUID, ownerKind string, ownerID uuid.UUID, limit, offset int32) ([]*entity.Submission, error)
	ListByHackathon(ctx context.Context, hackathonID uuid.UUID, limit, offset int32) ([]*entity.Submission, error)
	CountByOwner(ctx context.Context, hackathonID uuid.UUID, ownerKind string, ownerID uuid.UUID) (int64, error)
	GetFinalByOwner(ctx context.Context, hackathonID uuid.UUID, ownerKind string, ownerID uuid.UUID) (*entity.Submission, error)
	UpdateDescription(ctx context.Context, id uuid.UUID, description string) (*entity.Submission, error)
	GetTotalSize(ctx context.Context, submissionID uuid.UUID) (int64, error)
	UnsetFinalForOwner(ctx context.Context, hackathonID uuid.UUID, ownerKind string, ownerID uuid.UUID) error
}

type SubmissionFileRepository interface {
	Create(ctx context.Context, file *entity.SubmissionFile) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.SubmissionFile, error)
	GetByIDAndSubmissionID(ctx context.Context, id, submissionID uuid.UUID) (*entity.SubmissionFile, error)
	ListBySubmission(ctx context.Context, submissionID uuid.UUID) ([]*entity.SubmissionFile, error)
	CountBySubmission(ctx context.Context, submissionID uuid.UUID) (int64, error)
	CountCompletedBySubmission(ctx context.Context, submissionID uuid.UUID) (int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, completedAt *time.Time) (*entity.SubmissionFile, error)
	MarkExpiredFilesAsFailed(ctx context.Context, expiredBefore time.Duration) (int64, error)
}

type HackathonClient interface {
	GetHackathon(ctx context.Context, hackathonID string) (stage string, err error)
}

type ParticipationRolesClient interface {
	GetHackathonContext(ctx context.Context, hackathonID string) (userID, participationStatus string, roles []string, teamID string, err error)
}

type TeamClient interface {
	GetTeamCaptain(ctx context.Context, hackathonID, teamID string) (captainUserID string, err error)
	ListTeamMembers(ctx context.Context, hackathonID, teamID string) ([]string, error)
}
