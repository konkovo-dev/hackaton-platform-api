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

	resp, body = tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/validate", hackathonID), owner.AccessToken, nil)
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

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/publish", hackathonID), owner.AccessToken, map[string]interface{}{})
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
	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/publish", hackathonID), owner.AccessToken, map[string]interface{}{})

	listBody := map[string]interface{}{
		"pageSize": 10,
	}
	resp, body = tc.DoRequest("POST", "/v1/hackathons/list", listBody, nil)
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
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathonID), participant.AccessToken, registerBody)
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

func TestListHackathonsWithFilters(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	now := time.Now()

	hackathon1 := createAndPublishHackathonCustom(tc, owner, map[string]interface{}{
		"name":              "Running Hackathon Test",
		"short_description": "Test running",
		"description":       "Full description for running hackathon",
		"location": map[string]interface{}{
			"online":  true,
			"country": "Russia",
			"city":    "Moscow",
		},
		"dates": map[string]interface{}{
			"registration_opens_at":  now.Add(1 * time.Hour).Format(time.RFC3339),
			"registration_closes_at": now.Add(5 * 24 * time.Hour).Format(time.RFC3339),
			"starts_at":              now.Add(10 * 24 * time.Hour).Format(time.RFC3339),
			"ends_at":                now.Add(12 * 24 * time.Hour).Format(time.RFC3339),
			"judging_ends_at":        now.Add(15 * 24 * time.Hour).Format(time.RFC3339),
		},
		"registration_policy": map[string]interface{}{
			"allow_individual": true,
			"allow_team":       true,
		},
		"limits": map[string]interface{}{
			"team_size_max": 5,
		},
	})

	_, err := tc.DB.Exec(context.Background(), fmt.Sprintf(`
		UPDATE %s 
		SET starts_at = $1,
		    stage = 'running'
		WHERE id = $2
	`, tc.HackathonDBName), now.Add(-1*time.Hour), hackathon1)
	require.NoError(tc.T, err, "Failed to update hackathon1 to RUNNING stage")

	hackathon2 := createAndPublishHackathonCustom(tc, owner, map[string]interface{}{
		"name":              "Upcoming Hackathon Test",
		"short_description": "Test upcoming",
		"description":       "Full description for upcoming hackathon",
		"location": map[string]interface{}{
			"online":  false,
			"country": "USA",
			"city":    "New York",
		},
		"dates": map[string]interface{}{
			"registration_opens_at":  now.Add(1 * time.Hour).Format(time.RFC3339),
			"registration_closes_at": now.Add(10 * 24 * time.Hour).Format(time.RFC3339),
			"starts_at":              now.Add(15 * 24 * time.Hour).Format(time.RFC3339),
			"ends_at":                now.Add(17 * 24 * time.Hour).Format(time.RFC3339),
			"judging_ends_at":        now.Add(20 * 24 * time.Hour).Format(time.RFC3339),
		},
		"registration_policy": map[string]interface{}{
			"allow_individual": true,
			"allow_team":       false,
		},
		"limits": map[string]interface{}{
			"team_size_max": 3,
		},
	})

	time.Sleep(1 * time.Second)

	t.Run("filter by stage", func(t *testing.T) {
		listBody := map[string]interface{}{
			"query": map[string]interface{}{
				"filter_groups": []map[string]interface{}{
					{
						"filters": []map[string]interface{}{
							{
								"field":        "stage",
								"operation":    "FILTER_OPERATION_EQUAL",
								"string_value": "running",
							},
						},
					},
				},
				"page": map[string]interface{}{
					"page_size": 50,
				},
			},
		}

		resp, body := tc.DoRequest("POST", "/v1/hackathons/list", listBody, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list hackathons: %s", string(body))

		data := tc.ParseJSON(body)
		hackathons := data["hackathons"].([]interface{})

		foundRunning := false
		for _, h := range hackathons {
			hackathon := h.(map[string]interface{})
			if hackathon["hackathonId"].(string) == hackathon1 {
				foundRunning = true
				assert.Equal(t, "HACKATHON_STAGE_RUNNING", hackathon["stage"])
			}
			if hackathon["hackathonId"].(string) == hackathon2 {
				t.Errorf("Found upcoming hackathon in running filter")
			}
		}
		assert.True(t, foundRunning, "Running hackathon should be in results")
	})

	t.Run("filter by location_city", func(t *testing.T) {
		listBody := map[string]interface{}{
			"query": map[string]interface{}{
				"filter_groups": []map[string]interface{}{
					{
						"filters": []map[string]interface{}{
							{
								"field":        "location_city",
								"operation":    "FILTER_OPERATION_EQUAL",
								"string_value": "New York",
							},
						},
					},
				},
				"page": map[string]interface{}{
					"page_size": 50,
				},
			},
		}

		resp, body := tc.DoRequest("POST", "/v1/hackathons/list", listBody, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		data := tc.ParseJSON(body)
		hackathons := data["hackathons"].([]interface{})

		foundNY := false
		for _, h := range hackathons {
			hackathon := h.(map[string]interface{})
			if hackathon["hackathonId"].(string) == hackathon2 {
				foundNY = true
				location := hackathon["location"].(map[string]interface{})
				assert.Equal(t, "New York", location["city"])
			}
			if hackathon["hackathonId"].(string) == hackathon1 {
				t.Errorf("Found Moscow hackathon in New York filter")
			}
		}
		assert.True(t, foundNY, "New York hackathon should be in results")
	})

	t.Run("text search by name", func(t *testing.T) {
		listBody := map[string]interface{}{
			"query": map[string]interface{}{
				"q": "Running",
				"page": map[string]interface{}{
					"page_size": 50,
				},
			},
		}

		resp, body := tc.DoRequest("POST", "/v1/hackathons/list", listBody, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		data := tc.ParseJSON(body)
		hackathons := data["hackathons"].([]interface{})

		foundRunning := false
		for _, h := range hackathons {
			hackathon := h.(map[string]interface{})
			if hackathon["hackathonId"].(string) == hackathon1 {
				foundRunning = true
			}
		}
		assert.True(t, foundRunning, "Running hackathon should be found by text search")
	})
}

func TestListHackathonsWithRoleFilters(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	judge := tc.RegisterUser()

	hackathon1 := createAndPublishHackathonCustom(tc, owner, map[string]interface{}{
		"name":              "Owner Hackathon",
		"short_description": "Test",
		"description":       "Full description for owner hackathon",
		"location": map[string]interface{}{
			"online":  true,
			"country": "Russia",
			"city":    "Moscow",
		},
		"dates": map[string]interface{}{
			"registration_opens_at":  time.Now().Add(-5 * 24 * time.Hour).Format(time.RFC3339),
			"registration_closes_at": time.Now().Add(5 * 24 * time.Hour).Format(time.RFC3339),
			"starts_at":              time.Now().Add(10 * 24 * time.Hour).Format(time.RFC3339),
			"ends_at":                time.Now().Add(12 * 24 * time.Hour).Format(time.RFC3339),
			"judging_ends_at":        time.Now().Add(15 * 24 * time.Hour).Format(time.RFC3339),
		},
		"registration_policy": map[string]interface{}{
			"allow_individual": true,
			"allow_team":       true,
		},
		"limits": map[string]interface{}{
			"team_size_max": 5,
		},
	})

	hackathon2 := createAndPublishHackathonCustom(tc, owner, map[string]interface{}{
		"name":              "Judge Hackathon",
		"short_description": "Test",
		"description":       "Full description for judge hackathon",
		"location": map[string]interface{}{
			"online":  true,
			"country": "Russia",
			"city":    "Moscow",
		},
		"dates": map[string]interface{}{
			"registration_opens_at":  time.Now().Add(-5 * 24 * time.Hour).Format(time.RFC3339),
			"registration_closes_at": time.Now().Add(5 * 24 * time.Hour).Format(time.RFC3339),
			"starts_at":              time.Now().Add(10 * 24 * time.Hour).Format(time.RFC3339),
			"ends_at":                time.Now().Add(12 * 24 * time.Hour).Format(time.RFC3339),
			"judging_ends_at":        time.Now().Add(15 * 24 * time.Hour).Format(time.RFC3339),
		},
		"registration_policy": map[string]interface{}{
			"allow_individual": true,
			"allow_team":       true,
		},
		"limits": map[string]interface{}{
			"team_size_max": 5,
		},
	})

	tc.AssignRole(hackathon2, owner.AccessToken, judge.UserID, "judge")

	_, err := tc.DB.Exec(context.Background(),
		`UPDATE hackathon.hackathons SET state = 'published', published_at = NOW() WHERE id = $1`,
		hackathon1)
	require.NoError(t, err, "Failed to set hackathon1 to published")

	_, err = tc.DB.Exec(context.Background(),
		`UPDATE hackathon.hackathons SET state = 'published', published_at = NOW() WHERE id = $1`,
		hackathon2)
	require.NoError(t, err, "Failed to set hackathon2 to published")

	time.Sleep(2 * time.Second)

	t.Run("filter by my_role=owner", func(t *testing.T) {
		listBody := map[string]interface{}{
			"query": map[string]interface{}{
				"filter_groups": []map[string]interface{}{
					{
						"filters": []map[string]interface{}{
							{
								"field":        "my_role",
								"operation":    "FILTER_OPERATION_EQUAL",
								"string_value": "owner",
							},
						},
					},
				},
				"page": map[string]interface{}{
					"page_size": 50,
				},
			},
		}

		resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons/list", owner.AccessToken, listBody)
		require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list hackathons: %s", string(body))

		data := tc.ParseJSON(body)
		hackathons := data["hackathons"].([]interface{})

		foundHack1 := false
		foundHack2 := false
		for _, h := range hackathons {
			hackathon := h.(map[string]interface{})
			if hackathon["hackathonId"].(string) == hackathon1 {
				foundHack1 = true
			}
			if hackathon["hackathonId"].(string) == hackathon2 {
				foundHack2 = true
			}
		}
		assert.True(t, foundHack1, "Owner should see hackathon1")
		assert.True(t, foundHack2, "Owner should see hackathon2")
	})

	t.Run("filter by my_role=judge", func(t *testing.T) {
		listBody := map[string]interface{}{
			"query": map[string]interface{}{
				"filter_groups": []map[string]interface{}{
					{
						"filters": []map[string]interface{}{
							{
								"field":        "my_role",
								"operation":    "FILTER_OPERATION_EQUAL",
								"string_value": "judge",
							},
						},
					},
				},
				"page": map[string]interface{}{
					"page_size": 50,
				},
			},
		}

		resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons/list", judge.AccessToken, listBody)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		data := tc.ParseJSON(body)
		hackathons := data["hackathons"].([]interface{})

		foundHack1 := false
		foundHack2 := false
		for _, h := range hackathons {
			hackathon := h.(map[string]interface{})
			if hackathon["hackathonId"].(string) == hackathon1 {
				foundHack1 = true
			}
			if hackathon["hackathonId"].(string) == hackathon2 {
				foundHack2 = true
			}
		}
		assert.False(t, foundHack1, "Judge should not see hackathon1")
		assert.True(t, foundHack2, "Judge should see hackathon2")
	})
}

func TestListHackathonsWithParticipationFilters(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant1 := tc.RegisterUser()
	participant2 := tc.RegisterUser()
	nonParticipant := tc.RegisterUser()

	// Create hackathon 1 where participant1 will register as individual
	hackathon1 := createAndPublishHackathonCustom(tc, owner, map[string]interface{}{
		"name":              "Individual Hackathon",
		"short_description": "Test individual participation",
		"description":       "Full description for individual hackathon",
		"location": map[string]interface{}{
			"online":  true,
			"country": "Russia",
			"city":    "Moscow",
		},
		"dates": map[string]interface{}{
			"registration_opens_at":  time.Now().Add(-5 * 24 * time.Hour).Format(time.RFC3339),
			"registration_closes_at": time.Now().Add(5 * 24 * time.Hour).Format(time.RFC3339),
			"starts_at":              time.Now().Add(10 * 24 * time.Hour).Format(time.RFC3339),
			"ends_at":                time.Now().Add(12 * 24 * time.Hour).Format(time.RFC3339),
			"judging_ends_at":        time.Now().Add(15 * 24 * time.Hour).Format(time.RFC3339),
		},
		"registration_policy": map[string]interface{}{
			"allow_individual": true,
			"allow_team":       true,
		},
		"limits": map[string]interface{}{
			"team_size_max": 5,
		},
	})

	// Create hackathon 2 where participant1 will register as looking_for_team
	hackathon2 := createAndPublishHackathonCustom(tc, owner, map[string]interface{}{
		"name":              "Team Hackathon",
		"short_description": "Test team participation",
		"description":       "Full description for team hackathon",
		"location": map[string]interface{}{
			"online":  true,
			"country": "Russia",
			"city":    "Moscow",
		},
		"dates": map[string]interface{}{
			"registration_opens_at":  time.Now().Add(-5 * 24 * time.Hour).Format(time.RFC3339),
			"registration_closes_at": time.Now().Add(5 * 24 * time.Hour).Format(time.RFC3339),
			"starts_at":              time.Now().Add(10 * 24 * time.Hour).Format(time.RFC3339),
			"ends_at":                time.Now().Add(12 * 24 * time.Hour).Format(time.RFC3339),
			"judging_ends_at":        time.Now().Add(15 * 24 * time.Hour).Format(time.RFC3339),
		},
		"registration_policy": map[string]interface{}{
			"allow_individual": true,
			"allow_team":       true,
		},
		"limits": map[string]interface{}{
			"team_size_max": 5,
		},
	})

	// Manually set hackathons to published state
	_, err := tc.DB.Exec(context.Background(),
		`UPDATE hackathon.hackathons SET state = 'published', published_at = NOW() WHERE id = $1`,
		hackathon1)
	require.NoError(t, err, "Failed to set hackathon1 to published")

	_, err = tc.DB.Exec(context.Background(),
		`UPDATE hackathon.hackathons SET state = 'published', published_at = NOW() WHERE id = $1`,
		hackathon2)
	require.NoError(t, err, "Failed to set hackathon2 to published")

	// Participant1 registers as individual in hackathon1
	register1Body := map[string]interface{}{
		"desired_status":  "PART_INDIVIDUAL",
		"motivation_text": "Excited to participate individually!",
	}
	resp, body := tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathon1), participant1.AccessToken, register1Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to register participant1 in hackathon1: %s", string(body))

	// Participant1 registers as looking_for_team in hackathon2
	register2Body := map[string]interface{}{
		"desired_status":  "PART_LOOKING_FOR_TEAM",
		"motivation_text": "Looking for a team!",
	}
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathon2), participant1.AccessToken, register2Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to register participant1 in hackathon2: %s", string(body))

	// Participant2 registers as individual in hackathon2
	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/participations/register", hackathon2), participant2.AccessToken, register1Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to register participant2 in hackathon2: %s", string(body))

	// Wait for participations to be synced
	time.Sleep(2 * time.Second)

	t.Run("filter by my_participation_status=individual", func(t *testing.T) {
		listBody := map[string]interface{}{
			"query": map[string]interface{}{
				"filter_groups": []map[string]interface{}{
					{
						"filters": []map[string]interface{}{
							{
								"field":        "my_participation_status",
								"operation":    "FILTER_OPERATION_EQUAL",
								"string_value": "individual",
							},
						},
					},
				},
				"page": map[string]interface{}{
					"page_size": 50,
				},
			},
		}

		resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons/list", participant1.AccessToken, listBody)
		require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list hackathons: %s", string(body))

		data := tc.ParseJSON(body)
		hackathons := data["hackathons"].([]interface{})

		foundHack1 := false
		foundHack2 := false
		for _, h := range hackathons {
			hackathon := h.(map[string]interface{})
			if hackathon["hackathonId"].(string) == hackathon1 {
				foundHack1 = true
			}
			if hackathon["hackathonId"].(string) == hackathon2 {
				foundHack2 = true
			}
		}
		assert.True(t, foundHack1, "Participant1 should see hackathon1 (registered as individual)")
		assert.False(t, foundHack2, "Participant1 should not see hackathon2 (registered as looking_for_team)")
	})

	t.Run("filter by my_participation_status=looking_for_team", func(t *testing.T) {
		listBody := map[string]interface{}{
			"query": map[string]interface{}{
				"filter_groups": []map[string]interface{}{
					{
						"filters": []map[string]interface{}{
							{
								"field":        "my_participation_status",
								"operation":    "FILTER_OPERATION_EQUAL",
								"string_value": "looking_for_team",
							},
						},
					},
				},
				"page": map[string]interface{}{
					"page_size": 50,
				},
			},
		}

		resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons/list", participant1.AccessToken, listBody)
		require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list hackathons: %s", string(body))

		data := tc.ParseJSON(body)
		hackathons := data["hackathons"].([]interface{})

		foundHack1 := false
		foundHack2 := false
		for _, h := range hackathons {
			hackathon := h.(map[string]interface{})
			if hackathon["hackathonId"].(string) == hackathon1 {
				foundHack1 = true
			}
			if hackathon["hackathonId"].(string) == hackathon2 {
				foundHack2 = true
			}
		}
		assert.False(t, foundHack1, "Participant1 should not see hackathon1 (registered as individual)")
		assert.True(t, foundHack2, "Participant1 should see hackathon2 (registered as looking_for_team)")
	})

	t.Run("filter by my_participation=true (any participation)", func(t *testing.T) {
		listBody := map[string]interface{}{
			"query": map[string]interface{}{
				"filter_groups": []map[string]interface{}{
					{
						"filters": []map[string]interface{}{
							{
								"field":      "my_participation",
								"operation":  "FILTER_OPERATION_EQUAL",
								"bool_value": true,
							},
						},
					},
				},
				"page": map[string]interface{}{
					"page_size": 50,
				},
			},
		}

		resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons/list", participant1.AccessToken, listBody)
		require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list hackathons: %s", string(body))

		data := tc.ParseJSON(body)
		hackathons := data["hackathons"].([]interface{})

		foundHack1 := false
		foundHack2 := false
		for _, h := range hackathons {
			hackathon := h.(map[string]interface{})
			if hackathon["hackathonId"].(string) == hackathon1 {
				foundHack1 = true
			}
			if hackathon["hackathonId"].(string) == hackathon2 {
				foundHack2 = true
			}
		}
		assert.True(t, foundHack1, "Participant1 should see hackathon1 (participating)")
		assert.True(t, foundHack2, "Participant1 should see hackathon2 (participating)")
	})

	t.Run("non-participant sees no hackathons with participation filter", func(t *testing.T) {
		listBody := map[string]interface{}{
			"query": map[string]interface{}{
				"filter_groups": []map[string]interface{}{
					{
						"filters": []map[string]interface{}{
							{
								"field":      "my_participation",
								"operation":  "FILTER_OPERATION_EQUAL",
								"bool_value": true,
							},
						},
					},
				},
				"page": map[string]interface{}{
					"page_size": 50,
				},
			},
		}

		resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons/list", nonParticipant.AccessToken, listBody)
		require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list hackathons: %s", string(body))

		data := tc.ParseJSON(body)
		hackathons := data["hackathons"].([]interface{})

		assert.Equal(t, 0, len(hackathons), "Non-participant should see no hackathons with participation filter")
	})
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

	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/publish", hackathonID), owner.AccessToken, map[string]interface{}{})

	return hackathonID
}

func createAndPublishHackathonCustom(tc *TestContext, owner *UserCredentials, hackathonData map[string]interface{}) string {
	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons", owner.AccessToken, hackathonData)
	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to create hackathon: %s", string(body))

	data := tc.ParseJSON(body)
	hackathonID := data["hackathonId"].(string)

	tc.WaitForHackathonOwnerRole(hackathonID, owner.AccessToken)

	taskBody := map[string]interface{}{
		"task": "Build something cool",
	}
	tc.DoAuthenticatedRequest("PUT", fmt.Sprintf("/v1/hackathons/%s/task", hackathonID), owner.AccessToken, taskBody)

	tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/publish", hackathonID), owner.AccessToken, map[string]interface{}{})

	time.Sleep(1 * time.Second)

	return hackathonID
}
