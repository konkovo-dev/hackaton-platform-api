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
	BaseURL             string
	HTTPClient          *http.Client
	T                   *testing.T
	DB                  *pgxpool.Pool // Main DB connection (hackathon or hackathon_hackaton)
	MentorsDB           *pgxpool.Pool // Mentors DB connection
	SubmissionDB        *pgxpool.Pool // Submission DB connection
	TeamDB              *pgxpool.Pool // Team DB connection
	ParticipationDB     *pgxpool.Pool // Participation DB connection
	MatchmakingDB       *pgxpool.Pool // Matchmaking DB connection
	JudgingDB           *pgxpool.Pool // Judging DB connection
	HackathonDBName     string
	ParticipationDBName string
	MatchmakingDBName   string
	SubmissionDBName    string
	MentorsDBName       string
	TeamDBName          string
	JudgingDBName       string
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

	// Determine table naming based on database structure
	// Local: single DB "hackathon" with schemas (hackathon.hackathons, mentors.tickets, etc.)
	// Prod: separate DBs (hackathon_hackaton, hackathon_mentors, etc.) with schemas inside each DB
	hackathonTable := "hackathon.hackathons"
	participationTable := "participation_and_roles.staff_roles"
	matchmakingPrefix := "matchmaking"
	submissionPrefix := "submission"
	mentorsPrefix := "mentors"
	teamPrefix := "team"

	var mentorsDB, submissionDB, teamDB, participationDB, matchmakingDB, judgingDB *pgxpool.Pool

	// Check if we're on prod (separate databases)
	if contains(dbDSN, "hackathon_hackaton") {
		// Prod: connect to separate databases
		mentorsDSN := replaceDatabaseName(dbDSN, "hackathon_mentors")
		mentorsDB, err = pgxpool.New(context.Background(), mentorsDSN)
		if err != nil {
			t.Fatalf("Failed to connect to mentors database: %v", err)
		}

		submissionDSN := replaceDatabaseName(dbDSN, "hackathon_submission")
		submissionDB, err = pgxpool.New(context.Background(), submissionDSN)
		if err != nil {
			t.Fatalf("Failed to connect to submission database: %v", err)
		}

		teamDSN := replaceDatabaseName(dbDSN, "hackathon_team")
		teamDB, err = pgxpool.New(context.Background(), teamDSN)
		if err != nil {
			t.Fatalf("Failed to connect to team database: %v", err)
		}

		participationDSN := replaceDatabaseName(dbDSN, "hackathon_participation")
		participationDB, err = pgxpool.New(context.Background(), participationDSN)
		if err != nil {
			t.Fatalf("Failed to connect to participation database: %v", err)
		}

		matchmakingDSN := replaceDatabaseName(dbDSN, "hackathon_matchmaking")
		matchmakingDB, err = pgxpool.New(context.Background(), matchmakingDSN)
		if err != nil {
			t.Fatalf("Failed to connect to matchmaking database: %v", err)
		}

		judgingDSN := replaceDatabaseName(dbDSN, "hackathon_judging")
		judgingDB, err = pgxpool.New(context.Background(), judgingDSN)
		if err != nil {
			t.Fatalf("Failed to connect to judging database: %v", err)
		}
	} else {
		// Local: use same DB connection for all services
		mentorsDB = db
		submissionDB = db
		teamDB = db
		participationDB = db
		matchmakingDB = db
		judgingDB = db
	}

	return &TestContext{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		T:                   t,
		DB:                  db,
		MentorsDB:           mentorsDB,
		SubmissionDB:        submissionDB,
		TeamDB:              teamDB,
		ParticipationDB:     participationDB,
		MatchmakingDB:       matchmakingDB,
		JudgingDB:           judgingDB,
		HackathonDBName:     hackathonTable,
		ParticipationDBName: participationTable,
		MatchmakingDBName:   matchmakingPrefix,
		SubmissionDBName:    submissionPrefix,
		MentorsDBName:       mentorsPrefix,
		TeamDBName:          teamPrefix,
		JudgingDBName:       "judging",
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

func replaceDatabaseName(dsn, newDBName string) string {
	// Replace database name in DSN
	// Example: postgres://user:pass@host:5432/hackathon_hackaton?params
	//       -> postgres://user:pass@host:5432/hackathon_mentors?params
	parts := []rune(dsn)
	lastSlash := -1
	questionMark := len(parts)

	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == '?' && questionMark == len(parts) {
			questionMark = i
		}
		if parts[i] == '/' {
			lastSlash = i
			break
		}
	}

	if lastSlash == -1 {
		return dsn
	}

	return string(parts[:lastSlash+1]) + newDBName + string(parts[questionMark:])
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

func (tc *TestContext) AssignRole(hackathonID string, ownerToken string, userID string, role string) {
	_, err := tc.ParticipationDB.Exec(context.Background(),
		`INSERT INTO participation_and_roles.staff_roles (hackathon_id, user_id, role, created_at)
		 VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (hackathon_id, user_id, role) DO NOTHING`,
		hackathonID, userID, role)
	require.NoError(tc.T, err, "Failed to assign role via DB")
	time.Sleep(500 * time.Millisecond)
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

func (tc *TestContext) WaitForMatchmakingUserSync(userID string) {
	maxAttempts := 20
	for i := 0; i < maxAttempts; i++ {
		var count int
		err := tc.MatchmakingDB.QueryRow(context.Background(),
			fmt.Sprintf("SELECT COUNT(*) FROM %s.users WHERE user_id = $1", tc.MatchmakingDBName),
			userID,
		).Scan(&count)
		if err == nil && count > 0 {
			tc.T.Logf("User synced to matchmaking after %d attempts", i+1)
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	tc.T.Fatalf("Timeout waiting for user sync to matchmaking (user_id: %s)", userID)
}

func (tc *TestContext) WaitForMatchmakingParticipationSync(hackathonID, userID string) {
	maxAttempts := 20
	for i := 0; i < maxAttempts; i++ {
		var count int
		err := tc.MatchmakingDB.QueryRow(context.Background(),
			fmt.Sprintf("SELECT COUNT(*) FROM %s.participations WHERE hackathon_id = $1 AND user_id = $2", tc.MatchmakingDBName),
			hackathonID, userID,
		).Scan(&count)
		if err == nil && count > 0 {
			tc.T.Logf("Participation synced to matchmaking after %d attempts", i+1)
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	tc.T.Fatalf("Timeout waiting for participation sync to matchmaking (hackathon_id: %s, user_id: %s)", hackathonID, userID)
}

func (tc *TestContext) WaitForMatchmakingTeamSync(teamID string) {
	maxAttempts := 20
	for i := 0; i < maxAttempts; i++ {
		var count int
		err := tc.MatchmakingDB.QueryRow(context.Background(),
			fmt.Sprintf("SELECT COUNT(*) FROM %s.teams WHERE team_id = $1", tc.MatchmakingDBName),
			teamID,
		).Scan(&count)
		if err == nil && count > 0 {
			tc.T.Logf("Team synced to matchmaking after %d attempts", i+1)
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	tc.T.Fatalf("Timeout waiting for team sync to matchmaking (team_id: %s)", teamID)
}

func (tc *TestContext) WaitForMatchmakingVacancySync(vacancyID string) {
	maxAttempts := 20
	for i := 0; i < maxAttempts; i++ {
		var count int
		err := tc.MatchmakingDB.QueryRow(context.Background(),
			fmt.Sprintf("SELECT COUNT(*) FROM %s.vacancies WHERE vacancy_id = $1", tc.MatchmakingDBName),
			vacancyID,
		).Scan(&count)
		if err == nil && count > 0 {
			tc.T.Logf("Vacancy synced to matchmaking after %d attempts", i+1)
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	tc.T.Fatalf("Timeout waiting for vacancy sync to matchmaking (vacancy_id: %s)", vacancyID)
}

func (tc *TestContext) Close() {
	if tc.DB != nil {
		tc.DB.Close()
	}
	if tc.MentorsDB != nil && tc.MentorsDB != tc.DB {
		tc.MentorsDB.Close()
	}
	if tc.SubmissionDB != nil && tc.SubmissionDB != tc.DB {
		tc.SubmissionDB.Close()
	}
	if tc.TeamDB != nil && tc.TeamDB != tc.DB {
		tc.TeamDB.Close()
	}
	if tc.ParticipationDB != nil && tc.ParticipationDB != tc.DB {
		tc.ParticipationDB.Close()
	}
	if tc.MatchmakingDB != nil && tc.MatchmakingDB != tc.DB {
		tc.MatchmakingDB.Close()
	}
	if tc.JudgingDB != nil && tc.JudgingDB != tc.DB {
		tc.JudgingDB.Close()
	}
}
