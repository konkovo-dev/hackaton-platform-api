package judging

import (
	"context"

	judgingpolicy "github.com/belikoooova/hackaton-platform-api/internal/judging-service/policy"
	"github.com/google/uuid"
)

type GetJudgingPermissionsIn struct {
	HackathonID uuid.UUID
}

type GetJudgingPermissionsOut struct {
	ViewSubmissionEvaluations bool
	ViewMyJudgingAssignments  bool
	ViewLeaderboard           bool
	AssignJudging             bool
	SubmitVerdict             bool
}

func (s *Service) GetJudgingPermissions(ctx context.Context, in GetJudgingPermissionsIn) (*GetJudgingPermissionsOut, error) {
	out := &GetJudgingPermissionsOut{}

	// Check viewSubmissionEvaluations
	if pctx, err := s.getSubmissionEvalsPolicy.LoadContext(ctx, judgingpolicy.GetSubmissionEvaluationsParams{
		HackathonID:  in.HackathonID,
		SubmissionID: uuid.New(), // dummy value for permission check
	}); err == nil {
		decision := s.getSubmissionEvalsPolicy.Check(ctx, pctx)
		out.ViewSubmissionEvaluations = decision.Allowed
	}

	// Check viewMyJudgingAssignments
	if pctx, err := s.getMyAssignmentsPolicy.LoadContext(ctx, judgingpolicy.GetMyAssignmentsParams{
		HackathonID: in.HackathonID,
	}); err == nil {
		decision := s.getMyAssignmentsPolicy.Check(ctx, pctx)
		out.ViewMyJudgingAssignments = decision.Allowed
	}

	// Check viewLeaderboard
	if pctx, err := s.getLeaderboardPolicy.LoadContext(ctx, judgingpolicy.GetLeaderboardParams{
		HackathonID: in.HackathonID,
	}); err == nil {
		decision := s.getLeaderboardPolicy.Check(ctx, pctx)
		out.ViewLeaderboard = decision.Allowed
	}

	// Check assignJudging
	if pctx, err := s.assignPolicy.LoadContext(ctx, judgingpolicy.AssignSubmissionsParams{
		HackathonID: in.HackathonID,
	}); err == nil {
		decision := s.assignPolicy.Check(ctx, pctx)
		out.AssignJudging = decision.Allowed
	}

	// Check submitVerdict
	if pctx, err := s.submitEvaluationPolicy.LoadContext(ctx, judgingpolicy.SubmitEvaluationParams{
		HackathonID:  in.HackathonID,
		SubmissionID: uuid.New(), // dummy value for permission check
	}); err == nil {
		decision := s.submitEvaluationPolicy.Check(ctx, pctx)
		out.SubmitVerdict = decision.Allowed
	}

	return out, nil
}
