package submission

import (
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/client/hackathon"
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/client/participationroles"
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/client/team"
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/repository/postgres"
	"go.uber.org/fx"
)

var Module = fx.Module("submission-usecase",
	fx.Provide(
		func(sr *postgres.SubmissionRepository) SubmissionRepository {
			return sr
		},
		func(fr *postgres.SubmissionFileRepository) SubmissionFileRepository {
			return fr
		},
		func(hc *hackathon.Client) HackathonClient {
			return hc
		},
		func(prc *participationroles.Client) ParticipationRolesClient {
			return prc
		},
		func(tc *team.Client) TeamClient {
			return tc
		},
		NewService,
	),
)
