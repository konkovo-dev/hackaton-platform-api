package integration

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterWithUsernameAndEmail(t *testing.T) {
	tc := NewTestContext(t)

	email := tc.GenerateUniqueEmail()
	username := "user_" + uuid.New().String()[:8]
	password := "SecurePass123"

	reqBody := map[string]interface{}{
		"username":   username,
		"email":      email,
		"password":   password,
		"first_name": "Alice",
		"last_name":  "Smith",
		"timezone":   "Europe/Moscow",
	}

	resp, body := tc.DoRequest("POST", "/v1/auth/register", reqBody, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Registration failed: %s", string(body))

	data := tc.ParseJSON(body)
	assert.NotEmpty(t, data["accessToken"], "Access token should be returned")
	assert.NotEmpty(t, data["refreshToken"], "Refresh token should be returned")
	assert.NotEmpty(t, data["accessExpiresAt"], "Access expiry should be returned")
	assert.NotEmpty(t, data["refreshExpiresAt"], "Refresh expiry should be returned")
}

func TestRegisterWithIdempotencyKey(t *testing.T) {
	tc := NewTestContext(t)

	email := tc.GenerateUniqueEmail()
	username := "user_" + uuid.New().String()[:8]
	idempotencyKey := uuid.New().String()

	reqBody := map[string]interface{}{
		"username":        username,
		"email":           email,
		"password":        "SecurePass123",
		"first_name":      "Bob",
		"last_name":       "Test",
		"timezone":        "UTC",
		"idempotency_key": map[string]string{"key": idempotencyKey},
	}

	resp1, body1 := tc.DoRequest("POST", "/v1/auth/register", reqBody, nil)
	require.Equal(t, http.StatusOK, resp1.StatusCode, "First registration failed: %s", string(body1))

	data1 := tc.ParseJSON(body1)
	token1 := data1["accessToken"].(string)

	resp2, body2 := tc.DoRequest("POST", "/v1/auth/register", reqBody, nil)
	require.Equal(t, http.StatusOK, resp2.StatusCode, "Second registration failed: %s", string(body2))

	data2 := tc.ParseJSON(body2)
	token2 := data2["accessToken"].(string)

	assert.Equal(t, token1, token2, "Idempotent requests should return same token")
}

func TestRegisterDuplicateEmail(t *testing.T) {
	tc := NewTestContext(t)

	email := tc.GenerateUniqueEmail()
	reqBody := map[string]interface{}{
		"username":   "user1_" + uuid.New().String()[:8],
		"email":      email,
		"password":   "SecurePass123",
		"first_name": "Test",
		"last_name":  "User",
		"timezone":   "UTC",
	}

	resp, _ := tc.DoRequest("POST", "/v1/auth/register", reqBody, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	reqBody["username"] = "user2_" + uuid.New().String()[:8]
	resp, body := tc.DoRequest("POST", "/v1/auth/register", reqBody, nil)
	assert.Equal(t, http.StatusConflict, resp.StatusCode, "Should reject duplicate email: %s", string(body))
}

func TestLoginWithEmail(t *testing.T) {
	tc := NewTestContext(t)
	creds := tc.RegisterUser()

	reqBody := map[string]interface{}{
		"email":    creds.Email,
		"password": creds.Password,
	}

	resp, body := tc.DoRequest("POST", "/v1/auth/login", reqBody, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Login failed: %s", string(body))

	data := tc.ParseJSON(body)
	assert.NotEmpty(t, data["accessToken"])
	assert.NotEmpty(t, data["refreshToken"])
}

func TestLoginWithUsername(t *testing.T) {
	tc := NewTestContext(t)

	email := tc.GenerateUniqueEmail()
	username := "user_" + uuid.New().String()[:8]
	password := "SecurePass123"

	reqBody := map[string]interface{}{
		"username":   username,
		"email":      email,
		"password":   password,
		"first_name": "Test",
		"last_name":  "User",
		"timezone":   "UTC",
	}
	tc.DoRequest("POST", "/v1/auth/register", reqBody, nil)

	loginBody := map[string]interface{}{
		"username": username,
		"password": password,
	}

	resp, body := tc.DoRequest("POST", "/v1/auth/login", loginBody, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Login with username failed: %s", string(body))

	data := tc.ParseJSON(body)
	assert.NotEmpty(t, data["accessToken"])
	assert.NotEmpty(t, data["refreshToken"])
}

func TestLoginInvalidCredentials(t *testing.T) {
	tc := NewTestContext(t)

	reqBody := map[string]interface{}{
		"email":    "nonexistent@example.com",
		"password": "WrongPassword123!",
	}

	resp, body := tc.DoRequest("POST", "/v1/auth/login", reqBody, nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Should reject invalid credentials: %s", string(body))
}

func TestIntrospectValidToken(t *testing.T) {
	tc := NewTestContext(t)
	creds := tc.RegisterUser()

	reqBody := map[string]interface{}{
		"access_token": creds.AccessToken,
	}

	resp, body := tc.DoRequest("POST", "/v1/auth/introspect", reqBody, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Introspect failed: %s", string(body))

	data := tc.ParseJSON(body)
	assert.Equal(t, true, data["active"], "Token should be active")
	assert.Equal(t, creds.UserID, data["userId"], "User ID should match")
	assert.NotEmpty(t, data["expiresAt"], "Expiry should be present")
}

func TestIntrospectInvalidToken(t *testing.T) {
	tc := NewTestContext(t)

	reqBody := map[string]interface{}{
		"access_token": "invalid-token-12345",
	}

	resp, body := tc.DoRequest("POST", "/v1/auth/introspect", reqBody, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Introspect should return 200: %s", string(body))

	data := tc.ParseJSON(body)
	if active, ok := data["active"]; ok {
		assert.Equal(t, false, active, "Invalid token should be inactive")
	}
}

func TestRefreshToken(t *testing.T) {
	tc := NewTestContext(t)

	email := tc.GenerateUniqueEmail()
	username := "user_" + uuid.New().String()[:8]
	password := "SecurePass123"

	reqBody := map[string]interface{}{
		"username":   username,
		"email":      email,
		"password":   password,
		"first_name": "Test",
		"last_name":  "User",
		"timezone":   "UTC",
	}

	resp, body := tc.DoRequest("POST", "/v1/auth/register", reqBody, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	registerData := tc.ParseJSON(body)
	refreshToken := registerData["refreshToken"].(string)

	refreshBody := map[string]interface{}{
		"refresh_token": refreshToken,
	}

	resp, body = tc.DoRequest("POST", "/v1/auth/refresh", refreshBody, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Refresh failed: %s", string(body))

	data := tc.ParseJSON(body)
	assert.NotEmpty(t, data["accessToken"], "New access token should be returned")
	assert.NotEmpty(t, data["refreshToken"], "New refresh token should be returned")
}

func TestLogout(t *testing.T) {
	tc := NewTestContext(t)

	email := tc.GenerateUniqueEmail()
	username := "user_" + uuid.New().String()[:8]
	password := "SecurePass123"

	reqBody := map[string]interface{}{
		"username":   username,
		"email":      email,
		"password":   password,
		"first_name": "Test",
		"last_name":  "User",
		"timezone":   "UTC",
	}

	resp, body := tc.DoRequest("POST", "/v1/auth/register", reqBody, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	registerData := tc.ParseJSON(body)
	refreshToken := registerData["refreshToken"].(string)

	logoutBody := map[string]interface{}{
		"refresh_token": refreshToken,
	}

	resp, body = tc.DoRequest("POST", "/v1/auth/logout", logoutBody, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Logout failed: %s", string(body))

	refreshBody := map[string]interface{}{
		"refresh_token": refreshToken,
	}

	resp, _ = tc.DoRequest("POST", "/v1/auth/refresh", refreshBody, nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Refresh should fail after logout")
}

func TestUnauthorizedAccess(t *testing.T) {
	tc := NewTestContext(t)

	resp, _ := tc.DoRequest("GET", "/v1/users/me", nil, nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Should require authentication")

	resp, _ = tc.DoAuthenticatedRequest("GET", "/v1/users/me", "invalid-token-xyz", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Should reject invalid token")
}
