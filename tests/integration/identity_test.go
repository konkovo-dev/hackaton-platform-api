package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMe(t *testing.T) {
	tc := NewTestContext(t)
	creds := tc.RegisterUser()

	resp, body := tc.DoAuthenticatedRequest("GET", "/v1/users/me", creds.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get profile: %s", string(body))

	data := tc.ParseJSON(body)
	user := data["user"].(map[string]interface{})
	assert.Equal(t, creds.UserID, user["userId"], "User ID should match")
	assert.NotEmpty(t, user["username"], "Username should be present")
	assert.NotEmpty(t, user["firstName"], "First name should be present")
	assert.NotEmpty(t, user["lastName"], "Last name should be present")
}

func TestUpdateMe(t *testing.T) {
	tc := NewTestContext(t)
	creds := tc.RegisterUser()

	updateBody := map[string]interface{}{
		"first_name": "UpdatedFirst",
		"last_name":  "UpdatedLast",
		"avatar_url": "https://example.com/avatar.jpg",
		"timezone":   "Europe/Moscow",
	}

	resp, body := tc.DoAuthenticatedRequest("PUT", "/v1/users/me", creds.AccessToken, updateBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to update profile: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("GET", "/v1/users/me", creds.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	user := data["user"].(map[string]interface{})
	assert.Equal(t, "UpdatedFirst", user["firstName"])
	assert.Equal(t, "UpdatedLast", user["lastName"])
	assert.Equal(t, "https://example.com/avatar.jpg", user["avatarUrl"])
	assert.Equal(t, "Europe/Moscow", user["timezone"])
}

func TestUpdateMySkillsWithCatalogAndCustom(t *testing.T) {
	tc := NewTestContext(t)
	creds := tc.RegisterUser()

	resp, body := tc.DoRequest("POST", "/v1/skills/list", map[string]interface{}{
		"query": map[string]interface{}{
			"page": map[string]interface{}{
				"page_size": 5,
			},
		},
	}, map[string]string{"Authorization": "Bearer " + creds.AccessToken})
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
		"user_skills":       []string{"Go", "PostgreSQL", "Docker"},
		"skills_visibility": "VISIBILITY_LEVEL_PUBLIC",
	}

	if catalogSkillID != "" {
		updateBody["catalog_skill_ids"] = []string{catalogSkillID}
	}

	resp, body = tc.DoAuthenticatedRequest("PUT", "/v1/users/me/skills", creds.AccessToken, updateBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to update skills: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("GET", "/v1/users/me", creds.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	userSkills, ok := data["skills"].([]interface{})
	assert.True(t, ok && len(userSkills) > 0, "Skills should be updated")
}

func TestUpdateMyContactsWithMultipleTypes(t *testing.T) {
	tc := NewTestContext(t)
	creds := tc.RegisterUser()

	updateBody := map[string]interface{}{
		"contacts": []map[string]interface{}{
			{
				"contact": map[string]interface{}{
					"type":  "CONTACT_TYPE_EMAIL",
					"value": "work@example.com",
				},
				"visibility": "VISIBILITY_LEVEL_PUBLIC",
			},
			{
				"contact": map[string]interface{}{
					"type":  "CONTACT_TYPE_TELEGRAM",
					"value": "@testuser_tg",
				},
				"visibility": "VISIBILITY_LEVEL_PRIVATE",
			},
			{
				"contact": map[string]interface{}{
					"type":  "CONTACT_TYPE_GITHUB",
					"value": "github.com/testuser",
				},
				"visibility": "VISIBILITY_LEVEL_PUBLIC",
			},
		},
		"contacts_visibility": "VISIBILITY_LEVEL_PUBLIC",
	}

	resp, body := tc.DoAuthenticatedRequest("PUT", "/v1/users/me/contacts", creds.AccessToken, updateBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to update contacts: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("GET", "/v1/users/me", creds.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	contacts, ok := data["contacts"].([]interface{})
	require.True(t, ok, "Contacts should be present")
	assert.GreaterOrEqual(t, len(contacts), 3, "Should have at least 3 contacts")
}

func TestListSkillCatalogBasic(t *testing.T) {
	tc := NewTestContext(t)
	creds := tc.RegisterUser()

	reqBody := map[string]interface{}{
		"query": map[string]interface{}{
			"page": map[string]interface{}{
				"page_size": 10,
			},
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/skills/list", creds.AccessToken, reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to list skills: %s", string(body))

	data := tc.ParseJSON(body)
	skills, ok := data["skills"].([]interface{})
	assert.True(t, ok, "Skills array should be present")

	if len(skills) > 0 {
		skill := skills[0].(map[string]interface{})
		assert.NotEmpty(t, skill["id"], "Skill should have ID")
		assert.NotEmpty(t, skill["name"], "Skill should have name")
	}
}

func TestListSkillCatalogWithSearch(t *testing.T) {
	tc := NewTestContext(t)
	creds := tc.RegisterUser()

	reqBody := map[string]interface{}{
		"query": map[string]interface{}{
			"q": "script",
			"page": map[string]interface{}{
				"page_size": 10,
			},
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/skills/list", creds.AccessToken, reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to search skills: %s", string(body))

	data := tc.ParseJSON(body)
	skills, ok := data["skills"].([]interface{})
	assert.True(t, ok, "Skills array should be present")

	if len(skills) > 0 {
		for _, s := range skills {
			skill := s.(map[string]interface{})
			t.Logf("Found skill: %s", skill["name"])
		}
	}
}

func TestListSkillCatalogWithFilterContains(t *testing.T) {
	tc := NewTestContext(t)
	creds := tc.RegisterUser()

	reqBody := map[string]interface{}{
		"query": map[string]interface{}{
			"filter_groups": []map[string]interface{}{
				{
					"filters": []map[string]interface{}{
						{
							"field":        "name",
							"operation":    "FILTER_OPERATION_CONTAINS",
							"string_value": "go",
						},
					},
				},
			},
			"page": map[string]interface{}{
				"page_size": 10,
			},
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/skills/list", creds.AccessToken, reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to filter skills: %s", string(body))

	data := tc.ParseJSON(body)
	skills, ok := data["skills"].([]interface{})
	assert.True(t, ok, "Skills array should be present")

	if len(skills) > 0 {
		skill := skills[0].(map[string]interface{})
		t.Logf("Found skill: %s", skill["name"])
	}
}

func TestListSkillCatalogWithFilterPrefix(t *testing.T) {
	tc := NewTestContext(t)
	creds := tc.RegisterUser()

	reqBody := map[string]interface{}{
		"query": map[string]interface{}{
			"filter_groups": []map[string]interface{}{
				{
					"filters": []map[string]interface{}{
						{
							"field":        "name",
							"operation":    "FILTER_OPERATION_PREFIX",
							"string_value": "Ja",
						},
					},
				},
			},
			"page": map[string]interface{}{
				"page_size": 10,
			},
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/skills/list", creds.AccessToken, reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to filter skills by prefix: %s", string(body))

	data := tc.ParseJSON(body)
	skills, ok := data["skills"].([]interface{})
	assert.True(t, ok, "Skills array should be present")

	for _, s := range skills {
		skill := s.(map[string]interface{})
		t.Logf("Found skill with prefix 'Ja': %s", skill["name"])
	}
}

func TestListSkillCatalogWithSortDesc(t *testing.T) {
	tc := NewTestContext(t)
	creds := tc.RegisterUser()

	reqBody := map[string]interface{}{
		"query": map[string]interface{}{
			"sort": []map[string]interface{}{
				{
					"field":     "name",
					"direction": "SORT_DIRECTION_DESC",
				},
			},
			"page": map[string]interface{}{
				"page_size": 5,
			},
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/skills/list", creds.AccessToken, reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to sort skills: %s", string(body))

	data := tc.ParseJSON(body)
	skills, ok := data["skills"].([]interface{})
	assert.True(t, ok, "Skills array should be present")

	if len(skills) >= 2 {
		skill1 := skills[0].(map[string]interface{})
		skill2 := skills[1].(map[string]interface{})
		assert.GreaterOrEqual(t, skill1["name"].(string), skill2["name"].(string), "Should be sorted DESC")
	}
}

func TestListSkillCatalogWithPagination(t *testing.T) {
	tc := NewTestContext(t)
	creds := tc.RegisterUser()

	reqBody := map[string]interface{}{
		"query": map[string]interface{}{
			"page": map[string]interface{}{
				"page_size": 5,
			},
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/skills/list", creds.AccessToken, reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	page, ok := data["page"].(map[string]interface{})
	require.True(t, ok, "Page info should be present")

	nextToken, hasNext := page["nextPageToken"].(string)
	if !hasNext || nextToken == "" {
		t.Skip("Not enough skills for pagination test")
		return
	}

	reqBody["query"].(map[string]interface{})["page"].(map[string]interface{})["page_token"] = nextToken

	resp, body = tc.DoAuthenticatedRequest("POST", "/v1/skills/list", creds.AccessToken, reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get second page: %s", string(body))

	data2 := tc.ParseJSON(body)
	skills2, ok := data2["skills"].([]interface{})
	assert.True(t, ok && len(skills2) > 0, "Second page should have skills")
}

func TestGetUser(t *testing.T) {
	tc := NewTestContext(t)
	user1 := tc.RegisterUser()
	user2 := tc.RegisterUser()

	resp, body := tc.DoAuthenticatedRequest("GET", "/v1/users/"+user2.UserID, user1.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get user: %s", string(body))

	data := tc.ParseJSON(body)
	user := data["user"].(map[string]interface{})
	assert.Equal(t, user2.UserID, user["userId"])
	assert.NotEmpty(t, user["username"], "Username should be present")
}

func TestGetUserWithIncludeSkillsAndContacts(t *testing.T) {
	tc := NewTestContext(t)
	user1 := tc.RegisterUser()
	user2 := tc.RegisterUser()

	updateSkills := map[string]interface{}{
		"catalog_skill_ids": []string{},
		"user_skills":       []string{"React", "Node.js"},
		"skills_visibility": "VISIBILITY_LEVEL_PUBLIC",
	}
	tc.DoAuthenticatedRequest("PUT", "/v1/users/me/skills", user2.AccessToken, updateSkills)

	updateContacts := map[string]interface{}{
		"contacts": []map[string]interface{}{
			{
				"contact": map[string]interface{}{
					"type":  "CONTACT_TYPE_GITHUB",
					"value": "github.com/user2",
				},
				"visibility": "VISIBILITY_LEVEL_PUBLIC",
			},
		},
		"contacts_visibility": "VISIBILITY_LEVEL_PUBLIC",
	}
	tc.DoAuthenticatedRequest("PUT", "/v1/users/me/contacts", user2.AccessToken, updateContacts)

	resp, body := tc.DoAuthenticatedRequest("GET", "/v1/users/"+user2.UserID+"?include_skills=true&include_contacts=true", user1.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get user with details: %s", string(body))

	data := tc.ParseJSON(body)
	skills, ok := data["skills"].([]interface{})
	assert.True(t, ok && len(skills) > 0, "Skills should be present")

	contacts, ok := data["contacts"].([]interface{})
	assert.True(t, ok && len(contacts) > 0, "Contacts should be present")
}

func TestGetUserPrivateSkills(t *testing.T) {
	tc := NewTestContext(t)
	user1 := tc.RegisterUser()
	user2 := tc.RegisterUser()

	updateSkills := map[string]interface{}{
		"catalog_skill_ids": []string{},
		"user_skills":       []string{"SecretSkill"},
		"skills_visibility": "VISIBILITY_LEVEL_PRIVATE",
	}
	tc.DoAuthenticatedRequest("PUT", "/v1/users/me/skills", user2.AccessToken, updateSkills)

	resp, body := tc.DoAuthenticatedRequest("GET", "/v1/users/"+user2.UserID+"?include_skills=true", user1.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := tc.ParseJSON(body)
	skills, ok := data["skills"].([]interface{})
	assert.True(t, !ok || len(skills) == 0, "Private skills should not be visible")
}

func TestBatchGetUsers(t *testing.T) {
	tc := NewTestContext(t)
	user1 := tc.RegisterUser()
	user2 := tc.RegisterUser()
	user3 := tc.RegisterUser()

	reqBody := map[string]interface{}{
		"user_ids":       []string{user2.UserID, user3.UserID},
		"include_skills": false,
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/users/batchGet", user1.AccessToken, reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to batch get users: %s", string(body))

	data := tc.ParseJSON(body)
	users, ok := data["users"].([]interface{})
	require.True(t, ok, "Users array should be present")
	assert.GreaterOrEqual(t, len(users), 2, "Should return at least 2 users")
}

func TestListUsersWithSearch(t *testing.T) {
	tc := NewTestContext(t)
	creds := tc.RegisterUser()

	updateBody := map[string]interface{}{
		"first_name": "SearchableFirstName",
		"last_name":  "SearchableLastName",
		"timezone":   "UTC",
	}
	resp, body := tc.DoAuthenticatedRequest("PUT", "/v1/users/me", creds.AccessToken, updateBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to update profile: %s", string(body))

	reqBody := map[string]interface{}{
		"query": map[string]interface{}{
			"q": "SearchableFirst",
			"page": map[string]interface{}{
				"page_size": 10,
			},
		},
	}

	resp, body = tc.DoAuthenticatedRequest("POST", "/v1/users/list", creds.AccessToken, reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to search users: %s", string(body))

	data := tc.ParseJSON(body)
	users, ok := data["users"].([]interface{})
	assert.True(t, ok, "Users array should be present")

	found := false
	for _, u := range users {
		user := u.(map[string]interface{})
		userObj := user["user"].(map[string]interface{})
		if userObj["userId"] == creds.UserID {
			found = true
			break
		}
	}
	assert.True(t, found, "Should find user by search query")
}

func TestListUsersWithFilterUsernamePrefix(t *testing.T) {
	tc := NewTestContext(t)
	creds := tc.RegisterUser()

	reqBody := map[string]interface{}{
		"query": map[string]interface{}{
			"filter_groups": []map[string]interface{}{
				{
					"filters": []map[string]interface{}{
						{
							"field":        "username",
							"operation":    "FILTER_OPERATION_PREFIX",
							"string_value": "user_",
						},
					},
				},
			},
			"page": map[string]interface{}{
				"page_size": 20,
			},
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/users/list", creds.AccessToken, reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to filter users: %s", string(body))

	data := tc.ParseJSON(body)
	users, ok := data["users"].([]interface{})
	assert.True(t, ok && len(users) > 0, "Should find users with username prefix")
}

func TestListUsersWithFilterSkills(t *testing.T) {
	tc := NewTestContext(t)
	user1 := tc.RegisterUser()
	user2 := tc.RegisterUser()

	updateSkills := map[string]interface{}{
		"catalog_skill_ids": []string{},
		"user_skills":       []string{"Go"},
		"skills_visibility": "VISIBILITY_LEVEL_PUBLIC",
	}
	resp, body := tc.DoAuthenticatedRequest("PUT", "/v1/users/me/skills", user2.AccessToken, updateSkills)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to update skills: %s", string(body))

	// Wait longer for search index to update (eventual consistency)
	time.Sleep(2 * time.Second)

	reqBody := map[string]interface{}{
		"query": map[string]interface{}{
			"filter_groups": []map[string]interface{}{
				{
					"filters": []map[string]interface{}{
						{
							"field":     "user_id",
							"operation": "FILTER_OPERATION_IN",
							"string_list": map[string]interface{}{
								"values": []string{user2.UserID},
							},
						},
						{
							"field":        "skills",
							"operation":    "FILTER_OPERATION_CONTAINS",
							"string_value": "go",
						},
					},
				},
			},
			"page": map[string]interface{}{
				"page_size": 10,
			},
		},
		"include_skills": true,
	}

	// Retry search with exponential backoff (eventual consistency)
	found := false
	maxAttempts := 10
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(200*attempt) * time.Millisecond)
		}

		resp, body = tc.DoAuthenticatedRequest("POST", "/v1/users/list", user1.AccessToken, reqBody)
		require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to filter by skills: %s", string(body))

		data := tc.ParseJSON(body)
		users, ok := data["users"].([]interface{})
		assert.True(t, ok, "Users array should be present")

		if attempt == 0 || attempt == maxAttempts-1 {
			t.Logf("Attempt %d: Found %d users in response", attempt+1, len(users))
			if len(users) > 0 {
				t.Logf("First user: %+v", users[0])
			}
		}

		for _, u := range users {
			user := u.(map[string]interface{})
			userObj := user["user"].(map[string]interface{})
			if userObj["userId"] == user2.UserID {
				found = true
				break
			}
		}

		if found {
			t.Logf("Found user with Go skill after %d attempts", attempt+1)
			break
		}
	}

	if !found {
		t.Logf("User2 ID we're looking for: %s", user2.UserID)
	}

	assert.True(t, found, "Should find user with Go skill")
}

func TestListUsersWithFilterUserIDIn(t *testing.T) {
	tc := NewTestContext(t)
	user1 := tc.RegisterUser()
	user2 := tc.RegisterUser()
	user3 := tc.RegisterUser()

	reqBody := map[string]interface{}{
		"query": map[string]interface{}{
			"filter_groups": []map[string]interface{}{
				{
					"filters": []map[string]interface{}{
						{
							"field":     "user_id",
							"operation": "FILTER_OPERATION_IN",
							"string_list": map[string]interface{}{
								"values": []string{user2.UserID, user3.UserID},
							},
						},
					},
				},
			},
			"page": map[string]interface{}{
				"page_size": 50,
			},
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/users/list", user1.AccessToken, reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to filter by user IDs: %s", string(body))

	data := tc.ParseJSON(body)
	users, ok := data["users"].([]interface{})
	require.True(t, ok, "Users array should be present")
	assert.Equal(t, 2, len(users), "Should return exactly 2 users")
}

func TestGetUserNotFound(t *testing.T) {
	tc := NewTestContext(t)
	creds := tc.RegisterUser()

	resp, _ := tc.DoAuthenticatedRequest("GET", "/v1/users/00000000-0000-0000-0000-000000000001", creds.AccessToken, nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Should return 404 for non-existent user")
}
