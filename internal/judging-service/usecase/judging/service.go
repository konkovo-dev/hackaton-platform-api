package judging

import (
	"log/slog"

	judgingpolicy "github.com/belikoooova/hackaton-platform-api/internal/judging-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
)

type Service struct {
	logger *slog.Logger

	assignmentRepo  AssignmentRepository
	evaluationRepo  EvaluationRepository
	leaderboardRepo LeaderboardRepository

	hackathonClient HackathonClient
	prClient        ParticipationRolesClient
	submissionClient SubmissionClient

	txManager *pgxutil.TxManager

	assignPolicy              *judgingpolicy.AssignSubmissionsPolicy
	getMyAssignmentsPolicy    *judgingpolicy.GetMyAssignmentsPolicy
	submitEvaluationPolicy    *judgingpolicy.SubmitEvaluationPolicy
	getMyEvaluationsPolicy    *judgingpolicy.GetMyEvaluationsPolicy
	getSubmissionEvalsPolicy  *judgingpolicy.GetSubmissionEvaluationsPolicy
	getLeaderboardPolicy      *judgingpolicy.GetLeaderboardPolicy
	getMyResultPolicy         *judgingpolicy.GetMyEvaluationResultPolicy
}

func NewService(
	logger *slog.Logger,
	assignmentRepo AssignmentRepository,
	evaluationRepo EvaluationRepository,
	leaderboardRepo LeaderboardRepository,
	hackathonClient HackathonClient,
	prClient ParticipationRolesClient,
	submissionClient SubmissionClient,
	txManager *pgxutil.TxManager,
	assignPolicy *judgingpolicy.AssignSubmissionsPolicy,
	getMyAssignmentsPolicy *judgingpolicy.GetMyAssignmentsPolicy,
	submitEvaluationPolicy *judgingpolicy.SubmitEvaluationPolicy,
	getMyEvaluationsPolicy *judgingpolicy.GetMyEvaluationsPolicy,
	getSubmissionEvalsPolicy *judgingpolicy.GetSubmissionEvaluationsPolicy,
	getLeaderboardPolicy *judgingpolicy.GetLeaderboardPolicy,
	getMyResultPolicy *judgingpolicy.GetMyEvaluationResultPolicy,
) *Service {
	return &Service{
		logger:                    logger,
		assignmentRepo:            assignmentRepo,
		evaluationRepo:            evaluationRepo,
		leaderboardRepo:           leaderboardRepo,
		hackathonClient:           hackathonClient,
		prClient:                  prClient,
		submissionClient:          submissionClient,
		txManager:                 txManager,
		assignPolicy:              assignPolicy,
		getMyAssignmentsPolicy:    getMyAssignmentsPolicy,
		submitEvaluationPolicy:    submitEvaluationPolicy,
		getMyEvaluationsPolicy:    getMyEvaluationsPolicy,
		getSubmissionEvalsPolicy:  getSubmissionEvalsPolicy,
		getLeaderboardPolicy:      getLeaderboardPolicy,
		getMyResultPolicy:         getMyResultPolicy,
	}
}
