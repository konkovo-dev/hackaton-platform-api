package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendMessage_WithoutAuth_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)

	body := map[string]interface{}{
		"text": "Test message",
	}

	resp, respBody := tc.DoRequest("POST", fmt.Sprintf("/v1/hackathons/%s/support/messages", hackathonID), body, nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
		"Should reject unauthenticated request: %s", string(respBody))
}

func TestSendMessage_InRegistrationStage_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	body := map[string]interface{}{
		"text": "Test message",
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/messages", hackathonID),
		participant.AccessToken,
		body,
	)

	assert.NotEqual(t, http.StatusOK, resp.StatusCode,
		"Should reject operations in non-RUNNING stage: %s", string(respBody))
}

func TestSendMessage_AsIndividualParticipant_ShouldCreateUserTicket(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStageForMentors(tc, hackathonID)

	body := map[string]interface{}{
		"text": "I need help with the API",
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/messages", hackathonID),
		participant.AccessToken,
		body,
	)

	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to send message: %s", string(respBody))

	data := tc.ParseJSON(respBody)
	ticketID := data["ticketId"].(string)
	messageID := data["messageId"].(string)

	assert.NotEmpty(t, ticketID, "Should return ticket ID")
	assert.NotEmpty(t, messageID, "Should return message ID")

	var ownerKind string
	var ownerID string
	var status string
	err := tc.MentorsDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT owner_kind, owner_id, status FROM %s.tickets WHERE id = $1", tc.MentorsDBName),
		ticketID,
	).Scan(&ownerKind, &ownerID, &status)
	require.NoError(t, err)

	assert.Equal(t, "user", ownerKind, "Should create USER owner ticket")
	assert.Equal(t, participant.UserID, ownerID, "Owner should be the participant")
	assert.Equal(t, "open", status, "Ticket should be OPEN")

	var messageText string
	var authorRole string
	err = tc.MentorsDB.QueryRow(context.Background(),
		"SELECT text, author_role FROM mentors.messages WHERE id = $1",
		messageID,
	).Scan(&messageText, &authorRole)
	require.NoError(t, err)

	assert.Equal(t, "I need help with the API", messageText)
	assert.Equal(t, "participant", authorRole)

	var eventCount int
	err = tc.MentorsDB.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM mentors.outbox_events WHERE aggregate_id = $1 AND event_type = 'message.created'",
		messageID,
	).Scan(&eventCount)
	require.NoError(t, err)
	assert.Equal(t, 1, eventCount, "Should create outbox event for message")
}

func TestSendMessage_AsTeamMember_ShouldCreateTeamTicket(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	member := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, member, "PART_LOOKING_FOR_TEAM")

	teamID := createTeam(tc, hackathonID, captain, "Test Team")

	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)
	inviteBody := map[string]interface{}{
		"target_user_id": member.UserID,
		"vacancy_id":     vacancyID,
		"message":        "Join us!",
	}
	resp, body := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathonID, teamID),
		captain.AccessToken,
		inviteBody,
	)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	inviteData := tc.ParseJSON(body)
	invitationID := inviteData["invitationId"].(string)

	acceptResp, acceptBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/users/me/team-invitations/%s/accept", invitationID),
		member.AccessToken,
		map[string]interface{}{},
	)
	require.Equal(t, http.StatusOK, acceptResp.StatusCode, "Failed to accept invitation: %s", string(acceptBody))
	time.Sleep(500 * time.Millisecond)

	var memberCount int
	err := tc.TeamDB.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM team.memberships WHERE team_id = $1 AND user_id = $2",
		teamID, member.UserID,
	).Scan(&memberCount)
	require.NoError(t, err)
	require.Equal(t, 1, memberCount, "Member should be in team after accepting invitation")

	partResp, partBody := tc.DoAuthenticatedRequest(
		"GET",
		fmt.Sprintf("/v1/hackathons/%s/participations/me", hackathonID),
		member.AccessToken,
		nil,
	)
	require.Equal(t, http.StatusOK, partResp.StatusCode, "Failed to get participation: %s", string(partBody))
	partData := tc.ParseJSON(partBody)
	t.Logf("Participation data: %+v", partData)
	require.NotNil(t, partData["participation"], "Participation should exist")
	participation := partData["participation"].(map[string]interface{})
	require.NotEmpty(t, participation["teamId"], "Team ID should be set in participation")

	transitionToRunning(tc, hackathonID, owner)

	messageBody := map[string]interface{}{
		"text": "Our team needs help",
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/messages", hackathonID),
		member.AccessToken,
		messageBody,
	)

	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to send message: %s", string(respBody))

	data := tc.ParseJSON(respBody)
	ticketID := data["ticketId"].(string)
	messageID := data["messageId"].(string)

	var ownerKind string
	var ownerID string
	var authorUserID string
	err = tc.MentorsDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT t.owner_kind, t.owner_id, m.author_user_id FROM %s.tickets t JOIN %s.messages m ON t.id = m.ticket_id WHERE t.id = $1 AND m.id = $2", tc.MentorsDBName, tc.MentorsDBName),
		ticketID, messageID,
	).Scan(&ownerKind, &ownerID, &authorUserID)
	require.NoError(t, err)

	assert.Equal(t, "team", ownerKind, "Should create TEAM owner ticket")
	assert.Equal(t, teamID, ownerID, "Owner should be the team")
	assert.Equal(t, member.UserID, authorUserID, "Author should be the member who sent message")
}

func TestSendMessage_SecondMessageToOpenTicket_ShouldReuseTicket(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStageForMentors(tc, hackathonID)

	firstBody := map[string]interface{}{
		"text": "First message",
	}
	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/messages", hackathonID),
		participant.AccessToken,
		firstBody,
	)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	firstData := tc.ParseJSON(respBody)
	firstTicketID := firstData["ticketId"].(string)

	time.Sleep(300 * time.Millisecond)

	secondBody := map[string]interface{}{
		"text": "Second message",
	}
	resp, respBody = tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/messages", hackathonID),
		participant.AccessToken,
		secondBody,
	)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	secondData := tc.ParseJSON(respBody)
	secondTicketID := secondData["ticketId"].(string)

	assert.Equal(t, firstTicketID, secondTicketID, "Should reuse the same OPEN ticket")

	var openTicketCount int
	err := tc.MentorsDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT COUNT(*) FROM %s.tickets WHERE hackathon_id = $1 AND owner_id = $2 AND status = 'open'", tc.MentorsDBName),
		hackathonID, participant.UserID,
	).Scan(&openTicketCount)
	require.NoError(t, err)

	assert.Equal(t, 1, openTicketCount, "Should have exactly 1 OPEN ticket")

	var messageCount int
	err = tc.MentorsDB.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM mentors.messages WHERE ticket_id = $1",
		firstTicketID,
	).Scan(&messageCount)
	require.NoError(t, err)

	assert.Equal(t, 2, messageCount, "Should have 2 messages in the ticket")
}

func TestSendMessage_AsMentor_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	assignMentorRole(tc, hackathonID, mentor)

	body := map[string]interface{}{
		"text": "Mentor trying to send message",
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/messages", hackathonID),
		mentor.AccessToken,
		body,
	)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"Staff should not be able to send messages as participants: %s", string(respBody))
}

func TestSendMessage_WithClientMessageID_ShouldBeIdempotent(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStageForMentors(tc, hackathonID)

	clientMsgID := uuid.New().String()
	body := map[string]interface{}{
		"text":              "Test message",
		"client_message_id": clientMsgID,
	}

	resp1, respBody1 := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/messages", hackathonID),
		participant.AccessToken,
		body,
	)
	require.Equal(t, http.StatusOK, resp1.StatusCode)

	data1 := tc.ParseJSON(respBody1)
	messageID1 := data1["messageId"].(string)
	ticketID1 := data1["ticketId"].(string)

	time.Sleep(200 * time.Millisecond)

	resp2, respBody2 := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/messages", hackathonID),
		participant.AccessToken,
		body,
	)
	require.Equal(t, http.StatusOK, resp2.StatusCode)

	data2 := tc.ParseJSON(respBody2)
	messageID2 := data2["messageId"].(string)
	ticketID2 := data2["ticketId"].(string)

	assert.Equal(t, messageID1, messageID2, "Should return same message ID")
	assert.Equal(t, ticketID1, ticketID2, "Should return same ticket ID")

	var messageCount int
	err := tc.MentorsDB.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM mentors.messages WHERE client_message_id = $1",
		clientMsgID,
	).Scan(&messageCount)
	require.NoError(t, err)

	assert.Equal(t, 1, messageCount, "Should have exactly 1 message with this client_message_id")
}

func TestGetMyChatMessages_AsIndividualParticipant_ShouldReturnAllMessages(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	moveHackathonToRunningStageForMentors(tc, hackathonID)

	assignMentorRole(tc, hackathonID, mentor)

	// Participant sends first message
	sendMessage(tc, hackathonID, participant, "First message")
	time.Sleep(200 * time.Millisecond)

	// Participant sends second message
	sendMessage(tc, hackathonID, participant, "Second message")
	time.Sleep(200 * time.Millisecond)

	resp, respBody := tc.DoAuthenticatedRequest(
		"GET",
		fmt.Sprintf("/v1/hackathons/%s/support/my-messages?query.limit=10&query.offset=0", hackathonID),
		participant.AccessToken,
		nil,
	)

	require.Equal(t, http.StatusOK, resp.StatusCode, "Should return messages: %s", string(respBody))

	data := tc.ParseJSON(respBody)
	messages := data["messages"].([]interface{})

	assert.GreaterOrEqual(t, len(messages), 2, "Should have at least 2 messages")

	firstMsg := messages[0].(map[string]interface{})
	assert.Equal(t, "First message", firstMsg["text"])
	assert.Equal(t, "AUTHOR_ROLE_PARTICIPANT", firstMsg["authorRole"])

	secondMsg := messages[1].(map[string]interface{})
	assert.Equal(t, "Second message", secondMsg["text"])
}

func TestGetMyChatMessages_AsTeamMember_ShouldReturnAllTeamMessages(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	captain := tc.RegisterUser()
	member := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createHackathonInRegistration(tc, owner)
	registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
	registerParticipant(tc, hackathonID, member, "PART_LOOKING_FOR_TEAM")
	assignMentorRole(tc, hackathonID, mentor)

	teamID := createTeam(tc, hackathonID, captain, "Test Team")
	vacancyID := createVacancy(tc, hackathonID, teamID, captain, 2)
	inviteAndAccept(tc, hackathonID, teamID, captain, member, vacancyID)

	transitionToRunning(tc, hackathonID, owner)

	// Captain sends message
	sendMessage(tc, hackathonID, captain, "Captain message")
	time.Sleep(200 * time.Millisecond)

	// Member sends message
	sendMessage(tc, hackathonID, member, "Member message")
	time.Sleep(200 * time.Millisecond)

	// Both team members should see all messages
	resp, respBody := tc.DoAuthenticatedRequest(
		"GET",
		fmt.Sprintf("/v1/hackathons/%s/support/my-messages?query.limit=10&query.offset=0", hackathonID),
		member.AccessToken,
		nil,
	)

	require.Equal(t, http.StatusOK, resp.StatusCode, "Should return team messages: %s", string(respBody))

	data := tc.ParseJSON(respBody)
	messages := data["messages"].([]interface{})

	assert.GreaterOrEqual(t, len(messages), 2, "Should have at least 2 messages")

	// Verify both messages are present
	texts := make([]string, 0)
	for _, msgData := range messages {
		msg := msgData.(map[string]interface{})
		texts = append(texts, msg["text"].(string))
	}

	assert.Contains(t, texts, "Captain message")
	assert.Contains(t, texts, "Member message")
}

func TestGetMyChatMessages_WithPagination_ShouldReturnCorrectly(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStageForMentors(tc, hackathonID)

	// Send 10 messages to ensure we have enough for pagination
	for i := 1; i <= 10; i++ {
		sendMessage(tc, hackathonID, participant, fmt.Sprintf("Message %d", i))
		time.Sleep(100 * time.Millisecond)
	}

	// Get first page with limit=5
	resp1, respBody1 := tc.DoAuthenticatedRequest(
		"GET",
		fmt.Sprintf("/v1/hackathons/%s/support/my-messages?query.limit=5&query.offset=0", hackathonID),
		participant.AccessToken,
		nil,
	)

	require.Equal(t, http.StatusOK, resp1.StatusCode)

	data1 := tc.ParseJSON(respBody1)
	messages1 := data1["messages"].([]interface{})
	hasMore1 := data1["hasMore"].(bool)

	t.Logf("First page: returned %d messages, hasMore=%v", len(messages1), hasMore1)
	for i, msgData := range messages1 {
		msg := msgData.(map[string]interface{})
		t.Logf("  [%d] %s (role=%s)", i, msg["text"], msg["authorRole"])
	}

	assert.GreaterOrEqual(t, len(messages1), 5, "Should return at least 5 messages")
	assert.True(t, hasMore1, "Should have more messages (got %d messages, hasMore=%v)", len(messages1), hasMore1)

	// Get second page
	resp2, respBody2 := tc.DoAuthenticatedRequest(
		"GET",
		fmt.Sprintf("/v1/hackathons/%s/support/my-messages?query.limit=5&query.offset=5", hackathonID),
		participant.AccessToken,
		nil,
	)

	require.Equal(t, http.StatusOK, resp2.StatusCode)

	data2 := tc.ParseJSON(respBody2)
	messages2 := data2["messages"].([]interface{})
	hasMore2 := data2["hasMore"].(bool)

	t.Logf("Second page: returned %d messages, hasMore=%v", len(messages2), hasMore2)
	for i, msgData := range messages2 {
		msg := msgData.(map[string]interface{})
		t.Logf("  [%d] %s (role=%s)", i, msg["text"], msg["authorRole"])
	}

	assert.GreaterOrEqual(t, len(messages2), 5, "Should return at least 5 remaining messages")

	// Verify all 10 participant messages are present across both pages
	allMessages := append(messages1, messages2...)
	participantMsgCount := 0
	for _, msgData := range allMessages {
		msg := msgData.(map[string]interface{})
		if msg["authorRole"] == "AUTHOR_ROLE_PARTICIPANT" {
			participantMsgCount++
		}
	}

	assert.GreaterOrEqual(t, participantMsgCount, 10, "Should have at least 10 participant messages")
}

func TestGetMyChatMessages_IncludesClosedTickets_ShouldReturnAll(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	moveHackathonToRunningStageForMentors(tc, hackathonID)

	assignMentorRole(tc, hackathonID, mentor)

	// First ticket
	ticketID1 := sendMessage(tc, hackathonID, participant, "First ticket message")
	time.Sleep(200 * time.Millisecond)

	claimTicket(tc, hackathonID, ticketID1, mentor)
	time.Sleep(200 * time.Millisecond)

	replyInTicket(tc, hackathonID, ticketID1, mentor, "Mentor reply")
	time.Sleep(200 * time.Millisecond)

	closeTicket(tc, hackathonID, ticketID1, mentor)
	time.Sleep(200 * time.Millisecond)

	// Second ticket (after first is closed)
	sendMessage(tc, hackathonID, participant, "Second ticket message")
	time.Sleep(200 * time.Millisecond)

	// Get all messages
	resp, respBody := tc.DoAuthenticatedRequest(
		"GET",
		fmt.Sprintf("/v1/hackathons/%s/support/my-messages?query.limit=50&query.offset=0", hackathonID),
		participant.AccessToken,
		nil,
	)

	require.Equal(t, http.StatusOK, resp.StatusCode, "Should return all messages: %s", string(respBody))

	data := tc.ParseJSON(respBody)
	messages := data["messages"].([]interface{})

	// Should include: first message + system "Mentor joined" + mentor reply + system "Ticket closed" + second message
	assert.GreaterOrEqual(t, len(messages), 5, "Should have messages from both tickets")

	texts := make([]string, 0)
	for _, msgData := range messages {
		msg := msgData.(map[string]interface{})
		texts = append(texts, msg["text"].(string))
	}

	assert.Contains(t, texts, "First ticket message")
	assert.Contains(t, texts, "Second ticket message")
	assert.Contains(t, texts, "Mentor reply")
}

func TestGetTicketMessages_AsParticipant_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStageForMentors(tc, hackathonID)

	ticketID := sendMessage(tc, hackathonID, participant, "My message")
	time.Sleep(200 * time.Millisecond)

	resp, respBody := tc.DoAuthenticatedRequest(
		"GET",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/messages?limit=10&offset=0", hackathonID, ticketID),
		participant.AccessToken,
		nil,
	)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"Participants should not use GetTicketMessages, they should use GetMyChatMessages: %s", string(respBody))
}

func TestGetTicketMessages_AsUnassignedMentor_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	moveHackathonToRunningStageForMentors(tc, hackathonID)

	assignMentorRole(tc, hackathonID, mentor)

	ticketID := sendMessage(tc, hackathonID, participant, "Private message")
	time.Sleep(200 * time.Millisecond)

	resp, respBody := tc.DoAuthenticatedRequest(
		"GET",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/messages?limit=10&offset=0", hackathonID, ticketID),
		mentor.AccessToken,
		nil,
	)

	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Unassigned mentor should be able to read messages: %s", string(respBody))

	data := tc.ParseJSON(respBody)
	messages := data["messages"].([]interface{})
	assert.GreaterOrEqual(t, len(messages), 1, "Should have at least 1 message")

	firstMessage := messages[0].(map[string]interface{})
	assert.Equal(t, "Private message", firstMessage["text"])
}

func TestClaimTicket_AsMentor_ShouldAssignTicket(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	moveHackathonToRunningStageForMentors(tc, hackathonID)

	assignMentorRole(tc, hackathonID, mentor)

	ticketID := sendMessage(tc, hackathonID, participant, "Need help")
	time.Sleep(200 * time.Millisecond)

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/claim", hackathonID, ticketID),
		mentor.AccessToken,
		map[string]interface{}{},
	)

	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to claim ticket: %s", string(respBody))

	data := tc.ParseJSON(respBody)
	assert.Equal(t, ticketID, data["ticketId"])
	assert.Equal(t, mentor.UserID, data["assignedMentorUserId"])
	assert.NotEmpty(t, data["assignedAt"])

	var assignedMentorID string
	err := tc.MentorsDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT assigned_mentor_user_id FROM %s.tickets WHERE id = $1", tc.MentorsDBName),
		ticketID,
	).Scan(&assignedMentorID)
	require.NoError(t, err)

	assert.Equal(t, mentor.UserID, assignedMentorID, "Ticket should be assigned to mentor")

	var eventCount int
	err = tc.MentorsDB.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM mentors.outbox_events WHERE aggregate_id = $1 AND event_type = 'ticket.assigned'",
		ticketID,
	).Scan(&eventCount)
	require.NoError(t, err)
	assert.Equal(t, 1, eventCount, "Should create ticket.assigned event")

	// Check system message was created
	var systemMessageCount int
	err = tc.MentorsDB.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM mentors.messages WHERE ticket_id = $1 AND author_role = 'system' AND text = 'Mentor joined the chat'",
		ticketID,
	).Scan(&systemMessageCount)
	require.NoError(t, err)
	assert.Equal(t, 1, systemMessageCount, "Should create system message")
}

func TestClaimTicket_ParallelClaim_ShouldHaveOnlyOneWinner(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	mentor1 := tc.RegisterUser()
	mentor2 := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	moveHackathonToRunningStageForMentors(tc, hackathonID)

	assignMentorRole(tc, hackathonID, mentor1)
	assignMentorRole(tc, hackathonID, mentor2)

	ticketID := sendMessage(tc, hackathonID, participant, "Need help")
	time.Sleep(300 * time.Millisecond)

	results := make(chan *http.Response, 2)

	for _, mentor := range []*UserCredentials{mentor1, mentor2} {
		go func(m *UserCredentials) {
			resp, _ := tc.DoAuthenticatedRequest(
				"POST",
				fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/claim", hackathonID, ticketID),
				m.AccessToken,
				map[string]interface{}{},
			)
			results <- resp
		}(mentor)
	}

	resp1 := <-results
	resp2 := <-results

	successCount := 0
	conflictCount := 0

	for _, resp := range []*http.Response{resp1, resp2} {
		switch resp.StatusCode {
		case http.StatusOK:
			successCount++
		case http.StatusConflict:
			conflictCount++
		}
	}

	assert.Equal(t, 1, successCount, "Exactly one claim should succeed")
	assert.Equal(t, 1, conflictCount, "Exactly one claim should conflict")

	var assignedMentorID string
	err := tc.MentorsDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT assigned_mentor_user_id FROM %s.tickets WHERE id = $1", tc.MentorsDBName),
		ticketID,
	).Scan(&assignedMentorID)
	require.NoError(t, err)

	assert.True(t, assignedMentorID == mentor1.UserID || assignedMentorID == mentor2.UserID,
		"Should be assigned to one of the mentors")
}

func TestClaimTicket_AsParticipant_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStageForMentors(tc, hackathonID)

	ticketID := sendMessage(tc, hackathonID, participant, "Need help")
	time.Sleep(200 * time.Millisecond)

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/claim", hackathonID, ticketID),
		participant.AccessToken,
		map[string]interface{}{},
	)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"Participant should not be able to claim tickets: %s", string(respBody))
}

func TestClaimTicket_AlreadyClosed_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	moveHackathonToRunningStageForMentors(tc, hackathonID)

	assignMentorRole(tc, hackathonID, mentor)

	ticketID := sendMessage(tc, hackathonID, participant, "Need help")
	time.Sleep(200 * time.Millisecond)

	_, err := tc.MentorsDB.Exec(context.Background(),
		fmt.Sprintf("UPDATE %s.tickets SET status = 'closed', closed_at = NOW() WHERE id = $1", tc.MentorsDBName),
		ticketID,
	)
	require.NoError(t, err)

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/claim", hackathonID, ticketID),
		mentor.AccessToken,
		map[string]interface{}{},
	)

	assert.NotEqual(t, http.StatusOK, resp.StatusCode,
		"Should not be able to claim closed ticket: %s", string(respBody))
}

func TestReplyInTicket_UnassignedMentor_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	moveHackathonToRunningStageForMentors(tc, hackathonID)

	assignMentorRole(tc, hackathonID, mentor)

	ticketID := sendMessage(tc, hackathonID, participant, "Need help")
	time.Sleep(200 * time.Millisecond)

	replyBody := map[string]interface{}{
		"text": "Here's my answer",
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/reply", hackathonID, ticketID),
		mentor.AccessToken,
		replyBody,
	)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"Unassigned mentor should not be able to reply: %s", string(respBody))
}

func TestReplyInTicket_AssignedMentor_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	moveHackathonToRunningStageForMentors(tc, hackathonID)

	assignMentorRole(tc, hackathonID, mentor)

	ticketID := sendMessage(tc, hackathonID, participant, "Need help")
	time.Sleep(200 * time.Millisecond)

	claimTicket(tc, hackathonID, ticketID, mentor)
	time.Sleep(200 * time.Millisecond)

	replyBody := map[string]interface{}{
		"text": "Here's the solution",
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/reply", hackathonID, ticketID),
		mentor.AccessToken,
		replyBody,
	)

	require.Equal(t, http.StatusOK, resp.StatusCode, "Assigned mentor should be able to reply: %s", string(respBody))

	data := tc.ParseJSON(respBody)
	messageID := data["messageId"].(string)
	assert.NotEmpty(t, messageID)

	var authorRole string
	err := tc.MentorsDB.QueryRow(context.Background(),
		"SELECT author_role FROM mentors.messages WHERE id = $1",
		messageID,
	).Scan(&authorRole)
	require.NoError(t, err)

	assert.Equal(t, "mentor", authorRole, "Author role should be mentor")
}

func TestReplyInTicket_AsOrganizer_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStageForMentors(tc, hackathonID)

	ticketID := sendMessage(tc, hackathonID, participant, "Need help")
	time.Sleep(200 * time.Millisecond)

	replyBody := map[string]interface{}{
		"text": "Organizer trying to reply",
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/reply", hackathonID, ticketID),
		owner.AccessToken,
		replyBody,
	)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"Organizer should not be able to reply (read-only access): %s", string(respBody))
}

func TestReplyInTicket_AsParticipant_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	moveHackathonToRunningStageForMentors(tc, hackathonID)

	assignMentorRole(tc, hackathonID, mentor)

	ticketID := sendMessage(tc, hackathonID, participant, "Need help")
	time.Sleep(200 * time.Millisecond)

	claimTicket(tc, hackathonID, ticketID, mentor)
	time.Sleep(200 * time.Millisecond)

	replyBody := map[string]interface{}{
		"text": "Participant trying to reply via ReplyInTicket",
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/reply", hackathonID, ticketID),
		participant.AccessToken,
		replyBody,
	)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"Participant should not use ReplyInTicket, they should use SendMessage: %s", string(respBody))
}

func TestReplyInTicket_WithClientMessageID_ShouldBeIdempotent(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	moveHackathonToRunningStageForMentors(tc, hackathonID)

	assignMentorRole(tc, hackathonID, mentor)

	ticketID := sendMessage(tc, hackathonID, participant, "Need help")
	time.Sleep(200 * time.Millisecond)

	claimTicket(tc, hackathonID, ticketID, mentor)
	time.Sleep(200 * time.Millisecond)

	idempotencyKey := uuid.New().String()
	replyBody := map[string]interface{}{
		"text": "Idempotent reply",
		"idempotencyKey": map[string]interface{}{
			"key": idempotencyKey,
		},
	}

	resp1, respBody1 := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/reply", hackathonID, ticketID),
		mentor.AccessToken,
		replyBody,
	)
	require.Equal(t, http.StatusOK, resp1.StatusCode)

	data1 := tc.ParseJSON(respBody1)
	messageID1 := data1["messageId"].(string)

	time.Sleep(200 * time.Millisecond)

	resp2, respBody2 := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/reply", hackathonID, ticketID),
		mentor.AccessToken,
		replyBody,
	)
	require.Equal(t, http.StatusOK, resp2.StatusCode)

	data2 := tc.ParseJSON(respBody2)
	messageID2 := data2["messageId"].(string)

	assert.Equal(t, messageID1, messageID2, "Should return same message ID")

	var messageCount int
	err := tc.MentorsDB.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM mentors.messages WHERE id = $1",
		messageID1,
	).Scan(&messageCount)
	require.NoError(t, err)

	assert.Equal(t, 1, messageCount, "Should have exactly 1 message")
}

func TestCloseTicket_AsAssignedMentor_ShouldCloseAndCreateSystemMessage(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	moveHackathonToRunningStageForMentors(tc, hackathonID)

	assignMentorRole(tc, hackathonID, mentor)

	ticketID := sendMessage(tc, hackathonID, participant, "Need help")
	time.Sleep(200 * time.Millisecond)

	claimTicket(tc, hackathonID, ticketID, mentor)
	time.Sleep(200 * time.Millisecond)

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/close", hackathonID, ticketID),
		mentor.AccessToken,
		map[string]interface{}{},
	)

	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to close ticket: %s", string(respBody))

	var status string
	var closedAt *time.Time
	err := tc.MentorsDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT status, closed_at FROM %s.tickets WHERE id = $1", tc.MentorsDBName),
		ticketID,
	).Scan(&status, &closedAt)
	require.NoError(t, err)

	assert.Equal(t, "closed", status)
	assert.NotNil(t, closedAt, "closed_at should be set")

	// Check system message for "Mentor joined the chat" (from ClaimTicket)
	var mentorJoinedCount int
	err = tc.MentorsDB.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM mentors.messages WHERE ticket_id = $1 AND author_role = 'system' AND text = 'Mentor joined the chat'",
		ticketID,
	).Scan(&mentorJoinedCount)
	require.NoError(t, err)
	assert.Equal(t, 1, mentorJoinedCount, "Should create 'Mentor joined' system message")

	// Check system message for "Ticket closed" (from CloseTicket)
	var ticketClosedMsgCount int
	err = tc.MentorsDB.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM mentors.messages WHERE ticket_id = $1 AND author_role = 'system' AND text = 'Ticket closed'",
		ticketID,
	).Scan(&ticketClosedMsgCount)
	require.NoError(t, err)
	assert.Equal(t, 1, ticketClosedMsgCount, "Should create 'Ticket closed' system message")

	// Verify system messages via API have empty author_user_id
	messagesResp := getTicketMessages(tc, hackathonID, ticketID, mentor, 10, 0)
	require.NotNil(t, messagesResp)
	require.NotNil(t, messagesResp.Messages)

	var systemMessageFound bool
	for _, msg := range messagesResp.Messages {
		if msg.AuthorRole == "AUTHOR_ROLE_SYSTEM" && msg.Text == "Ticket closed" {
			systemMessageFound = true
			assert.Equal(t, "", msg.AuthorUserId, "System message should have empty author_user_id in API response")
			break
		}
	}
	assert.True(t, systemMessageFound, "Should find 'Ticket closed' system message in API response")

	var ticketClosedCount int
	err = tc.MentorsDB.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM mentors.outbox_events WHERE aggregate_id = $1 AND event_type = 'ticket.closed'",
		ticketID,
	).Scan(&ticketClosedCount)
	require.NoError(t, err)

	assert.Equal(t, 1, ticketClosedCount, "Should create ticket.closed event")
}

func TestCloseTicket_AsUnassignedMentor_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	moveHackathonToRunningStageForMentors(tc, hackathonID)

	assignMentorRole(tc, hackathonID, mentor)

	ticketID := sendMessage(tc, hackathonID, participant, "Need help")
	time.Sleep(200 * time.Millisecond)

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/close", hackathonID, ticketID),
		mentor.AccessToken,
		map[string]interface{}{},
	)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"Unassigned mentor should not be able to close ticket: %s", string(respBody))
}

func TestCloseTicket_AsOrganizer_ShouldFail(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStageForMentors(tc, hackathonID)

	ticketID := sendMessage(tc, hackathonID, participant, "Need help")
	time.Sleep(200 * time.Millisecond)

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/close", hackathonID, ticketID),
		owner.AccessToken,
		map[string]interface{}{},
	)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"Organizer should not be able to close ticket without being assigned: %s", string(respBody))
}

func TestCloseTicket_Idempotent_ShouldReturnSuccessOnSecondCall(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	moveHackathonToRunningStageForMentors(tc, hackathonID)

	assignMentorRole(tc, hackathonID, mentor)

	ticketID := sendMessage(tc, hackathonID, participant, "Need help")
	time.Sleep(200 * time.Millisecond)

	claimTicket(tc, hackathonID, ticketID, mentor)
	time.Sleep(200 * time.Millisecond)

	resp1, _ := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/close", hackathonID, ticketID),
		mentor.AccessToken,
		map[string]interface{}{},
	)
	require.Equal(t, http.StatusOK, resp1.StatusCode)

	time.Sleep(200 * time.Millisecond)

	resp2, respBody2 := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/close", hackathonID, ticketID),
		mentor.AccessToken,
		map[string]interface{}{},
	)

	assert.Equal(t, http.StatusOK, resp2.StatusCode,
		"Second close should be idempotent: %s", string(respBody2))

	var systemMessageCount int
	err := tc.MentorsDB.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM mentors.messages WHERE ticket_id = $1 AND author_role = 'system' AND text = 'Ticket closed'",
		ticketID,
	).Scan(&systemMessageCount)
	require.NoError(t, err)

	assert.Equal(t, 1, systemMessageCount, "Should have exactly 1 'Ticket closed' system message")
}

func TestSystemMessages_ClaimTicket_ShouldHaveNullAuthorUserId(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	moveHackathonToRunningStageForMentors(tc, hackathonID)

	assignMentorRole(tc, hackathonID, mentor)

	ticketID := sendMessage(tc, hackathonID, participant, "Need help")
	time.Sleep(200 * time.Millisecond)

	claimTicket(tc, hackathonID, ticketID, mentor)
	time.Sleep(200 * time.Millisecond)

	messagesResp := getTicketMessages(tc, hackathonID, ticketID, mentor, 10, 0)
	require.NotNil(t, messagesResp)
	require.NotNil(t, messagesResp.Messages)

	var systemMessageFound bool
	for _, msg := range messagesResp.Messages {
		if msg.AuthorRole == "AUTHOR_ROLE_SYSTEM" && msg.Text == "Mentor joined the chat" {
			systemMessageFound = true
			assert.Equal(t, "", msg.AuthorUserId, "System message should have empty author_user_id in API response")
			break
		}
	}
	assert.True(t, systemMessageFound, "Should find 'Mentor joined the chat' system message in API response")
}

func TestSystemMessages_AssignTicket_ShouldCreateSystemMessage(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	moveHackathonToRunningStageForMentors(tc, hackathonID)

	assignMentorRole(tc, hackathonID, mentor)

	ticketID := sendMessage(tc, hackathonID, participant, "Need help")
	time.Sleep(200 * time.Millisecond)

	assignBody := map[string]interface{}{
		"mentor_user_id": mentor.UserID,
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/assign", hackathonID, ticketID),
		owner.AccessToken,
		assignBody,
	)

	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to assign ticket: %s", string(respBody))

	time.Sleep(200 * time.Millisecond)

	messagesResp := getTicketMessages(tc, hackathonID, ticketID, mentor, 10, 0)
	require.NotNil(t, messagesResp)
	require.NotNil(t, messagesResp.Messages)

	var systemMessageFound bool
	for _, msg := range messagesResp.Messages {
		if msg.AuthorRole == "AUTHOR_ROLE_SYSTEM" && msg.Text == "Mentor joined the chat" {
			systemMessageFound = true
			assert.Equal(t, "", msg.AuthorUserId, "System message should have empty author_user_id in API response")
			break
		}
	}
	assert.True(t, systemMessageFound, "Should find 'Mentor joined the chat' system message in API response")
}

func TestSendMessage_AfterClose_ShouldCreateNewTicket(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	moveHackathonToRunningStageForMentors(tc, hackathonID)

	assignMentorRole(tc, hackathonID, mentor)

	firstTicketID := sendMessage(tc, hackathonID, participant, "First issue")
	time.Sleep(200 * time.Millisecond)

	claimTicket(tc, hackathonID, firstTicketID, mentor)
	time.Sleep(200 * time.Millisecond)

	closeTicket(tc, hackathonID, firstTicketID, mentor)
	time.Sleep(200 * time.Millisecond)

	secondTicketID := sendMessage(tc, hackathonID, participant, "Second issue")
	time.Sleep(200 * time.Millisecond)

	assert.NotEqual(t, firstTicketID, secondTicketID, "Should create new ticket")

	var ticketCount int
	err := tc.MentorsDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT COUNT(*) FROM %s.tickets WHERE hackathon_id = $1 AND owner_id = $2", tc.MentorsDBName),
		hackathonID, participant.UserID,
	).Scan(&ticketCount)
	require.NoError(t, err)

	assert.Equal(t, 2, ticketCount, "Should have 2 tickets total")

	var firstStatus, secondStatus string
	err = tc.MentorsDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT status FROM %s.tickets WHERE id = $1", tc.MentorsDBName),
		firstTicketID,
	).Scan(&firstStatus)
	require.NoError(t, err)

	err = tc.MentorsDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT status FROM %s.tickets WHERE id = $1", tc.MentorsDBName),
		secondTicketID,
	).Scan(&secondStatus)
	require.NoError(t, err)

	assert.Equal(t, "closed", firstStatus)
	assert.Equal(t, "open", secondStatus)
}

func TestListAssignedTickets_AsMentor_ShouldReturnOnlyAssignedTickets(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant1 := tc.RegisterUser()
	participant2 := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant1, "PART_INDIVIDUAL")
	registerParticipant(tc, hackathonID, participant2, "PART_INDIVIDUAL")
	moveHackathonToRunningStageForMentors(tc, hackathonID)

	assignMentorRole(tc, hackathonID, mentor)

	ticket1ID := sendMessage(tc, hackathonID, participant1, "Ticket 1")
	_ = sendMessage(tc, hackathonID, participant2, "Ticket 2")
	time.Sleep(300 * time.Millisecond)

	claimTicket(tc, hackathonID, ticket1ID, mentor)
	time.Sleep(200 * time.Millisecond)

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/assigned/list", hackathonID),
		mentor.AccessToken,
		map[string]interface{}{
			"query": map[string]interface{}{
				"limit":  10,
				"offset": 0,
			},
		},
	)

	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get assigned tickets: %s", string(respBody))

	data := tc.ParseJSON(respBody)
	tickets := data["tickets"].([]interface{})

	assert.Equal(t, 1, len(tickets), "Should have exactly 1 assigned ticket")

	ticket := tickets[0].(map[string]interface{})
	assert.Equal(t, ticket1ID, ticket["ticketId"])
	assert.Equal(t, mentor.UserID, ticket["assignedMentorUserId"])
}

func TestListAllTickets_AsOrganizer_ShouldReturnAllTickets(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant1 := tc.RegisterUser()
	participant2 := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant1, "PART_INDIVIDUAL")
	registerParticipant(tc, hackathonID, participant2, "PART_INDIVIDUAL")
	moveHackathonToRunningStageForMentors(tc, hackathonID)

	assignMentorRole(tc, hackathonID, mentor)

	ticket1ID := sendMessage(tc, hackathonID, participant1, "Ticket 1")
	ticket2ID := sendMessage(tc, hackathonID, participant2, "Ticket 2")
	time.Sleep(300 * time.Millisecond)

	claimTicket(tc, hackathonID, ticket1ID, mentor)
	time.Sleep(200 * time.Millisecond)

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/all/list", hackathonID),
		owner.AccessToken,
		map[string]interface{}{
			"query": map[string]interface{}{
				"limit":  10,
				"offset": 0,
			},
		},
	)

	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get all tickets: %s", string(respBody))

	data := tc.ParseJSON(respBody)
	tickets := data["tickets"].([]interface{})

	assert.GreaterOrEqual(t, len(tickets), 2, "Should have at least 2 tickets")

	foundAssigned := false
	foundUnassigned := false

	for _, ticketData := range tickets {
		ticket := ticketData.(map[string]interface{})
		ticketID := ticket["ticketId"].(string)

		if ticketID == ticket1ID {
			assert.Equal(t, mentor.UserID, ticket["assignedMentorUserId"], "Ticket 1 should be assigned")
			foundAssigned = true
		}
		if ticketID == ticket2ID {
			assignedMentor := ticket["assignedMentorUserId"]
			assert.True(t, assignedMentor == nil || assignedMentor == "", "Ticket 2 should be unassigned")
			foundUnassigned = true
		}
	}

	assert.True(t, foundAssigned, "Should include assigned tickets")
	assert.True(t, foundUnassigned, "Should include unassigned tickets")
}

func TestListAllTickets_AsMentor_ShouldSucceed(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()
	mentor := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")
	moveHackathonToRunningStageForMentors(tc, hackathonID)

	assignMentorRole(tc, hackathonID, mentor)

	ticketID := sendMessage(tc, hackathonID, participant, "Test ticket")
	time.Sleep(200 * time.Millisecond)

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/all/list", hackathonID),
		mentor.AccessToken,
		map[string]interface{}{
			"query": map[string]interface{}{
				"limit":  10,
				"offset": 0,
			},
		},
	)

	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Mentor should be able to list all tickets: %s", string(respBody))

	data := tc.ParseJSON(respBody)
	tickets := data["tickets"].([]interface{})
	assert.GreaterOrEqual(t, len(tickets), 1, "Should have at least 1 ticket")

	found := false
	for _, ticketData := range tickets {
		ticket := ticketData.(map[string]interface{})
		if ticket["ticketId"].(string) == ticketID {
			found = true
			break
		}
	}
	assert.True(t, found, "Should find the created ticket")
}

func TestInvariant_OnlyOneOpenTicketPerOwner(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStageForMentors(tc, hackathonID)

	for i := 1; i <= 3; i++ {
		sendMessage(tc, hackathonID, participant, fmt.Sprintf("Message %d", i))
		time.Sleep(100 * time.Millisecond)
	}

	var openTicketCount int
	err := tc.MentorsDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT COUNT(*) FROM %s.tickets WHERE hackathon_id = $1 AND owner_id = $2 AND status = 'open'", tc.MentorsDBName),
		hackathonID, participant.UserID,
	).Scan(&openTicketCount)
	require.NoError(t, err)

	assert.Equal(t, 1, openTicketCount, "Should have exactly 1 OPEN ticket despite 3 messages")

	var messageCount int
	err = tc.MentorsDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT COUNT(*) FROM %s.messages m JOIN %s.tickets t ON m.ticket_id = t.id WHERE t.hackathon_id = $1 AND t.owner_id = $2", tc.MentorsDBName, tc.MentorsDBName),
		hackathonID, participant.UserID,
	).Scan(&messageCount)
	require.NoError(t, err)

	assert.Equal(t, 3, messageCount, "Should have 3 messages")
}

func TestSendMessage_ParallelFirstMessages_ShouldCreateOnlyOneTicket(t *testing.T) {
	tc := NewTestContext(t)
	owner := tc.RegisterUser()
	participant := tc.RegisterUser()

	hackathonID := createHackathonInRunning(tc, owner)
	registerParticipant(tc, hackathonID, participant, "PART_INDIVIDUAL")

	moveHackathonToRunningStageForMentors(tc, hackathonID)

	results := make(chan string, 5)

	for i := 1; i <= 5; i++ {
		go func(index int) {
			body := map[string]interface{}{
				"text": fmt.Sprintf("Parallel message %d", index),
			}
			resp, respBody := tc.DoAuthenticatedRequest(
				"POST",
				fmt.Sprintf("/v1/hackathons/%s/support/messages", hackathonID),
				participant.AccessToken,
				body,
			)
			if resp.StatusCode == http.StatusOK {
				data := tc.ParseJSON(respBody)
				results <- data["ticketId"].(string)
			} else {
				results <- ""
			}
		}(i)
	}

	ticketIDs := make(map[string]int)
	for i := 0; i < 5; i++ {
		ticketID := <-results
		if ticketID != "" {
			ticketIDs[ticketID]++
		}
	}

	assert.Equal(t, 1, len(ticketIDs), "Should create exactly 1 ticket despite parallel requests")

	var openTicketCount int
	err := tc.MentorsDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT COUNT(*) FROM %s.tickets WHERE hackathon_id = $1 AND owner_id = $2 AND status = 'open'", tc.MentorsDBName),
		hackathonID, participant.UserID,
	).Scan(&openTicketCount)
	require.NoError(t, err)

	assert.Equal(t, 1, openTicketCount, "DB should have exactly 1 OPEN ticket")

	var messageCount int
	err = tc.MentorsDB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT COUNT(*) FROM %s.messages m JOIN %s.tickets t ON m.ticket_id = t.id WHERE t.hackathon_id = $1 AND t.owner_id = $2", tc.MentorsDBName, tc.MentorsDBName),
		hackathonID, participant.UserID,
	).Scan(&messageCount)
	require.NoError(t, err)

	assert.Equal(t, 5, messageCount, "Should have 5 messages")
}

// sendMessage sends a message and returns ticket ID
func sendMessage(tc *TestContext, hackathonID string, user *UserCredentials, text string) string {
	body := map[string]interface{}{
		"text": text,
	}

	resp, respBody := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/messages", hackathonID),
		user.AccessToken,
		body,
	)

	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to send message: %s", string(respBody))

	data := tc.ParseJSON(respBody)
	return data["ticketId"].(string)
}

// claimTicket claims a ticket as a mentor
func claimTicket(tc *TestContext, hackathonID, ticketID string, mentor *UserCredentials) {
	resp, body := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/claim", hackathonID, ticketID),
		mentor.AccessToken,
		map[string]interface{}{},
	)

	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to claim ticket: %s", string(body))
}

// replyInTicket sends a reply to a ticket as a mentor
func replyInTicket(tc *TestContext, hackathonID, ticketID string, user *UserCredentials, text string) {
	replyBody := map[string]interface{}{
		"text": text,
	}

	resp, body := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/reply", hackathonID, ticketID),
		user.AccessToken,
		replyBody,
	)

	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to reply in ticket: %s", string(body))
}

// closeTicket closes a ticket
func closeTicket(tc *TestContext, hackathonID, ticketID string, user *UserCredentials) {
	resp, body := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/close", hackathonID, ticketID),
		user.AccessToken,
		map[string]interface{}{},
	)

	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to close ticket: %s", string(body))
}

type GetTicketMessagesResponse struct {
	Messages []struct {
		MessageId    string `json:"messageId"`
		TicketId     string `json:"ticketId"`
		AuthorUserId string `json:"authorUserId"`
		AuthorRole   string `json:"authorRole"`
		Text         string `json:"text"`
		CreatedAt    string `json:"createdAt"`
	} `json:"messages"`
	HasMore bool `json:"hasMore"`
}

func getTicketMessages(tc *TestContext, hackathonID, ticketID string, user *UserCredentials, limit, offset int) *GetTicketMessagesResponse {
	resp, body := tc.DoAuthenticatedRequest(
		"GET",
		fmt.Sprintf("/v1/hackathons/%s/support/tickets/%s/messages?query.limit=%d&query.offset=%d", hackathonID, ticketID, limit, offset),
		user.AccessToken,
		nil,
	)

	require.Equal(tc.T, http.StatusOK, resp.StatusCode, "Failed to get ticket messages: %s", string(body))

	var result GetTicketMessagesResponse
	err := json.Unmarshal(body, &result)
	require.NoError(tc.T, err)

	return &result
}

// inviteAndAccept invites a user to team and accepts the invitation
func inviteAndAccept(tc *TestContext, hackathonID, teamID string, captain, invitee *UserCredentials, vacancyID string) {
	inviteBody := map[string]interface{}{
		"target_user_id": invitee.UserID,
		"vacancy_id":     vacancyID,
		"message":        "Join us!",
	}

	resp, body := tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/hackathons/%s/teams/%s/team-invitations", hackathonID, teamID),
		captain.AccessToken,
		inviteBody,
	)
	require.Equal(tc.T, http.StatusOK, resp.StatusCode)

	inviteData := tc.ParseJSON(body)
	invitationID := inviteData["invitationId"].(string)

	tc.DoAuthenticatedRequest(
		"POST",
		fmt.Sprintf("/v1/users/me/team-invitations/%s/accept", invitationID),
		invitee.AccessToken,
		map[string]interface{}{},
	)

	time.Sleep(500 * time.Millisecond)
}

// assignMentorRole assigns mentor role to a user in a hackathon
func assignMentorRole(tc *TestContext, hackathonID string, user *UserCredentials) {
	// For testing, we assign mentor role via DB directly
	// In production, this would be done through staff API by organizer

	_, err := tc.ParticipationDB.Exec(context.Background(),
		fmt.Sprintf("INSERT INTO %s (hackathon_id, user_id, role) VALUES ($1, $2, 'mentor') ON CONFLICT DO NOTHING",
			tc.ParticipationDBName),
		hackathonID, user.UserID,
	)
	require.NoError(tc.T, err, "Failed to assign mentor role")

	time.Sleep(300 * time.Millisecond)
}

func transitionToRunning(tc *TestContext, hackathonID string, owner *UserCredentials) {
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
		now.Add(-2*24*time.Hour),  // registration closed 2 days ago
		now.Add(-1*24*time.Hour),  // started 1 day ago
		now.Add(5*24*time.Hour),   // ends in 5 days
		hackathonID)
	require.NoError(tc.T, err, "Failed to update hackathon to RUNNING stage")
}

func createHackathonInRunning(tc *TestContext, owner *UserCredentials) string {
	now := time.Now()
	hackathonBody := map[string]interface{}{
		"name":              "Mentors Test Hackathon",
		"short_description": "Test hackathon for mentors service",
		"description":       "Full description for mentors testing",
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

	// First, put hackathon in REGISTRATION stage by setting registration_opens_at to past
	_, err := tc.DB.Exec(context.Background(), fmt.Sprintf(`
		UPDATE %s 
		SET registration_opens_at = $1
		WHERE id = $2
	`, tc.HackathonDBName), now.Add(-10*24*time.Hour), hackathonID)
	require.NoError(tc.T, err, "Failed to update hackathon dates in DB")

	return hackathonID
}

func moveHackathonToRunningStageForMentors(tc *TestContext, hackathonID string) {
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
		now.Add(-2*24*time.Hour),  // registration closed 2 days ago
		now.Add(-1*24*time.Hour),  // started 1 day ago
		now.Add(5*24*time.Hour),   // ends in 5 days
		hackathonID)
	require.NoError(tc.T, err, "Failed to move hackathon to RUNNING stage")
}
