package judging

import (
	"context"

	judgingpolicy "github.com/belikoooova/hackaton-platform-api/internal/judging-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
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

	// Get user context
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return out, nil
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return out, nil
	}

	// Get hackathon context
	stage, _, err := s.hackathonClient.GetHackathon(ctx, in.HackathonID.String())
	if err != nil {
		return out, nil
	}

	// Get participation and roles
	_, _, roles, _, err := s.prClient.GetHackathonContext(ctx, in.HackathonID.String())
	if err != nil {
		return out, nil
	}

	// Check viewSubmissionEvaluations
	if pctx, err := s.getSubmissionEvalsPolicy.LoadContext(ctx, judgingpolicy.GetSubmissionEvaluationsParams{
		HackathonID:  in.HackathonID,
		SubmissionID: uuid.New(), // dummy value for permission check
	}); err == nil {
		pctx.SetAuthenticated(true)
		pctx.SetActorUserID(userUUID)
		pctx.SetHackathonStage(stage)
		pctx.SetActorRoles(roles)
		decision := s.getSubmissionEvalsPolicy.Check(ctx, pctx)
		out.ViewSubmissionEvaluations = decision.Allowed
	}

	// Check viewMyJudgingAssignments
	if pctx, err := s.getMyAssignmentsPolicy.LoadContext(ctx, judgingpolicy.GetMyAssignmentsParams{
		HackathonID: in.HackathonID,
	}); err == nil {
		pctx.SetAuthenticated(true)
		pctx.SetActorUserID(userUUID)
		pctx.SetHackathonStage(stage)
		pctx.SetActorRoles(roles)
		decision := s.getMyAssignmentsPolicy.Check(ctx, pctx)
		out.ViewMyJudgingAssignments = decision.Allowed
	}

	// Check viewLeaderboard
	if pctx, err := s.getLeaderboardPolicy.LoadContext(ctx, judgingpolicy.GetLeaderboardParams{
		HackathonID: in.HackathonID,
	}); err == nil {
		pctx.SetAuthenticated(true)
		pctx.SetActorUserID(userUUID)
		pctx.SetHackathonStage(stage)
		pctx.SetActorRoles(roles)
		decision := s.getLeaderboardPolicy.Check(ctx, pctx)
		out.ViewLeaderboard = decision.Allowed
	}

	// Check assignJudging
	if pctx, err := s.assignPolicy.LoadContext(ctx, judgingpolicy.AssignSubmissionsParams{
		HackathonID: in.HackathonID,
	}); err == nil {
		pctx.SetAuthenticated(true)
		pctx.SetActorUserID(userUUID)
		pctx.SetHackathonStage(stage)
		pctx.SetActorRoles(roles)
		decision := s.assignPolicy.Check(ctx, pctx)
		out.AssignJudging = decision.Allowed
	}

	// Check submitVerdict
	if pctx, err := s.submitEvaluationPolicy.LoadContext(ctx, judgingpolicy.SubmitEvaluationParams{
		HackathonID:  in.HackathonID,
		SubmissionID: uuid.New(), // dummy value for permission check
	}); err == nil {
		pctx.SetAuthenticated(true)
		pctx.SetActorUserID(userUUID)
		pctx.SetHackathonStage(stage)
		pctx.SetActorRoles(roles)
		decision := s.submitEvaluationPolicy.Check(ctx, pctx)
		out.SubmitVerdict = decision.Allowed
	}

	return out, nil
}
