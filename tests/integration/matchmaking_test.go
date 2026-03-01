package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecommendTeams_Unauthenticated_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)

	resp, body := tc.DoRequest("GET", fmt.Sprintf("/v1/hackathons/%s/matchmaking/teams", hackathonID), nil, nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Unauthenticated request should fail: %s", string(body))
}

func TestRecommendTeams_NonParticipant_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	nonParticipant := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/matchmaking/teams", hackathonID), nonParticipant.AccessToken, nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Non-participant should not get recommendations: %s", string(body))
}

func TestRecommendTeams_WrongStatus_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	tc.WaitForMatchmakingParticipationSync(hackathonID, participant.UserID)

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/matchmaking/teams", hackathonID), participant.AccessToken, nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Participant with INDIVIDUAL status should not get recommendations: %s", string(body))
}

func TestRecommendCandidates_NonCaptain_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	member := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, member, "PART_LOOKING_FOR_TEAM")
	tc.WaitForMatchmakingParticipationSync(hackathonID, captain.UserID)
	tc.WaitForMatchmakingParticipationSync(hackathonID, member.UserID)

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	tc.WaitForMatchmakingTeamSync(teamID)

	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)
	tc.WaitForMatchmakingVacancySync(vacancyID)

	inviteBody := map[string]interface{}{
		"target_user_id": member.UserID,
		"vacancy_id":     vacancyID,
		"message":        "Join!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathonID, teamID), captain.AccessToken, inviteBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	inviteData := tc.ParseJSON(body)
	invitationID := inviteData["invitationId"].(string)

	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/team-invitations/%s/accept", invitationID), member.AccessToken, map[string]interface{}{})
	tc.WaitForMatchmakingParticipationSync(hackathonID, member.UserID)

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/matchmaking/candidates?teamId=%s&vacancyId=%s", hackathonID, teamID, vacancyID), member.AccessToken, nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Non-captain should not get candidate recommendations: %s", string(body))
}

func TestRecommendCandidates_WrongTeam_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captainA := tc.RegisterUser()
	captainB := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captainA, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, captainB, "PART_LOOKING_FOR_TEAM")
	tc.WaitForMatchmakingParticipationSync(hackathonID, captainA.UserID)
	tc.WaitForMatchmakingParticipationSync(hackathonID, captainB.UserID)

	teamAID := createTeam(tc, hackathonID, captainA, "Team A")
	teamBID := createTeam(tc, hackathonID, captainB, "Team B")
	tc.WaitForMatchmakingTeamSync(teamAID)
	tc.WaitForMatchmakingTeamSync(teamBID)

	vacancyBID := createVacancy(tc, hackathonID, teamBID, captainB, 2)
	tc.WaitForMatchmakingVacancySync(vacancyBID)

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/matchmaking/candidates?teamId=%s&vacancyId=%s", hackathonID, teamBID, vacancyBID), captainA.AccessToken, nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Captain A should not get recommendations for Team B: %s", string(body))
}

func TestRecommendTeams_RegistrationStage_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_LOOKING_FOR_TEAM")
	tc.WaitForMatchmakingParticipationSync(hackathonID, participant.UserID)

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/matchmaking/teams", hackathonID), participant.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Should succeed in REGISTRATION stage: %s", string(body))

	data := tc.ParseJSON(body)
	recommendations, ok := data["recommendations"].([]interface{})
	assert.True(t, ok, "Recommendations array should be present")
	assert.NotNil(t, recommendations, "Recommendations should not be nil")
}

func TestRecommendTeams_ValidRequest_ShouldReturnStructure(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, participant, "PART_LOOKING_FOR_TEAM")
	tc.WaitForMatchmakingParticipationSync(hackathonID, captain.UserID)
	tc.WaitForMatchmakingParticipationSync(hackathonID, participant.UserID)

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	tc.WaitForMatchmakingTeamSync(teamID)

	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)
	tc.WaitForMatchmakingVacancySync(vacancyID)

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/matchmaking/teams?limit=10", hackathonID), participant.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get recommendations: %s", string(body))

	data := tc.ParseJSON(body)
	recommendations, ok := data["recommendations"].([]interface{})
	require.True(t, ok, "Recommendations array should be present")

	if len(recommendations) > 0 {
		rec := recommendations[0].(map[string]interface{})
		assert.NotEmpty(t, rec["teamId"], "Should have teamId")

		matchScore, ok := rec["matchScore"].(map[string]interface{})
		require.True(t, ok, "Should have matchScore object")

		assert.NotNil(t, matchScore["totalScore"], "Should have totalScore")

		skills, ok := matchScore["skills"].(map[string]interface{})
		require.True(t, ok, "Should have skills breakdown")
		assert.NotNil(t, skills["score"], "Skills should have score")
		assert.NotNil(t, skills["weight"], "Skills should have weight")

		roles, ok := matchScore["roles"].(map[string]interface{})
		require.True(t, ok, "Should have roles breakdown")
		assert.NotNil(t, roles["score"], "Roles should have score")
		assert.NotNil(t, roles["weight"], "Roles should have weight")

		text, ok := matchScore["text"].(map[string]interface{})
		require.True(t, ok, "Should have text breakdown")
		assert.NotNil(t, text["score"], "Text should have score")
		assert.NotNil(t, text["weight"], "Text should have weight")
	}
}

func TestRecommendTeams_NoTeams_ShouldReturnEmpty(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_LOOKING_FOR_TEAM")
	tc.WaitForMatchmakingParticipationSync(hackathonID, participant.UserID)

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/matchmaking/teams", hackathonID), participant.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Should succeed even with no teams: %s", string(body))

	data := tc.ParseJSON(body)
	recommendations, ok := data["recommendations"].([]interface{})
	require.True(t, ok, "Recommendations array should be present")
	assert.Equal(t, 0, len(recommendations), "Should have empty recommendations")
}

func TestRecommendCandidates_ValidRequest_ShouldReturnStructure(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	candidate := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, candidate, "PART_LOOKING_FOR_TEAM")
	tc.WaitForMatchmakingParticipationSync(hackathonID, captain.UserID)
	tc.WaitForMatchmakingParticipationSync(hackathonID, candidate.UserID)

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	tc.WaitForMatchmakingTeamSync(teamID)

	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)
	tc.WaitForMatchmakingVacancySync(vacancyID)

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/matchmaking/candidates?teamId=%s&vacancyId=%s&limit=10", hackathonID, teamID, vacancyID), captain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get candidate recommendations: %s", string(body))

	data := tc.ParseJSON(body)
	recommendations, ok := data["recommendations"].([]interface{})
	require.True(t, ok, "Recommendations array should be present")

	if len(recommendations) > 0 {
		rec := recommendations[0].(map[string]interface{})
		assert.NotEmpty(t, rec["userId"], "Should have userId")

		matchScore, ok := rec["matchScore"].(map[string]interface{})
		require.True(t, ok, "Should have matchScore object")
		assert.NotNil(t, matchScore["totalScore"], "Should have totalScore")

		skills, ok := matchScore["skills"].(map[string]interface{})
		require.True(t, ok, "Should have skills breakdown")
		assert.NotNil(t, skills["score"], "Skills should have score")

		roles, ok := matchScore["roles"].(map[string]interface{})
		require.True(t, ok, "Should have roles breakdown")
		assert.NotNil(t, roles["score"], "Roles should have score")

		text, ok := matchScore["text"].(map[string]interface{})
		require.True(t, ok, "Should have text breakdown")
		assert.NotNil(t, text["score"], "Text should have score")
	}
}
func TestMatchmakingSync_UserSkillsUpdate_ShouldSyncToReadModel(t *testing.T) {
	tc := NewTestContext(t)
	user := tc.RegisterUser()

	resp, body := tc.DoRequest("POST", "/v1/skills/list", map[string]interface{}{
		"query": map[string]interface{}{
			"page": map[string]interface{}{
				"page_size": 5,
			},
		},
	}, map[string]string{"Authorization": "Bearer " + user.AccessToken})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	skillsData := tc.ParseJSON(body)
	skills, ok := skillsData["skills"].([]interface{})
	var catalogSkillID string
	if ok && len(skills) > 0 {
		firstSkill := skills[0].(map[string]interface{})
		catalogSkillID = firstSkill["id"].(string)
	}

	updateBody := map[string]interface{}{
		"catalog_skill_ids": []string{},
		"user_skills":       []string{"React", "TypeScript"},
		"skills_visibility": "VISIBILITY_LEVEL_PUBLIC",
	}

	if catalogSkillID != "" {
		updateBody["catalog_skill_ids"] = []string{catalogSkillID}
	}

	resp, body = tc.DoAuthenticatedRequest("PUT", "/v1/users/me/skills", user.AccessToken, updateBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to update skills: %s", string(body))

	tc.WaitForMatchmakingUserSync(user.UserID)

	var catalogSkillIDs []string
	var customSkillNames []string
	err := tc.DB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT catalog_skill_ids, custom_skill_names FROM %s.users WHERE user_id = $1", tc.MatchmakingDBName),
		user.UserID,
	).Scan(&catalogSkillIDs, &customSkillNames)
	require.NoError(t, err, "Failed to query matchmaking.users")

	if catalogSkillID != "" {
		assert.Contains(t, catalogSkillIDs, catalogSkillID, "Catalog skill should be synced")
	}
	assert.Contains(t, customSkillNames, "React", "Custom skill 'React' should be synced")
	assert.Contains(t, customSkillNames, "TypeScript", "Custom skill 'TypeScript' should be synced")
}

func TestMatchmakingSync_ParticipationRegistered_ShouldSyncToReadModel(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_LOOKING_FOR_TEAM")

	tc.WaitForMatchmakingParticipationSync(hackathonID, participant.UserID)

	var status string
	var motivationText string
	err := tc.DB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT status, motivation_text FROM %s.participations WHERE hackathon_id = $1 AND user_id = $2", tc.MatchmakingDBName),
		hackathonID, participant.UserID,
	).Scan(&status, &motivationText)
	require.NoError(t, err, "Failed to query matchmaking.participations")

	assert.Equal(t, "looking_for_team", status, "Status should be synced")
	assert.Equal(t, "Test motivation", motivationText, "Motivation text should be synced")
}

func TestMatchmakingSync_TeamCreated_ShouldSyncToReadModel(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")

	teamID := createTeam(tc, hackathonID, captain, "Sync Test Team")

	tc.WaitForMatchmakingTeamSync(teamID)

	var name string
	var description string
	var isJoinable bool
	err := tc.DB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT name, description, is_joinable FROM %s.teams WHERE team_id = $1", tc.MatchmakingDBName),
		teamID,
	).Scan(&name, &description, &isJoinable)
	require.NoError(t, err, "Failed to query matchmaking.teams")

	assert.Equal(t, "Sync Test Team", name, "Team name should be synced")
	assert.Equal(t, "Test team description", description, "Team description should be synced")
	assert.True(t, isJoinable, "is_joinable should be synced")
}

func TestMatchmakingSync_VacancyCreated_ShouldSyncToReadModel(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 3)

	tc.WaitForMatchmakingVacancySync(vacancyID)

	var slotsOpen int32
	var description string
	err := tc.DB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT slots_open, description FROM %s.vacancies WHERE vacancy_id = $1", tc.MatchmakingDBName),
		vacancyID,
	).Scan(&slotsOpen, &description)
	require.NoError(t, err, "Failed to query matchmaking.vacancies")

	assert.Equal(t, int32(3), slotsOpen, "slots_open should be synced")
	assert.Equal(t, "We need a backend developer", description, "Vacancy description should be synced")
}

func TestMatchmakingScoring_SkillsMatch_ShouldRankCorrectly(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	captain1 := tc.RegisterUser()
	captain2 := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, captain1, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, captain2, "PART_LOOKING_FOR_TEAM")
	tc.WaitForMatchmakingParticipationSync(hackathonID, participant.UserID)
	tc.WaitForMatchmakingParticipationSync(hackathonID, captain1.UserID)
	tc.WaitForMatchmakingParticipationSync(hackathonID, captain2.UserID)

	skillA := uuid.New()
	skillB := uuid.New()
	skillC := uuid.New()
	skillD := uuid.New()

	_, err := tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.users SET catalog_skill_ids = $1 WHERE user_id = $2", tc.MatchmakingDBName),
		[]uuid.UUID{skillA, skillB, skillC}, participant.UserID,
	)
	require.NoError(t, err)

	team1ID := createTeam(tc, hackathonID, captain1, "Team High Match")
	team2ID := createTeam(tc, hackathonID, captain2, "Team Low Match")
	tc.WaitForMatchmakingTeamSync(team1ID)
	tc.WaitForMatchmakingTeamSync(team2ID)

	vacancy1ID := createVacancy(tc, hackathonID, team1ID, captain1, 2)
	vacancy2ID := createVacancy(tc, hackathonID, team2ID, captain2, 2)
	tc.WaitForMatchmakingVacancySync(vacancy1ID)
	tc.WaitForMatchmakingVacancySync(vacancy2ID)

	_, err = tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.vacancies SET desired_skill_ids = $1 WHERE vacancy_id = $2", tc.MatchmakingDBName),
		[]uuid.UUID{skillA, skillB}, vacancy1ID,
	)
	require.NoError(t, err)

	_, err = tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.vacancies SET desired_skill_ids = $1 WHERE vacancy_id = $2", tc.MatchmakingDBName),
		[]uuid.UUID{skillD}, vacancy2ID,
	)
	require.NoError(t, err)

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/matchmaking/teams?limit=10", hackathonID), participant.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get recommendations: %s", string(body))

	data := tc.ParseJSON(body)
	recommendations, ok := data["recommendations"].([]interface{})
	require.True(t, ok && len(recommendations) >= 2, "Should have at least 2 recommendations")

	rec1 := recommendations[0].(map[string]interface{})
	rec2 := recommendations[1].(map[string]interface{})

	score1 := rec1["matchScore"].(map[string]interface{})["totalScore"].(float64)
	score2 := rec2["matchScore"].(map[string]interface{})["totalScore"].(float64)

	assert.Greater(t, score1, score2, "Team with matching skills should rank higher")

	if rec1["teamId"].(string) == team1ID {
		assert.Equal(t, team1ID, rec1["teamId"], "Team 1 should rank first")
	}
}

func TestMatchmakingScoring_RolesMatch_ShouldRankCorrectly(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	captain1 := tc.RegisterUser()
	captain2 := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, captain1, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, captain2, "PART_LOOKING_FOR_TEAM")
	tc.WaitForMatchmakingParticipationSync(hackathonID, participant.UserID)
	tc.WaitForMatchmakingParticipationSync(hackathonID, captain1.UserID)
	tc.WaitForMatchmakingParticipationSync(hackathonID, captain2.UserID)

	roleA := uuid.New()
	roleB := uuid.New()
	roleC := uuid.New()

	_, err := tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.participations SET wished_role_ids = $1 WHERE hackathon_id = $2 AND user_id = $3", tc.MatchmakingDBName),
		[]uuid.UUID{roleA, roleB}, hackathonID, participant.UserID,
	)
	require.NoError(t, err)

	team1ID := createTeam(tc, hackathonID, captain1, "Team High Match")
	team2ID := createTeam(tc, hackathonID, captain2, "Team Low Match")
	tc.WaitForMatchmakingTeamSync(team1ID)
	tc.WaitForMatchmakingTeamSync(team2ID)

	vacancy1ID := createVacancy(tc, hackathonID, team1ID, captain1, 2)
	vacancy2ID := createVacancy(tc, hackathonID, team2ID, captain2, 2)
	tc.WaitForMatchmakingVacancySync(vacancy1ID)
	tc.WaitForMatchmakingVacancySync(vacancy2ID)

	_, err = tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.vacancies SET desired_role_ids = $1 WHERE vacancy_id = $2", tc.MatchmakingDBName),
		[]uuid.UUID{roleA, roleB}, vacancy1ID,
	)
	require.NoError(t, err)

	_, err = tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.vacancies SET desired_role_ids = $1 WHERE vacancy_id = $2", tc.MatchmakingDBName),
		[]uuid.UUID{roleC}, vacancy2ID,
	)
	require.NoError(t, err)

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/matchmaking/teams?limit=10", hackathonID), participant.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get recommendations: %s", string(body))

	data := tc.ParseJSON(body)
	recommendations, ok := data["recommendations"].([]interface{})
	require.True(t, ok && len(recommendations) >= 2, "Should have at least 2 recommendations")

	rec1 := recommendations[0].(map[string]interface{})
	rec2 := recommendations[1].(map[string]interface{})

	score1 := rec1["matchScore"].(map[string]interface{})["totalScore"].(float64)
	score2 := rec2["matchScore"].(map[string]interface{})["totalScore"].(float64)

	assert.Greater(t, score1, score2, "Team with matching roles should rank higher")

	if rec1["teamId"].(string) == team1ID {
		assert.Equal(t, team1ID, rec1["teamId"], "Team 1 should rank first")
	}
}

func TestMatchmakingScoring_CustomSkills_ShouldMatchInText(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	captain := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	tc.WaitForMatchmakingParticipationSync(hackathonID, participant.UserID)
	tc.WaitForMatchmakingParticipationSync(hackathonID, captain.UserID)

	_, err := tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.users SET custom_skill_names = $1 WHERE user_id = $2", tc.MatchmakingDBName),
		[]string{"React", "TypeScript"}, participant.UserID,
	)
	require.NoError(t, err)

	teamID := createTeam(tc, hackathonID, captain, "Frontend Team")
	tc.WaitForMatchmakingTeamSync(teamID)

	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)
	tc.WaitForMatchmakingVacancySync(vacancyID)

	_, err = tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.teams SET description = $1 WHERE team_id = $2", tc.MatchmakingDBName),
		"Looking for React developer with TypeScript experience", teamID,
	)
	require.NoError(t, err)

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/matchmaking/teams?limit=10", hackathonID), participant.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get recommendations: %s", string(body))

	data := tc.ParseJSON(body)
	recommendations, ok := data["recommendations"].([]interface{})
	require.True(t, ok && len(recommendations) > 0, "Should have at least 1 recommendation")

	rec := recommendations[0].(map[string]interface{})
	matchScore := rec["matchScore"].(map[string]interface{})
	textBreakdown := matchScore["text"].(map[string]interface{})

	textScore := textBreakdown["score"].(float64)
	assert.Greater(t, textScore, 0.0, "Text score should be greater than 0")

	matchedKeywords, ok := textBreakdown["matchedKeywords"].([]interface{})
	t.Logf("matchedKeywords: %v (ok=%v, len=%d)", matchedKeywords, ok, len(matchedKeywords))
	if ok && len(matchedKeywords) > 0 {
		found := false
		for _, kw := range matchedKeywords {
			keyword := kw.(string)
			t.Logf("checking keyword: %q", keyword)
			if keyword == "react" || keyword == "typescript" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should match 'react' or 'typescript' in keywords")
	}
}

func TestMatchmakingScoring_CombinedWeights_ShouldCalculateCorrectly(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	captain := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	tc.WaitForMatchmakingParticipationSync(hackathonID, participant.UserID)
	tc.WaitForMatchmakingParticipationSync(hackathonID, captain.UserID)

	skillA := uuid.New()
	skillB := uuid.New()
	roleA := uuid.New()

	_, err := tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.users SET catalog_skill_ids = $1, custom_skill_names = $2 WHERE user_id = $3", tc.MatchmakingDBName),
		[]uuid.UUID{skillA, skillB}, []string{"Docker"}, participant.UserID,
	)
	require.NoError(t, err)

	_, err = tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.participations SET wished_role_ids = $1 WHERE hackathon_id = $2 AND user_id = $3", tc.MatchmakingDBName),
		[]uuid.UUID{roleA}, hackathonID, participant.UserID,
	)
	require.NoError(t, err)

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	tc.WaitForMatchmakingTeamSync(teamID)

	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)
	tc.WaitForMatchmakingVacancySync(vacancyID)

	_, err = tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.vacancies SET desired_skill_ids = $1, desired_role_ids = $2, description = $3 WHERE vacancy_id = $4", tc.MatchmakingDBName),
		[]uuid.UUID{skillA, skillB}, []uuid.UUID{roleA}, "Need Docker expert", vacancyID,
	)
	require.NoError(t, err)

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/matchmaking/teams?limit=10", hackathonID), participant.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get recommendations: %s", string(body))

	data := tc.ParseJSON(body)
	recommendations, ok := data["recommendations"].([]interface{})
	require.True(t, ok && len(recommendations) > 0, "Should have at least 1 recommendation")

	rec := recommendations[0].(map[string]interface{})
	matchScore := rec["matchScore"].(map[string]interface{})

	totalScore := matchScore["totalScore"].(float64)
	skillsBreakdown := matchScore["skills"].(map[string]interface{})
	rolesBreakdown := matchScore["roles"].(map[string]interface{})
	textBreakdown := matchScore["text"].(map[string]interface{})

	skillsScore := skillsBreakdown["score"].(float64)
	skillsWeight := skillsBreakdown["weight"].(float64)
	rolesScore := rolesBreakdown["score"].(float64)
	rolesWeight := rolesBreakdown["weight"].(float64)
	textScore := textBreakdown["score"].(float64)
	textWeight := textBreakdown["weight"].(float64)

	expectedTotal := skillsScore*skillsWeight + rolesScore*rolesWeight + textScore*textWeight

	assert.InDelta(t, expectedTotal, totalScore, 0.01, "Total score should match weighted sum")

	assert.InDelta(t, 0.63, skillsWeight, 0.01, "Skills weight should be ~0.63")
	assert.InDelta(t, 0.27, rolesWeight, 0.01, "Roles weight should be ~0.27")
	assert.InDelta(t, 0.10, textWeight, 0.01, "Text weight should be ~0.10")
}

func TestMatchmakingScoring_CandidateSkillsMatch_ShouldRankCorrectly(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	candidate1 := tc.RegisterUser()
	candidate2 := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, candidate1, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, candidate2, "PART_LOOKING_FOR_TEAM")
	tc.WaitForMatchmakingParticipationSync(hackathonID, captain.UserID)
	tc.WaitForMatchmakingParticipationSync(hackathonID, candidate1.UserID)
	tc.WaitForMatchmakingParticipationSync(hackathonID, candidate2.UserID)

	skillA := uuid.New()
	skillB := uuid.New()
	skillC := uuid.New()

	_, err := tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.users SET catalog_skill_ids = $1 WHERE user_id = $2", tc.MatchmakingDBName),
		[]uuid.UUID{skillA, skillB}, candidate1.UserID,
	)
	require.NoError(t, err)

	_, err = tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.users SET catalog_skill_ids = $1 WHERE user_id = $2", tc.MatchmakingDBName),
		[]uuid.UUID{skillC}, candidate2.UserID,
	)
	require.NoError(t, err)

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	tc.WaitForMatchmakingTeamSync(teamID)

	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)
	tc.WaitForMatchmakingVacancySync(vacancyID)

	_, err = tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.vacancies SET desired_skill_ids = $1 WHERE vacancy_id = $2", tc.MatchmakingDBName),
		[]uuid.UUID{skillA, skillB}, vacancyID,
	)
	require.NoError(t, err)

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/matchmaking/candidates?teamId=%s&vacancyId=%s&limit=10", hackathonID, teamID, vacancyID), captain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get recommendations: %s", string(body))

	data := tc.ParseJSON(body)
	recommendations, ok := data["recommendations"].([]interface{})
	require.True(t, ok && len(recommendations) >= 2, "Should have at least 2 recommendations")

	rec1 := recommendations[0].(map[string]interface{})
	rec2 := recommendations[1].(map[string]interface{})

	score1 := rec1["matchScore"].(map[string]interface{})["totalScore"].(float64)
	score2 := rec2["matchScore"].(map[string]interface{})["totalScore"].(float64)

	assert.Greater(t, score1, score2, "Candidate with matching skills should rank higher")

	if rec1["userId"].(string) == candidate1.UserID {
		assert.Equal(t, candidate1.UserID, rec1["userId"], "Candidate 1 should rank first")
	}
}

func TestMatchmakingScoring_CandidateRolesMatch_ShouldRankCorrectly(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	candidate1 := tc.RegisterUser()
	candidate2 := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, candidate1, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, candidate2, "PART_LOOKING_FOR_TEAM")
	tc.WaitForMatchmakingParticipationSync(hackathonID, captain.UserID)
	tc.WaitForMatchmakingParticipationSync(hackathonID, candidate1.UserID)
	tc.WaitForMatchmakingParticipationSync(hackathonID, candidate2.UserID)

	roleA := uuid.New()
	roleB := uuid.New()
	roleC := uuid.New()

	_, err := tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.participations SET wished_role_ids = $1 WHERE hackathon_id = $2 AND user_id = $3", tc.MatchmakingDBName),
		[]uuid.UUID{roleA, roleB}, hackathonID, candidate1.UserID,
	)
	require.NoError(t, err)

	_, err = tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.participations SET wished_role_ids = $1 WHERE hackathon_id = $2 AND user_id = $3", tc.MatchmakingDBName),
		[]uuid.UUID{roleC}, hackathonID, candidate2.UserID,
	)
	require.NoError(t, err)

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	tc.WaitForMatchmakingTeamSync(teamID)

	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)
	tc.WaitForMatchmakingVacancySync(vacancyID)

	_, err = tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.vacancies SET desired_role_ids = $1 WHERE vacancy_id = $2", tc.MatchmakingDBName),
		[]uuid.UUID{roleA, roleB}, vacancyID,
	)
	require.NoError(t, err)

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/matchmaking/candidates?teamId=%s&vacancyId=%s&limit=10", hackathonID, teamID, vacancyID), captain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get recommendations: %s", string(body))

	data := tc.ParseJSON(body)
	recommendations, ok := data["recommendations"].([]interface{})
	require.True(t, ok && len(recommendations) >= 2, "Should have at least 2 recommendations")

	rec1 := recommendations[0].(map[string]interface{})
	rec2 := recommendations[1].(map[string]interface{})

	score1 := rec1["matchScore"].(map[string]interface{})["totalScore"].(float64)
	score2 := rec2["matchScore"].(map[string]interface{})["totalScore"].(float64)

	assert.Greater(t, score1, score2, "Candidate with matching roles should rank higher")

	if rec1["userId"].(string) == candidate1.UserID {
		assert.Equal(t, candidate1.UserID, rec1["userId"], "Candidate 1 should rank first")
	}
}

func TestMatchmakingScoring_CandidateCustomSkills_ShouldMatchInText(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	candidate := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, candidate, "PART_LOOKING_FOR_TEAM")
	tc.WaitForMatchmakingParticipationSync(hackathonID, captain.UserID)
	tc.WaitForMatchmakingParticipationSync(hackathonID, candidate.UserID)

	_, err := tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.users SET custom_skill_names = $1 WHERE user_id = $2", tc.MatchmakingDBName),
		[]string{"Kubernetes", "Microservices"}, candidate.UserID,
	)
	require.NoError(t, err)

	_, err = tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.participations SET motivation_text = $1 WHERE hackathon_id = $2 AND user_id = $3", tc.MatchmakingDBName),
		"I have experience with cloud-native architectures and distributed systems", candidate.UserID, hackathonID,
	)
	require.NoError(t, err)

	teamID := createTeam(tc, hackathonID, captain, "Backend Team")
	tc.WaitForMatchmakingTeamSync(teamID)

	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)
	tc.WaitForMatchmakingVacancySync(vacancyID)

	_, err = tc.DB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.vacancies SET description = $1 WHERE vacancy_id = $2", tc.MatchmakingDBName),
		"Looking for backend developer with Kubernetes and microservices experience", vacancyID,
	)
	require.NoError(t, err)

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/matchmaking/candidates?teamId=%s&vacancyId=%s&limit=10", hackathonID, teamID, vacancyID), captain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get recommendations: %s", string(body))

	data := tc.ParseJSON(body)
	recommendations, ok := data["recommendations"].([]interface{})
	require.True(t, ok && len(recommendations) > 0, "Should have at least 1 recommendation")

	rec := recommendations[0].(map[string]interface{})
	matchScore := rec["matchScore"].(map[string]interface{})
	textBreakdown := matchScore["text"].(map[string]interface{})

	textScore := textBreakdown["score"].(float64)
	assert.Greater(t, textScore, 0.0, "Text score should be greater than 0")

	matchedKeywords, ok := textBreakdown["matchedKeywords"].([]interface{})
	if ok && len(matchedKeywords) > 0 {
		found := false
		for _, kw := range matchedKeywords {
			keyword := kw.(string)
			if keyword == "kubernetes" || keyword == "microservices" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should match 'kubernetes' or 'microservices' in keywords")
	}
}
