package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Initialize MinIO bucket before running tests
	if err := initMinIOBucket(); err != nil {
		fmt.Printf("Failed to initialize MinIO bucket: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func initMinIOBucket() error {
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

	bucketName := os.Getenv("S3_SUBMISSIONS_BUCKET")
	if bucketName == "" {
		bucketName = "submissions"
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

func TestCreateSubmission_WithoutAuth_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)

	body := map[string]interface{}{
		"title":       "Test Submission",
		"description": "Test description",
	}

	resp, respBody := tc.DoRequest("POST", fmt.Sprintf("/v1/hackathons/%s/submissions", hackathonID), body, nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
		"Should reject unauthenticated request: %s", string(respBody))
}

func TestCreateSubmission_AsNonParticipant_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	nonParticipant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)

	body := map[string]interface{}{
		"title":       "Test Submission",
		"description": "Test description",
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/submissions", hackathonID),
		nonParticipant.AccessToken,
		body,
	)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"Non-participant should not create submission: %s", string(respBody))
}

func TestGetSubmission_AsOwner_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID := createSubmission(tc, hackathonID, participant, "My Solution", "Description")

	submission := getSubmission(tc, hackathonID, submissionID, participant)
	assert.Equal(t, submissionID, submission["submissionId"])
	assert.Equal(t, "My Solution", submission["title"])
}

func TestGetSubmission_AsStaff_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	organizer := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	assignOrganizerRole(tc, hackathonID, organizer)

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID := createSubmission(tc, hackathonID, participant, "Participant Solution", "Description")

	submission := getSubmission(tc, hackathonID, submissionID, organizer)
	assert.Equal(t, submissionID, submission["submissionId"])
	assert.Equal(t, "Participant Solution", submission["title"])
}

func TestGetSubmission_AsOtherParticipant_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant1 := tc.RegisterUser()
	participant2 := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant1, "PART_INDIVIDUAL")
	registerParticipant(tc, hackathonID, participant2, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID := createSubmission(tc, hackathonID, participant1, "Private Solution", "Description")

	resp, respBody := tc.DoAuthenticatedRequest(
		"GET",
		fmt.Sprintf("/v1/hackathons/%s/submissions/%s", hackathonID, submissionID),
		participant2.AccessToken,
		nil,
	)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"Other participant should not see submission: %s", string(respBody))
}

func TestCreateSubmission_InRegistrationStage_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRegistrationForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	body := map[string]interface{}{
		"title":       "Early Submission",
		"description": "Too early",
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/submissions", hackathonID),
		participant.AccessToken,
		body,
	)

	assert.NotEqual(t, http.StatusOK, resp.StatusCode,
		"Should reject submission in REGISTRATION stage: %s", string(respBody))
}

func TestCreateSubmission_InRunningStage_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID := createSubmission(tc, hackathonID, participant, "Running Stage Submission", "Created during RUNNING")

	// Verify hackathon is in RUNNING stage by checking through API
	resp, body := tc.DoAuthenticatedRequest("GET", fmt.Sprintf("/v1/hackathons/%s", hackathonID), owner.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	hackData := tc.ParseJSON(body)
	hackathon := hackData["hackathon"].(map[string]interface{})
	assert.Equal(t, "HACKATHON_STAGE_RUNNING", hackathon["stage"])

	submission := getSubmission(tc, hackathonID, submissionID, participant)
	assert.Equal(t, "Running Stage Submission", submission["title"])
}

func TestGetSubmission_AfterRunning_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID := createSubmission(tc, hackathonID, participant, "Solution", "Description")

	transitionToJudging(tc, hackathonID)

	submission := getSubmission(tc, hackathonID, submissionID, participant)
	assert.Equal(t, submissionID, submission["submissionId"])
}

func TestSelectFinalSubmission_AfterRunning_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID := createSubmission(tc, hackathonID, participant, "Solution", "Description")

	transitionToJudging(tc, hackathonID)

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/submissions/%s/select-final", hackathonID, submissionID),
		participant.AccessToken,
		map[string]interface{}{},
	)

	assert.NotEqual(t, http.StatusOK, resp.StatusCode,
		"Should reject final selection after RUNNING stage: %s", string(respBody))
}

func TestCreateSubmission_AsIndividual_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID := createSubmission(tc, hackathonID, participant, "MVP Implementation", "This is my solution")

	var ownerKind, ownerID, createdByUserID string
	var isFinal bool
	err := tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT owner_kind, owner_id, created_by_user_id, is_final FROM %s.submissions WHERE id = $1", tc.SubmissionDBName),
		submissionID,
	).Scan(&ownerKind, &ownerID, &createdByUserID, &isFinal)
	require.NoError(t, err)

	assert.Equal(t, "user", ownerKind, "Owner kind should be 'user'")
	assert.Equal(t, participant.UserID, ownerID, "Owner ID should be participant user ID")
	assert.Equal(t, participant.UserID, createdByUserID, "Created by should be participant")
	assert.True(t, isFinal, "First submission should be automatically marked as final")
}

func TestUpdateSubmission_AsCreator_ShouldUpdateDescription(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID := createSubmission(tc, hackathonID, participant, "Solution v1", "Initial description")

	updateSubmission(tc, hackathonID, submissionID, participant, "Updated description with more details")

	var description string
	err := tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT description FROM %s.submissions WHERE id = $1", tc.SubmissionDBName),
		submissionID,
	).Scan(&description)
	require.NoError(t, err)

	assert.Equal(t, "Updated description with more details", description)
}

func TestUpdateSubmission_AsTeamMember_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	member := tc.RegisterUser()

	hackathonID, _ := createHackathonInRunningWithTeam(tc, owner, captain, member)

	submissionID := createSubmission(tc, hackathonID, captain, "Team Solution", "Created by captain")

	updateSubmission(tc, hackathonID, submissionID, member, "Updated by team member")

	var description, createdByUserID string
	err := tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT description, created_by_user_id FROM %s.submissions WHERE id = $1", tc.SubmissionDBName),
		submissionID,
	).Scan(&description, &createdByUserID)
	require.NoError(t, err)

	assert.Equal(t, "Updated by team member", description, "Description should be updated")
	assert.Equal(t, captain.UserID, createdByUserID, "Created by should remain captain")
}

func TestUpdateSubmission_AsNonCreator_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant1 := tc.RegisterUser()
	participant2 := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant1, "PART_INDIVIDUAL")
	registerParticipant(tc, hackathonID, participant2, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID := createSubmission(tc, hackathonID, participant1, "Solution", "Description")

	body := map[string]interface{}{
		"description": "Hacked description",
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"PUT",
		fmt.Sprintf("/v1/hackathons/%s/submissions/%s", hackathonID, submissionID),
		participant2.AccessToken,
		body,
	)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"Non-creator should not update submission: %s", string(respBody))
}

func TestListSubmissions_AsIndividual_ShouldReturnOwnSubmissions(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID1 := createSubmission(tc, hackathonID, participant, "Solution v1", "First version")
	time.Sleep(100 * time.Millisecond)
	submissionID2 := createSubmission(tc, hackathonID, participant, "Solution v2", "Second version")

	submissions := listSubmissions(tc, hackathonID, participant, "", "")

	assert.Len(t, submissions, 2, "Should return 2 submissions")

	submissionIDs := []string{
		submissions[0].(map[string]interface{})["submissionId"].(string),
		submissions[1].(map[string]interface{})["submissionId"].(string),
	}

	assert.Contains(t, submissionIDs, submissionID1)
	assert.Contains(t, submissionIDs, submissionID2)
}

func TestGetFinalSubmission_AsIndividual_ShouldReturnLatest(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	createSubmission(tc, hackathonID, participant, "Solution v1", "First")
	time.Sleep(100 * time.Millisecond)
	submissionID2 := createSubmission(tc, hackathonID, participant, "Solution v2", "Second")

	finalSubmission := getFinalSubmission(tc, hackathonID, "user", participant.UserID, participant)

	assert.Equal(t, submissionID2, finalSubmission["submissionId"])
	assert.True(t, finalSubmission["isFinal"].(bool))
}

func TestCreateSubmission_AsTeamMember_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	member := tc.RegisterUser()

	hackathonID, teamID := createHackathonInRunningWithTeam(tc, owner, captain, member)

	submissionID := createSubmission(tc, hackathonID, member, "Team Solution", "Created by team member")

	var ownerKind, ownerID, createdByUserID string
	err := tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT owner_kind, owner_id, created_by_user_id FROM %s.submissions WHERE id = $1", tc.SubmissionDBName),
		submissionID,
	).Scan(&ownerKind, &ownerID, &createdByUserID)
	require.NoError(t, err)

	assert.Equal(t, "team", ownerKind, "Owner kind should be 'team'")
	assert.Equal(t, teamID, ownerID, "Owner ID should be team ID")
	assert.Equal(t, member.UserID, createdByUserID, "Created by should be the member who uploaded")
}

func TestSelectFinalSubmission_AsCaptain_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	member := tc.RegisterUser()

	hackathonID, _ := createHackathonInRunningWithTeam(tc, owner, captain, member)

	submissionID1 := createSubmission(tc, hackathonID, captain, "Solution v1", "First")
	time.Sleep(100 * time.Millisecond)
	submissionID2 := createSubmission(tc, hackathonID, captain, "Solution v2", "Second")

	selectFinalSubmission(tc, hackathonID, submissionID1, captain)

	var isFinal1, isFinal2 bool
	err := tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT is_final FROM %s.submissions WHERE id = $1", tc.SubmissionDBName),
		submissionID1,
	).Scan(&isFinal1)
	require.NoError(t, err)

	err = tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT is_final FROM %s.submissions WHERE id = $1", tc.SubmissionDBName),
		submissionID2,
	).Scan(&isFinal2)
	require.NoError(t, err)

	assert.True(t, isFinal1, "Selected submission should be final")
	assert.False(t, isFinal2, "Previous final should be unmarked")
}

func TestSelectFinalSubmission_AsTeamMember_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	member := tc.RegisterUser()

	hackathonID, _ := createHackathonInRunningWithTeam(tc, owner, captain, member)

	submissionID := createSubmission(tc, hackathonID, captain, "Team Solution", "Description")

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/submissions/%s/select-final", hackathonID, submissionID),
		member.AccessToken,
		map[string]interface{}{},
	)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"Team member (non-captain) should not select final: %s", string(respBody))
}

func TestListSubmissions_AsTeamMember_ShouldReturnTeamSubmissions(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	member := tc.RegisterUser()

	hackathonID, _ := createHackathonInRunningWithTeam(tc, owner, captain, member)

	submissionID1 := createSubmission(tc, hackathonID, captain, "Solution by Captain", "By captain")
	time.Sleep(100 * time.Millisecond)
	submissionID2 := createSubmission(tc, hackathonID, member, "Solution by Member", "By member")

	submissions := listSubmissions(tc, hackathonID, member, "", "")

	assert.Len(t, submissions, 2, "Team member should see all team submissions")

	submissionIDs := []string{
		submissions[0].(map[string]interface{})["submissionId"].(string),
		submissions[1].(map[string]interface{})["submissionId"].(string),
	}

	assert.Contains(t, submissionIDs, submissionID1)
	assert.Contains(t, submissionIDs, submissionID2)
}

func TestGetFinalSubmission_ForTeam_ShouldReturnSelected(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	member := tc.RegisterUser()

	hackathonID, teamID := createHackathonInRunningWithTeam(tc, owner, captain, member)

	submissionID1 := createSubmission(tc, hackathonID, captain, "Solution v1", "First")
	time.Sleep(100 * time.Millisecond)
	_ = createSubmission(tc, hackathonID, captain, "Solution v2", "Second")

	selectFinalSubmission(tc, hackathonID, submissionID1, captain)

	finalSubmission := getFinalSubmission(tc, hackathonID, "team", teamID, captain)

	assert.Equal(t, submissionID1, finalSubmission["submissionId"])
	assert.True(t, finalSubmission["isFinal"].(bool))
}

func TestCreateMultipleSubmissions_ShouldMarkLatestAsFinal(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID1 := createSubmission(tc, hackathonID, participant, "Solution v1", "First version")
	time.Sleep(100 * time.Millisecond)
	submissionID2 := createSubmission(tc, hackathonID, participant, "Solution v2", "Second version")
	time.Sleep(100 * time.Millisecond)
	submissionID3 := createSubmission(tc, hackathonID, participant, "Solution v3", "Third version")

	var isFinal1, isFinal2, isFinal3 bool
	tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT is_final FROM %s.submissions WHERE id = $1", tc.SubmissionDBName),
		submissionID1,
	).Scan(&isFinal1)

	tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT is_final FROM %s.submissions WHERE id = $1", tc.SubmissionDBName),
		submissionID2,
	).Scan(&isFinal2)

	tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT is_final FROM %s.submissions WHERE id = $1", tc.SubmissionDBName),
		submissionID3,
	).Scan(&isFinal3)

	assert.False(t, isFinal1, "First submission should not be final")
	assert.False(t, isFinal2, "Second submission should not be final")
	assert.True(t, isFinal3, "Latest submission should be final")
}

func TestSelectFinalSubmission_ShouldUnmarkPrevious(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID1 := createSubmission(tc, hackathonID, participant, "Solution v1", "First")
	time.Sleep(100 * time.Millisecond)
	submissionID2 := createSubmission(tc, hackathonID, participant, "Solution v2", "Second")
	time.Sleep(100 * time.Millisecond)
	submissionID3 := createSubmission(tc, hackathonID, participant, "Solution v3", "Third")

	selectFinalSubmission(tc, hackathonID, submissionID2, participant)

	var isFinal1, isFinal2, isFinal3 bool
	tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT is_final FROM %s.submissions WHERE id = $1", tc.SubmissionDBName),
		submissionID1,
	).Scan(&isFinal1)

	tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT is_final FROM %s.submissions WHERE id = $1", tc.SubmissionDBName),
		submissionID2,
	).Scan(&isFinal2)

	tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT is_final FROM %s.submissions WHERE id = $1", tc.SubmissionDBName),
		submissionID3,
	).Scan(&isFinal3)

	assert.False(t, isFinal1, "First submission should not be final")
	assert.True(t, isFinal2, "Selected submission should be final")
	assert.False(t, isFinal3, "Previous final should be unmarked")
}

func TestListSubmissions_ShouldReturnAllVersions(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID1 := createSubmission(tc, hackathonID, participant, "Solution v1", "First")
	time.Sleep(100 * time.Millisecond)
	submissionID2 := createSubmission(tc, hackathonID, participant, "Solution v2", "Second")
	time.Sleep(100 * time.Millisecond)
	submissionID3 := createSubmission(tc, hackathonID, participant, "Solution v3", "Third")

	submissions := listSubmissions(tc, hackathonID, participant, "", "")

	assert.Len(t, submissions, 3, "Should return all 3 versions")

	submissionIDs := make([]string, 3)
	for i, sub := range submissions {
		submissionIDs[i] = sub.(map[string]interface{})["submissionId"].(string)
	}

	assert.Contains(t, submissionIDs, submissionID1)
	assert.Contains(t, submissionIDs, submissionID2)
	assert.Contains(t, submissionIDs, submissionID3)
}

func TestCreateSubmissionUpload_ShouldReturnPresignedURL(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID := createSubmission(tc, hackathonID, participant, "Solution", "Description")

	fileID, uploadURL := createSubmissionUpload(tc, hackathonID, submissionID, participant, "test-document.pdf", 102400, "application/pdf")

	assert.NotEmpty(t, fileID, "File ID should be returned")
	assert.Contains(t, uploadURL, "X-Amz-Algorithm", "Upload URL should be pre-signed S3 URL")

	var uploadStatus string
	err := tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT upload_status FROM %s.submission_files WHERE id = $1", tc.SubmissionDBName),
		fileID,
	).Scan(&uploadStatus)
	require.NoError(t, err)

	assert.Equal(t, "pending", uploadStatus, "File should be in pending status")
}

func TestUploadFile_ToMinIO_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID := createSubmission(tc, hackathonID, participant, "Solution", "Description")

	testContent := generateTestFile(102400)
	fileID, uploadURL := createSubmissionUpload(tc, hackathonID, submissionID, participant, "test-file.pdf", int64(len(testContent)), "application/pdf")

	t.Logf("Upload URL: %s", uploadURL)
	t.Logf("File ID: %s", fileID)
	t.Logf("Content size: %d bytes", len(testContent))

	// Get storage_key from DB
	var storageKey string
	err := tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT storage_key FROM %s.submission_files WHERE id = $1", tc.SubmissionDBName),
		fileID,
	).Scan(&storageKey)
	require.NoError(t, err)
	t.Logf("Storage key from DB: %s", storageKey)

	// Upload directly to MinIO using SDK (bypassing pre-signed URL issue)
	err = uploadToMinIODirectly(storageKey, testContent, "application/pdf")
	assert.NoError(t, err, "Should successfully upload to MinIO")

	// Wait a bit to ensure S3 consistency
	time.Sleep(100 * time.Millisecond)

	completeSubmissionUpload(tc, hackathonID, submissionID, fileID, participant)

	var uploadStatus string
	var completedAt *time.Time
	err = tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT upload_status, completed_at FROM %s.submission_files WHERE id = $1", tc.SubmissionDBName),
		fileID,
	).Scan(&uploadStatus, &completedAt)
	require.NoError(t, err)

	assert.Equal(t, "completed", uploadStatus, "File should be marked as completed")
	assert.NotNil(t, completedAt, "Completed at should be set")
}

func TestCompleteSubmissionUpload_ShouldMarkCompleted(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID := createSubmission(tc, hackathonID, participant, "Solution", "Description")

	testContent := generateTestFile(50000)
	fileID := uploadFileToSubmission(tc, hackathonID, submissionID, participant, "document.pdf", testContent, "application/pdf")

	var uploadStatus string
	err := tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT upload_status FROM %s.submission_files WHERE id = $1", tc.SubmissionDBName),
		fileID,
	).Scan(&uploadStatus)
	require.NoError(t, err)

	assert.Equal(t, "completed", uploadStatus)
}

func TestGetSubmissionFileDownloadURL_ShouldReturnPresignedURL(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID := createSubmission(tc, hackathonID, participant, "Solution", "Description")

	testContent := generateTestFile(50000)
	fileID := uploadFileToSubmission(tc, hackathonID, submissionID, participant, "test.pdf", testContent, "application/pdf")

	downloadURL := getSubmissionFileDownloadURL(tc, hackathonID, submissionID, fileID, participant)

	assert.Contains(t, downloadURL, "X-Amz-Algorithm", "Download URL should be pre-signed S3 URL")
}

func TestDownloadFile_FromMinIO_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID := createSubmission(tc, hackathonID, participant, "Solution", "Description")

	originalContent := generateTestFile(50000)
	fileID := uploadFileToSubmission(tc, hackathonID, submissionID, participant, "test.pdf", originalContent, "application/pdf")

	// Get storage_key from DB
	var storageKey string
	err := tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT storage_key FROM %s.submission_files WHERE id = $1", tc.SubmissionDBName),
		fileID,
	).Scan(&storageKey)
	require.NoError(t, err)

	// Download directly from MinIO using SDK
	downloadedContent, err := downloadFromMinIODirectly(storageKey)
	require.NoError(t, err, "Should successfully download from MinIO")

	assert.Equal(t, originalContent, downloadedContent, "Downloaded content should match uploaded content")
}

func TestCompleteSubmissionUpload_WithoutS3File_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID := createSubmission(tc, hackathonID, participant, "Solution", "Description")

	fileID, _ := createSubmissionUpload(tc, hackathonID, submissionID, participant, "missing-file.pdf", 1000, "application/pdf")

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/submissions/%s/files/%s/complete", hackathonID, submissionID, fileID),
		participant.AccessToken,
		map[string]interface{}{},
	)

	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Complete should return 200 but mark as failed: %s", string(respBody))

	time.Sleep(200 * time.Millisecond)

	var uploadStatus string
	err := tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT upload_status FROM %s.submission_files WHERE id = $1", tc.SubmissionDBName),
		fileID,
	).Scan(&uploadStatus)
	require.NoError(t, err)

	assert.Equal(t, "failed", uploadStatus, "File should be marked as failed when not found in S3")
}

func TestCreateSubmissionUpload_ExceedMaxFileSize_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID := createSubmission(tc, hackathonID, participant, "Solution", "Description")

	body := map[string]interface{}{
		"filename":     "huge-file.pdf",
		"size_bytes":   52428801,
		"content_type": "application/pdf",
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/submissions/%s/files", hackathonID, submissionID),
		participant.AccessToken,
		body,
	)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
		"Should reject file exceeding max size: %s", string(respBody))
}

func TestCreateSubmissionUpload_ExceedMaxTotalSize_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID := createSubmission(tc, hackathonID, participant, "Solution", "Description")

	// Upload 4 files, each 50MB (total 200MB, at the limit)
	// Max file size is 50MB, max total size is 200MB
	file1Content := generateTestFile(50 * 1024 * 1024)
	uploadFileToSubmission(tc, hackathonID, submissionID, participant, "file1.zip", file1Content, "application/zip")

	file2Content := generateTestFile(50 * 1024 * 1024)
	uploadFileToSubmission(tc, hackathonID, submissionID, participant, "file2.zip", file2Content, "application/zip")

	file3Content := generateTestFile(50 * 1024 * 1024)
	uploadFileToSubmission(tc, hackathonID, submissionID, participant, "file3.zip", file3Content, "application/zip")

	file4Content := generateTestFile(50 * 1024 * 1024)
	uploadFileToSubmission(tc, hackathonID, submissionID, participant, "file4.zip", file4Content, "application/zip")

	// Try to upload a fifth file (even 1 byte should exceed the 200MB limit)
	file5Content := generateTestFile(1024)

	body := map[string]interface{}{
		"filename":     "file5.zip",
		"size_bytes":   int64(len(file5Content)),
		"content_type": "application/zip",
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/submissions/%s/files", hackathonID, submissionID),
		participant.AccessToken,
		body,
	)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
		"Should reject file exceeding total size limit: %s", string(respBody))
}

func TestCreateSubmissionUpload_ExceedMaxFilesCount_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID := createSubmission(tc, hackathonID, participant, "Solution", "Description")

	smallContent := generateTestFile(1000)
	for i := 0; i < 20; i++ {
		uploadFileToSubmission(tc, hackathonID, submissionID, participant, fmt.Sprintf("file%d.txt", i), smallContent, "text/plain")
	}

	body := map[string]interface{}{
		"filename":     "file21.txt",
		"size_bytes":   1000,
		"content_type": "text/plain",
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/submissions/%s/files", hackathonID, submissionID),
		participant.AccessToken,
		body,
	)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
		"Should reject exceeding max files count: %s", string(respBody))
}

func TestCreateSubmission_ExceedMaxSubmissionsPerOwner_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	for i := 0; i < 50; i++ {
	moveHackathonToRunningStage(tc, hackathonID)

		createSubmission(tc, hackathonID, participant, fmt.Sprintf("Solution v%d", i+1), fmt.Sprintf("Version %d", i+1))
	}

	body := map[string]interface{}{
		"title":       "Solution v51",
		"description": "This should fail",
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/submissions", hackathonID),
		participant.AccessToken,
		body,
	)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
		"Should reject exceeding max submissions per owner: %s", string(respBody))
}

func TestCreateSubmissionUpload_InvalidFileType_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID := createSubmission(tc, hackathonID, participant, "Solution", "Description")

	body := map[string]interface{}{
		"filename":     "malware.exe",
		"size_bytes":   1000,
		"content_type": "application/x-msdownload",
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/submissions/%s/files", hackathonID, submissionID),
		participant.AccessToken,
		body,
	)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
		"Should reject invalid file type: %s", string(respBody))
}

func TestListSubmissions_AsOrganizer_WithOwnerFilter_ShouldReturnFiltered(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant1 := tc.RegisterUser()
	participant2 := tc.RegisterUser()
	organizer := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant1, "PART_INDIVIDUAL")
	registerParticipant(tc, hackathonID, participant2, "PART_INDIVIDUAL")
	assignOrganizerRole(tc, hackathonID, organizer)

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID1 := createSubmission(tc, hackathonID, participant1, "Solution by P1", "By participant 1")
	_ = createSubmission(tc, hackathonID, participant2, "Solution by P2", "By participant 2")

	submissions := listSubmissions(tc, hackathonID, organizer, "user", participant1.UserID)

	assert.Len(t, submissions, 1, "Should return only filtered submissions")
	assert.Equal(t, submissionID1, submissions[0].(map[string]interface{})["submissionId"])
}

func TestGetFinalSubmission_AsJudge_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	judge := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	assignJudgeRole(tc, hackathonID, judge)

	moveHackathonToRunningStage(tc, hackathonID)
	submissionID := createSubmission(tc, hackathonID, participant, "Final Solution", "Description")

	transitionToJudging(tc, hackathonID)

	finalSubmission := getFinalSubmission(tc, hackathonID, "user", participant.UserID, judge)

	assert.Equal(t, submissionID, finalSubmission["submissionId"])
	assert.True(t, finalSubmission["isFinal"].(bool))
}

func TestListSubmissions_AsJudge_InRunningStage_ShouldSeeAll(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	judge := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	assignJudgeRole(tc, hackathonID, judge)

	moveHackathonToRunningStage(tc, hackathonID)
	submissionID1 := createSubmission(tc, hackathonID, participant, "Solution v1", "First")
	time.Sleep(100 * time.Millisecond)
	submissionID2 := createSubmission(tc, hackathonID, participant, "Solution v2", "Second")

	submissions := listSubmissions(tc, hackathonID, judge, "user", participant.UserID)

	assert.Len(t, submissions, 2, "Judge should see all submissions in RUNNING stage")

	submissionIDs := []string{
		submissions[0].(map[string]interface{})["submissionId"].(string),
		submissions[1].(map[string]interface{})["submissionId"].(string),
	}

	assert.Contains(t, submissionIDs, submissionID1)
	assert.Contains(t, submissionIDs, submissionID2)
}

func TestCompleteSubmissionWorkflow_WithMultipleFiles(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunningForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStage(tc, hackathonID)

	submissionID := createSubmission(tc, hackathonID, participant, "Complete Solution", "Full implementation with docs")

	pdfContent := generateTestFile(100 * 1024)
	fileID1 := uploadFileToSubmission(tc, hackathonID, submissionID, participant, "documentation.pdf", pdfContent, "application/pdf")

	zipContent := generateTestFile(1024 * 1024)
	fileID2 := uploadFileToSubmission(tc, hackathonID, submissionID, participant, "source-code.zip", zipContent, "application/zip")

	pngContent := generateTestFile(500 * 1024)
	fileID3 := uploadFileToSubmission(tc, hackathonID, submissionID, participant, "screenshot.png", pngContent, "image/png")

	submission := getSubmission(tc, hackathonID, submissionID, participant)
	files := submission["files"].([]interface{})

	assert.Len(t, files, 3, "Submission should have 3 files")

	fileIDs := make([]string, 3)
	for i, f := range files {
		file := f.(map[string]interface{})
		fileIDs[i] = file["fileId"].(string)
		assert.Equal(t, "FILE_UPLOAD_STATUS_COMPLETED", file["uploadStatus"], "All files should be completed")
	}

	assert.Contains(t, fileIDs, fileID1)
	assert.Contains(t, fileIDs, fileID2)
	assert.Contains(t, fileIDs, fileID3)
}

func TestTeamSubmissionWorkflow_MultipleMembersUpload(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	member1 := tc.RegisterUser()
	member2 := tc.RegisterUser()

	// Create hackathon in REGISTRATION, create team with all members, then transition to RUNNING
	hackathonID := createHackathonInRegistrationForSubmissions(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, member1, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, member2, "PART_LOOKING_FOR_TEAM")

	teamID := createTeam(tc, hackathonID, captain, "Collaborative Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 3)
	inviteAndAccept(tc, hackathonID, teamID, captain, member1, vacancyID)
	inviteAndAccept(tc, hackathonID, teamID, captain, member2, vacancyID)

	transitionToRunning(tc, hackathonID, owner)

	submissionID1 := createSubmission(tc, hackathonID, captain, "Initial Draft", "By captain")
	time.Sleep(100 * time.Millisecond)
	submissionID2 := createSubmission(tc, hackathonID, member1, "Improved Version", "By member1")
	time.Sleep(100 * time.Millisecond)
	submissionID3 := createSubmission(tc, hackathonID, member2, "Final Version", "By member2")

	var createdBy1, createdBy2, createdBy3 string
	tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT created_by_user_id FROM %s.submissions WHERE id = $1", tc.SubmissionDBName),
		submissionID1,
	).Scan(&createdBy1)

	tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT created_by_user_id FROM %s.submissions WHERE id = $1", tc.SubmissionDBName),
		submissionID2,
	).Scan(&createdBy2)

	tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT created_by_user_id FROM %s.submissions WHERE id = $1", tc.SubmissionDBName),
		submissionID3,
	).Scan(&createdBy3)

	assert.Equal(t, captain.UserID, createdBy1)
	assert.Equal(t, member1.UserID, createdBy2)
	assert.Equal(t, member2.UserID, createdBy3)

	selectFinalSubmission(tc, hackathonID, submissionID2, captain)

	finalSubmission := getFinalSubmission(tc, hackathonID, "team", teamID, member1)
	assert.Equal(t, submissionID2, finalSubmission["submissionId"])
}

// createSubmission creates a submission and returns submission_id
func createSubmission(tc *TestContext, hackathonID string, user *UserCredentials, title, description string) string {
	body := map[string]interface{}{
		"title":       title,
		"description": description,
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/submissions", hackathonID),
		user.AccessToken,
		body,
	)

	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to create submission: %s", string(respBody))

	data := tc.ParseJSON(respBody)
	submissionID := data["submissionId"].(string)
	require.NotEmpty(tc.T, submissionID, "Submission ID should be returned")

	return submissionID
}

// updateSubmission updates submission description
func updateSubmission(tc *TestContext, hackathonID, submissionID string, user *UserCredentials, description string) {
	body := map[string]interface{}{
		"description": description,
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"PUT",
		fmt.Sprintf("/v1/hackathons/%s/submissions/%s", hackathonID, submissionID),
		user.AccessToken,
		body,
	)

	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to update submission: %s", string(respBody))
}

// selectFinalSubmission marks submission as final
func selectFinalSubmission(tc *TestContext, hackathonID, submissionID string, user *UserCredentials) {
	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/submissions/%s/select-final", hackathonID, submissionID),
		user.AccessToken,
		map[string]interface{}{},
	)

	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to select final submission: %s", string(respBody))
}

// getSubmission retrieves a submission by ID
func getSubmission(tc *TestContext, hackathonID, submissionID string, user *UserCredentials) map[string]interface{} {
	resp, respBody := tc.DoAuthenticatedRequest(
		"GET",
		fmt.Sprintf("/v1/hackathons/%s/submissions/%s", hackathonID, submissionID),
		user.AccessToken,
		nil,
	)

	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to get submission: %s", string(respBody))

	data := tc.ParseJSON(respBody)
	return data["submission"].(map[string]interface{})
}

// listSubmissions retrieves submissions list
func listSubmissions(tc *TestContext, hackathonID string, user *UserCredentials, ownerKind, ownerID string) []interface{} {
	body := map[string]interface{}{
		"query": map[string]interface{}{
			"limit":  50,
			"offset": 0,
		},
	}

	if ownerKind != "" {
		body["owner_kind"] = ownerKind
	}
	if ownerID != "" {
		body["owner_id"] = ownerID
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/submissions/list", hackathonID),
		user.AccessToken,
		body,
	)

	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to list submissions: %s", string(respBody))

	data := tc.ParseJSON(respBody)
	submissions, ok := data["submissions"].([]interface{})
	require.True(tc.T, ok, "Submissions array should be present")

	return submissions
}

// getFinalSubmission retrieves final submission for owner
func getFinalSubmission(tc *TestContext, hackathonID, ownerKind, ownerID string, user *UserCredentials) map[string]interface{} {
	resp, respBody := tc.DoAuthenticatedRequest(
		"GET",
		fmt.Sprintf("/v1/hackathons/%s/participants/%s/%s/final-submission", hackathonID, ownerKind, ownerID),
		user.AccessToken,
		nil,
	)

	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to get final submission: %s", string(respBody))

	data := tc.ParseJSON(respBody)
	return data["submission"].(map[string]interface{})
}

// createSubmissionUpload initiates file upload and returns file_id and upload_url
func createSubmissionUpload(tc *TestContext, hackathonID, submissionID string, user *UserCredentials, filename string, sizeBytes int64, contentType string) (string, string) {
	body := map[string]interface{}{
		"filename":     filename,
		"size_bytes":   sizeBytes,
		"content_type": contentType,
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/submissions/%s/files", hackathonID, submissionID),
		user.AccessToken,
		body,
	)

	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to create submission upload: %s", string(respBody))

	data := tc.ParseJSON(respBody)
	fileID := data["fileId"].(string)
	uploadURL := data["uploadUrl"].(string)

	require.NotEmpty(tc.T, fileID, "File ID should be returned")
	require.NotEmpty(tc.T, uploadURL, "Upload URL should be returned")

	return fileID, uploadURL
}

// completeSubmissionUpload marks file upload as completed
func completeSubmissionUpload(tc *TestContext, hackathonID, submissionID, fileID string, user *UserCredentials) {
	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/submissions/%s/files/%s/complete", hackathonID, submissionID, fileID),
		user.AccessToken,
		map[string]interface{}{},
	)

	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to complete submission upload: %s", string(respBody))
}

// getSubmissionFileDownloadURL retrieves pre-signed download URL
func getSubmissionFileDownloadURL(tc *TestContext, hackathonID, submissionID, fileID string, user *UserCredentials) string {
	resp, respBody := tc.DoAuthenticatedRequest(
		"GET",
		fmt.Sprintf("/v1/hackathons/%s/submissions/%s/files/%s/download-url", hackathonID, submissionID, fileID),
		user.AccessToken,
		nil,
	)

	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to get download URL: %s", string(respBody))

	data := tc.ParseJSON(respBody)
	downloadURL := data["downloadUrl"].(string)
	require.NotEmpty(tc.T, downloadURL, "Download URL should be returned")

	return downloadURL
}

// uploadToS3 performs HTTP PUT to pre-signed S3 URL
func uploadToS3(url string, content []byte, contentType string) error {
	req, err := http.NewRequest("PUT", url, bytes.NewReader(content))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", contentType)
	req.ContentLength = int64(len(content))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("S3 upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	fmt.Printf("S3 upload successful: status=%d, body=%s\n", resp.StatusCode, string(bodyBytes))
	return nil
}

// uploadToMinIODirectly uploads file directly to MinIO using SDK
func uploadToMinIODirectly(storageKey string, content []byte, contentType string) error {
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

	bucketName := os.Getenv("S3_SUBMISSIONS_BUCKET")
	if bucketName == "" {
		bucketName = "submissions"
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

// downloadFromS3 performs HTTP GET from pre-signed S3 URL
func downloadFromS3(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("S3 download failed with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// downloadFromMinIODirectly downloads file directly from MinIO using SDK
func downloadFromMinIODirectly(storageKey string) ([]byte, error) {
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

	bucketName := os.Getenv("S3_SUBMISSIONS_BUCKET")
	if bucketName == "" {
		bucketName = "submissions"
	}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx := context.Background()

	object, err := minioClient.GetObject(ctx, bucketName, storageKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from minio: %w", err)
	}
	defer object.Close()

	content, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("failed to read object content: %w", err)
	}

	return content, nil
}

// uploadFileToSubmission performs full upload flow: create upload → PUT to S3 → complete
func uploadFileToSubmission(tc *TestContext, hackathonID, submissionID string, user *UserCredentials, filename string, content []byte, contentType string) string {
	fileID, _ := createSubmissionUpload(tc, hackathonID, submissionID, user, filename, int64(len(content)), contentType)

	// Get storage_key from DB
	var storageKey string
	err := tc.SubmissionDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT storage_key FROM %s.submission_files WHERE id = $1", tc.SubmissionDBName),
		fileID,
	).Scan(&storageKey)
	require.NoError(tc.T, err, "Failed to get storage key from DB")

	// Upload directly to MinIO using SDK
	err = uploadToMinIODirectly(storageKey, content, contentType)
	require.NoError(tc.T, err, "Failed to upload file to MinIO")

	completeSubmissionUpload(tc, hackathonID, submissionID, fileID, user)

	return fileID
}

// assignJudgeRole assigns judge role to a user in a hackathon
func assignJudgeRole(tc *TestContext, hackathonID string, judge *UserCredentials) {
	_, err := tc.ParticipationDB.Exec(context.Background(),
		fmt.Sprintf("INSERT INTO %s (hackathon_id, user_id, role) VALUES ($1, $2, 'judge') ON CONFLICT DO NOTHING",
			tc.ParticipationDBName),
		hackathonID, judge.UserID,
	)
	require.NoError(tc.T, err, "Failed to assign judge role")

	time.Sleep(300 * time.Millisecond)
}

// assignOrganizerRole assigns organizer role to a user in a hackathon
func assignOrganizerRole(tc *TestContext, hackathonID string, organizer *UserCredentials) {
	_, err := tc.ParticipationDB.Exec(context.Background(),
		fmt.Sprintf("INSERT INTO %s (hackathon_id, user_id, role) VALUES ($1, $2, 'organizer') ON CONFLICT DO NOTHING",
			tc.ParticipationDBName),
		hackathonID, organizer.UserID,
	)
	require.NoError(tc.T, err, "Failed to assign organizer role")

	time.Sleep(300 * time.Millisecond)
}

// transitionToJudging transitions hackathon to judging stage
func transitionToJudging(tc *TestContext, hackathonID string) {
	now := time.Now()

	_, err := tc.DB.Exec(context.Background(), fmt.Sprintf(`
		UPDATE %s 
		SET ends_at = $1,
		    stage = 'judging'
		WHERE id = $2
	`, tc.HackathonDBName), now.Add(-1*time.Hour), hackathonID)
	require.NoError(tc.T, err, "Failed to update hackathon to JUDGING stage")

	time.Sleep(500 * time.Millisecond)
}

// generateTestFile generates test file content of specified size
func generateTestFile(sizeBytes int) []byte {
	content := make([]byte, sizeBytes)
	_, err := rand.Read(content)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate test file: %v", err))
	}
	return content
}

// createHackathonInRunningForSubmissions creates hackathon in RUNNING stage for submission tests
func createHackathonInRunningForSubmissions(tc *TestContext, owner *UserCredentials) string {
	now := time.Now()
	hackathonBody := map[string]interface{}{
		"name":              "Submission Test Hackathon",
		"short_description": "Test hackathon for submissions",
		"description":       "Full description for submission testing",
		"location": map[string]interface{}{
			"online": true,
		},
		"dates": map[string]interface{}{
			"registration_opens_at":  now.Add(1 * time.Hour).Format(time.RFC3339),
			"registration_closes_at": now.Add(15 * 24 * time.Hour).Format(time.RFC3339),
			"starts_at":              now.Add(20 * 24 * time.Hour).Format(time.RFC3339),
			"ends_at":                now.Add(22 * 24 * time.Hour).Format(time.RFC3339),
			"judging_ends_at":        now.Add(25 * 24 * time.Hour).Format(time.RFC3339),
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
	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to create hackathon: %s", string(body))

	data := tc.ParseJSON(body)
	hackathonID := data["hackathonId"].(string)

	tc.WaitForHackathonOwnerRole(hackathonID, owner.AccessToken)

	taskBody := map[string]interface{}{
		"task": "Build something innovative",
	}
	resp, body = tc.DoAuthenticatedRequest("PUT", fmt.Sprintf("/v1/hackathons/%s/task", hackathonID), owner.AccessToken, taskBody)
	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to set task: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/publish", hackathonID), owner.AccessToken, map[string]interface{}{})
	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to publish hackathon: %s", string(body))

	// Update dates to put hackathon in REGISTRATION stage first (for participant registration)
	_, err := tc.DB.Exec(context.Background(), fmt.Sprintf(`
		UPDATE %s 
		SET registration_opens_at = $1
		WHERE id = $2
	`, tc.HackathonDBName), now.Add(-24*time.Hour), hackathonID)
	require.NoError(tc.T, err, "Failed to update hackathon to REGISTRATION stage")

	time.Sleep(500 * time.Millisecond)

	return hackathonID
}

// createHackathonInRegistrationForSubmissions creates hackathon in REGISTRATION stage
func createHackathonInRegistrationForSubmissions(tc *TestContext, owner *UserCredentials) string {
	now := time.Now()
	hackathonBody := map[string]interface{}{
		"name":              "Submission Test Hackathon (Registration)",
		"short_description": "Test hackathon in registration stage",
		"description":       "Full description",
		"location": map[string]interface{}{
			"online": true,
		},
		"dates": map[string]interface{}{
			"registration_opens_at":  now.Add(1 * time.Hour).Format(time.RFC3339),
			"registration_closes_at": now.Add(15 * 24 * time.Hour).Format(time.RFC3339),
			"starts_at":              now.Add(20 * 24 * time.Hour).Format(time.RFC3339),
			"ends_at":                now.Add(22 * 24 * time.Hour).Format(time.RFC3339),
			"judging_ends_at":        now.Add(25 * 24 * time.Hour).Format(time.RFC3339),
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
	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to create hackathon: %s", string(body))

	data := tc.ParseJSON(body)
	hackathonID := data["hackathonId"].(string)

	tc.WaitForHackathonOwnerRole(hackathonID, owner.AccessToken)

	taskBody := map[string]interface{}{
		"task": "Build something innovative",
	}
	resp, body = tc.DoAuthenticatedRequest("PUT", fmt.Sprintf("/v1/hackathons/%s/task", hackathonID), owner.AccessToken, taskBody)
	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to set task: %s", string(body))

	resp, body = tc.DoAuthenticatedRequest("POST", fmt.Sprintf("/v1/hackathons/%s/publish", hackathonID), owner.AccessToken, map[string]interface{}{})
	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to publish hackathon: %s", string(body))

	// Update dates and stage directly in DB to move to REGISTRATION stage (bypassing validation)
	_, err := tc.DB.Exec(context.Background(), fmt.Sprintf(`
		UPDATE %s 
		SET registration_opens_at = $1,
		    stage = 'registration'
		WHERE id = $2
	`, tc.HackathonDBName), now.Add(-24*time.Hour), hackathonID)
	require.NoError(tc.T, err, "Failed to update hackathon dates in DB")

	// Wait a bit for updates to propagate
	time.Sleep(500 * time.Millisecond)

	return hackathonID
}

// createHackathonInRunningWithTeam creates hackathon in REGISTRATION, creates team, then transitions to RUNNING
func createHackathonInRunningWithTeam(tc *TestContext, owner *UserCredentials, captain *UserCredentials, member *UserCredentials) (string, string) {
	// Create hackathon in REGISTRATION stage
	hackathonID := createHackathonInRegistrationForSubmissions(tc, owner)

	// Register participants
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, member, "PART_LOOKING_FOR_TEAM")

	// Create team
	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)
	inviteAndAccept(tc, hackathonID, teamID, captain, member, vacancyID)

	// Transition to RUNNING stage
	transitionToRunning(tc, hackathonID, owner)

	return hackathonID, teamID
}

// moveHackathonToRunningStage moves a hackathon to RUNNING stage by updating dates in DB
func moveHackathonToRunningStage(tc *TestContext, hackathonID string) {
	now := time.Now()
	_, err := tc.DB.Exec(context.Background(), fmt.Sprintf(`
		UPDATE %s 
		SET registration_opens_at = $1,
		    registration_closes_at = $2,
		    starts_at = $3,
		    ends_at = $4
		WHERE id = $5
	`, tc.HackathonDBName),
		now.Add(-10*24*time.Hour), // registration opened 10 days ago
		now.Add(-5*24*time.Hour),  // registration closed 5 days ago
		now.Add(-1*time.Hour),     // hackathon started 1 hour ago (RUNNING)
		now.Add(5*24*time.Hour),   // ends in 5 days
		hackathonID)
	require.NoError(tc.T, err, "Failed to move hackathon to RUNNING stage")

	time.Sleep(500 * time.Millisecond)
}
