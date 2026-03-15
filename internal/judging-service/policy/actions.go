package policy

import "github.com/belikoooova/hackaton-platform-api/pkg/policy"

const (
	ActionAssignSubmissionsToJudges policy.Action = "judging.assign_submissions_to_judges"
	ActionGetMyAssignments          policy.Action = "judging.get_my_assignments"
	ActionSubmitEvaluation          policy.Action = "judging.submit_evaluation"
	ActionGetMyEvaluations          policy.Action = "judging.get_my_evaluations"
	ActionGetSubmissionEvaluations  policy.Action = "judging.get_submission_evaluations"
	ActionGetLeaderboard            policy.Action = "judging.get_leaderboard"
	ActionGetMyEvaluationResult     policy.Action = "judging.get_my_evaluation_result"
)
