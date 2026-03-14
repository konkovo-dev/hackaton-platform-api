package judging

import (
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/client/hackathon"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/client/participationroles"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/client/submission"
	judgingpolicy "github.com/belikoooova/hackaton-platform-api/internal/judging-service/policy"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/repository/postgres"
	"go.uber.org/fx"
)

var Module = fx.Module("judging-usecase",
	fx.Provide(
		func(ar *postgres.AssignmentRepository) AssignmentRepository {
			return ar
		},
		func(er *postgres.EvaluationRepository) EvaluationRepository {
			return er
		},
		func(lr *postgres.LeaderboardRepository) LeaderboardRepository {
			return lr
		},
		func(hc *hackathon.Client) HackathonClient {
			return hc
		},
		func(prc *participationroles.Client) ParticipationRolesClient {
			return prc
		},
		func(sc *submission.Client) SubmissionClient {
			return sc
		},
		NewService,
		judgingpolicy.NewAssignSubmissionsPolicy,
		judgingpolicy.NewGetMyAssignmentsPolicy,
		judgingpolicy.NewSubmitEvaluationPolicy,
		judgingpolicy.NewGetMyEvaluationsPolicy,
		judgingpolicy.NewGetSubmissionEvaluationsPolicy,
		judgingpolicy.NewGetLeaderboardPolicy,
		judgingpolicy.NewGetMyEvaluationResultPolicy,
	),
)
