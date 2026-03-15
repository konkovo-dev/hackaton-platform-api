package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssignSubmissionsToJudges_AsOrganizer_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant1 := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant1, "PART_INDIVIDUAL")
	submission1 := createSubmissionForJudging(tc, hackathonID, participant1)

	participant2 := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant2, "PART_INDIVIDUAL")
	submission2 := createSubmissionForJudging(tc, hackathonID, participant2)

	judge1 := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge1)

	judge2 := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge2)

	judge3 := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge3)

	body := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}

	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, body)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to assign submissions: %s", string(respBody))

	assignResp := tc.ParseJSON(respBody)

	assert.Equal(t, float64(6), assignResp["assignmentsCount"], "Should create 6 assignments (2 submissions * 3 judges)")
	assert.Equal(t, float64(3), assignResp["judgesCount"], "Should have 3 judges")
	assert.Equal(t, float64(2), assignResp["submissionsCount"], "Should have 2 submissions")

	var assignmentCount int64
	err := tc.JudgingDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT COUNT(*) FROM %s.assignments WHERE hackathon_id = $1", tc.JudgingDBName),
		hackathonID,
	).Scan(&assignmentCount)
	require.NoError(t, err)
	assert.Equal(t, int64(6), assignmentCount, "Should have 6 assignments in database")

	_ = submission1
	_ = submission2
}

func TestAssignSubmissionsToJudges_Idempotent_ShouldReturnSuccess(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	createSubmissionForJudging(tc, hackathonID, participant)

	judge := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge)

	idempotencyKey := uuid.New().String()
	body := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": idempotencyKey,
		},
	}

	resp1, respBody1 := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, body)
	require.Equal(t, http.StatusOK, resp1.StatusCode,
		"First assignment should succeed: %s", string(respBody1))

	resp2, respBody2 := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, body)
	require.Equal(t, http.StatusOK, resp2.StatusCode,
		"Second assignment should also succeed (idempotent): %s", string(respBody2))

	var assignmentCount int64
	err := tc.JudgingDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT COUNT(*) FROM %s.assignments WHERE hackathon_id = $1", tc.JudgingDBName),
		hackathonID,
	).Scan(&assignmentCount)
	require.NoError(t, err)
	assert.Equal(t, int64(1), assignmentCount, "Should still have only 1 assignment")
}

func TestAssignSubmissionsToJudges_AsParticipant_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	body := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}

	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), participant.AccessToken, body)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"Participant should not be able to assign: %s", string(respBody))
}

func TestAssignSubmissionsToJudges_WrongStage_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInRunningForSubmissions(tc, owner)

	body := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}

	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, body)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"Should fail in non-judging stage: %s", string(respBody))
}

func TestAssignSubmissionsToJudges_NoJudges_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	createSubmissionForJudging(tc, hackathonID, participant)

	body := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}

	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, body)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
		"Should fail when no judges: %s", string(respBody))
}

func TestAssignSubmissionsToJudges_NoSubmissions_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	judge := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge)

	body := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}

	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, body)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
		"Should fail when no submissions: %s", string(respBody))
}

func TestGetMyAssignments_AsJudge_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	submissionID := createSubmissionForJudging(tc, hackathonID, participant)

	judge := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge)

	assignBody := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, assignBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to assign submissions: %s", string(respBody))

	listBody := map[string]interface{}{
		"query": map[string]interface{}{
			"limit":  10,
			"offset": 0,
		},
	}

	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/my-assignments/list", hackathonID), judge.AccessToken, listBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to get assignments: %s", string(respBody))

	listResp := tc.ParseJSON(respBody)

	assignments := listResp["assignments"].([]interface{})
	assert.Len(t, assignments, 1, "Judge should have 1 assignment")

	assignment := assignments[0].(map[string]interface{})
	assignmentData := assignment["assignment"].(map[string]interface{})
	assert.Equal(t, submissionID, assignmentData["submissionId"], "Assignment should be for created submission")
	assert.False(t, assignmentData["isEvaluated"].(bool), "Should not be evaluated yet")
}

func TestGetMyAssignments_FilterEvaluated_ShouldWork(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant1 := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant1, "PART_INDIVIDUAL")
	submission1 := createSubmissionForJudging(tc, hackathonID, participant1)

	participant2 := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant2, "PART_INDIVIDUAL")
	submission2 := createSubmissionForJudging(tc, hackathonID, participant2)

	judge := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge)

	assignBody := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, assignBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to assign submissions: %s", string(respBody))

	evalBody := map[string]interface{}{
		"score":   8,
		"comment": "Great work!",
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submission1), judge.AccessToken, evalBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to submit evaluation: %s", string(respBody))

	listEvaluatedBody := map[string]interface{}{
		"evaluated": true,
		"query": map[string]interface{}{
			"limit":  10,
			"offset": 0,
		},
	}
	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/my-assignments/list", hackathonID), judge.AccessToken, listEvaluatedBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to get evaluated assignments: %s", string(respBody))

	evaluatedResp := tc.ParseJSON(respBody)
	evaluatedAssignments := evaluatedResp["assignments"].([]interface{})
	assert.Len(t, evaluatedAssignments, 1, "Should have 1 evaluated assignment")

	listNotEvaluatedBody := map[string]interface{}{
		"evaluated": false,
		"query": map[string]interface{}{
			"limit":  10,
			"offset": 0,
		},
	}
	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/my-assignments/list", hackathonID), judge.AccessToken, listNotEvaluatedBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to get not evaluated assignments: %s", string(respBody))

	notEvaluatedResp := tc.ParseJSON(respBody)
	notEvaluatedAssignments := notEvaluatedResp["assignments"].([]interface{})
	assert.Len(t, notEvaluatedAssignments, 1, "Should have 1 not evaluated assignment")

	_ = submission2
}

func TestGetMyAssignments_AsParticipant_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	listBody := map[string]interface{}{
		"query": map[string]interface{}{
			"limit":  10,
			"offset": 0,
		},
	}

	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/my-assignments/list", hackathonID), participant.AccessToken, listBody)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"Participant should not see assignments: %s", string(respBody))
}

func TestSubmitEvaluation_AsJudge_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	submissionID := createSubmissionForJudging(tc, hackathonID, participant)

	judge := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge)

	assignBody := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, assignBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to assign submissions: %s", string(respBody))

	evalBody := map[string]interface{}{
		"score":   9,
		"comment": "Excellent submission!",
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}

	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submissionID), judge.AccessToken, evalBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to submit evaluation: %s", string(respBody))

	evalResp := tc.ParseJSON(respBody)
	assert.NotEmpty(t, evalResp["evaluationId"], "Should return evaluation_id")
	assert.NotEmpty(t, evalResp["evaluatedAt"], "Should return evaluated_at")

	var score int32
	var comment string
	err := tc.JudgingDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT score, comment FROM %s.evaluations WHERE submission_id = $1 AND judge_user_id = $2", tc.JudgingDBName),
		submissionID, judge.UserID,
	).Scan(&score, &comment)
	require.NoError(t, err)
	assert.Equal(t, int32(9), score, "Score should be saved")
	assert.Equal(t, "Excellent submission!", comment, "Comment should be saved")
}

func TestSubmitEvaluation_UpdateExisting_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	submissionID := createSubmissionForJudging(tc, hackathonID, participant)

	judge := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge)

	assignBody := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, assignBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to assign submissions: %s", string(respBody))

	evalBody1 := map[string]interface{}{
		"score":   7,
		"comment": "Good work",
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submissionID), judge.AccessToken, evalBody1)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to submit first evaluation: %s", string(respBody))

	evalResp1 := tc.ParseJSON(respBody)
	evaluationID := evalResp1["evaluationId"].(string)

	time.Sleep(100 * time.Millisecond)

	evalBody2 := map[string]interface{}{
		"score":   9,
		"comment": "Actually excellent!",
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submissionID), judge.AccessToken, evalBody2)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to update evaluation: %s", string(respBody))

	evalResp2 := tc.ParseJSON(respBody)
	assert.Equal(t, evaluationID, evalResp2["evaluationId"], "Should return same evaluation_id")

	var score int32
	var comment string
	err := tc.JudgingDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT score, comment FROM %s.evaluations WHERE id = $1", tc.JudgingDBName),
		evaluationID,
	).Scan(&score, &comment)
	require.NoError(t, err)
	assert.Equal(t, int32(9), score, "Score should be updated")
	assert.Equal(t, "Actually excellent!", comment, "Comment should be updated")
}

func TestSubmitEvaluation_InvalidScore_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	submissionID := createSubmissionForJudging(tc, hackathonID, participant)

	judge := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge)

	assignBody := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, assignBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to assign submissions: %s", string(respBody))

	evalBody := map[string]interface{}{
		"score":   11,
		"comment": "Invalid score",
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}

	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submissionID), judge.AccessToken, evalBody)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
		"Should reject invalid score: %s", string(respBody))
}

func TestSubmitEvaluation_EmptyComment_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	submissionID := createSubmissionForJudging(tc, hackathonID, participant)

	judge := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge)

	assignBody := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, assignBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to assign submissions: %s", string(respBody))

	evalBody := map[string]interface{}{
		"score":   8,
		"comment": "",
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}

	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submissionID), judge.AccessToken, evalBody)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
		"Should reject empty comment: %s", string(respBody))
}

func TestSubmitEvaluation_NotAssigned_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	submissionID := createSubmissionForJudging(tc, hackathonID, participant)

	judge1 := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge1)

	judge2 := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge2)

	assignBody := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, assignBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to assign submissions: %s", string(respBody))

	_, err := tc.JudgingDB.Exec(context.Background(),
		fmt.Sprintf("DELETE FROM %s.assignments WHERE judge_user_id = $1", tc.JudgingDBName),
		judge2.UserID,
	)
	require.NoError(t, err)

	evalBody := map[string]interface{}{
		"score":   8,
		"comment": "Trying to evaluate unassigned submission",
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}

	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submissionID), judge2.AccessToken, evalBody)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"Should reject evaluation for unassigned submission: %s", string(respBody))
}

func TestGetMyEvaluations_AsJudge_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	submissionID := createSubmissionForJudging(tc, hackathonID, participant)

	judge := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge)

	assignBody := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, assignBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to assign submissions: %s", string(respBody))

	evalBody := map[string]interface{}{
		"score":   8,
		"comment": "Good work!",
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submissionID), judge.AccessToken, evalBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to submit evaluation: %s", string(respBody))

	listBody := map[string]interface{}{
		"query": map[string]interface{}{
			"limit":  10,
			"offset": 0,
		},
	}

	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/my-evaluations/list", hackathonID), judge.AccessToken, listBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to get my evaluations: %s", string(respBody))

	listResp := tc.ParseJSON(respBody)

	evaluations := listResp["evaluations"].([]interface{})
	assert.Len(t, evaluations, 1, "Judge should have 1 evaluation")

	evaluation := evaluations[0].(map[string]interface{})
	evalData := evaluation["evaluation"].(map[string]interface{})
	assert.Equal(t, float64(8), evalData["score"], "Score should match")
	assert.Equal(t, "Good work!", evalData["comment"], "Comment should match")
}

func TestGetSubmissionEvaluations_AsOrganizer_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	submissionID := createSubmissionForJudging(tc, hackathonID, participant)

	judge1 := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge1)

	judge2 := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge2)

	assignBody := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, assignBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to assign submissions: %s", string(respBody))

	evalBody1 := map[string]interface{}{
		"score":   7,
		"comment": "Judge 1 comment",
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submissionID), judge1.AccessToken, evalBody1)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to submit evaluation from judge1: %s", string(respBody))

	evalBody2 := map[string]interface{}{
		"score":   9,
		"comment": "Judge 2 comment",
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submissionID), judge2.AccessToken, evalBody2)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to submit evaluation from judge2: %s", string(respBody))

	resp, respBody = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluations", hackathonID, submissionID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to get submission evaluations: %s", string(respBody))

	evalsResp := tc.ParseJSON(respBody)

	evaluations := evalsResp["evaluations"].([]interface{})
	assert.Len(t, evaluations, 2, "Should have 2 evaluations")
}

func TestGetSubmissionEvaluations_AsJudge_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	submissionID := createSubmissionForJudging(tc, hackathonID, participant)

	judge := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge)

	assignBody := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, assignBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to assign submissions: %s", string(respBody))

	evalBody := map[string]interface{}{
		"score":   8,
		"comment": "Judge comment",
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submissionID), judge.AccessToken, evalBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to submit evaluation: %s", string(respBody))

	resp, respBody = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluations", hackathonID, submissionID), judge.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Judge should be able to see submission evaluations: %s", string(respBody))

	evalsResp := tc.ParseJSON(respBody)

	evaluations := evalsResp["evaluations"].([]interface{})
	assert.Len(t, evaluations, 1, "Should have 1 evaluation")
}

func TestGetSubmissionEvaluations_AsParticipant_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	submissionID := createSubmissionForJudging(tc, hackathonID, participant)

	resp, respBody := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluations", hackathonID, submissionID), participant.AccessToken, nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"Participant should not see evaluations: %s", string(respBody))
}

func TestGetLeaderboard_AsOrganizer_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant1 := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant1, "PART_INDIVIDUAL")
	submission1 := createSubmissionForJudging(tc, hackathonID, participant1)

	participant2 := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant2, "PART_INDIVIDUAL")
	submission2 := createSubmissionForJudging(tc, hackathonID, participant2)

	judge1 := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge1)

	judge2 := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge2)

	assignBody := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, assignBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to assign submissions: %s", string(respBody))

	evalBody1 := map[string]interface{}{
		"score":   9,
		"comment": "Excellent!",
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submission1), judge1.AccessToken, evalBody1)
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submission1), judge2.AccessToken, evalBody1)

	evalBody2 := map[string]interface{}{
		"score":   6,
		"comment": "Good effort",
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submission2), judge1.AccessToken, evalBody2)
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submission2), judge2.AccessToken, evalBody2)

	listBody := map[string]interface{}{
		"query": map[string]interface{}{
			"limit":  10,
			"offset": 0,
		},
	}

	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/leaderboard", hackathonID), owner.AccessToken, listBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to get leaderboard: %s", string(respBody))

	leaderboardResp := tc.ParseJSON(respBody)

	entries := leaderboardResp["entries"].([]interface{})
	assert.Len(t, entries, 2, "Should have 2 entries")

	entry1 := entries[0].(map[string]interface{})
	assert.Equal(t, submission1, entry1["submissionId"], "First place should be submission1")
	assert.Equal(t, float64(9), entry1["averageScore"], "Average score should be 9")
	assert.Equal(t, float64(1), entry1["rank"], "Rank should be 1")

	entry2 := entries[1].(map[string]interface{})
	assert.Equal(t, submission2, entry2["submissionId"], "Second place should be submission2")
	assert.Equal(t, float64(6), entry2["averageScore"], "Average score should be 6")
	assert.Equal(t, float64(2), entry2["rank"], "Rank should be 2")
}

func TestGetLeaderboard_AsJudge_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	createSubmissionForJudging(tc, hackathonID, participant)

	judge := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge)

	assignBody := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, assignBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to assign submissions: %s", string(respBody))

	listBody := map[string]interface{}{
		"query": map[string]interface{}{
			"limit":  10,
			"offset": 0,
		},
	}

	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/leaderboard", hackathonID), judge.AccessToken, listBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Judge should be able to see leaderboard: %s", string(respBody))
}

func TestGetLeaderboard_AsParticipant_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	listBody := map[string]interface{}{
		"query": map[string]interface{}{
			"limit":  10,
			"offset": 0,
		},
	}

	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/leaderboard", hackathonID), participant.AccessToken, listBody)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"Participant should not see leaderboard: %s", string(respBody))
}

func TestGetMyEvaluationResult_BeforePublish_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	submissionID := createSubmissionForJudging(tc, hackathonID, participant)

	judge := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge)

	assignBody := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, assignBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to assign submissions: %s", string(respBody))

	evalBody := map[string]interface{}{
		"score":   8,
		"comment": "Good work!",
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submissionID), judge.AccessToken, evalBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to submit evaluation: %s", string(respBody))

	resp, respBody = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/judging/my-result", hackathonID), participant.AccessToken, nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"Should fail before result publication: %s", string(respBody))
}

func TestGetMyEvaluationResult_AfterPublish_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	submissionID := createSubmissionForJudging(tc, hackathonID, participant)

	judge1 := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge1)

	judge2 := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge2)

	assignBody := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, assignBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to assign submissions: %s", string(respBody))

	evalBody1 := map[string]interface{}{
		"score":   7,
		"comment": "Judge 1 comment",
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submissionID), judge1.AccessToken, evalBody1)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to submit evaluation from judge1: %s", string(respBody))

	evalBody2 := map[string]interface{}{
		"score":   9,
		"comment": "Judge 2 comment",
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submissionID), judge2.AccessToken, evalBody2)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to submit evaluation from judge2: %s", string(respBody))

	_, err := tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s SET result_published_at = NOW() WHERE id = $1", tc.HackathonDBName),
		hackathonID,
	)
	require.NoError(t, err, "Failed to publish results")

	resp, respBody = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/judging/my-result", hackathonID), participant.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to get my result: %s", string(respBody))

	resultResp := tc.ParseJSON(respBody)

	result := resultResp["result"].(map[string]interface{})
	assert.Equal(t, submissionID, result["submissionId"], "Should return correct submission")
	assert.Equal(t, float64(8), result["averageScore"], "Average score should be 8 ((7+9)/2)")
	assert.Equal(t, float64(2), result["evaluationCount"], "Should have 2 evaluations")
	assert.Equal(t, float64(1), result["rank"], "Should be rank 1")

	comments := result["comments"].([]interface{})
	assert.Len(t, comments, 2, "Should have 2 comments")
}

func TestGetMyEvaluationResult_AsNonOwner_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	submissionID := createSubmissionForJudging(tc, hackathonID, participant)

	otherUser := tc.RegisterUser()
	registerParticipant(tc, hackathonID, otherUser, "PART_INDIVIDUAL")

	judge := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge)

	assignBody := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, assignBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to assign submissions: %s", string(respBody))

	evalBody := map[string]interface{}{
		"score":   8,
		"comment": "Good work!",
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submissionID), judge.AccessToken, evalBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to submit evaluation: %s", string(respBody))

	_, err := tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s SET result_published_at = NOW() WHERE id = $1", tc.HackathonDBName),
		hackathonID,
	)
	require.NoError(t, err, "Failed to publish results")

	resp, respBody = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/judging/my-result", hackathonID), otherUser.AccessToken, nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode,
		"Other user should not see participant's result: %s", string(respBody))
}

func TestLeaderboardRanking_WithTiedScores_ShouldSortByCreatedAt(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant1 := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant1, "PART_INDIVIDUAL")
	submission1 := createSubmissionForJudging(tc, hackathonID, participant1)
	time.Sleep(100 * time.Millisecond)

	participant2 := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant2, "PART_INDIVIDUAL")
	submission2 := createSubmissionForJudging(tc, hackathonID, participant2)

	judge := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge)

	assignBody := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, assignBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to assign submissions: %s", string(respBody))

	evalBody := map[string]interface{}{
		"score":   8,
		"comment": "Same score for both",
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submission1), judge.AccessToken, evalBody)

	evalBody2 := map[string]interface{}{
		"score":   8,
		"comment": "Same score for both",
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, submission2), judge.AccessToken, evalBody2)

	listBody := map[string]interface{}{
		"query": map[string]interface{}{
			"limit":  10,
			"offset": 0,
		},
	}

	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/leaderboard", hackathonID), owner.AccessToken, listBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to get leaderboard: %s", string(respBody))

	leaderboardResp := tc.ParseJSON(respBody)

	entries := leaderboardResp["entries"].([]interface{})
	assert.Len(t, entries, 2, "Should have 2 entries")

	entry1 := entries[0].(map[string]interface{})
	entry2 := entries[1].(map[string]interface{})

	assert.Equal(t, float64(8), entry1["averageScore"], "Both should have score 8")
	assert.Equal(t, float64(8), entry2["averageScore"], "Both should have score 8")
	assert.Equal(t, submission1, entry1["submissionId"], "Earlier submission should rank higher")
	assert.Equal(t, submission2, entry2["submissionId"], "Later submission should rank lower")
}

func TestAssignSubmissionsToJudges_WithFewerJudgesThanMin_ShouldAssignAll(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	createSubmissionForJudging(tc, hackathonID, participant)

	judge1 := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge1)

	judge2 := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge2)

	body := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}

	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, body)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to assign submissions: %s", string(respBody))

	assignResp := tc.ParseJSON(respBody)

	assert.Equal(t, float64(2), assignResp["assignmentsCount"], "Should create 2 assignments (1 submission * 2 judges, since < 3)")
	assert.Equal(t, float64(2), assignResp["judgesCount"], "Should have 2 judges")
}

func TestCompleteJudgingWorkflow_MultipleSubmissions_ShouldWork(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	owner := tc.RegisterUser()
	hackathonID := createHackathonInJudgingStage(tc, owner)

	participant1 := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant1, "PART_INDIVIDUAL")
	submission1 := createSubmissionForJudging(tc, hackathonID, participant1)

	participant2 := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant2, "PART_INDIVIDUAL")
	submission2 := createSubmissionForJudging(tc, hackathonID, participant2)

	participant3 := tc.RegisterUser()
	registerParticipant(tc, hackathonID, participant3, "PART_INDIVIDUAL")
	submission3 := createSubmissionForJudging(tc, hackathonID, participant3)

	judge1 := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge1)

	judge2 := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge2)

	judge3 := tc.RegisterUser()
	addJudgeToHackathon(tc, hackathonID, judge3)

	assignBody := map[string]interface{}{
		"idempotency_key": map[string]interface{}{
			"key": uuid.New().String(),
		},
	}
	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/assign", hackathonID), owner.AccessToken, assignBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to assign submissions: %s", string(respBody))

	scores := map[string][]int32{
		submission1: {10, 9, 10},
		submission2: {7, 8, 7},
		submission3: {5, 6, 5},
	}

	judges := []*UserCredentials{judge1, judge2, judge3}

	for subID, judgeScores := range scores {
		for i, judge := range judges {
			evalBody := map[string]interface{}{
				"score":   judgeScores[i],
				"comment": fmt.Sprintf("Judge %d comment for submission", i+1),
				"idempotency_key": map[string]interface{}{
					"key": uuid.New().String(),
				},
			}
			resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions/%s/evaluate", hackathonID, subID), judge.AccessToken, evalBody)
			require.Equal(t, http.StatusOK, resp.StatusCode,
				"Failed to submit evaluation: %s", string(respBody))
		}
	}

	listBody := map[string]interface{}{
		"query": map[string]interface{}{
			"limit":  10,
			"offset": 0,
		},
	}

	resp, respBody = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/judging/leaderboard", hackathonID), owner.AccessToken, listBody)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to get leaderboard: %s", string(respBody))

	leaderboardResp := tc.ParseJSON(respBody)

	entries := leaderboardResp["entries"].([]interface{})
	assert.Len(t, entries, 3, "Should have 3 entries")

	entry1 := entries[0].(map[string]interface{})
	assert.Equal(t, submission1, entry1["submissionId"], "First place: submission1")
	assert.InDelta(t, 9.67, entry1["averageScore"], 0.1, "Average should be ~9.67")

	entry2 := entries[1].(map[string]interface{})
	assert.Equal(t, submission2, entry2["submissionId"], "Second place: submission2")
	assert.InDelta(t, 7.33, entry2["averageScore"], 0.1, "Average should be ~7.33")

	entry3 := entries[2].(map[string]interface{})
	assert.Equal(t, submission3, entry3["submissionId"], "Third place: submission3")
	assert.InDelta(t, 5.33, entry3["averageScore"], 0.1, "Average should be ~5.33")

	_, err := tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s SET result_published_at = NOW() WHERE id = $1", tc.HackathonDBName),
		hackathonID,
	)
	require.NoError(t, err, "Failed to publish results")

	resp, respBody = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/judging/my-result", hackathonID), participant1.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Failed to get result for participant1: %s", string(respBody))

	resultResp := tc.ParseJSON(respBody)

	result := resultResp["result"].(map[string]interface{})
	assert.Equal(t, float64(1), result["rank"], "Participant1 should be rank 1")
}

func createHackathonInJudgingStage(tc *TestContext, owner *UserCredentials) string {
	now := time.Now()
	hackathonBody := map[string]interface{}{
		"name":              fmt.Sprintf("Judging Test Hackathon %s", uuid.New().String()[:8]),
		"short_description": "Test hackathon for judging",
		"description":       "Full description for judging testing",
		"location": map[string]interface{}{
			"online": true,
		},
		"dates": map[string]interface{}{
			"registration_opens_at":  now.Add(1 * time.Hour).Format(time.RFC3339),
			"registration_closes_at": now.Add(15 * 24 * time.Hour).Format(time.RFC3339),
			"starts_at":              now.Add(20 * 24 * time.Hour).Format(time.RFC3339),
			"ends_at":                now.Add(22 * 24 * time.Hour).Format(time.RFC3339),
			"judging_ends_at":        now.Add(25 * 24 * time.Hour).Format(time.RFC3339),
		},
		"registration_policy": map[string]interface{}{
			"allow_individual": true,
			"allow_team":       true,
		},
		"limits": map[string]interface{}{
			"team_size_max": 5,
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons", owner.AccessToken, hackathonBody)
	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to create hackathon: %s", string(body))

	data := tc.ParseJSON(body)
	hackathonID := data["hackathonId"].(string)

	tc.WaitForHackathonOwnerRole(hackathonID, owner.AccessToken)

	taskBody := map[string]interface{}{
		"task": "Build something innovative",
	}
	resp, body = tc.DoAuthenticatedRequest("PUT", fmt.Sprintf("/v1/hackathons/%s/task", hackathonID), owner.AccessToken, taskBody)
	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to set task: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/publish", hackathonID), owner.AccessToken, map[string]interface{}{})
	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to publish hackathon: %s", string(body))

	_, err := tc.DB.Exec(context.Background(), fmt.Sprintf(`
		UPDATE %s 
		SET ends_at = $1,
		    stage = 'judging'
		WHERE id = $2
	`, tc.HackathonDBName), now.Add(-1*time.Hour), hackathonID)
	require.NoError(tc.T, err, "Failed to update hackathon to JUDGING stage")

	time.Sleep(500 * time.Millisecond)

	return hackathonID
}

func createSubmissionForJudging(tc *TestContext, hackathonID string, participant *UserCredentials) string {
	_, err := tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s SET stage = 'running' WHERE id = $1", tc.HackathonDBName),
		hackathonID,
	)
	require.NoError(tc.T, err, "Failed to update hackathon stage to running")

	time.Sleep(300 * time.Millisecond)

	body := map[string]interface{}{
		"title":       fmt.Sprintf("Test Submission %s", uuid.New().String()[:8]),
		"description": "Test submission for judging",
	}

	resp, respBody := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions", hackathonID), participant.AccessToken, body)
	require.Equal(tc.T, http.StatusOK, resp.StatusCode,
		"Failed to create submission: %s", string(respBody))

	data := tc.ParseJSON(respBody)
	submissionID := data["submissionId"].(string)

	_, err = tc.SubmissionDB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.submissions SET is_final = true WHERE id = $1", tc.SubmissionDBName),
		submissionID,
	)
	require.NoError(tc.T, err, "Failed to mark submission as final")

	_, err = tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s SET stage = 'judging' WHERE id = $1", tc.HackathonDBName),
		hackathonID,
	)
	require.NoError(tc.T, err, "Failed to update hackathon stage back to judging")

	time.Sleep(300 * time.Millisecond)

	return submissionID
}

func addJudgeToHackathon(tc *TestContext, hackathonID string, judge *UserCredentials) {
	_, err := tc.ParticipationDB.Exec(context.Background(),
		fmt.Sprintf("INSERT INTO %s (hackathon_id, user_id, role) VALUES ($1, $2, 'judge') ON CONFLICT DO NOTHING", tc.ParticipationDBName),
		hackathonID, judge.UserID,
	)
	require.NoError(tc.T, err, "Failed to add judge")
}
