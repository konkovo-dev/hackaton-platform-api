package integration

import (
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

	hackathonID := createAndPublishHackathon(tc, owner)

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

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations:register", hackathonID), participant.AccessToken, registerBody)
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

	hackathonID := createAndPublishHackathon(tc, owner)

	registerBody := map[string]interface{}{
		"desired_status":  "PART_LOOKING_FOR_TEAM",
		"wished_role_ids": []string{},
		"motivation_text": "Want to find a great team",
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations:register", hackathonID), participant.AccessToken, registerBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to register: %s", string(body))

	data := tc.ParseJSON(body)
	participation := data["participation"].(map[string]interface{})
	assert.Equal(t, "PART_LOOKING_FOR_TEAM", participation["status"])
}

func TestRegisterTwiceShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	registerBody := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "First registration",
	}
	resp, _ := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations:register", hackathonID), participant.AccessToken, registerBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	registerBody["motivation_text"] = "Second registration"
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations:register", hackathonID), participant.AccessToken, registerBody)
	assert.Equal(t, http.StatusConflict, resp.StatusCode, "Should reject duplicate registration: %s", string(body))
}

func TestGetMyParticipation(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	registerBody := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "Test motivation",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations:register", hackathonID), participant.AccessToken, registerBody)

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

	hackathonID := createAndPublishHackathon(tc, owner)

	registerBody := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "Original motivation",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations:register", hackathonID), participant.AccessToken, registerBody)

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

	hackathonID := createAndPublishHackathon(tc, owner)

	registerBody := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "Test",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations:register", hackathonID), participant.AccessToken, registerBody)

	switchBody := map[string]interface{}{
		"new_status": "PART_LOOKING_FOR_TEAM",
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/me:switchMode", hackathonID), participant.AccessToken, switchBody)
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

	hackathonID := createAndPublishHackathon(tc, owner)

	registerBody := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "Test",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations:register", hackathonID), participant.AccessToken, registerBody)

	switchBody := map[string]interface{}{
		"new_status": "PART_INDIVIDUAL",
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/me:switchMode", hackathonID), participant.AccessToken, switchBody)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should reject switch to same status: %s", string(body))
}

func TestGetUserParticipation(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant1 := tc.RegisterUser()
	participant2 := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	registerBody := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "Test",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations:register", hackathonID), participant1.AccessToken, registerBody)
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations:register", hackathonID), participant2.AccessToken, registerBody)

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

	hackathonID := createAndPublishHackathon(tc, owner)

	registerBody := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "Test",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations:register", hackathonID), participant1.AccessToken, registerBody)

	registerBody["desired_status"] = "PART_LOOKING_FOR_TEAM"
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations:register", hackathonID), participant2.AccessToken, registerBody)

	listBody := map[string]interface{}{}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations:list", hackathonID), participant1.AccessToken, listBody)
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

	hackathonID := createAndPublishHackathon(tc, owner)

	registerBody1 := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "Test",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations:register", hackathonID), participant1.AccessToken, registerBody1)

	registerBody2 := map[string]interface{}{
		"desired_status":  "PART_LOOKING_FOR_TEAM",
		"motivation_text": "Test",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations:register", hackathonID), participant2.AccessToken, registerBody2)

	listBody := map[string]interface{}{
		"status_filter": map[string]interface{}{
			"statuses": []string{"PART_LOOKING_FOR_TEAM"},
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations:list", hackathonID), participant1.AccessToken, listBody)
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

	hackathonID := createAndPublishHackathon(tc, owner)

	listBody := map[string]interface{}{}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations:list", hackathonID), nonParticipant.AccessToken, listBody)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Non-participant should not list participants: %s", string(body))
}

func TestUnregisterFromHackathon(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	registerBody := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "Test",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations:register", hackathonID), participant.AccessToken, registerBody)

	unregisterBody := map[string]interface{}{}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/me:unregister", hackathonID), participant.AccessToken, unregisterBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to unregister: %s", string(body))

	resp, _ = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/participations/me", hackathonID), participant.AccessToken, nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Should not find participation after unregister")
}
