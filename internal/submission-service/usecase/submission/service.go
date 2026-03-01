package submission

import (
	"log/slog"

	submissionservice "github.com/belikoooova/hackaton-platform-api/internal/submission-service"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/belikoooova/hackaton-platform-api/pkg/s3"
)

type Service struct {
	submissionRepo SubmissionRepository
	fileRepo       SubmissionFileRepository
	hackathonClient HackathonClient
	prClient       ParticipationRolesClient
	teamClient     TeamClient
	txManager      *pgxutil.TxManager
	idempotencyHelper *idempotency.Helper
	idempotencyRepo idempotency.Repository
	s3Client       *s3.Client
	config         *submissionservice.Config
	logger         *slog.Logger
}

func NewService(
	submissionRepo SubmissionRepository,
	fileRepo SubmissionFileRepository,
	hackathonClient HackathonClient,
	prClient ParticipationRolesClient,
	teamClient TeamClient,
	txManager *pgxutil.TxManager,
	idempotencyHelper *idempotency.Helper,
	idempotencyRepo idempotency.Repository,
	s3Client *s3.Client,
	config *submissionservice.Config,
	logger *slog.Logger,
) *Service {
	return &Service{
		submissionRepo: submissionRepo,
		fileRepo:       fileRepo,
		hackathonClient: hackathonClient,
		prClient:       prClient,
		teamClient:     teamClient,
		txManager:      txManager,
		idempotencyHelper: idempotencyHelper,
		idempotencyRepo: idempotencyRepo,
		s3Client:       s3Client,
		config:         config,
		logger:         logger,
	}
}
