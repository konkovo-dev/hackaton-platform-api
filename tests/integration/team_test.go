package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// 1) Teams CRUD
// ============================================================================

func TestListTeams_AsParticipant_ShouldReturnTeams(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain1 := tc.RegisterUser()
	captain2 := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain1, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, captain2, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	team1ID := createTeam(tc, hackathonID, captain1, "Team Alpha")
	createTeam(tc, hackathonID, captain2, "Team Beta")

	createVacancy(tc, hackathonID, team1ID, captain1, 2)

	listBody := map[string]interface{}{
		"include_vacancies": false,
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/list", hackathonID), participant.AccessToken, listBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list teams: %s", string(body))

	data := tc.ParseJSON(body)
	teams, ok := data["teams"].([]interface{})
	require.True(t, ok, "Teams array should be present")
	assert.GreaterOrEqual(t, len(teams), 2, "Should have at least 2 teams")

	for _, teamData := range teams {
		teamWithVacancies := teamData.(map[string]interface{})
		team := teamWithVacancies["team"].(map[string]interface{})
		assert.NotEmpty(t, team["teamId"], "Team should have ID")
		assert.NotEmpty(t, team["hackathonId"], "Team should have hackathon ID")
		assert.NotEmpty(t, team["name"], "Team should have name")
	}

	listBodyWithVacancies := map[string]interface{}{
		"include_vacancies": true,
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/list", hackathonID), participant.AccessToken, listBodyWithVacancies)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list teams with vacancies: %s", string(body))

	dataWithVacancies := tc.ParseJSON(body)
	teamsWithVacancies := dataWithVacancies["teams"].([]interface{})

	foundTeamWithVacancies := false
	for _, teamData := range teamsWithVacancies {
		teamWithVacancies := teamData.(map[string]interface{})
		team := teamWithVacancies["team"].(map[string]interface{})
		if team["teamId"] == team1ID {
			vacancies, ok := teamWithVacancies["vacancies"].([]interface{})
			if ok && len(vacancies) > 0 {
				foundTeamWithVacancies = true
				vacancy := vacancies[0].(map[string]interface{})
				assert.NotEmpty(t, vacancy["vacancyId"], "Vacancy should have ID")
				assert.NotNil(t, vacancy["slotsTotal"], "Vacancy should have slotsTotal")
				assert.NotNil(t, vacancy["slotsOpen"], "Vacancy should have slotsOpen")
			}
		}
	}
	assert.True(t, foundTeamWithVacancies, "Should find team with vacancies")
}

func TestListTeams_AsNonParticipant_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	nonParticipant := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	createTeam(tc, hackathonID, captain, "Test Team")

	listBody := map[string]interface{}{}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/list", hackathonID), nonParticipant.AccessToken, listBody)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Non-participant should not list teams: %s", string(body))
}

func TestGetTeam_ShouldReturnTeam(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 3)

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s?include_vacancies=false", hackathonID, teamID), participant.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get team: %s", string(body))

	data := tc.ParseJSON(body)
	teamWithVacancies := data["team"].(map[string]interface{})
	team := teamWithVacancies["team"].(map[string]interface{})
	assert.Equal(t, teamID, team["teamId"])
	assert.Equal(t, hackathonID, team["hackathonId"])
	assert.Equal(t, "Test Team", team["name"])

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s?include_vacancies=true", hackathonID, teamID), participant.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get team with vacancies: %s", string(body))

	dataWithVacancies := tc.ParseJSON(body)
	teamWithVacanciesData := dataWithVacancies["team"].(map[string]interface{})
	vacancies, ok := teamWithVacanciesData["vacancies"].([]interface{})
	require.True(t, ok && len(vacancies) > 0, "Should have vacancies")

	vacancy := vacancies[0].(map[string]interface{})
	assert.Equal(t, vacancyID, vacancy["vacancyId"])

	assert.Equal(t, "3", vacancy["slotsTotal"])
	assert.Equal(t, "3", vacancy["slotsOpen"])
}

func TestGetTeam_WrongHackathon_ShouldNotLeak(t *testing.T) {
	tc := NewTestContext(t)
	owner1 := tc.RegisterUser()
	owner2 := tc.RegisterUser()
	captain := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathon1ID := createHackathonInRegistration(tc, owner1)
	hackathon2ID := createHackathonInRegistration(tc, owner2)

	registerParticipant(tc, hackathon1ID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathon2ID, participant, "PART_INDIVIDUAL")

	team1ID := createTeam(tc, hackathon1ID, captain, "Team in Hackathon 1")

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s", hackathon2ID, team1ID), participant.AccessToken, nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Should not leak team from another hackathon: %s", string(body))

	fakeTeamID := "00000000-0000-0000-0000-000000000000"
	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s", hackathon1ID, fakeTeamID), participant.AccessToken, nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Should return 403 when not participant: %s", string(body))
}

func TestCreateTeam_FromLookingForTeam_ShouldBecomeCaptain(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")

	teamBody := map[string]interface{}{
		"name":        "New Team",
		"description": "Test team",
		"is_joinable": true,
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams", hackathonID), captain.AccessToken, teamBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to create team: %s", string(body))

	data := tc.ParseJSON(body)
	teamID := data["teamId"].(string)
	assert.NotEmpty(t, teamID, "Team ID should be returned")

	time.Sleep(500 * time.Millisecond)

	participation := getParticipation(tc, hackathonID, captain)
	assert.Equal(t, "PART_TEAM_CAPTAIN", participation["status"])
	assert.Equal(t, teamID, participation["teamId"])

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s/members", hackathonID, teamID), captain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get members: %s", string(body))

	membersData := tc.ParseJSON(body)
	members, ok := membersData["members"].([]interface{})
	require.True(t, ok && len(members) == 1, "Should have 1 member")

	member := members[0].(map[string]interface{})
	assert.Equal(t, captain.UserID, member["userId"])
	assert.Equal(t, true, member["isCaptain"])
}

func TestCreateTeam_DuplicateName_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain1 := tc.RegisterUser()
	captain2 := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain1, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, captain2, "PART_INDIVIDUAL")

	teamName := "Duplicate Team Name"
	createTeam(tc, hackathonID, captain1, teamName)

	teamBody := map[string]interface{}{
		"name":        teamName,
		"description": "Another team",
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams", hackathonID), captain2.AccessToken, teamBody)
	assert.Equal(t, http.StatusConflict, resp.StatusCode, "Should reject duplicate team name: %s", string(body))

	participation := getParticipation(tc, hackathonID, captain2)
	assert.Equal(t, "PART_INDIVIDUAL", participation["status"])
}

func TestUpdateTeam_AsCaptain_ShouldUpdateFields(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")

	teamID := createTeam(tc, hackathonID, captain, "Original Name")

	updateBody := map[string]interface{}{
		"name":        "Updated Name",
		"description": "Updated description",
		"is_joinable": false,
	}

	resp, body := tc.DoAuthenticatedRequest("PUT", fmt.Sprintf("/v1/hackathons/%s/teams/%s", hackathonID, teamID), captain.AccessToken, updateBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to update team: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s", hackathonID, teamID), captain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	teamWithVacancies := data["team"].(map[string]interface{})
	team := teamWithVacancies["team"].(map[string]interface{})
	assert.Equal(t, "Updated Name", team["name"])
	assert.Equal(t, "Updated description", team["description"])
	assert.Equal(t, false, team["isJoinable"])
}

func TestUpdateTeam_AsNonCaptain_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	member := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, member, "PART_LOOKING_FOR_TEAM")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)

	inviteBody := map[string]interface{}{
		"target_user_id": member.UserID,
		"vacancy_id":     vacancyID,
		"message":        "Join us!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathonID, teamID), captain.AccessToken, inviteBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	inviteData := tc.ParseJSON(body)
	invitationID := inviteData["invitationId"].(string)

	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/team-invitations/%s/accept", invitationID), member.AccessToken, map[string]interface{}{})
	time.Sleep(500 * time.Millisecond)

	updateBody := map[string]interface{}{
		"name": "Hacked Name",
	}

	resp, body = tc.DoAuthenticatedRequest("PUT", fmt.Sprintf("/v1/hackathons/%s/teams/%s", hackathonID, teamID), member.AccessToken, updateBody)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Non-captain should not update team: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s", hackathonID, teamID), captain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	teamWithVacancies := data["team"].(map[string]interface{})
	team := teamWithVacancies["team"].(map[string]interface{})
	assert.Equal(t, "Test Team", team["name"])
}

func TestDeleteTeam_SoleMemberCaptain_ShouldDeleteAndConvertParticipation(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")

	teamID := createTeam(tc, hackathonID, captain, "Team To Delete")

	resp, body := tc.DoAuthenticatedRequest("DELETE", fmt.Sprintf("/v1/hackathons/%s/teams/%s", hackathonID, teamID), captain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to delete team: %s", string(body))

	time.Sleep(500 * time.Millisecond)

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s", hackathonID, teamID), captain.AccessToken, nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Team should be deleted: %s", string(body))

	participation := getParticipation(tc, hackathonID, captain)
	assert.Equal(t, "PART_LOOKING_FOR_TEAM", participation["status"])
	teamIDVal, hasTeamID := participation["teamId"]
	teamIDStr := ""
	if hasTeamID && teamIDVal != nil {
		teamIDStr = teamIDVal.(string)
	}
	assert.True(t, !hasTeamID || teamIDStr == "", "Team ID should be empty or absent")
}

func TestDeleteTeam_WithMoreThanOneMember_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	member := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, member, "PART_LOOKING_FOR_TEAM")

	teamID := createTeam(tc, hackathonID, captain, "Team With Members")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)

	inviteBody := map[string]interface{}{
		"target_user_id": member.UserID,
		"vacancy_id":     vacancyID,
		"message":        "Join us!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathonID, teamID), captain.AccessToken, inviteBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	inviteData := tc.ParseJSON(body)
	invitationID := inviteData["invitationId"].(string)

	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/team-invitations/%s/accept", invitationID), member.AccessToken, map[string]interface{}{})
	time.Sleep(500 * time.Millisecond)

	resp, body = tc.DoAuthenticatedRequest("DELETE", fmt.Sprintf("/v1/hackathons/%s/teams/%s", hackathonID, teamID), captain.AccessToken, nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Should not delete team with multiple members: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s", hackathonID, teamID), captain.AccessToken, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Team should still exist: %s", string(body))
}

// ============================================================================
// 2) Vacancies
// ============================================================================

func TestUpsertVacancy_Create_AsCaptain_ShouldCreate(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")

	vacancyBody := map[string]interface{}{
		"vacancy_id":  "",
		"name":        "Frontend Developer",
		"description": "We need a frontend expert",
		"slots_total": 2,
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/vacancies/upsert", hackathonID, teamID), captain.AccessToken, vacancyBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to create vacancy: %s", string(body))

	data := tc.ParseJSON(body)
	vacancyID := data["vacancyId"].(string)
	assert.NotEmpty(t, vacancyID, "Vacancy ID should be returned")

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s?include_vacancies=true", hackathonID, teamID), captain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	teamData := tc.ParseJSON(body)
	team := teamData["team"].(map[string]interface{})
	vacancies := team["vacancies"].([]interface{})
	require.GreaterOrEqual(t, len(vacancies), 1, "Should have at least 1 vacancy")

	vacancy := vacancies[0].(map[string]interface{})
	assert.Equal(t, vacancyID, vacancy["vacancyId"])
	assert.Equal(t, "2", vacancy["slotsTotal"])
	assert.Equal(t, "2", vacancy["slotsOpen"])
}

func TestUpsertVacancy_Update_DecreaseBelowOccupied_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	member := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, member, "PART_LOOKING_FOR_TEAM")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)

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
	time.Sleep(500 * time.Millisecond)

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s?include_vacancies=true", hackathonID, teamID), captain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	teamData := tc.ParseJSON(body)
	team := teamData["team"].(map[string]interface{})
	vacancies := team["vacancies"].([]interface{})
	vacancy := vacancies[0].(map[string]interface{})
	assert.Equal(t, "1", vacancy["slotsOpen"], "Should have 1 open slot after member joined")

	updateVacancyBody := map[string]interface{}{
		"vacancy_id":  vacancyID,
		"name":        "Backend Developer",
		"slots_total": 0,
	}

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/vacancies:upsert", hackathonID, teamID), captain.AccessToken, updateVacancyBody)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Should not allow decreasing slots below occupied: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s?include_vacancies=true", hackathonID, teamID), captain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	teamData2 := tc.ParseJSON(body)
	team2 := teamData2["team"].(map[string]interface{})
	vacancies2 := team2["vacancies"].([]interface{})
	vacancy2 := vacancies2[0].(map[string]interface{})
	assert.Equal(t, "2", vacancy2["slotsTotal"], "Slots total should remain 2")
}

// ============================================================================
// 3) Members
// ============================================================================

func TestListTeamMembers_AsParticipant_ShouldReturnRoster(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s/members", hackathonID, teamID), participant.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list members: %s", string(body))

	data := tc.ParseJSON(body)
	members, ok := data["members"].([]interface{})
	require.True(t, ok && len(members) == 1, "Should have 1 member")

	member := members[0].(map[string]interface{})
	assert.Equal(t, captain.UserID, member["userId"])
	assert.Equal(t, true, member["isCaptain"])
	assert.NotNil(t, member["joinedAt"])
}

func TestKickMember_AsCaptain_ShouldRemoveMemberAndReopenSlot(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	member := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, member, "PART_LOOKING_FOR_TEAM")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)

	inviteBody := map[string]interface{}{
		"target_user_id": member.UserID,
		"vacancy_id":     vacancyID,
		"message":        "Join us!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathonID, teamID), captain.AccessToken, inviteBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	inviteData := tc.ParseJSON(body)
	invitationID := inviteData["invitationId"].(string)

	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/team-invitations/%s/accept", invitationID), member.AccessToken, map[string]interface{}{})
	time.Sleep(500 * time.Millisecond)

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s/members", hackathonID, teamID), captain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	membersData := tc.ParseJSON(body)
	members := membersData["members"].([]interface{})
	assert.Equal(t, 2, len(members), "Should have 2 members")

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/members/%s/kick", hackathonID, teamID, member.UserID), captain.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to kick member: %s", string(body))

	time.Sleep(500 * time.Millisecond)

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s/members", hackathonID, teamID), captain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	membersData2 := tc.ParseJSON(body)
	members2 := membersData2["members"].([]interface{})
	assert.Equal(t, 1, len(members2), "Should have 1 member after kick")

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s?include_vacancies=true", hackathonID, teamID), captain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	teamData := tc.ParseJSON(body)
	team := teamData["team"].(map[string]interface{})
	vacancies := team["vacancies"].([]interface{})
	vacancy := vacancies[0].(map[string]interface{})
	assert.Equal(t, "2", vacancy["slotsOpen"], "Slot should be reopened")

	participation := getParticipation(tc, hackathonID, member)
	assert.Equal(t, "PART_LOOKING_FOR_TEAM", participation["status"])
}

func TestLeaveTeam_AsMember_ShouldLeaveAndReopenSlot(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	member := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, member, "PART_LOOKING_FOR_TEAM")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)

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
	time.Sleep(500 * time.Millisecond)

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/members/leave", hackathonID, teamID), member.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to leave team: %s", string(body))

	time.Sleep(500 * time.Millisecond)

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s/members", hackathonID, teamID), captain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	membersData := tc.ParseJSON(body)
	members := membersData["members"].([]interface{})
	assert.Equal(t, 1, len(members), "Should have 1 member after leave")

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s?include_vacancies=true", hackathonID, teamID), captain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	teamData := tc.ParseJSON(body)
	team := teamData["team"].(map[string]interface{})
	vacancies := team["vacancies"].([]interface{})
	vacancy := vacancies[0].(map[string]interface{})
	assert.Equal(t, "2", vacancy["slotsOpen"], "Slot should be reopened")

	participation := getParticipation(tc, hackathonID, member)
	assert.Equal(t, "PART_LOOKING_FOR_TEAM", participation["status"])
}

func TestLeaveTeam_AsCaptain_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/members/leave", hackathonID, teamID), captain.AccessToken, map[string]interface{}{})
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Captain should not be able to leave: %s", string(body))

	participation := getParticipation(tc, hackathonID, captain)
	assert.Equal(t, "PART_TEAM_CAPTAIN", participation["status"])
}

func TestTransferCaptain_AsCaptain_ShouldTransferAndSwapStatuses(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	oldCaptain := tc.RegisterUser()
	newCaptain := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, oldCaptain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, newCaptain, "PART_LOOKING_FOR_TEAM")

	teamID := createTeam(tc, hackathonID, oldCaptain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, oldCaptain, 2)

	inviteBody := map[string]interface{}{
		"target_user_id": newCaptain.UserID,
		"vacancy_id":     vacancyID,
		"message":        "Join!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathonID, teamID), oldCaptain.AccessToken, inviteBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	inviteData := tc.ParseJSON(body)
	invitationID := inviteData["invitationId"].(string)

	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/team-invitations/%s/accept", invitationID), newCaptain.AccessToken, map[string]interface{}{})
	time.Sleep(500 * time.Millisecond)

	transferBody := map[string]interface{}{
		"target_user_id": newCaptain.UserID,
	}

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/transferCaptain", hackathonID, teamID), oldCaptain.AccessToken, transferBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to transfer captain: %s", string(body))

	time.Sleep(500 * time.Millisecond)

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s/members", hackathonID, teamID), newCaptain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	membersData := tc.ParseJSON(body)
	members := membersData["members"].([]interface{})

	var oldCaptainMember, newCaptainMember map[string]interface{}
	for _, m := range members {
		member := m.(map[string]interface{})
		if member["userId"] == oldCaptain.UserID {
			oldCaptainMember = member
		}
		if member["userId"] == newCaptain.UserID {
			newCaptainMember = member
		}
	}

	assert.False(t, oldCaptainMember["isCaptain"].(bool), "Old captain should not be captain")
	assert.True(t, newCaptainMember["isCaptain"].(bool), "New captain should be captain")

	oldParticipation := getParticipation(tc, hackathonID, oldCaptain)
	assert.Equal(t, "PART_TEAM_MEMBER", oldParticipation["status"])

	newParticipation := getParticipation(tc, hackathonID, newCaptain)
	assert.Equal(t, "PART_TEAM_CAPTAIN", newParticipation["status"])
}

// ============================================================================
// 4) Team Invitations
// ============================================================================

func TestListTeamInvitations_AsCaptain_ShouldReturnOutgoing(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	target := tc.RegisterUser()
	nonCaptain := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, target, "PART_INDIVIDUAL")
	registerParticipant(tc, hackathonID, nonCaptain, "PART_INDIVIDUAL")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)

	inviteBody := map[string]interface{}{
		"target_user_id": target.UserID,
		"vacancy_id":     vacancyID,
		"message":        "Join us!",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathonID, teamID), captain.AccessToken, inviteBody)

	listBody := map[string]interface{}{}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations/list", hackathonID, teamID), captain.AccessToken, listBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list invitations: %s", string(body))

	data := tc.ParseJSON(body)
	invitations, ok := data["invitations"].([]interface{})
	require.True(t, ok && len(invitations) >= 1, "Should have at least 1 invitation")

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations/list", hackathonID, teamID), nonCaptain.AccessToken, listBody)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Non-captain should not list invitations: %s", string(body))
}

func TestCreateTeamInvitation_ShouldCreatePending(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	target := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, target, "PART_INDIVIDUAL")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)

	inviteBody := map[string]interface{}{
		"target_user_id": target.UserID,
		"vacancy_id":     vacancyID,
		"message":        "We would love to have you!",
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathonID, teamID), captain.AccessToken, inviteBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to create invitation: %s", string(body))

	data := tc.ParseJSON(body)
	invitationID := data["invitationId"].(string)
	assert.NotEmpty(t, invitationID, "Invitation ID should be returned")

	listBody := map[string]interface{}{}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations/list", hackathonID, teamID), captain.AccessToken, listBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	listData := tc.ParseJSON(body)
	invitations := listData["invitations"].([]interface{})
	foundInvitation := false
	for _, inv := range invitations {
		invitation := inv.(map[string]interface{})
		if invitation["invitationId"] == invitationID {
			foundInvitation = true
			assert.Equal(t, "TEAM_INBOX_PENDING", invitation["status"])
		}
	}
	assert.True(t, foundInvitation, "Invitation should be in outgoing list")

	resp, body = tc.DoAuthenticatedRequest("POST", "/v1/users/me/team-invitations/list", target.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	targetListData := tc.ParseJSON(body)
	targetInvitations := targetListData["invitations"].([]interface{})
	foundInTarget := false
	for _, inv := range targetInvitations {
		invitation := inv.(map[string]interface{})
		if invitation["invitationId"] == invitationID {
			foundInTarget = true
			assert.Equal(t, "TEAM_INBOX_PENDING", invitation["status"])
		}
	}
	assert.True(t, foundInTarget, "Invitation should be in target's inbox")
}

func TestCancelTeamInvitation_ShouldMarkCanceled(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	target := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, target, "PART_INDIVIDUAL")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)

	inviteBody := map[string]interface{}{
		"target_user_id": target.UserID,
		"vacancy_id":     vacancyID,
		"message":        "Join!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathonID, teamID), captain.AccessToken, inviteBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	inviteData := tc.ParseJSON(body)
	invitationID := inviteData["invitationId"].(string)

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations/%s/cancel", hackathonID, teamID, invitationID), captain.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to cancel invitation: %s", string(body))

	listBody := map[string]interface{}{}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations/list", hackathonID, teamID), captain.AccessToken, listBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	listData := tc.ParseJSON(body)
	invitations := listData["invitations"].([]interface{})
	for _, inv := range invitations {
		invitation := inv.(map[string]interface{})
		if invitation["invitationId"] == invitationID {
			assert.Equal(t, "TEAM_INBOX_CANCELED", invitation["status"])
		}
	}

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/team-invitations/%s/accept", invitationID), target.AccessToken, map[string]interface{}{})
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Should not accept canceled invitation: %s", string(body))
}

func TestAcceptTeamInvitation_ShouldJoinConsumeSlotAndCancelCompeting(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captainA := tc.RegisterUser()
	captainB := tc.RegisterUser()
	captainC := tc.RegisterUser()
	target := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captainA, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, captainB, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, captainC, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, target, "PART_LOOKING_FOR_TEAM")

	teamAID := createTeam(tc, hackathonID, captainA, "Team A")
	teamBID := createTeam(tc, hackathonID, captainB, "Team B")
	teamCID := createTeam(tc, hackathonID, captainC, "Team C")

	vacancyAID := createVacancy(tc, hackathonID, teamAID, captainA, 2)
	vacancyBID := createVacancy(tc, hackathonID, teamBID, captainB, 2)
	vacancyCID := createVacancy(tc, hackathonID, teamCID, captainC, 2)

	inviteABody := map[string]interface{}{
		"target_user_id": target.UserID,
		"vacancy_id":     vacancyAID,
		"message":        "Join A!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathonID, teamAID), captainA.AccessToken, inviteABody)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	inviteAData := tc.ParseJSON(body)
	invitationAID := inviteAData["invitationId"].(string)

	inviteBBody := map[string]interface{}{
		"target_user_id": target.UserID,
		"vacancy_id":     vacancyBID,
		"message":        "Join B!",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathonID, teamBID), captainB.AccessToken, inviteBBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	inviteBData := tc.ParseJSON(body)
	invitationBID := inviteBData["invitationId"].(string)

	requestBody := map[string]interface{}{
		"vacancy_id": vacancyCID,
		"message":    "I want to join C!",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests", hackathonID, teamCID), target.AccessToken, requestBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	requestData := tc.ParseJSON(body)
	requestCID := requestData["requestId"].(string)

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/team-invitations/%s/accept", invitationAID), target.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to accept invitation: %s", string(body))

	time.Sleep(500 * time.Millisecond)

	participation := getParticipation(tc, hackathonID, target)
	assert.Equal(t, "PART_TEAM_MEMBER", participation["status"])
	assert.Equal(t, teamAID, participation["teamId"])

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s?include_vacancies=true", hackathonID, teamAID), captainA.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	teamData := tc.ParseJSON(body)
	team := teamData["team"].(map[string]interface{})
	vacancies := team["vacancies"].([]interface{})
	vacancy := vacancies[0].(map[string]interface{})
	assert.Equal(t, "1", vacancy["slotsOpen"], "Slot should be consumed")

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s/members", hackathonID, teamAID), captainA.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	membersData := tc.ParseJSON(body)
	members := membersData["members"].([]interface{})
	foundMember := false
	for _, m := range members {
		member := m.(map[string]interface{})
		if member["userId"] == target.UserID {
			foundMember = true
			assert.Equal(t, vacancyAID, member["assignedVacancyId"])
		}
	}
	assert.True(t, foundMember, "Target should be in team")

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations/list", hackathonID, teamBID), captainB.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	listBData := tc.ParseJSON(body)
	invitationsB := listBData["invitations"].([]interface{})
	for _, inv := range invitationsB {
		invitation := inv.(map[string]interface{})
		if invitation["invitationId"] == invitationBID {
			assert.Equal(t, "TEAM_INBOX_CANCELED", invitation["status"], "Competing invitation should be canceled")
		}
	}

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests/list", hackathonID, teamCID), captainC.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	listCData := tc.ParseJSON(body)
	requestsC := listCData["requests"].([]interface{})
	for _, req := range requestsC {
		request := req.(map[string]interface{})
		if request["requestId"] == requestCID {
			assert.Equal(t, "TEAM_INBOX_CANCELED", request["status"], "Competing join request should be canceled")
		}
	}
}

func TestRejectTeamInvitation_ShouldMarkDeclined(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	target := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, target, "PART_INDIVIDUAL")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)

	inviteBody := map[string]interface{}{
		"target_user_id": target.UserID,
		"vacancy_id":     vacancyID,
		"message":        "Join!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathonID, teamID), captain.AccessToken, inviteBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	inviteData := tc.ParseJSON(body)
	invitationID := inviteData["invitationId"].(string)

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/team-invitations/%s/reject", invitationID), target.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to reject invitation: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("POST", "/v1/users/me/team-invitations/list", target.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	listData := tc.ParseJSON(body)
	invitations := listData["invitations"].([]interface{})
	for _, inv := range invitations {
		invitation := inv.(map[string]interface{})
		if invitation["invitationId"] == invitationID {
			assert.Equal(t, "TEAM_INBOX_DECLINED", invitation["status"])
		}
	}

	participation := getParticipation(tc, hackathonID, target)
	assert.Equal(t, "PART_INDIVIDUAL", participation["status"])

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s?include_vacancies=true", hackathonID, teamID), captain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	teamData := tc.ParseJSON(body)
	team := teamData["team"].(map[string]interface{})
	vacancies := team["vacancies"].([]interface{})
	vacancy := vacancies[0].(map[string]interface{})
	assert.Equal(t, "2", vacancy["slotsOpen"], "Slots should remain unchanged")
}

// ============================================================================
// 5) Join Requests
// ============================================================================

func TestListJoinRequests_AsCaptain_ShouldReturnIncoming(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	requester := tc.RegisterUser()
	nonCaptain := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, requester, "PART_INDIVIDUAL")
	registerParticipant(tc, hackathonID, nonCaptain, "PART_INDIVIDUAL")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)

	requestBody := map[string]interface{}{
		"vacancy_id": vacancyID,
		"message":    "I want to join!",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests", hackathonID, teamID), requester.AccessToken, requestBody)

	listBody := map[string]interface{}{}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests/list", hackathonID, teamID), captain.AccessToken, listBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list join requests: %s", string(body))

	data := tc.ParseJSON(body)
	requests, ok := data["requests"].([]interface{})
	require.True(t, ok && len(requests) >= 1, "Should have at least 1 request")

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests/list", hackathonID, teamID), nonCaptain.AccessToken, listBody)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Non-captain should not list requests: %s", string(body))
}

func TestCreateJoinRequest_TeamNotJoinable_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	requester := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, requester, "PART_INDIVIDUAL")

	teamBody := map[string]interface{}{
		"name":        "Private Team",
		"description": "Not joinable",
		"is_joinable": false,
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams", hackathonID), captain.AccessToken, teamBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	teamData := tc.ParseJSON(body)
	teamID := teamData["teamId"].(string)

	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)

	requestBody := map[string]interface{}{
		"vacancy_id": vacancyID,
		"message":    "I want to join!",
	}

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests", hackathonID, teamID), requester.AccessToken, requestBody)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Should not create request for non-joinable team: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("POST", "/v1/users/me/join-requests/list", requester.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	listData := tc.ParseJSON(body)
	requests, ok := listData["requests"].([]interface{})
	if ok {
		assert.Equal(t, 0, len(requests), "Should have no requests")
	}
}

func TestCreateJoinRequest_ShouldCreatePending(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	requester := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, requester, "PART_INDIVIDUAL")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)

	requestBody := map[string]interface{}{
		"vacancy_id": vacancyID,
		"message":    "I have great skills!",
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests", hackathonID, teamID), requester.AccessToken, requestBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to create join request: %s", string(body))

	data := tc.ParseJSON(body)
	requestID := data["requestId"].(string)
	assert.NotEmpty(t, requestID, "Request ID should be returned")

	resp, body = tc.DoAuthenticatedRequest("POST", "/v1/users/me/join-requests/list", requester.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	listData := tc.ParseJSON(body)
	requests := listData["requests"].([]interface{})
	foundRequest := false
	for _, req := range requests {
		request := req.(map[string]interface{})
		if request["requestId"] == requestID {
			foundRequest = true
			assert.Equal(t, "TEAM_INBOX_PENDING", request["status"])
		}
	}
	assert.True(t, foundRequest, "Request should be in requester's list")

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests/list", hackathonID, teamID), captain.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	captainListData := tc.ParseJSON(body)
	captainRequests := captainListData["requests"].([]interface{})
	foundInCaptainList := false
	for _, req := range captainRequests {
		request := req.(map[string]interface{})
		if request["requestId"] == requestID {
			foundInCaptainList = true
			assert.Equal(t, "TEAM_INBOX_PENDING", request["status"])
		}
	}
	assert.True(t, foundInCaptainList, "Request should be in captain's incoming list")
}

func TestCancelJoinRequest_ShouldMarkCanceled(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	requester := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, requester, "PART_INDIVIDUAL")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)

	requestBody := map[string]interface{}{
		"vacancy_id": vacancyID,
		"message":    "Join!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests", hackathonID, teamID), requester.AccessToken, requestBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	requestData := tc.ParseJSON(body)
	requestID := requestData["requestId"].(string)

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/join-requests/%s/cancel", requestID), requester.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to cancel request: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("POST", "/v1/users/me/join-requests/list", requester.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	listData := tc.ParseJSON(body)
	requests := listData["requests"].([]interface{})
	for _, req := range requests {
		request := req.(map[string]interface{})
		if request["requestId"] == requestID {
			assert.Equal(t, "TEAM_INBOX_CANCELED", request["status"])
		}
	}
}

func TestAcceptJoinRequest_ShouldJoinConsumeSlotAndCancelCompeting(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captainA := tc.RegisterUser()
	captainB := tc.RegisterUser()
	captainC := tc.RegisterUser()
	requester := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captainA, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, captainB, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, captainC, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, requester, "PART_LOOKING_FOR_TEAM")

	teamAID := createTeam(tc, hackathonID, captainA, "Team A")
	teamBID := createTeam(tc, hackathonID, captainB, "Team B")
	teamCID := createTeam(tc, hackathonID, captainC, "Team C")

	vacancyAID := createVacancy(tc, hackathonID, teamAID, captainA, 2)
	vacancyBID := createVacancy(tc, hackathonID, teamBID, captainB, 2)
	vacancyCID := createVacancy(tc, hackathonID, teamCID, captainC, 2)

	requestABody := map[string]interface{}{
		"vacancy_id": vacancyAID,
		"message":    "Join A!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests", hackathonID, teamAID), requester.AccessToken, requestABody)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	requestAData := tc.ParseJSON(body)
	requestAID := requestAData["requestId"].(string)

	inviteBBody := map[string]interface{}{
		"target_user_id": requester.UserID,
		"vacancy_id":     vacancyBID,
		"message":        "Join B!",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathonID, teamBID), captainB.AccessToken, inviteBBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	inviteBData := tc.ParseJSON(body)
	invitationBID := inviteBData["invitationId"].(string)

	requestCBody := map[string]interface{}{
		"vacancy_id": vacancyCID,
		"message":    "Join C!",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests", hackathonID, teamCID), requester.AccessToken, requestCBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	requestCData := tc.ParseJSON(body)
	requestCID := requestCData["requestId"].(string)

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests/%s/accept", hackathonID, teamAID, requestAID), captainA.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to accept join request: %s", string(body))

	time.Sleep(500 * time.Millisecond)

	participation := getParticipation(tc, hackathonID, requester)
	assert.Equal(t, "PART_TEAM_MEMBER", participation["status"])
	assert.Equal(t, teamAID, participation["teamId"])

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s?include_vacancies=true", hackathonID, teamAID), captainA.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	teamData := tc.ParseJSON(body)
	team := teamData["team"].(map[string]interface{})
	vacancies := team["vacancies"].([]interface{})
	vacancy := vacancies[0].(map[string]interface{})
	assert.Equal(t, "1", vacancy["slotsOpen"], "Slot should be consumed")

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s/members", hackathonID, teamAID), captainA.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	membersData := tc.ParseJSON(body)
	members := membersData["members"].([]interface{})
	foundMember := false
	for _, m := range members {
		member := m.(map[string]interface{})
		if member["userId"] == requester.UserID {
			foundMember = true
			assert.Equal(t, vacancyAID, member["assignedVacancyId"])
		}
	}
	assert.True(t, foundMember, "Requester should be in team")

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations/list", hackathonID, teamBID), captainB.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	listBData := tc.ParseJSON(body)
	invitationsB := listBData["invitations"].([]interface{})
	for _, inv := range invitationsB {
		invitation := inv.(map[string]interface{})
		if invitation["invitationId"] == invitationBID {
			assert.Equal(t, "TEAM_INBOX_CANCELED", invitation["status"], "Competing invitation should be canceled")
		}
	}

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests/list", hackathonID, teamCID), captainC.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	listCData := tc.ParseJSON(body)
	requestsC := listCData["requests"].([]interface{})
	for _, req := range requestsC {
		request := req.(map[string]interface{})
		if request["requestId"] == requestCID {
			assert.Equal(t, "TEAM_INBOX_CANCELED", request["status"], "Competing join request should be canceled")
		}
	}
}

func TestRejectJoinRequest_ShouldMarkDeclined(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	requester := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, requester, "PART_INDIVIDUAL")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)

	requestBody := map[string]interface{}{
		"vacancy_id": vacancyID,
		"message":    "Join!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests", hackathonID, teamID), requester.AccessToken, requestBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	requestData := tc.ParseJSON(body)
	requestID := requestData["requestId"].(string)

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests/%s/reject", hackathonID, teamID, requestID), captain.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to reject request: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("POST", "/v1/users/me/join-requests/list", requester.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	listData := tc.ParseJSON(body)
	requests := listData["requests"].([]interface{})
	for _, req := range requests {
		request := req.(map[string]interface{})
		if request["requestId"] == requestID {
			assert.Equal(t, "TEAM_INBOX_DECLINED", request["status"])
		}
	}

	participation := getParticipation(tc, hackathonID, requester)
	assert.Equal(t, "PART_INDIVIDUAL", participation["status"])

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s?include_vacancies=true", hackathonID, teamID), captain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	teamData := tc.ParseJSON(body)
	team := teamData["team"].(map[string]interface{})
	vacancies := team["vacancies"].([]interface{})
	vacancy := vacancies[0].(map[string]interface{})
	assert.Equal(t, "2", vacancy["slotsOpen"], "Slots should remain unchanged")
}

func TestListMyTeamInvitations_ShouldReturnAllHackathons(t *testing.T) {
	tc := NewTestContext(t)
	owner1 := tc.RegisterUser()
	owner2 := tc.RegisterUser()
	captain1 := tc.RegisterUser()
	captain2 := tc.RegisterUser()
	target := tc.RegisterUser()

	hackathon1ID := createHackathonInRegistration(tc, owner1)
	hackathon2ID := createHackathonInRegistration(tc, owner2)

	registerParticipant(tc, hackathon1ID, captain1, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathon1ID, target, "PART_INDIVIDUAL")
	registerParticipant(tc, hackathon2ID, captain2, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathon2ID, target, "PART_INDIVIDUAL")

	team1ID := createTeam(tc, hackathon1ID, captain1, "Team in Hackathon 1")
	team2ID := createTeam(tc, hackathon2ID, captain2, "Team in Hackathon 2")

	vacancy1ID := createVacancy(tc, hackathon1ID, team1ID, captain1, 2)
	vacancy2ID := createVacancy(tc, hackathon2ID, team2ID, captain2, 2)

	invite1Body := map[string]interface{}{
		"target_user_id": target.UserID,
		"vacancy_id":     vacancy1ID,
		"message":        "Join hackathon 1!",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathon1ID, team1ID), captain1.AccessToken, invite1Body)

	invite2Body := map[string]interface{}{
		"target_user_id": target.UserID,
		"vacancy_id":     vacancy2ID,
		"message":        "Join hackathon 2!",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathon2ID, team2ID), captain2.AccessToken, invite2Body)

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/users/me/team-invitations/list", target.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list my invitations: %s", string(body))

	data := tc.ParseJSON(body)
	invitations, ok := data["invitations"].([]interface{})
	require.True(t, ok, "Invitations array should be present")
	assert.GreaterOrEqual(t, len(invitations), 2, "Should have invitations from both hackathons")

	foundHackathon1 := false
	foundHackathon2 := false
	for _, inv := range invitations {
		invitation := inv.(map[string]interface{})
		if invitation["hackathonId"] == hackathon1ID {
			foundHackathon1 = true
		}
		if invitation["hackathonId"] == hackathon2ID {
			foundHackathon2 = true
		}
	}
	assert.True(t, foundHackathon1, "Should have invitation from hackathon 1")
	assert.True(t, foundHackathon2, "Should have invitation from hackathon 2")
}

func TestListMyJoinRequests_ShouldReturnAllHackathons(t *testing.T) {
	tc := NewTestContext(t)
	owner1 := tc.RegisterUser()
	owner2 := tc.RegisterUser()
	captain1 := tc.RegisterUser()
	captain2 := tc.RegisterUser()
	requester := tc.RegisterUser()

	hackathon1ID := createHackathonInRegistration(tc, owner1)
	hackathon2ID := createHackathonInRegistration(tc, owner2)

	registerParticipant(tc, hackathon1ID, captain1, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathon1ID, requester, "PART_INDIVIDUAL")
	registerParticipant(tc, hackathon2ID, captain2, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathon2ID, requester, "PART_INDIVIDUAL")

	team1ID := createTeam(tc, hackathon1ID, captain1, "Team in Hackathon 1")
	team2ID := createTeam(tc, hackathon2ID, captain2, "Team in Hackathon 2")

	vacancy1ID := createVacancy(tc, hackathon1ID, team1ID, captain1, 2)
	vacancy2ID := createVacancy(tc, hackathon2ID, team2ID, captain2, 2)

	request1Body := map[string]interface{}{
		"vacancy_id": vacancy1ID,
		"message":    "Join hackathon 1!",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests", hackathon1ID, team1ID), requester.AccessToken, request1Body)

	request2Body := map[string]interface{}{
		"vacancy_id": vacancy2ID,
		"message":    "Join hackathon 2!",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests", hackathon2ID, team2ID), requester.AccessToken, request2Body)

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/users/me/join-requests/list", requester.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list my join requests: %s", string(body))

	data := tc.ParseJSON(body)
	requests, ok := data["requests"].([]interface{})
	require.True(t, ok, "Requests array should be present")
	assert.GreaterOrEqual(t, len(requests), 2, "Should have requests from both hackathons")

	foundHackathon1 := false
	foundHackathon2 := false
	for _, req := range requests {
		request := req.(map[string]interface{})
		if request["hackathonId"] == hackathon1ID {
			foundHackathon1 = true
		}
		if request["hackathonId"] == hackathon2ID {
			foundHackathon2 = true
		}
	}
	assert.True(t, foundHackathon1, "Should have request from hackathon 1")
	assert.True(t, foundHackathon2, "Should have request from hackathon 2")
}

func TestCreateTeamInvitation_NoOpenSlots_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	member := tc.RegisterUser()
	target := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, member, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, target, "PART_INDIVIDUAL")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 1)

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
	time.Sleep(500 * time.Millisecond)

	invite2Body := map[string]interface{}{
		"target_user_id": target.UserID,
		"vacancy_id":     vacancyID,
		"message":        "Join!",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathonID, teamID), captain.AccessToken, invite2Body)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Should not create invitation when no slots available: %s", string(body))
}

func TestCreateJoinRequest_NoOpenSlots_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	member := tc.RegisterUser()
	requester := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, member, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, requester, "PART_INDIVIDUAL")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 1)

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
	time.Sleep(500 * time.Millisecond)

	requestBody := map[string]interface{}{
		"vacancy_id": vacancyID,
		"message":    "I want to join!",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests", hackathonID, teamID), requester.AccessToken, requestBody)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Should not create request when no slots available: %s", string(body))
}

func TestAcceptTeamInvitation_AlreadyInTeam_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain1 := tc.RegisterUser()
	captain2 := tc.RegisterUser()
	target := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain1, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, captain2, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, target, "PART_LOOKING_FOR_TEAM")

	team1ID := createTeam(tc, hackathonID, captain1, "Team 1")
	team2ID := createTeam(tc, hackathonID, captain2, "Team 2")

	vacancy1ID := createVacancy(tc, hackathonID, team1ID, captain1, 2)
	vacancy2ID := createVacancy(tc, hackathonID, team2ID, captain2, 2)

	invite1Body := map[string]interface{}{
		"target_user_id": target.UserID,
		"vacancy_id":     vacancy1ID,
		"message":        "Join 1!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathonID, team1ID), captain1.AccessToken, invite1Body)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	invite1Data := tc.ParseJSON(body)
	invitation1ID := invite1Data["invitationId"].(string)

	invite2Body := map[string]interface{}{
		"target_user_id": target.UserID,
		"vacancy_id":     vacancy2ID,
		"message":        "Join 2!",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathonID, team2ID), captain2.AccessToken, invite2Body)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	invite2Data := tc.ParseJSON(body)
	invitation2ID := invite2Data["invitationId"].(string)

	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/team-invitations/%s/accept", invitation1ID), target.AccessToken, map[string]interface{}{})
	time.Sleep(500 * time.Millisecond)

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/team-invitations/%s/accept", invitation2ID), target.AccessToken, map[string]interface{}{})
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Should not accept invitation when already in team: %s", string(body))

	participation := getParticipation(tc, hackathonID, target)
	assert.Equal(t, team1ID, participation["teamId"])
}

func TestAcceptJoinRequest_RequesterAlreadyInTeam_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain1 := tc.RegisterUser()
	captain2 := tc.RegisterUser()
	requester := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain1, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, captain2, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, requester, "PART_LOOKING_FOR_TEAM")

	team1ID := createTeam(tc, hackathonID, captain1, "Team 1")
	team2ID := createTeam(tc, hackathonID, captain2, "Team 2")

	vacancy1ID := createVacancy(tc, hackathonID, team1ID, captain1, 2)
	vacancy2ID := createVacancy(tc, hackathonID, team2ID, captain2, 2)

	request1Body := map[string]interface{}{
		"vacancy_id": vacancy1ID,
		"message":    "Join 1!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests", hackathonID, team1ID), requester.AccessToken, request1Body)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	request1Data := tc.ParseJSON(body)
	request1ID := request1Data["requestId"].(string)

	request2Body := map[string]interface{}{
		"vacancy_id": vacancy2ID,
		"message":    "Join 2!",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests", hackathonID, team2ID), requester.AccessToken, request2Body)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	request2Data := tc.ParseJSON(body)
	request2ID := request2Data["requestId"].(string)

	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests/%s/accept", hackathonID, team1ID, request1ID), captain1.AccessToken, map[string]interface{}{})
	time.Sleep(500 * time.Millisecond)

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests/%s/accept", hackathonID, team2ID, request2ID), captain2.AccessToken, map[string]interface{}{})
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Should not accept request when requester already in team: %s", string(body))

	participation := getParticipation(tc, hackathonID, requester)
	assert.Equal(t, team1ID, participation["teamId"])
}

func TestKickMember_CannotKickCaptain_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	member := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, member, "PART_LOOKING_FOR_TEAM")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)

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
	time.Sleep(500 * time.Millisecond)

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/members/%s/kick", hackathonID, teamID, captain.UserID), captain.AccessToken, map[string]interface{}{})
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Should not kick captain: %s", string(body))

	participation := getParticipation(tc, hackathonID, captain)
	assert.Equal(t, "PART_TEAM_CAPTAIN", participation["status"])
}

func TestTransferCaptain_ToNonMember_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	outsider := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, outsider, "PART_INDIVIDUAL")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")

	transferBody := map[string]interface{}{
		"target_user_id": outsider.UserID,
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/transferCaptain", hackathonID, teamID), captain.AccessToken, transferBody)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Should not transfer to non-member: %s", string(body))

	participation := getParticipation(tc, hackathonID, captain)
	assert.Equal(t, "PART_TEAM_CAPTAIN", participation["status"])
}

// ============================================================================
// 6) Stage/Policy
// ============================================================================

func TestWriteOperations_OutsideRegistration_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	target := tc.RegisterUser()

	hackathonBody := map[string]interface{}{
		"name":              "Past Registration Hackathon",
		"short_description": "Test",
		"description":       "Test",
		"location": map[string]interface{}{
			"online": true,
		},
		"dates": map[string]interface{}{
			"registration_opens_at":  time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339), // 30 days ago
			"registration_closes_at": time.Now().Add(-10 * 24 * time.Hour).Format(time.RFC3339), // 10 days ago (closed)
			"starts_at":              time.Now().Add(-5 * 24 * time.Hour).Format(time.RFC3339),  // Started 5 days ago
			"ends_at":                time.Now().Add(2 * 24 * time.Hour).Format(time.RFC3339),   // Ends in 2 days
			"judging_ends_at":        time.Now().Add(5 * 24 * time.Hour).Format(time.RFC3339),
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
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	hackathonID := data["hackathonId"].(string)

	tc.WaitForHackathonOwnerRole(hackathonID, owner.AccessToken)

	taskBody := map[string]interface{}{
		"task": "Build something",
	}
	tc.DoAuthenticatedRequest("PUT", fmt.Sprintf("/v1/hackathons/%s/task", hackathonID), owner.AccessToken, taskBody)
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/publish", hackathonID), owner.AccessToken, map[string]interface{}{})
	time.Sleep(500 * time.Millisecond)

	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, target, "PART_INDIVIDUAL")

	teamBody := map[string]interface{}{
		"name": "Test Team",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams", hackathonID), captain.AccessToken, teamBody)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Should not create team outside REGISTRATION: %s", string(body))

	listBody := map[string]interface{}{}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/list", hackathonID), captain.AccessToken, listBody)
	if resp.StatusCode == http.StatusOK {
		listData := tc.ParseJSON(body)
		teams, ok := listData["teams"].([]interface{})
		if ok {
			assert.Equal(t, 0, len(teams), "Should have no teams")
		}
	}

	participation := getParticipation(tc, hackathonID, captain)
	assert.Equal(t, "PART_LOOKING_FOR_TEAM", participation["status"])
}

func TestAllowTeamFalse_ShouldDenyWrites(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()

	hackathonBody := map[string]interface{}{
		"name":              "No Teams Hackathon",
		"short_description": "Individual only",
		"description":       "Test",
		"location": map[string]interface{}{
			"online": true,
		},
		"dates": map[string]interface{}{
			"registration_opens_at":  time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			"registration_closes_at": time.Now().Add(15 * 24 * time.Hour).Format(time.RFC3339),
			"starts_at":              time.Now().Add(20 * 24 * time.Hour).Format(time.RFC3339),
			"ends_at":                time.Now().Add(22 * 24 * time.Hour).Format(time.RFC3339),
			"judging_ends_at":        time.Now().Add(25 * 24 * time.Hour).Format(time.RFC3339),
		},
		"registration_policy": map[string]interface{}{
			"allow_individual": true,
			"allow_team":       false, // Teams not allowed
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons", owner.AccessToken, hackathonBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	hackathonID := data["hackathonId"].(string)

	tc.WaitForHackathonOwnerRole(hackathonID, owner.AccessToken)

	taskBody := map[string]interface{}{
		"task": "Build individually",
	}
	tc.DoAuthenticatedRequest("PUT", fmt.Sprintf("/v1/hackathons/%s/task", hackathonID), owner.AccessToken, taskBody)
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/publish", hackathonID), owner.AccessToken, map[string]interface{}{})
	time.Sleep(500 * time.Millisecond)

	registerParticipant(tc, hackathonID, captain, "PART_INDIVIDUAL")

	teamBody := map[string]interface{}{
		"name": "Forbidden Team",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams", hackathonID), captain.AccessToken, teamBody)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Should not create team when allow_team=false: %s", string(body))

	participation := getParticipation(tc, hackathonID, captain)
	assert.Equal(t, "PART_INDIVIDUAL", participation["status"])
}

func TestCompetingCancellation_OnlyWithinSameHackathon(t *testing.T) {
	tc := NewTestContext(t)
	owner1 := tc.RegisterUser()
	owner2 := tc.RegisterUser()
	captain1 := tc.RegisterUser()
	captain2 := tc.RegisterUser()
	target := tc.RegisterUser()

	hackathon1ID := createHackathonInRegistration(tc, owner1)
	hackathon2ID := createHackathonInRegistration(tc, owner2)

	registerParticipant(tc, hackathon1ID, captain1, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathon1ID, target, "PART_INDIVIDUAL")
	registerParticipant(tc, hackathon2ID, captain2, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathon2ID, target, "PART_INDIVIDUAL")

	team1ID := createTeam(tc, hackathon1ID, captain1, "Team in H1")
	team2ID := createTeam(tc, hackathon2ID, captain2, "Team in H2")

	vacancy1ID := createVacancy(tc, hackathon1ID, team1ID, captain1, 2)
	vacancy2ID := createVacancy(tc, hackathon2ID, team2ID, captain2, 2)

	invite1Body := map[string]interface{}{
		"target_user_id": target.UserID,
		"vacancy_id":     vacancy1ID,
		"message":        "Join H1!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathon1ID, team1ID), captain1.AccessToken, invite1Body)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	invite1Data := tc.ParseJSON(body)
	invitation1ID := invite1Data["invitationId"].(string)

	invite2Body := map[string]interface{}{
		"target_user_id": target.UserID,
		"vacancy_id":     vacancy2ID,
		"message":        "Join H2!",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathon2ID, team2ID), captain2.AccessToken, invite2Body)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	invite2Data := tc.ParseJSON(body)
	invitation2ID := invite2Data["invitationId"].(string)

	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/team-invitations/%s/accept", invitation1ID), target.AccessToken, map[string]interface{}{})
	time.Sleep(500 * time.Millisecond)

	resp, body = tc.DoAuthenticatedRequest("POST", "/v1/users/me/team-invitations/list", target.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	listData := tc.ParseJSON(body)
	invitations := listData["invitations"].([]interface{})
	for _, inv := range invitations {
		invitation := inv.(map[string]interface{})
		if invitation["invitationId"] == invitation2ID {
			assert.Equal(t, "TEAM_INBOX_PENDING", invitation["status"], "Invitation in different hackathon should remain PENDING")
		}
	}

	participation1 := getParticipation(tc, hackathon1ID, target)
	assert.Equal(t, "PART_TEAM_MEMBER", participation1["status"])
	assert.Equal(t, team1ID, participation1["teamId"])

	participation2 := getParticipation(tc, hackathon2ID, target)
	assert.Equal(t, "PART_INDIVIDUAL", participation2["status"])
}

func TestUpsertVacancy_ExceedTeamSizeMax_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()

	now := time.Now()
	hackathonBody := map[string]interface{}{
		"name":              "Small Team Hackathon",
		"short_description": "Max 3 per team",
		"description":       "Test",
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
			"team_size_max": 3,
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons", owner.AccessToken, hackathonBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	hackathonID := data["hackathonId"].(string)

	tc.WaitForHackathonOwnerRole(hackathonID, owner.AccessToken)

	taskBody := map[string]interface{}{
		"task": "Build with small team",
	}
	resp, body = tc.DoAuthenticatedRequest("PUT", fmt.Sprintf("/v1/hackathons/%s/task", hackathonID), owner.AccessToken, taskBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to set task: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/publish", hackathonID), owner.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to publish hackathon: %s", string(body))

	_, err := tc.DB.Exec(context.Background(), `
		UPDATE hackathon.hackathons 
		SET registration_opens_at = $1,
		    stage = 'registration'
		WHERE id = $2
	`, now.Add(-24*time.Hour), hackathonID)
	require.NoError(t, err, "Failed to update hackathon dates in DB")

	time.Sleep(500 * time.Millisecond)

	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")

	teamID := createTeam(tc, hackathonID, captain, "Small Team")

	vacancyBody := map[string]interface{}{
		"vacancy_id":  "",
		"name":        "Too Many Slots",
		"slots_total": 3,
	}

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/vacancies:upsert", hackathonID, teamID), captain.AccessToken, vacancyBody)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Should not create vacancy exceeding team_size_max: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s?include_vacancies=true", hackathonID, teamID), captain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	teamData := tc.ParseJSON(body)
	team := teamData["team"].(map[string]interface{})
	vacancies, ok := team["vacancies"].([]interface{})
	if ok {
		assert.Equal(t, 0, len(vacancies), "Should have no vacancies")
	}
}

func TestCreateTeamInvitation_TargetAlreadyInTeam_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain1 := tc.RegisterUser()
	captain2 := tc.RegisterUser()
	target := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain1, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, captain2, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, target, "PART_LOOKING_FOR_TEAM")

	team1ID := createTeam(tc, hackathonID, captain1, "Team 1")
	team2ID := createTeam(tc, hackathonID, captain2, "Team 2")

	vacancy1ID := createVacancy(tc, hackathonID, team1ID, captain1, 2)
	vacancy2ID := createVacancy(tc, hackathonID, team2ID, captain2, 2)

	invite1Body := map[string]interface{}{
		"target_user_id": target.UserID,
		"vacancy_id":     vacancy1ID,
		"message":        "Join!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathonID, team1ID), captain1.AccessToken, invite1Body)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	invite1Data := tc.ParseJSON(body)
	invitation1ID := invite1Data["invitationId"].(string)

	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/team-invitations/%s/accept", invitation1ID), target.AccessToken, map[string]interface{}{})
	time.Sleep(500 * time.Millisecond)

	invite2Body := map[string]interface{}{
		"target_user_id": target.UserID,
		"vacancy_id":     vacancy2ID,
		"message":        "Join us!",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathonID, team2ID), captain2.AccessToken, invite2Body)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Should not invite user already in team: %s", string(body))
}

func TestCreateJoinRequest_AlreadyInTeam_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain1 := tc.RegisterUser()
	captain2 := tc.RegisterUser()
	user := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain1, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, captain2, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, user, "PART_LOOKING_FOR_TEAM")

	team1ID := createTeam(tc, hackathonID, captain1, "Team 1")
	team2ID := createTeam(tc, hackathonID, captain2, "Team 2")

	vacancy1ID := createVacancy(tc, hackathonID, team1ID, captain1, 2)
	vacancy2ID := createVacancy(tc, hackathonID, team2ID, captain2, 2)

	// User joins team1
	invite1Body := map[string]interface{}{
		"target_user_id": user.UserID,
		"vacancy_id":     vacancy1ID,
		"message":        "Join!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathonID, team1ID), captain1.AccessToken, invite1Body)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	invite1Data := tc.ParseJSON(body)
	invitation1ID := invite1Data["invitationId"].(string)

	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/team-invitations/%s/accept", invitation1ID), user.AccessToken, map[string]interface{}{})
	time.Sleep(500 * time.Millisecond)

	request2Body := map[string]interface{}{
		"vacancy_id": vacancy2ID,
		"message":    "Join 2!",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/join-requests", hackathonID, team2ID), user.AccessToken, request2Body)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Should not create request when already in team: %s", string(body))
}

func TestUpdateTeam_DuplicateName_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain1 := tc.RegisterUser()
	captain2 := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain1, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, captain2, "PART_LOOKING_FOR_TEAM")

	team1ID := createTeam(tc, hackathonID, captain1, "Team Alpha")
	createTeam(tc, hackathonID, captain2, "Team Beta")

	updateBody := map[string]interface{}{
		"name": "Team Beta",
	}

	resp, body := tc.DoAuthenticatedRequest("PUT", fmt.Sprintf("/v1/hackathons/%s/teams/%s", hackathonID, team1ID), captain1.AccessToken, updateBody)
	assert.Equal(t, http.StatusConflict, resp.StatusCode, "Should not allow duplicate team name: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s", hackathonID, team1ID), captain1.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	teamWithVacancies := data["team"].(map[string]interface{})
	team := teamWithVacancies["team"].(map[string]interface{})
	assert.Equal(t, "Team Alpha", team["name"])
}

func TestUpsertVacancy_NegativeSlotsTotal_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")

	vacancyBody := map[string]interface{}{
		"vacancy_id":  "",
		"name":        "Invalid Vacancy",
		"slots_total": -1,
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/vacancies/upsert", hackathonID, teamID), captain.AccessToken, vacancyBody)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Should not create vacancy with negative slots: %s", string(body))
}

func TestKickMember_AsNonCaptain_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	member1 := tc.RegisterUser()
	member2 := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, member1, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, member2, "PART_LOOKING_FOR_TEAM")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 3)

	for _, member := range []*UserCredentials{member1, member2} {
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
		time.Sleep(300 * time.Millisecond)
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/members/%s/kick", hackathonID, teamID, member2.UserID), member1.AccessToken, map[string]interface{}{})
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Non-captain should not kick members: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/teams/%s/members", hackathonID, teamID), captain.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	membersData := tc.ParseJSON(body)
	members := membersData["members"].([]interface{})
	assert.Equal(t, 3, len(members), "Should still have 3 members")
}

// createHackathonInRegistration creates a hackathon that is currently in REGISTRATION stage
// This means: registration_opens_at is in the past, registration_closes_at is in the future
func createHackathonInRegistration(tc *TestContext, owner *UserCredentials) string {
	now := time.Now()
	hackathonBody := map[string]interface{}{
		"name":              "Team Test Hackathon",
		"short_description": "Test hackathon for teams",
		"description":       "Full description for team testing",
		"location": map[string]interface{}{
			"online": true,
		},
		"dates": map[string]interface{}{
			"registration_opens_at":  now.Add(1 * time.Hour).Format(time.RFC3339), // Opens in 1 hour (future for publish)
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

	// Update dates and stage directly in DB to move to REGISTRATION stage (bypassing validation)
	_, err := tc.DB.Exec(context.Background(), `
		UPDATE hackathon.hackathons 
		SET registration_opens_at = $1,
		    stage = 'registration'
		WHERE id = $2
	`, now.Add(-24*time.Hour), hackathonID)
	require.NoError(tc.T, err, "Failed to update hackathon dates in DB")

	// Wait a bit for updates to propagate
	time.Sleep(500 * time.Millisecond)

	return hackathonID
}

// registerParticipant registers a user for a hackathon with given status
func registerParticipant(tc *TestContext, hackathonID string, user *UserCredentials, status string) {
	registerBody := map[string]interface{}{
		"desired_status":  status,
		"motivation_text": "Test motivation",
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), user.AccessToken, registerBody)
	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to register participant: %s", string(body))

	tc.WaitForParticipationRegistered(hackathonID, user.AccessToken)
}

// createTeam creates a team and returns team ID
func createTeam(tc *TestContext, hackathonID string, captain *UserCredentials, name string) string {
	teamBody := map[string]interface{}{
		"name":        name,
		"description": "Test team description",
		"is_joinable": true,
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams", hackathonID), captain.AccessToken, teamBody)
	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to create team: %s", string(body))

	data := tc.ParseJSON(body)
	teamID := data["teamId"].(string)

	// Wait for team creation to propagate
	time.Sleep(300 * time.Millisecond)

	return teamID
}

// createVacancy creates a vacancy and returns vacancy ID
func createVacancy(tc *TestContext, hackathonID, teamID string, captain *UserCredentials, slotsTotal int64) string {
	vacancyBody := map[string]interface{}{
		"vacancy_id":  "",
		"name":        "Backend Developer",
		"description": "We need a backend developer",
		"slots_total": slotsTotal,
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/teams/%s/vacancies/upsert", hackathonID, teamID), captain.AccessToken, vacancyBody)
	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to create vacancy: %s", string(body))

	data := tc.ParseJSON(body)
	vacancyID := data["vacancyId"].(string)

	return vacancyID
}

// getParticipation gets user's participation status
func getParticipation(tc *TestContext, hackathonID string, user *UserCredentials) map[string]interface{} {
	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/participations/me", hackathonID), user.AccessToken, nil)
	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to get participation: %s", string(body))

	data := tc.ParseJSON(body)
	return data["participation"].(map[string]interface{})
}
