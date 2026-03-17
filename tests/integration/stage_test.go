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

func TestStageCalculation_Upcoming(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	now := time.Now().UTC()
	regOpens := now.Add(5 * 24 * time.Hour)  // +5 дней
	regCloses := now.Add(10 * 24 * time.Hour) // +10 дней
	starts := now.Add(11 * 24 * time.Hour)    // +11 дней
	ends := now.Add(13 * 24 * time.Hour)      // +13 дней
	judgingEnds := now.Add(15 * 24 * time.Hour) // +15 дней

	hackathonID := createDraftHackathon(tc, owner)

	// Manually update DB to simulate upcoming hackathon
	_, err := tc.DB.Exec(context.Background(), `
		UPDATE hackathon.hackathons 
		SET state = 'published', 
		    published_at = $2,
		    registration_opens_at = $3,
		    registration_closes_at = $4,
		    starts_at = $5,
		    ends_at = $6,
		    judging_ends_at = $7,
		    stage = 'upcoming'
		WHERE id = $1
	`, hackathonID, now, regOpens, regCloses, starts, ends, judgingEnds)
	require.NoError(t, err)

	// Get hackathon and check stage is recalculated
	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	hackathonData := tc.ParseJSON(body)
	hackathon := hackathonData["hackathon"].(map[string]interface{})
	assert.Equal(t, "HACKATHON_STAGE_UPCOMING", hackathon["stage"], "Stage should be UPCOMING when now < registration_opens_at")
}

func TestStageCalculation_Registration(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	now := time.Now().UTC()
	regOpens := now.Add(-1 * time.Hour)       // -1 час (уже началась)
	regCloses := now.Add(5 * 24 * time.Hour)  // +5 дней
	starts := now.Add(6 * 24 * time.Hour)     // +6 дней
	ends := now.Add(8 * 24 * time.Hour)       // +8 дней
	judgingEnds := now.Add(10 * 24 * time.Hour) // +10 дней

	hackathonID := createDraftHackathon(tc, owner)

	// Manually update DB to simulate registration stage hackathon
	_, err := tc.DB.Exec(context.Background(), `
		UPDATE hackathon.hackathons 
		SET state = 'published', 
		    published_at = $2,
		    registration_opens_at = $3,
		    registration_closes_at = $4,
		    starts_at = $5,
		    ends_at = $6,
		    judging_ends_at = $7,
		    stage = 'registration'
		WHERE id = $1
	`, hackathonID, now, regOpens, regCloses, starts, ends, judgingEnds)
	require.NoError(t, err)

	// Get hackathon and check stage is recalculated
	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	hackathonData := tc.ParseJSON(body)
	hackathon := hackathonData["hackathon"].(map[string]interface{})
	assert.Equal(t, "HACKATHON_STAGE_REGISTRATION", hackathon["stage"], "Stage should be REGISTRATION when registration_opens_at <= now < registration_closes_at")
}

func TestStageCalculation_Running(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	now := time.Now().UTC()
	regOpens := now.Add(-10 * 24 * time.Hour)  // -10 дней
	regCloses := now.Add(-5 * 24 * time.Hour)  // -5 дней
	starts := now.Add(-2 * 24 * time.Hour)     // -2 дня (уже началось)
	ends := now.Add(3 * 24 * time.Hour)        // +3 дня (еще идет)
	judgingEnds := now.Add(5 * 24 * time.Hour) // +5 дней

	hackathonID := createDraftHackathon(tc, owner)

	// Manually update DB to simulate running hackathon
	_, err := tc.DB.Exec(context.Background(), `
		UPDATE hackathon.hackathons 
		SET state = 'published', 
		    published_at = $2,
		    registration_opens_at = $3,
		    registration_closes_at = $4,
		    starts_at = $5,
		    ends_at = $6,
		    judging_ends_at = $7,
		    stage = 'running'
		WHERE id = $1
	`, hackathonID, now, regOpens, regCloses, starts, ends, judgingEnds)
	require.NoError(t, err)

	// Get hackathon and check stage is recalculated
	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	hackathonData := tc.ParseJSON(body)
	hackathon := hackathonData["hackathon"].(map[string]interface{})
	assert.Equal(t, "HACKATHON_STAGE_RUNNING", hackathon["stage"], "Stage should be RUNNING when starts_at <= now < ends_at")
}

func TestStageCalculation_Finished_ByResultPublished(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	now := time.Now().UTC()
	regOpens := now.Add(-10 * 24 * time.Hour)
	regCloses := now.Add(-5 * 24 * time.Hour)
	starts := now.Add(-3 * 24 * time.Hour)
	ends := now.Add(-1 * 24 * time.Hour)
	judgingEnds := now.Add(2 * 24 * time.Hour) // Judging еще не закончилось!
	resultPublished := now.Add(-1 * time.Hour)  // Но результаты опубликованы

	hackathonID := createDraftHackathon(tc, owner)

	// Manually update DB
	_, err := tc.DB.Exec(context.Background(), `
		UPDATE hackathon.hackathons 
		SET state = 'published', 
		    published_at = $2,
		    registration_opens_at = $3,
		    registration_closes_at = $4,
		    starts_at = $5,
		    ends_at = $6,
		    judging_ends_at = $7,
		    result_published_at = $8,
		    stage = 'judging'
		WHERE id = $1
	`, hackathonID, now, regOpens, regCloses, starts, ends, judgingEnds, resultPublished)
	require.NoError(t, err)

	// Get hackathon and check stage is recalculated to FINISHED
	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	hackathonData := tc.ParseJSON(body)
	hackathon := hackathonData["hackathon"].(map[string]interface{})
	assert.Equal(t, "HACKATHON_STAGE_FINISHED", hackathon["stage"], "Stage should be FINISHED when result_published_at != null (priority over judging_ends_at)")
}

func TestStageCalculation_InList(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	now := time.Now().UTC()

	// Create UPCOMING hackathon (manually)
	hackathonID1 := createDraftHackathon(tc, owner)
	regOpens1 := now.Add(5 * 24 * time.Hour)
	regCloses1 := now.Add(10 * 24 * time.Hour)
	starts1 := now.Add(11 * 24 * time.Hour)
	ends1 := now.Add(13 * 24 * time.Hour)
	judgingEnds1 := now.Add(15 * 24 * time.Hour)

	_, err := tc.DB.Exec(context.Background(), `
		UPDATE hackathon.hackathons 
		SET state = 'published', 
		    published_at = $2,
		    registration_opens_at = $3,
		    registration_closes_at = $4,
		    starts_at = $5,
		    ends_at = $6,
		    judging_ends_at = $7,
		    stage = 'upcoming'
		WHERE id = $1
	`, hackathonID1, now, regOpens1, regCloses1, starts1, ends1, judgingEnds1)
	require.NoError(t, err)

	// Create RUNNING hackathon (manually)
	hackathonID2 := createDraftHackathon(tc, owner)
	regOpens2 := now.Add(-10 * 24 * time.Hour)
	regCloses2 := now.Add(-5 * 24 * time.Hour)
	starts2 := now.Add(-2 * 24 * time.Hour)
	ends2 := now.Add(3 * 24 * time.Hour)
	judgingEnds2 := now.Add(5 * 24 * time.Hour)

	_, err = tc.DB.Exec(context.Background(), `
		UPDATE hackathon.hackathons 
		SET state = 'published', 
		    published_at = $2,
		    registration_opens_at = $3,
		    registration_closes_at = $4,
		    starts_at = $5,
		    ends_at = $6,
		    judging_ends_at = $7,
		    stage = 'running'
		WHERE id = $1
	`, hackathonID2, now, regOpens2, regCloses2, starts2, ends2, judgingEnds2)
	require.NoError(t, err)

	// List hackathons
	resp, body := tc.DoAuthenticatedRequest("GET", "/v1/hackathons", owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	listData := tc.ParseJSON(body)
	hackathons := listData["hackathons"].([]interface{})

	// Find our hackathons and verify stages
	foundUpcoming := false
	foundRunning := false

	for _, h := range hackathons {
		hackathon := h.(map[string]interface{})
		hID := hackathon["id"].(string)

		if hID == hackathonID1 {
			assert.Equal(t, "HACKATHON_STAGE_UPCOMING", hackathon["stage"], "First hackathon should be UPCOMING")
			foundUpcoming = true
		}
		if hID == hackathonID2 {
			assert.Equal(t, "HACKATHON_STAGE_RUNNING", hackathon["stage"], "Second hackathon should be RUNNING")
			foundRunning = true
		}
	}

	assert.True(t, foundUpcoming, "Should find upcoming hackathon in list")
	assert.True(t, foundRunning, "Should find running hackathon in list")
}

func createDraftHackathon(tc *TestContext, owner *UserCredentials) string {
	hackathonBody := map[string]interface{}{
		"name":              "Draft Hackathon",
		"short_description": "Draft",
		"allow_individual":  true,
		"allow_team":        true,
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/hackathons", owner.AccessToken, hackathonBody)
	if resp.StatusCode != http.StatusOK {
		tc.T.Fatalf("Failed to create hackathon: %s", string(body))
	}

	data := tc.ParseJSON(body)
	hackathonID := data["hackathonId"].(string)

	tc.WaitForHackathonOwnerRole(hackathonID, owner.AccessToken)

	return hackathonID
}
