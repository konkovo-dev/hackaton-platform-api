package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func uploadAvatarToMinIODirectly(storageKey string, content []byte, contentType string) error {
	endpoint := os.Getenv("S3_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:9000"
	}

	accessKey := os.Getenv("S3_ACCESS_KEY_ID")
	if accessKey == "" {
		accessKey = "minioadmin"
	}

	secretKey := os.Getenv("S3_SECRET_ACCESS_KEY")
	if secretKey == "" {
		secretKey = "minioadmin"
	}

	bucketName := os.Getenv("S3_AVATARS_BUCKET")
	if bucketName == "" {
		bucketName = "avatars"
	}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx := context.Background()
	_, err = minioClient.PutObject(ctx, bucketName, storageKey, bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload to minio: %w", err)
	}

	return nil
}

func initAvatarsBucket() error {
	endpoint := os.Getenv("S3_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:9000"
	}

	accessKey := os.Getenv("S3_ACCESS_KEY_ID")
	if accessKey == "" {
		accessKey = "minioadmin"
	}

	secretKey := os.Getenv("S3_SECRET_ACCESS_KEY")
	if secretKey == "" {
		secretKey = "minioadmin"
	}

	bucketName := os.Getenv("S3_AVATARS_BUCKET")
	if bucketName == "" {
		bucketName = "avatars"
	}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx := context.Background()

	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		fmt.Printf("Created MinIO bucket: %s\n", bucketName)
	} else {
		fmt.Printf("MinIO bucket already exists: %s\n", bucketName)
	}

	return nil
}

func TestAvatarUpload_FullFlow(t *testing.T) {
	if err := initAvatarsBucket(); err != nil {
		t.Fatalf("Failed to initialize avatars bucket: %v", err)
	}

	tc := NewTestContext(t)
	user := tc.RegisterUser()

	// Step 1: Request avatar upload URL
	idempotencyKey := fmt.Sprintf("avatar-upload-test-%d", time.Now().UnixNano())
	reqBody := map[string]interface{}{
		"filename":     "avatar.png",
		"sizeBytes":    "1024",
		"contentType":  "image/png",
		"idempotencyKey": map[string]interface{}{
			"key": idempotencyKey,
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/users/me/avatar/upload", user.AccessToken, reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to request avatar upload: %s", string(body))

	uploadData := tc.ParseJSON(body)
	uploadID, ok := uploadData["uploadId"].(string)
	require.True(t, ok && uploadID != "", "Should return uploadId")

	uploadURL, ok := uploadData["uploadUrl"].(string)
	require.True(t, ok && uploadURL != "", "Should return uploadUrl")

	expiresAt, ok := uploadData["expiresAt"].(string)
	require.True(t, ok && expiresAt != "", "Should return expiresAt")

	// Verify presigned URL uses public endpoint
	assert.Contains(t, uploadURL, "api.hackplatform.ru:9000", "Upload URL should use public endpoint")

	// Step 2: Upload file directly to MinIO (presigned URLs don't work in test environment)
	fileContent := make([]byte, 1024)
	_, err := rand.Read(fileContent)
	require.NoError(t, err, "Failed to generate random file content")

	// Get storage_key from DB
	var storageKey string
	err = tc.DB.QueryRow(context.Background(), `
		SELECT storage_key FROM identity.avatar_uploads WHERE upload_id = $1
	`, uploadID).Scan(&storageKey)
	require.NoError(t, err, "Failed to get storage key")
	t.Logf("Storage key: %s", storageKey)

	// Upload directly to MinIO
	err = uploadAvatarToMinIODirectly(storageKey, fileContent, "image/png")
	require.NoError(t, err, "Failed to upload to MinIO")

	// Wait for S3 consistency
	time.Sleep(100 * time.Millisecond)

	// Verify file was uploaded
	endpoint := os.Getenv("S3_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:9000"
	}
	accessKey := os.Getenv("S3_ACCESS_KEY_ID")
	if accessKey == "" {
		accessKey = "minioadmin"
	}
	secretKey := os.Getenv("S3_SECRET_ACCESS_KEY")
	if secretKey == "" {
		secretKey = "minioadmin"
	}
	bucketName := os.Getenv("S3_AVATARS_BUCKET")
	if bucketName == "" {
		bucketName = "avatars"
	}
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	require.NoError(t, err)
	objInfo, err := minioClient.StatObject(context.Background(), bucketName, storageKey, minio.StatObjectOptions{})
	require.NoError(t, err, "File should exist in MinIO")
	t.Logf("File uploaded successfully: size=%d", objInfo.Size)

	// Step 3: Complete avatar upload
	completeIdempotencyKey := fmt.Sprintf("avatar-complete-test-%d", time.Now().UnixNano())
	completeBody := map[string]interface{}{
		"uploadId": uploadID,
		"idempotencyKey": map[string]interface{}{
			"key": completeIdempotencyKey,
		},
	}

	resp, body = tc.DoAuthenticatedRequest("POST", "/v1/users/me/avatar/complete", user.AccessToken, completeBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to complete avatar upload: %s", string(body))

	completeData := tc.ParseJSON(body)
	avatarURL, ok := completeData["avatarUrl"].(string)
	require.True(t, ok && avatarURL != "", "Should return avatarUrl")

	// Step 4: Verify avatar URL is set in user profile
	resp, body = tc.DoAuthenticatedRequest("GET", "/v1/users/me", user.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get user profile: %s", string(body))

	userData := tc.ParseJSON(body)
	userObj, ok := userData["user"].(map[string]interface{})
	require.True(t, ok, "Should have user object")

	profileAvatarURL, ok := userObj["avatarUrl"].(string)
	require.True(t, ok, "Should have avatarUrl in profile")
	assert.Equal(t, avatarURL, profileAvatarURL, "Avatar URL should match")
}

func TestAvatarUpload_WithoutAuth_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)

	reqBody := map[string]interface{}{
		"filename":    "avatar.png",
		"sizeBytes":   "1024",
		"contentType": "image/png",
	}

	resp, body := tc.DoRequest("POST", "/v1/users/me/avatar/upload", reqBody, nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
		"Should reject unauthenticated request: %s", string(body))
}

func TestAvatarUpload_InvalidContentType_ShouldFail(t *testing.T) {
	if err := initAvatarsBucket(); err != nil {
		t.Fatalf("Failed to initialize avatars bucket: %v", err)
	}

	tc := NewTestContext(t)
	user := tc.RegisterUser()

	reqBody := map[string]interface{}{
		"filename":    "document.pdf",
		"sizeBytes":   "1024",
		"contentType": "application/pdf",
		"idempotencyKey": map[string]interface{}{
			"key": "avatar-invalid-type-test",
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/users/me/avatar/upload", user.AccessToken, reqBody)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
		"Should reject non-image content type: %s", string(body))
}

func TestAvatarUpload_TooLarge_ShouldFail(t *testing.T) {
	if err := initAvatarsBucket(); err != nil {
		t.Fatalf("Failed to initialize avatars bucket: %v", err)
	}

	tc := NewTestContext(t)
	user := tc.RegisterUser()

	// 11MB file (assuming 10MB limit)
	reqBody := map[string]interface{}{
		"filename":    "large-avatar.png",
		"sizeBytes":   "11534336",
		"contentType": "image/png",
		"idempotencyKey": map[string]interface{}{
			"key": "avatar-too-large-test",
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/users/me/avatar/upload", user.AccessToken, reqBody)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
		"Should reject file that is too large: %s", string(body))
}

func TestCompleteAvatarUpload_WithoutCreating_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	user := tc.RegisterUser()

	completeBody := map[string]interface{}{
		"uploadId": "non-existent-upload-id",
		"idempotencyKey": map[string]interface{}{
			"key": "avatar-complete-nonexistent-test",
		},
	}

	resp, body := tc.DoAuthenticatedRequest("POST", "/v1/users/me/avatar/complete", user.AccessToken, completeBody)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode,
		"Should fail to complete non-existent upload: %s", string(body))
}

func TestAvatarUpload_Idempotency(t *testing.T) {
	if err := initAvatarsBucket(); err != nil {
		t.Fatalf("Failed to initialize avatars bucket: %v", err)
	}

	tc := NewTestContext(t)
	user := tc.RegisterUser()

	idempotencyKey := fmt.Sprintf("avatar-idempotency-test-%d", time.Now().UnixNano())

	reqBody := map[string]interface{}{
		"filename":    "avatar.png",
		"sizeBytes":   "1024",
		"contentType": "image/png",
		"idempotencyKey": map[string]interface{}{
			"key": idempotencyKey,
		},
	}

	// First request
	resp1, body1 := tc.DoAuthenticatedRequest("POST", "/v1/users/me/avatar/upload", user.AccessToken, reqBody)
	require.Equal(t, http.StatusOK, resp1.StatusCode, "First request should succeed: %s", string(body1))

	uploadData1 := tc.ParseJSON(body1)
	uploadID1 := uploadData1["uploadId"].(string)

	// Second request with same idempotency key
	resp2, body2 := tc.DoAuthenticatedRequest("POST", "/v1/users/me/avatar/upload", user.AccessToken, reqBody)
	require.Equal(t, http.StatusOK, resp2.StatusCode, "Second request should succeed: %s", string(body2))

	uploadData2 := tc.ParseJSON(body2)
	uploadID2 := uploadData2["uploadId"].(string)

	// Should return the same upload ID
	assert.Equal(t, uploadID1, uploadID2, "Idempotent requests should return same uploadId")
}


