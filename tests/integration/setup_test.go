package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type TestContext struct {
	BaseURL              string
	HTTPClient           *http.Client
	T                    *testing.T
	DB                   *pgxpool.Pool
	HackathonDBName      string
	ParticipationDBName  string
}

type UserCredentials struct {
	Email        string
	Password     string
	AccessToken  string
	RefreshToken string
	UserID       string
}

func NewTestContext(t *testing.T) *TestContext {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		dbDSN = "postgres://hackathon:hackathon_dev_password@localhost:5432/hackathon?sslmode=disable"
	}

	db, err := pgxpool.New(context.Background(), dbDSN)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	hackathonTable := "hackathon.hackathons"
	participationTable := "participation_and_roles.staff_roles"
	if dbDSN != "" && (dbDSN == "postgres://hackathon:hackathon_dev_password@localhost:5432/hackathon?sslmode=disable" ||
		!contains(dbDSN, "hackathon_hackaton")) {
		hackathonTable = "hackathon.hackathons"
		participationTable = "participation_and_roles.staff_roles"
	} else {
		hackathonTable = "hackathons"
		participationTable = "staff_roles"
	}

	return &TestContext{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		T:                   t,
		DB:                  db,
		HackathonDBName:     hackathonTable,
		ParticipationDBName: participationTable,
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (tc *TestContext) GenerateUniqueEmail() string {
	return fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8])
}

func (tc *TestContext) DoRequest(method, path string, body interface{}, headers map[string]string) (*http.Response, []byte) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		require.NoError(tc.T, err, "Failed to marshal request body")
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, tc.BaseURL+path, reqBody)
	require.NoError(tc.T, err, "Failed to create request")

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := tc.HTTPClient.Do(req)
	require.NoError(tc.T, err, "Failed to perform request")

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(tc.T, err, "Failed to read response body")
	resp.Body.Close()

	return resp, respBody
}

func (tc *TestContext) DoAuthenticatedRequest(method, path string, token string, body interface{}) (*http.Response, []byte) {
	tc.T.Logf("→ %s %s", method, path)
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}
	resp, respBody := tc.DoRequest(method, path, body, headers)
	tc.T.Logf("← %d %s", resp.StatusCode, path)
	return resp, respBody
}

func (tc *TestContext) RegisterUser() *UserCredentials {
	email := tc.GenerateUniqueEmail()
	username := fmt.Sprintf("user_%s", uuid.New().String()[:8])
	password := "SecurePassword123!"

	reqBody := map[string]interface{}{
		"username":   username,
		"email":      email,
		"password":   password,
		"first_name": "Test",
		"last_name":  "User",
		"timezone":   "UTC",
	}

	resp, body := tc.DoRequest("POST", "/v1/auth/register", reqBody, nil)
	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Registration failed: %s", string(body))

	var registerResp struct {
		AccessToken      string `json:"accessToken"`
		RefreshToken     string `json:"refreshToken"`
		AccessExpiresAt  string `json:"accessExpiresAt"`
		RefreshExpiresAt string `json:"refreshExpiresAt"`
	}
	err := json.Unmarshal(body, &registerResp)
	require.NoError(tc.T, err, "Failed to parse registration response")

	introspectBody := map[string]interface{}{
		"access_token": registerResp.AccessToken,
	}
	resp, body = tc.DoRequest("POST", "/v1/auth/introspect", introspectBody, nil)
	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Introspect failed: %s", string(body))

	var introspectResp struct {
		UserID string `json:"userId"`
	}
	json.Unmarshal(body, &introspectResp)

	creds := &UserCredentials{
		Email:        email,
		Password:     password,
		AccessToken:  registerResp.AccessToken,
		RefreshToken: registerResp.RefreshToken,
		UserID:       introspectResp.UserID,
	}

	tc.WaitForUserInIdentityService(creds.AccessToken)

	return creds
}

func (tc *TestContext) Login(creds *UserCredentials) {
	reqBody := map[string]interface{}{
		"email":    creds.Email,
		"password": creds.Password,
	}

	resp, body := tc.DoRequest("POST", "/v1/auth/login", reqBody, nil)
	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Login failed: %s", string(body))

	var loginResp struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}
	err := json.Unmarshal(body, &loginResp)
	require.NoError(tc.T, err, "Failed to parse login response")

	creds.AccessToken = loginResp.AccessToken
	creds.RefreshToken = loginResp.RefreshToken
}

func (tc *TestContext) RegisterAndLogin() *UserCredentials {
	creds := tc.RegisterUser()
	tc.Login(creds)
	return creds
}

func (tc *TestContext) AssertJSONField(body []byte, field string, expected interface{}) {
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	require.NoError(tc.T, err, "Failed to parse JSON response")
	require.Equal(tc.T, expected, data[field], "Field %s mismatch", field)
}

func (tc *TestContext) ParseJSON(body []byte) map[string]interface{} {
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	require.NoError(tc.T, err, "Failed to parse JSON response")
	return data
}

func (tc *TestContext) WaitForUserInIdentityService(token string) {
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		resp, _ := tc.DoAuthenticatedRequest("GET", "/v1/users/me", token, nil)
		if resp.StatusCode == http.StatusOK {
			return // User found in identity service
		}
		time.Sleep(200 * time.Millisecond) // Wait 200ms before retry
	}
	tc.T.Logf("Warning: User not found in identity service after %d attempts", maxAttempts)
}

func (tc *TestContext) WaitForHackathonOwnerRole(hackathonID string, token string) {
	maxAttempts := 20
	for i := 0; i < maxAttempts; i++ {
		resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s", hackathonID), token, nil)
		if resp.StatusCode == http.StatusOK {
			tc.T.Logf("Owner role assigned after %d attempts", i+1)
			time.Sleep(500 * time.Millisecond)

			resp2, _ := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s", hackathonID), token, nil)
			if resp2.StatusCode == http.StatusOK {
				return
			}
			tc.T.Logf("Warning: Second GET failed after successful first GET, retrying...")
		}
		if i == 0 || i == 5 || i == 10 || i == 15 {
			tc.T.Logf("Waiting for owner role (attempt %d/%d), status: %d, body: %s", i+1, maxAttempts, resp.StatusCode, string(body))
		}
		time.Sleep(300 * time.Millisecond)
	}
	tc.T.Fatalf("Owner role not assigned after %d attempts (6 seconds)", maxAttempts)
}

func (tc *TestContext) WaitForParticipationRegistered(hackathonID string, token string) {
	maxAttempts := 15
	for i := 0; i < maxAttempts; i++ {
		resp, _ := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s/participations/me", hackathonID), token, nil)
		if resp.StatusCode == http.StatusOK {
			tc.T.Logf("Participation registered after %d attempts", i+1)
			time.Sleep(500 * time.Millisecond)
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	tc.T.Logf("Warning: Participation not confirmed after %d attempts", maxAttempts)
}
