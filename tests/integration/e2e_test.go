package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullHackathonFlow simulates a complete hackathon lifecycle:
// 1. Owner creates and publishes hackathon
// 2. Owner invites mentor and judge
// 3. Participants register (individual and looking for team)
// 4. Participants update profiles and switch modes
// 5. Participants view each other
// 6. Owner creates announcements
// 7. Participants unregister
func TestFullHackathonFlow(t *testing.T) {
	tc := NewTestContext(t)

	// Step 1: Register users
	t.Log("Step 1: Registering users...")
	owner := tc.RegisterUser()
	mentor := tc.RegisterUser()
	judge := tc.RegisterUser()
	participant1 := tc.RegisterUser()
	participant2 := tc.RegisterUser()
	participant3 := tc.RegisterUser()

	// Step 2: Owner creates hackathon
	t.Log("Step 2: Creating hackathon...")
	hackathonBody := map[string]interface{}{
		"name":              "Full E2E Test Hackathon",
		"short_description": "Complete integration test",
		"description":       "This hackathon tests the full platform flow from creation to completion.",
		"location": map[string]interface{}{
			"online":  false,
			"country": "Russia",
			"city":    "Moscow",
			"venue":   "Test Venue",
		},
		"dates": map[string]interface{}{
			"registration_opens_at":  time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			"registration_closes_at": time.Now().Add(15 * 24 * time.Hour).Format(time.RFC3339),
			"starts_at":              time.Now().Add(20 * 24 * time.Hour).Format(time.RFC3339),
			"ends_at":                time.Now().Add(22 * 24 * time.Hour).Format(time.RFC3339),
			"judging_ends_at":        time.Now().Add(25 * 24 * time.Hour).Format(time.RFC3339),
		},
		"registration_policy": map[string]interface{}{
			"allow_individual": true,
			"allow_team":       true,
		},
		"limits": map[string]interface{}{
			"team_size_max": 5,
		},
		"links": []map[string]interface{}{
			{
				"title": "Official Website",
				"url":   "https://test-hackathon.example.com",
			},
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons", owner.AccessToken, hackathonBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to create hackathon: %s", string(body))

	data := tc.ParseJSON(body)
	hackathonID := data["hackathonId"].(string)
	t.Logf("Created hackathon: %s", hackathonID)

	tc.WaitForHackathonOwnerRole(hackathonID, owner.AccessToken)

	// Step 3: Owner sets task
	t.Log("Step 3: Setting hackathon task...")
	taskBody := map[string]interface{}{
		"task": "Build an innovative solution that demonstrates creativity and technical excellence. Your project should solve a real-world problem.",
	}
	resp, body = tc.DoAuthenticatedRequest("PUT", fmt.Sprintf("/v1/hackathons/%s/task", hackathonID), owner.AccessToken, taskBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to set task: %s", string(body))

	// Step 4: Owner publishes hackathon
	t.Log("Step 4: Publishing hackathon...")
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/publish", hackathonID), owner.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to publish: %s", string(body))

	// Step 5: Owner invites staff
	t.Log("Step 5: Inviting staff...")

	// Invite mentor
	inviteMentorBody := map[string]interface{}{
		"target_user_id": mentor.UserID,
		"requested_role": "HACKATHON_ROLE_MENTOR",
		"message":        "We would love to have you as a mentor for this hackathon!",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff-invitations", hackathonID), owner.AccessToken, inviteMentorBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to invite mentor: %s", string(body))

	mentorInviteData := tc.ParseJSON(body)
	mentorInvitationID := mentorInviteData["invitationId"].(string)

	// Invite judge
	inviteJudgeBody := map[string]interface{}{
		"target_user_id": judge.UserID,
		"requested_role": "HACKATHON_ROLE_JUDGE",
		"message":        "We need your expertise as a judge!",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff-invitations", hackathonID), owner.AccessToken, inviteJudgeBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to invite judge: %s", string(body))

	judgeInviteData := tc.ParseJSON(body)
	judgeInvitationID := judgeInviteData["invitationId"].(string)

	// Step 6: Mentor and judge accept invitations
	t.Log("Step 6: Staff accepting invitations...")

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/staff-invitations/%s/accept", mentorInvitationID), mentor.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode, "Mentor failed to accept: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/staff-invitations/%s/accept", judgeInvitationID), judge.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode, "Judge failed to accept: %s", string(body))

	// Step 7: Verify staff list
	t.Log("Step 7: Verifying staff list...")
	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/staff", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	staffData := tc.ParseJSON(body)
	staff := staffData["staff"].([]interface{})
	assert.GreaterOrEqual(t, len(staff), 3, "Should have owner, mentor, and judge")

	// Step 8: Participants register
	t.Log("Step 8: Participants registering...")

	// Get team roles
	resp, body = tc.DoAuthenticatedRequest("GET", "/v1/team-roles", participant1.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	rolesData := tc.ParseJSON(body)
	roles := rolesData["teamRoles"].([]interface{})
	var frontendRoleID, backendRoleID string
	for _, r := range roles {
		role := r.(map[string]interface{})
		if role["name"] == "Frontend" {
			frontendRoleID = role["id"].(string)
		}
		if role["name"] == "Backend" {
			backendRoleID = role["id"].(string)
		}
	}

	// Participant1 registers as individual
	register1Body := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"wished_role_ids": []string{frontendRoleID},
		"motivation_text": "I'm passionate about frontend development and want to build something amazing!",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant1.AccessToken, register1Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Participant1 failed to register: %s", string(body))

	// Participant2 registers as looking for team
	register2Body := map[string]interface{}{
		"desired_status":  "PART_LOOKING_FOR_TEAM",
		"wished_role_ids": []string{backendRoleID},
		"motivation_text": "Looking for a team to collaborate with on backend development.",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant2.AccessToken, register2Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Participant2 failed to register: %s", string(body))

	// Participant3 registers as individual
	register3Body := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "Excited to participate!",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant3.AccessToken, register3Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Participant3 failed to register: %s", string(body))

	// Step 9: Participant1 updates profile
	t.Log("Step 9: Updating participant profiles...")
	updateBody := map[string]interface{}{
		"first_name": "UpdatedName",
		"last_name":  "UpdatedLastName",
	}
	tc.DoAuthenticatedRequest("PUT", "/v1/users/me", participant1.AccessToken, updateBody)

	// Step 10: Participant1 switches to looking for team
	t.Log("Step 10: Participant1 switching mode...")
	switchBody := map[string]interface{}{
		"new_status": "PART_LOOKING_FOR_TEAM",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/me/switchMode", hackathonID), participant1.AccessToken, switchBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to switch mode: %s", string(body))

	// Step 11: Participants view each other
	t.Log("Step 11: Participants viewing each other...")
	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/participations/users/%s", hackathonID, participant2.UserID), participant1.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to view other participant: %s", string(body))

	// Step 12: List all participants
	t.Log("Step 12: Listing all participants...")
	listBody := map[string]interface{}{}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/list", hackathonID), participant1.AccessToken, listBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list participants: %s", string(body))

	participantsData := tc.ParseJSON(body)
	participants := participantsData["participants"].([]interface{})
	assert.GreaterOrEqual(t, len(participants), 3, "Should have at least 3 participants")

	// Step 13: Filter participants by status
	t.Log("Step 13: Filtering participants by status...")
	filterBody := map[string]interface{}{
		"status_filter": map[string]interface{}{
			"statuses": []string{"PART_LOOKING_FOR_TEAM"},
		},
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/list", hackathonID), participant1.AccessToken, filterBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to filter participants: %s", string(body))

	filteredData := tc.ParseJSON(body)
	filteredParticipants := filteredData["participants"].([]interface{})
	assert.GreaterOrEqual(t, len(filteredParticipants), 2, "Should have at least 2 participants looking for team")

	// Step 14: Owner creates announcement
	t.Log("Step 14: Creating announcement...")
	announcementBody := map[string]interface{}{
		"title":   "Important Update!",
		"content": "Registration is going great! We have many talented participants joining us.",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/announcements", hackathonID), owner.AccessToken, announcementBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to create announcement: %s", string(body))

	// Step 15: Participants view announcements
	t.Log("Step 15: Viewing announcements...")
	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/announcements", hackathonID), participant1.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to view announcements: %s", string(body))

	announcementsData := tc.ParseJSON(body)
	announcements := announcementsData["announcements"].([]interface{})
	assert.GreaterOrEqual(t, len(announcements), 1, "Should have at least 1 announcement")

	// Step 16: Participant3 unregisters
	t.Log("Step 16: Participant3 unregistering...")
	unregisterBody := map[string]interface{}{}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/me/unregister", hackathonID), participant3.AccessToken, unregisterBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to unregister: %s", string(body))

	// Step 17: Verify participant3 is no longer registered
	t.Log("Step 17: Verifying unregistration...")
	resp, _ = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/participations/me", hackathonID), participant3.AccessToken, nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Unregistered user should not access participation")

	// Step 18: List hackathons (public)
	t.Log("Step 18: Listing public hackathons...")
	resp, body = tc.DoRequest("GET", "/v1/hackathons?page_size=10", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list hackathons: %s", string(body))

	listHackData := tc.ParseJSON(body)
	hackathons := listHackData["hackathons"].([]interface{})

	found := false
	for _, h := range hackathons {
		hack := h.(map[string]interface{})
		if hack["hackathonId"] == hackathonID {
			found = true
			assert.Equal(t, "Full E2E Test Hackathon", hack["name"])
			break
		}
	}
	assert.True(t, found, "Published hackathon should appear in public list")

	// Step 19: Mentor self-removes
	t.Log("Step 19: Mentor self-removing...")
	selfRemoveBody := map[string]interface{}{
		"role": "HACKATHON_ROLE_MENTOR",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff/selfRemoveRole", hackathonID), mentor.AccessToken, selfRemoveBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to self-remove: %s", string(body))

	// Step 20: Verify final state
	t.Log("Step 20: Verifying final state...")

	// Check staff list
	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/staff", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	finalStaffData := tc.ParseJSON(body)
	finalStaff := finalStaffData["staff"].([]interface{})

	mentorFound := false
	for _, s := range finalStaff {
		member := s.(map[string]interface{})
		if member["userId"] == mentor.UserID {
			mentorFound = true
			break
		}
	}
	assert.False(t, mentorFound, "Mentor should be removed from staff")

	// Check participants list
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/list", hackathonID), participant1.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	finalParticipantsData := tc.ParseJSON(body)
	finalParticipants := finalParticipantsData["participants"].([]interface{})
	assert.Equal(t, 2, len(finalParticipants), "Should have 2 participants after unregistration")

	t.Log("✓ Full E2E test completed successfully!")
}
