package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateHackathon(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	hackathonBody := map[string]interface{}{
		"name":              "AI Innovation Hackathon 2026",
		"short_description": "Build the future with AI",
		"description":       "Join us for an exciting 48-hour hackathon focused on AI and machine learning innovations.",
		"location": map[string]interface{}{
			"online":  false,
			"country": "Russia",
			"city":    "Moscow",
			"venue":   "Digital October Center",
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
				"url":   "https://ai-hackathon.example.com",
			},
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons", owner.AccessToken, hackathonBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to create hackathon: %s", string(body))

	data := tc.ParseJSON(body)
	assert.NotEmpty(t, data["hackathonId"], "Hackathon ID should be returned")
}

func TestGetHackathonDraft(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	hackathonBody := map[string]interface{}{
		"name":              "Test Hackathon",
		"short_description": "Test",
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
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons", owner.AccessToken, hackathonBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	hackathonID := data["hackathonId"].(string)

	tc.WaitForHackathonOwnerRole(hackathonID, owner.AccessToken)

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s?include_description=true&include_links=true", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get draft hackathon: %s", string(body))

	hackData := tc.ParseJSON(body)
	hackathon := hackData["hackathon"].(map[string]interface{})
	assert.Equal(t, "Test Hackathon", hackathon["name"])
	assert.Equal(t, "HACKATHON_STAGE_DRAFT", hackathon["stage"])
}

func TestGetHackathonDraftForbiddenForNonOwner(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	other := tc.RegisterUser()

	hackathonBody := map[string]interface{}{
		"name": "Private Draft Hackathon",
		"location": map[string]interface{}{
			"online": true,
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
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons", owner.AccessToken, hackathonBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	hackathonID := data["hackathonId"].(string)

	resp, _ = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s", hackathonID), other.AccessToken, nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Non-owner should not access DRAFT hackathon")
}

func TestUpdateHackathon(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	hackathonBody := map[string]interface{}{
		"name": "Original Name",
		"location": map[string]interface{}{
			"online": false,
			"city":   "Moscow",
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
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons", owner.AccessToken, hackathonBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	hackathonID := data["hackathonId"].(string)

	tc.WaitForHackathonOwnerRole(hackathonID, owner.AccessToken)

	updateBody := map[string]interface{}{
		"name":              "Updated Name",
		"short_description": "Updated description",
		"location": map[string]interface{}{
			"online": false,
			"city":   "Saint Petersburg",
		},
		"dates": hackathonBody["dates"],
		"registration_policy": map[string]interface{}{
			"allow_individual": true,
			"allow_team":       true,
		},
		"limits": map[string]interface{}{
			"team_size_max": 6,
		},
	}

	resp, body = tc.DoAuthenticatedRequest("PUT", fmt.Sprintf("/v1/hackathons/%s", hackathonID), owner.AccessToken, updateBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to update hackathon: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	hackData := tc.ParseJSON(body)
	hackathon := hackData["hackathon"].(map[string]interface{})
	assert.Equal(t, "Updated Name", hackathon["name"])
}

func TestValidateHackathon(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	hackathonBody := map[string]interface{}{
		"name": "Incomplete Hackathon",
		"location": map[string]interface{}{
			"online": false,
			"city":   "Moscow",
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
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons", owner.AccessToken, hackathonBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	hackathonID := data["hackathonId"].(string)

	tc.WaitForHackathonOwnerRole(hackathonID, owner.AccessToken)

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s:validate", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Validate should return 200: %s", string(body))

	validateData := tc.ParseJSON(body)
	errors, ok := validateData["validationErrors"].([]interface{})
	assert.True(t, ok && len(errors) > 0, "Should return validation errors for incomplete hackathon")
}

func TestPublishHackathon(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	hackathonBody := map[string]interface{}{
		"name":              "Complete Hackathon",
		"short_description": "Test",
		"description":       "Full description",
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
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons", owner.AccessToken, hackathonBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	hackathonID := data["hackathonId"].(string)

	tc.WaitForHackathonOwnerRole(hackathonID, owner.AccessToken)

	taskBody := map[string]interface{}{
		"task": "Build an innovative AI solution that solves a real-world problem.",
	}
	resp, body = tc.DoAuthenticatedRequest("PUT", fmt.Sprintf("/v1/hackathons/%s/task", hackathonID), owner.AccessToken, taskBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to set task: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s:publish", hackathonID), owner.AccessToken, map[string]interface{}{})
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to publish hackathon: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	hackData := tc.ParseJSON(body)
	hackathon := hackData["hackathon"].(map[string]interface{})
	assert.NotEqual(t, "HACKATHON_STAGE_DRAFT", hackathon["stage"], "Should not be DRAFT after publish")
	assert.NotEmpty(t, hackathon["publishedAt"], "Published timestamp should be set")
}

func TestGetHackathonTask(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	hackathonBody := map[string]interface{}{
		"name": "Hackathon with Task",
		"location": map[string]interface{}{
			"online": true,
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
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons", owner.AccessToken, hackathonBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	hackathonID := data["hackathonId"].(string)

	tc.WaitForHackathonOwnerRole(hackathonID, owner.AccessToken)

	taskBody := map[string]interface{}{
		"task": "Build a cool AI project",
	}
	tc.DoAuthenticatedRequest("PUT", fmt.Sprintf("/v1/hackathons/%s/task", hackathonID), owner.AccessToken, taskBody)

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/task", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get task: %s", string(body))

	taskData := tc.ParseJSON(body)
	task := taskData["task"].(string)
	assert.Equal(t, "Build a cool AI project", task)
}

func TestUpdateHackathonTask(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	hackathonBody := map[string]interface{}{
		"name": "Hackathon",
		"location": map[string]interface{}{
			"online": true,
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
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons", owner.AccessToken, hackathonBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	hackathonID := data["hackathonId"].(string)

	tc.WaitForHackathonOwnerRole(hackathonID, owner.AccessToken)

	taskBody := map[string]interface{}{
		"task": "Updated task description",
	}
	resp, body = tc.DoAuthenticatedRequest("PUT", fmt.Sprintf("/v1/hackathons/%s/task", hackathonID), owner.AccessToken, taskBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to update task: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/task", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	taskData := tc.ParseJSON(body)
	task := taskData["task"].(string)
	assert.Equal(t, "Updated task description", task)
}

func TestListHackathons(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	hackathonBody := map[string]interface{}{
		"name":              "Public Hackathon",
		"short_description": "Test",
		"description":       "Full description",
		"location": map[string]interface{}{
			"online": true,
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
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons", owner.AccessToken, hackathonBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	hackathonID := data["hackathonId"].(string)

	tc.WaitForHackathonOwnerRole(hackathonID, owner.AccessToken)

	taskBody := map[string]interface{}{
		"task": "Build something cool",
	}
	tc.DoAuthenticatedRequest("PUT", fmt.Sprintf("/v1/hackathons/%s/task", hackathonID), owner.AccessToken, taskBody)
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s:publish", hackathonID), owner.AccessToken, map[string]interface{}{})

	listBody := map[string]interface{}{
		"pageSize": 10,
	}
	resp, body = tc.DoRequest("POST", "/v1/hackathons:list", listBody, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list hackathons: %s", string(body))

	listData := tc.ParseJSON(body)
	hackathons, ok := listData["hackathons"].([]interface{})
	assert.True(t, ok, "Hackathons array should be present")

	found := false
	for _, h := range hackathons {
		hack := h.(map[string]interface{})
		if hack["hackathonId"] == hackathonID {
			found = true
			break
		}
	}
	assert.True(t, found, "Published hackathon should appear in list")
}

func TestCreateHackathonAnnouncement(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	announcementBody := map[string]interface{}{
		"title":   "Registration Opening Soon!",
		"content": "We are excited to announce that registration will open on March 1st, 2026.",
	}

	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/announcements", hackathonID), owner.AccessToken, announcementBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to create announcement: %s", string(body))

	data := tc.ParseJSON(body)
	announcementID := data["announcementId"].(string)
	assert.NotEmpty(t, announcementID, "Announcement ID should be returned")
}

func TestListHackathonAnnouncements(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	announcementBody := map[string]interface{}{
		"title":   "Test Announcement",
		"content": "Test content",
	}
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/announcements", hackathonID), owner.AccessToken, announcementBody)

	registerBody := map[string]interface{}{
		"desired_status": "PART_INDIVIDUAL",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations:register", hackathonID), participant.AccessToken, registerBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to register participant: %s", string(body))

	tc.WaitForParticipationRegistered(hackathonID, participant.AccessToken)

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/announcements?page_size=10", hackathonID), participant.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list announcements: %s", string(body))

	data := tc.ParseJSON(body)
	announcements, ok := data["announcements"].([]interface{})
	assert.True(t, ok && len(announcements) > 0, "Should have announcements")
}

func TestUpdateHackathonAnnouncement(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	announcementBody := map[string]interface{}{
		"title":   "Original Title",
		"content": "Original content",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/announcements", hackathonID), owner.AccessToken, announcementBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	announcementID := data["announcementId"].(string)

	updateBody := map[string]interface{}{
		"title":   "Updated Title",
		"content": "Updated content",
	}
	resp, body = tc.DoAuthenticatedRequest("PUT", fmt.Sprintf("/v1/hackathons/%s/announcements/%s", hackathonID, announcementID), owner.AccessToken, updateBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to update announcement: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/announcements", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	listData := tc.ParseJSON(body)
	announcements := listData["announcements"].([]interface{})
	updated := announcements[0].(map[string]interface{})
	assert.Equal(t, "Updated Title", updated["title"])
}

func TestDeleteHackathonAnnouncement(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	hackathonID := createAndPublishHackathon(tc, owner)

	announcementBody := map[string]interface{}{
		"title":   "To Be Deleted",
		"content": "Test",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/announcements", hackathonID), owner.AccessToken, announcementBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	announcementID := data["announcementId"].(string)

	resp, _ = tc.DoAuthenticatedRequest("DELETE", fmt.Sprintf("/v1/hackathons/%s/announcements/%s", hackathonID, announcementID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to delete announcement")

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/announcements", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	listData := tc.ParseJSON(body)
	announcements, ok := listData["announcements"].([]interface{})
	assert.True(t, !ok || len(announcements) == 0, "Announcement should be deleted")
}

func createAndPublishHackathon(tc *TestContext, owner *UserCredentials) string {
	hackathonBody := map[string]interface{}{
		"name":              "Test Hackathon",
		"short_description": "Test",
		"description":       "Full description",
		"location": map[string]interface{}{
			"online": true,
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
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons", owner.AccessToken, hackathonBody)
	require.Equal(tc.T, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	hackathonID := data["hackathonId"].(string)

	tc.WaitForHackathonOwnerRole(hackathonID, owner.AccessToken)

	taskBody := map[string]interface{}{
		"task": "Build something cool",
	}
	tc.DoAuthenticatedRequest("PUT", fmt.Sprintf("/v1/hackathons/%s/task", hackathonID), owner.AccessToken, taskBody)

	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s:publish", hackathonID), owner.AccessToken, map[string]interface{}{})

	return hackathonID
}
