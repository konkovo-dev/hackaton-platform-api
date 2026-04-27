package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTeamRoles(t *testing.T) {
	tc := NewTestContext(t)
	creds := tc.RegisterUser()

	resp, body := tc.DoAuthenticatedRequest("GET", "/v1/team-roles", creds.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list team roles: %s", string(body))

	data := tc.ParseJSON(body)
	roles, ok := data["teamRoles"].([]interface{})
	require.True(t, ok, "Team roles array should be present")
	assert.GreaterOrEqual(t, len(roles), 1, "Should have at least one team role")

	if len(roles) > 0 {
		role := roles[0].(map[string]interface{})
		assert.NotEmpty(t, role["id"], "Role should have ID")
		assert.NotEmpty(t, role["name"], "Role should have name")
	}
}

func TestRegisterForHackathonIndividual(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createAndPublishHackathonForRegistration(tc, owner)

	resp, body := tc.DoAuthenticatedRequest("GET", "/v1/team-roles", participant.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	rolesData := tc.ParseJSON(body)
	roles := rolesData["teamRoles"].([]interface{})
	var frontendRoleID string
	for _, r := range roles {
		role := r.(map[string]interface{})
		if role["name"] == "Frontend" {
			frontendRoleID = role["id"].(string)
			break
		}
	}

	registerBody := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"wished_role_ids": []string{frontendRoleID},
		"motivation_text": "I love frontend development!",
	}

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant.AccessToken, registerBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to register: %s", string(body))

	data := tc.ParseJSON(body)
	participation := data["participation"].(map[string]interface{})
	assert.Equal(t, hackathonID, participation["hackathonId"])
	assert.Equal(t, participant.UserID, participation["userId"])
	assert.Equal(t, "PART_INDIVIDUAL", participation["status"])
}

func TestRegisterForHackathonLookingForTeam(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createAndPublishHackathonForRegistration(tc, owner)

	registerBody := map[string]interface{}{
		"desired_status":  "PART_LOOKING_FOR_TEAM",
		"wished_role_ids": []string{},
		"motivation_text": "Want to find a great team",
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant.AccessToken, registerBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to register: %s", string(body))

	data := tc.ParseJSON(body)
	participation := data["participation"].(map[string]interface{})
	assert.Equal(t, "PART_LOOKING_FOR_TEAM", participation["status"])
}

func TestRegisterTwiceShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createAndPublishHackathonForRegistration(tc, owner)

	registerBody := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "First registration",
	}
	resp, _ := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant.AccessToken, registerBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	registerBody["motivation_text"] = "Second registration"
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant.AccessToken, registerBody)
	assert.Equal(t, http.StatusConflict, resp.StatusCode, "Should reject duplicate registration: %s", string(body))
}

func TestGetMyParticipation(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createAndPublishHackathonForRegistration(tc, owner)

	registerBody := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "Test motivation",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant.AccessToken, registerBody)

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/participations/me", hackathonID), participant.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get participation: %s", string(body))

	data := tc.ParseJSON(body)
	participation := data["participation"].(map[string]interface{})
	assert.Equal(t, participant.UserID, participation["userId"])
	assert.Equal(t, "PART_INDIVIDUAL", participation["status"])

	profile := participation["profile"].(map[string]interface{})
	assert.Equal(t, "Test motivation", profile["motivationText"])
}

func TestUpdateMyParticipation(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createAndPublishHackathonForRegistration(tc, owner)

	registerBody := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "Original motivation",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant.AccessToken, registerBody)

	resp, body := tc.DoAuthenticatedRequest("GET", "/v1/team-roles", participant.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	rolesData := tc.ParseJSON(body)
	roles := rolesData["teamRoles"].([]interface{})
	var frontendRoleID, designerRoleID string
	for _, r := range roles {
		role := r.(map[string]interface{})
		if role["name"] == "Frontend" {
			frontendRoleID = role["id"].(string)
		}
		if role["name"] == "Designer" {
			designerRoleID = role["id"].(string)
		}
	}

	updateBody := map[string]interface{}{
		"wished_role_ids": []string{frontendRoleID, designerRoleID},
		"motivation_text": "Updated motivation",
	}

	resp, body = tc.DoAuthenticatedRequest("PUT", fmt.Sprintf("/v1/hackathons/%s/participations/me", hackathonID), participant.AccessToken, updateBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to update participation: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/participations/me", hackathonID), participant.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	participation := data["participation"].(map[string]interface{})
	profile := participation["profile"].(map[string]interface{})
	assert.Equal(t, "Updated motivation", profile["motivationText"])

	wishedRoles := profile["wishedRoles"].([]interface{})
	assert.Equal(t, 2, len(wishedRoles), "Should have 2 wished roles")
}

func TestSwitchParticipationMode(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createAndPublishHackathonForRegistration(tc, owner)

	registerBody := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "Test",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant.AccessToken, registerBody)

	switchBody := map[string]interface{}{
		"new_status": "PART_LOOKING_FOR_TEAM",
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/me/switchMode", hackathonID), participant.AccessToken, switchBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to switch mode: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/participations/me", hackathonID), participant.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	participation := data["participation"].(map[string]interface{})
	assert.Equal(t, "PART_LOOKING_FOR_TEAM", participation["status"])
}

func TestSwitchToSameStatusShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createAndPublishHackathonForRegistration(tc, owner)

	registerBody := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "Test",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant.AccessToken, registerBody)

	switchBody := map[string]interface{}{
		"new_status": "PART_INDIVIDUAL",
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/me/switchMode", hackathonID), participant.AccessToken, switchBody)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should reject switch to same status: %s", string(body))
}

func TestGetUserParticipation(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant1 := tc.RegisterUser()
	participant2 := tc.RegisterUser()

	hackathonID := createAndPublishHackathonForRegistration(tc, owner)

	registerBody := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "Test",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant1.AccessToken, registerBody)
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant2.AccessToken, registerBody)

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/participations/users/%s", hackathonID, participant2.UserID), participant1.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get user participation: %s", string(body))

	data := tc.ParseJSON(body)
	participation := data["participation"].(map[string]interface{})
	assert.Equal(t, participant2.UserID, participation["userId"])
}

func TestListHackathonParticipants(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant1 := tc.RegisterUser()
	participant2 := tc.RegisterUser()

	hackathonID := createAndPublishHackathonForRegistration(tc, owner)

	registerBody := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "Test",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant1.AccessToken, registerBody)

	registerBody["desired_status"] = "PART_LOOKING_FOR_TEAM"
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant2.AccessToken, registerBody)

	listBody := map[string]interface{}{}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/list", hackathonID), participant1.AccessToken, listBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list participants: %s", string(body))

	data := tc.ParseJSON(body)
	participants, ok := data["participants"].([]interface{})
	require.True(t, ok, "Participants array should be present")
	assert.GreaterOrEqual(t, len(participants), 2, "Should have at least 2 participants")
}

func TestListHackathonParticipantsWithStatusFilter(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant1 := tc.RegisterUser()
	participant2 := tc.RegisterUser()

	hackathonID := createAndPublishHackathonForRegistration(tc, owner)

	registerBody1 := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "Test",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant1.AccessToken, registerBody1)

	registerBody2 := map[string]interface{}{
		"desired_status":  "PART_LOOKING_FOR_TEAM",
		"motivation_text": "Test",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant2.AccessToken, registerBody2)

	listBody := map[string]interface{}{
		"status_filter": map[string]interface{}{
			"statuses": []string{"PART_LOOKING_FOR_TEAM"},
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/list", hackathonID), participant1.AccessToken, listBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to filter participants: %s", string(body))

	data := tc.ParseJSON(body)
	participants, ok := data["participants"].([]interface{})
	require.True(t, ok, "Participants array should be present")

	for _, p := range participants {
		part := p.(map[string]interface{})
		assert.Equal(t, "PART_LOOKING_FOR_TEAM", part["status"])
	}
}

func TestListParticipantsAsNonParticipantShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	nonParticipant := tc.RegisterUser()

	hackathonID := createAndPublishHackathonForRegistration(tc, owner)

	listBody := map[string]interface{}{}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/list", hackathonID), nonParticipant.AccessToken, listBody)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Non-participant should not list participants: %s", string(body))
}

func TestUnregisterFromHackathon(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createAndPublishHackathonForRegistration(tc, owner)

	registerBody := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "Test",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant.AccessToken, registerBody)

	unregisterBody := map[string]interface{}{}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/me/unregister", hackathonID), participant.AccessToken, unregisterBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to unregister: %s", string(body))

	resp, _ = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/participations/me", hackathonID), participant.AccessToken, nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Should not find participation after unregister")
}

func TestRegisterForRunningHackathonShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	// Create and publish hackathon
	hackathonID := createAndPublishHackathon(tc, owner)

	// Manually update hackathon stage to RUNNING in the database
	_, err := tc.DB.Exec(context.Background(), `
		UPDATE hackathon.hackathons 
		SET stage = 'running' 
		WHERE id = $1
	`, hackathonID)
	require.NoError(t, err, "Failed to update hackathon stage")

	// Try to register - should fail because stage is not REGISTRATION
	registerBody := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "Trying to register for running hackathon",
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant.AccessToken, registerBody)
	require.Equal(t, http.StatusForbidden, resp.StatusCode, "Should not allow registration during RUNNING stage: %s", string(body))

	data := tc.ParseJSON(body)
	message, ok := data["message"].(string)
	require.True(t, ok, "Response should have message field")
	assert.Contains(t, message, "registration is only allowed during REGISTRATION stage", "Error message should mention stage restriction")
}

func TestGetParticipationPermissions_CanRegisterDuringRegistrationStage(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	regularUser := tc.RegisterUser()

	hackathonID := createDraftHackathon(tc, owner)

	// Manually set hackathon to published with registration stage dates
	_, err := tc.DB.Exec(context.Background(), `
		UPDATE hackathon.hackathons 
		SET state = 'published', 
		    published_at = NOW(),
		    registration_opens_at = NOW() - INTERVAL '5 days',
		    registration_closes_at = NOW() + INTERVAL '5 days',
		    starts_at = NOW() + INTERVAL '10 days',
		    ends_at = NOW() + INTERVAL '12 days',
		    judging_ends_at = NOW() + INTERVAL '15 days',
		    allow_individual = true,
		    allow_team = true
		WHERE id = $1
	`, hackathonID)
	require.NoError(t, err, "Failed to publish hackathon")

	// Get permissions for regular user (not staff, not participant)
	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/participations/permissions", hackathonID), regularUser.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get permissions: %s", string(body))

	data := tc.ParseJSON(body)
	permissions := data["permissions"].(map[string]interface{})

	assert.True(t, permissions["register"].(bool), "User should be able to register during REGISTRATION stage")
	assert.False(t, permissions["unregister"].(bool), "User should not be able to unregister if not registered")
	assert.False(t, permissions["switchParticipationMode"].(bool), "User should not be able to switch mode if not registered")
	assert.False(t, permissions["updateParticipationProfile"].(bool), "User should not be able to update profile if not registered")
	assert.False(t, permissions["inviteStaff"].(bool), "Regular user should not be able to invite staff")
	assert.False(t, permissions["listParticipants"].(bool), "Regular user should not be able to list participants")
}

func TestGetParticipationPermissions_CannotRegisterOutsideRegistrationStage(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	regularUser := tc.RegisterUser()

	hackathonID := createDraftHackathon(tc, owner)

	// Manually set hackathon to published with running stage dates
	_, err := tc.DB.Exec(context.Background(), `
		UPDATE hackathon.hackathons 
		SET state = 'published', 
		    published_at = NOW(),
		    registration_opens_at = NOW() - INTERVAL '10 days',
		    registration_closes_at = NOW() - INTERVAL '5 days',
		    starts_at = NOW() - INTERVAL '2 days',
		    ends_at = NOW() + INTERVAL '3 days',
		    judging_ends_at = NOW() + INTERVAL '5 days',
		    allow_individual = true,
		    allow_team = true
		WHERE id = $1
	`, hackathonID)
	require.NoError(t, err, "Failed to publish hackathon")

	// Get permissions for regular user
	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/participations/permissions", hackathonID), regularUser.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get permissions: %s", string(body))

	data := tc.ParseJSON(body)
	permissions := data["permissions"].(map[string]interface{})

	assert.False(t, permissions["register"].(bool), "User should NOT be able to register during RUNNING stage")
}

func TestGetParticipationPermissions_StaffCannotRegisterAsParticipant(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	hackathonID := createDraftHackathon(tc, owner)

	// Manually set hackathon to published with registration stage dates
	_, err := tc.DB.Exec(context.Background(), `
		UPDATE hackathon.hackathons 
		SET state = 'published', 
		    published_at = NOW(),
		    registration_opens_at = NOW() - INTERVAL '5 days',
		    registration_closes_at = NOW() + INTERVAL '5 days',
		    starts_at = NOW() + INTERVAL '10 days',
		    ends_at = NOW() + INTERVAL '12 days',
		    judging_ends_at = NOW() + INTERVAL '15 days',
		    allow_individual = true,
		    allow_team = true
		WHERE id = $1
	`, hackathonID)
	require.NoError(t, err, "Failed to publish hackathon")

	// Get permissions for owner (who is staff by default)
	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/participations/permissions", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get permissions: %s", string(body))

	data := tc.ParseJSON(body)
	permissions := data["permissions"].(map[string]interface{})

	assert.False(t, permissions["register"].(bool), "Owner (staff member) should NOT be able to register as participant")
	assert.True(t, permissions["inviteStaff"].(bool), "Owner should be able to invite staff")
	assert.True(t, permissions["listParticipants"].(bool), "Owner should be able to list participants")
}

func TestGetParticipationPermissions_ParticipantCanUnregisterAndUpdate(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createDraftHackathon(tc, owner)

	// Manually set hackathon to published with registration stage dates
	_, err := tc.DB.Exec(context.Background(), `
		UPDATE hackathon.hackathons 
		SET state = 'published', 
		    published_at = NOW(),
		    registration_opens_at = NOW() - INTERVAL '5 days',
		    registration_closes_at = NOW() + INTERVAL '5 days',
		    starts_at = NOW() + INTERVAL '10 days',
		    ends_at = NOW() + INTERVAL '12 days',
		    judging_ends_at = NOW() + INTERVAL '15 days',
		    allow_individual = true,
		    allow_team = true
		WHERE id = $1
	`, hackathonID)
	require.NoError(t, err, "Failed to publish hackathon")

	// Register as participant
	registerBody := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "Test",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant.AccessToken, registerBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to register: %s", string(body))

	// Get permissions for participant
	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/participations/permissions", hackathonID), participant.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get permissions: %s", string(body))

	data := tc.ParseJSON(body)
	permissions := data["permissions"].(map[string]interface{})

	assert.False(t, permissions["register"].(bool), "Already registered user should not be able to register again")
	assert.True(t, permissions["unregister"].(bool), "Participant should be able to unregister")
	assert.True(t, permissions["switchParticipationMode"].(bool), "Participant should be able to switch mode")
	assert.True(t, permissions["updateParticipationProfile"].(bool), "Participant should be able to update profile")
	assert.True(t, permissions["listParticipants"].(bool), "Participant should be able to list participants")
}
