package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListHackathonStaff(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/staff", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list staff: %s", string(body))

	data := tc.ParseJSON(body)
	staff, ok := data["staff"].([]interface{})
	require.True(t, ok, "Staff array should be present")
	assert.GreaterOrEqual(t, len(staff), 1, "Should have at least owner")

	found := false
	for _, s := range staff {
		member := s.(map[string]interface{})
		if member["userId"] == owner.UserID {
			roles := member["roles"].([]interface{})
			assert.Contains(t, roles, "HX_ROLE_OWNER")
			found = true
			break
		}
	}
	assert.True(t, found, "Owner should be in staff list")
}

func TestListHackathonStaffAsNonStaffShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	nonStaff := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/staff", hackathonID), nonStaff.AccessToken, nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Non-staff should not list staff: %s", string(body))
}

func TestCreateStaffInvitation(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	invitee := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	inviteBody := map[string]interface{}{
		"target_user_id": invitee.UserID,
		"requested_role": "HX_ROLE_MENTOR",
		"message":        "We would love to have you as a mentor!",
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff-invitations", hackathonID), owner.AccessToken, inviteBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to create invitation: %s", string(body))

	data := tc.ParseJSON(body)
	assert.NotEmpty(t, data["invitationId"], "Invitation ID should be returned")
}

func TestCreateStaffInvitationAsNonOwnerShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	nonOwner := tc.RegisterUser()
	invitee := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	inviteBody := map[string]interface{}{
		"target_user_id": invitee.UserID,
		"requested_role": "HX_ROLE_MENTOR",
		"message":        "Test",
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff-invitations", hackathonID), nonOwner.AccessToken, inviteBody)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Non-owner should not create invitations: %s", string(body))
}

func TestListMyStaffInvitations(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	invitee := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	inviteBody := map[string]interface{}{
		"target_user_id": invitee.UserID,
		"requested_role": "HX_ROLE_MENTOR",
		"message":        "Join us!",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff-invitations", hackathonID), owner.AccessToken, inviteBody)

	listBody := map[string]interface{}{}
	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/users/me/staff-invitations/list", invitee.AccessToken, listBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list invitations: %s", string(body))

	data := tc.ParseJSON(body)
	invitations, ok := data["invitations"].([]interface{})
	require.True(t, ok, "Invitations array should be present")
	assert.GreaterOrEqual(t, len(invitations), 1, "Should have at least 1 invitation")

	invitation := invitations[0].(map[string]interface{})
	assert.Equal(t, invitee.UserID, invitation["targetUserId"])
	assert.Equal(t, "HX_ROLE_MENTOR", invitation["requestedRole"])
	assert.Equal(t, "STAFF_INVITATION_PENDING", invitation["status"])
}

func TestAcceptStaffInvitation(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	invitee := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	inviteBody := map[string]interface{}{
		"target_user_id": invitee.UserID,
		"requested_role": "HX_ROLE_MENTOR",
		"message":        "Join us!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff-invitations", hackathonID), owner.AccessToken, inviteBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	inviteData := tc.ParseJSON(body)
	invitationID := inviteData["invitationId"].(string)

	acceptBody := map[string]interface{}{}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/staff-invitations/%s/accept", invitationID), invitee.AccessToken, acceptBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to accept invitation: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/staff", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	staffData := tc.ParseJSON(body)
	staff := staffData["staff"].([]interface{})

	found := false
	for _, s := range staff {
		member := s.(map[string]interface{})
		if member["userId"] == invitee.UserID {
			roles := member["roles"].([]interface{})
			assert.Contains(t, roles, "HX_ROLE_MENTOR")
			found = true
			break
		}
	}
	assert.True(t, found, "Invitee should be in staff after accepting")
}

func TestRejectStaffInvitation(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	invitee := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	inviteBody := map[string]interface{}{
		"target_user_id": invitee.UserID,
		"requested_role": "HX_ROLE_JUDGE",
		"message":        "Join us!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff-invitations", hackathonID), owner.AccessToken, inviteBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	inviteData := tc.ParseJSON(body)
	invitationID := inviteData["invitationId"].(string)

	rejectBody := map[string]interface{}{}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/staff-invitations/%s/reject", invitationID), invitee.AccessToken, rejectBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to reject invitation: %s", string(body))

	listBody := map[string]interface{}{}
	resp, body = tc.DoAuthenticatedRequest("POST", "/v1/users/me/staff-invitations/list", invitee.AccessToken, listBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	listData := tc.ParseJSON(body)
	invitations := listData["invitations"].([]interface{})

	found := false
	for _, inv := range invitations {
		invitation := inv.(map[string]interface{})
		if invitation["invitationId"] == invitationID {
			assert.Equal(t, "STAFF_INVITATION_DECLINED", invitation["status"])
			found = true
			break
		}
	}
	assert.True(t, found, "Rejected invitation should be in list with DECLINED status")
}

func TestCancelStaffInvitation(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	invitee := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	inviteBody := map[string]interface{}{
		"target_user_id": invitee.UserID,
		"requested_role": "HX_ROLE_MENTOR",
		"message":        "Join us!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff-invitations", hackathonID), owner.AccessToken, inviteBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	inviteData := tc.ParseJSON(body)
	invitationID := inviteData["invitationId"].(string)

	cancelBody := map[string]interface{}{}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff-invitations/%s/cancel", hackathonID, invitationID), owner.AccessToken, cancelBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to cancel invitation: %s", string(body))

	listBody := map[string]interface{}{}
	resp, body = tc.DoAuthenticatedRequest("POST", "/v1/users/me/staff-invitations/list", invitee.AccessToken, listBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	listData := tc.ParseJSON(body)
	invitations := listData["invitations"].([]interface{})

	found := false
	for _, inv := range invitations {
		invitation := inv.(map[string]interface{})
		if invitation["invitationId"] == invitationID {
			assert.Equal(t, "STAFF_INVITATION_CANCELED", invitation["status"])
			found = true
			break
		}
	}
	assert.True(t, found, "Cancelled invitation should be in list with CANCELLED status")
}

func TestRemoveHackathonRole(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	inviteBody := map[string]interface{}{
		"target_user_id": mentor.UserID,
		"requested_role": "HX_ROLE_MENTOR",
		"message":        "Join us!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff-invitations", hackathonID), owner.AccessToken, inviteBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	inviteData := tc.ParseJSON(body)
	invitationID := inviteData["invitationId"].(string)

	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/staff-invitations/%s/accept", invitationID), mentor.AccessToken, map[string]interface{}{})

	removeBody := map[string]interface{}{
		"user_id": mentor.UserID,
		"role":    "HX_ROLE_MENTOR",
	}

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff/removeRole", hackathonID), owner.AccessToken, removeBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to remove role: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/staff", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	staffData := tc.ParseJSON(body)
	staff := staffData["staff"].([]interface{})

	for _, s := range staff {
		member := s.(map[string]interface{})
		assert.NotEqual(t, mentor.UserID, member["userId"], "Mentor should be removed from staff")
	}
}

func TestRemoveOwnerRoleShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	removeBody := map[string]interface{}{
		"user_id": owner.UserID,
		"role":    "HX_ROLE_OWNER",
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff/removeRole", hackathonID), owner.AccessToken, removeBody)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Should not allow removing owner role: %s", string(body))
}

func TestSelfRemoveHackathonRole(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	inviteBody := map[string]interface{}{
		"target_user_id": mentor.UserID,
		"requested_role": "HX_ROLE_MENTOR",
		"message":        "Join us!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff-invitations", hackathonID), owner.AccessToken, inviteBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	inviteData := tc.ParseJSON(body)
	invitationID := inviteData["invitationId"].(string)

	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/staff-invitations/%s/accept", invitationID), mentor.AccessToken, map[string]interface{}{})

	selfRemoveBody := map[string]interface{}{
		"role": "HX_ROLE_MENTOR",
	}

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff/selfRemoveRole", hackathonID), mentor.AccessToken, selfRemoveBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to self-remove role: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/staff", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	staffData := tc.ParseJSON(body)
	staff := staffData["staff"].([]interface{})

	for _, s := range staff {
		member := s.(map[string]interface{})
		assert.NotEqual(t, mentor.UserID, member["userId"], "Mentor should be removed after self-remove")
	}
}

func TestSelfRemoveOwnerRoleShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	selfRemoveBody := map[string]interface{}{
		"role": "HACKATHON_ROLE_OWNER",
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff/selfRemoveRole", hackathonID), owner.AccessToken, selfRemoveBody)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should not allow self-removing owner role: %s", string(body))
}

func TestInviteToOwnerRoleShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	invitee := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	inviteBody := map[string]interface{}{
		"target_user_id": invitee.UserID,
		"requested_role": "HX_ROLE_OWNER",
		"message":        "Become owner!",
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff-invitations", hackathonID), owner.AccessToken, inviteBody)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should not allow inviting to owner role: %s", string(body))
}

func TestAcceptInvitationNotAddressedToYouShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	invitee := tc.RegisterUser()
	other := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	inviteBody := map[string]interface{}{
		"target_user_id": invitee.UserID,
		"requested_role": "HX_ROLE_MENTOR",
		"message":        "Join us!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff-invitations", hackathonID), owner.AccessToken, inviteBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	inviteData := tc.ParseJSON(body)
	invitationID := inviteData["invitationId"].(string)

	acceptBody := map[string]interface{}{}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/staff-invitations/%s/accept", invitationID), other.AccessToken, acceptBody)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Should not allow accepting invitation not addressed to you: %s", string(body))
}

func TestListHackathonStaffInvitations(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	mentor1 := tc.RegisterUser()
	mentor2 := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	// Create first invitation
	inviteBody1 := map[string]interface{}{
		"target_user_id": mentor1.UserID,
		"requested_role": "HX_ROLE_MENTOR",
		"message":        "Join us as mentor!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff-invitations", hackathonID), owner.AccessToken, inviteBody1)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to create first invitation: %s", string(body))

	// Create second invitation
	inviteBody2 := map[string]interface{}{
		"target_user_id": mentor2.UserID,
		"requested_role": "HX_ROLE_JUDGE",
		"message":        "Join us as judge!",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff-invitations", hackathonID), owner.AccessToken, inviteBody2)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to create second invitation: %s", string(body))

	// List invitations as owner
	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/staff-invitations", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list invitations: %s", string(body))

	data := tc.ParseJSON(body)
	invitations, ok := data["invitations"].([]interface{})
	require.True(t, ok, "Invitations array should be present")
	assert.GreaterOrEqual(t, len(invitations), 2, "Should have at least 2 invitations")

	// Verify invitation details
	foundMentor := false
	foundJudge := false
	for _, inv := range invitations {
		invitation := inv.(map[string]interface{})
		assert.Equal(t, hackathonID, invitation["hackathonId"], "Hackathon ID should match")
		assert.Equal(t, owner.UserID, invitation["createdByUserId"], "Creator should be owner")
		assert.Equal(t, "STAFF_INVITATION_PENDING", invitation["status"], "Status should be pending")

		if invitation["targetUserId"] == mentor1.UserID {
			assert.Equal(t, "HX_ROLE_MENTOR", invitation["requestedRole"])
			foundMentor = true
		}
		if invitation["targetUserId"] == mentor2.UserID {
			assert.Equal(t, "HX_ROLE_JUDGE", invitation["requestedRole"])
			foundJudge = true
		}
	}
	assert.True(t, foundMentor, "Should find mentor invitation")
	assert.True(t, foundJudge, "Should find judge invitation")
}

func TestListHackathonStaffInvitationsAsNonOwnerShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	nonOwner := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	// Create invitation
	inviteBody := map[string]interface{}{
		"target_user_id": mentor.UserID,
		"requested_role": "HX_ROLE_MENTOR",
		"message":        "Join us!",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff-invitations", hackathonID), owner.AccessToken, inviteBody)

	// Try to list as non-owner
	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/staff-invitations", hackathonID), nonOwner.AccessToken, nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Non-owner should not list invitations: %s", string(body))
}

func TestListHackathonStaffInvitationsAsOrganizerShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	organizer := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	// Invite organizer
	inviteOrganizerBody := map[string]interface{}{
		"target_user_id": organizer.UserID,
		"requested_role": "HX_ROLE_ORGANIZER",
		"message":        "Be our organizer!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff-invitations", hackathonID), owner.AccessToken, inviteOrganizerBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	inviteData := tc.ParseJSON(body)
	organizerInvitationID := inviteData["invitationId"].(string)

	// Accept organizer invitation
	acceptBody := map[string]interface{}{}
	resp, _ = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/users/me/staff-invitations/%s/accept", organizerInvitationID), organizer.AccessToken, acceptBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Create mentor invitation as owner
	inviteMentorBody := map[string]interface{}{
		"target_user_id": mentor.UserID,
		"requested_role": "HX_ROLE_MENTOR",
		"message":        "Join us!",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/staff-invitations", hackathonID), owner.AccessToken, inviteMentorBody)

	// List invitations as organizer - should succeed
	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/staff-invitations", hackathonID), organizer.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Organizer should be able to list invitations: %s", string(body))

	data := tc.ParseJSON(body)
	invitations, ok := data["invitations"].([]interface{})
	require.True(t, ok, "Invitations array should be present")
	assert.GreaterOrEqual(t, len(invitations), 2, "Should have at least 2 invitations (organizer + mentor)")
}
