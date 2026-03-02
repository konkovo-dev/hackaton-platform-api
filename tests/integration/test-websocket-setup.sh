#!/bin/bash
set -e

echo "🚀 Setting up WebSocket test environment..."
echo ""

# Database connection (same as in tests)
# Local: postgres://hackathon:hackathon_dev_password@localhost:5432/hackathon?sslmode=disable
# Prod: postgres://hackathon:FBTZYAXcQ0gxfj5Y8b7gd3jRKzwq8jG@178.154.192.57:5432/hackathon_hackaton?sslmode=disable
DB_DSN="${DB_DSN:-postgres://hackathon:hackathon_dev_password@localhost:5432/hackathon?sslmode=disable}"

# API base URL
API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"

# WebSocket URL
WS_URL="${WS_URL:-ws://localhost:8000}"

# Generate unique identifiers
TIMESTAMP=$(date +%s)
RANDOM_SUFFIX=$(openssl rand -hex 4)

# ============================================================================
# Step 1: Register users
# ============================================================================
echo "📝 Step 1: Registering users..."

# Participant
PARTICIPANT_EMAIL="participant_${TIMESTAMP}_${RANDOM_SUFFIX}@test.com"
PARTICIPANT_USERNAME="participant_${RANDOM_SUFFIX}"

curl -s -X POST $API_BASE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$PARTICIPANT_USERNAME\",\"email\":\"$PARTICIPANT_EMAIL\",\"password\":\"Pass123!\",\"first_name\":\"Part\",\"last_name\":\"User\",\"timezone\":\"UTC\"}" > /dev/null

PARTICIPANT_TOKEN=$(curl -s -X POST $API_BASE_URL/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$PARTICIPANT_EMAIL\",\"password\":\"Pass123!\"}" | jq -r '.accessToken')

PARTICIPANT_ID=$(curl -s -X POST $API_BASE_URL/v1/auth/introspect \
  -H "Content-Type: application/json" \
  -d "{\"access_token\":\"$PARTICIPANT_TOKEN\"}" | jq -r '.userId')

echo "✅ Participant: $PARTICIPANT_ID"

# Mentor
MENTOR_EMAIL="mentor_${TIMESTAMP}_${RANDOM_SUFFIX}@test.com"
MENTOR_USERNAME="mentor_${RANDOM_SUFFIX}"

curl -s -X POST $API_BASE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$MENTOR_USERNAME\",\"email\":\"$MENTOR_EMAIL\",\"password\":\"Pass123!\",\"first_name\":\"Mentor\",\"last_name\":\"User\",\"timezone\":\"UTC\"}" > /dev/null

MENTOR_TOKEN=$(curl -s -X POST $API_BASE_URL/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$MENTOR_EMAIL\",\"password\":\"Pass123!\"}" | jq -r '.accessToken')

MENTOR_ID=$(curl -s -X POST $API_BASE_URL/v1/auth/introspect \
  -H "Content-Type: application/json" \
  -d "{\"access_token\":\"$MENTOR_TOKEN\"}" | jq -r '.userId')

echo "✅ Mentor: $MENTOR_ID"

# Owner/Organizer
OWNER_EMAIL="owner_${TIMESTAMP}_${RANDOM_SUFFIX}@test.com"
OWNER_USERNAME="owner_${RANDOM_SUFFIX}"

curl -s -X POST $API_BASE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$OWNER_USERNAME\",\"email\":\"$OWNER_EMAIL\",\"password\":\"Pass123!\",\"first_name\":\"Owner\",\"last_name\":\"User\",\"timezone\":\"UTC\"}" > /dev/null

OWNER_TOKEN=$(curl -s -X POST $API_BASE_URL/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$OWNER_EMAIL\",\"password\":\"Pass123!\"}" | jq -r '.accessToken')

OWNER_ID=$(curl -s -X POST $API_BASE_URL/v1/auth/introspect \
  -H "Content-Type: application/json" \
  -d "{\"access_token\":\"$OWNER_TOKEN\"}" | jq -r '.userId')

echo "✅ Owner: $OWNER_ID"
echo ""

# ============================================================================
# Step 2: Create hackathon
# ============================================================================
echo "🏆 Step 2: Creating hackathon..."

NOW_RFC3339=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
REG_OPENS=$(date -u -v+1H +"%Y-%m-%dT%H:%M:%SZ")
REG_CLOSES=$(date -u -v+15d +"%Y-%m-%dT%H:%M:%SZ")
STARTS=$(date -u -v+20d +"%Y-%m-%dT%H:%M:%SZ")
ENDS=$(date -u -v+22d +"%Y-%m-%dT%H:%M:%SZ")
JUDGING_ENDS=$(date -u -v+25d +"%Y-%m-%dT%H:%M:%SZ")

HACKATHON_ID=$(curl -s -X POST $API_BASE_URL/v1/hackathons \
  -H "Authorization: Bearer $OWNER_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\":\"WebSocket Test Hackathon\",
    \"short_description\":\"Testing real-time support\",
    \"description\":\"Full description for testing\",
    \"location\":{\"online\":true},
    \"dates\":{
      \"registration_opens_at\":\"$REG_OPENS\",
      \"registration_closes_at\":\"$REG_CLOSES\",
      \"starts_at\":\"$STARTS\",
      \"ends_at\":\"$ENDS\",
      \"judging_ends_at\":\"$JUDGING_ENDS\"
    },
    \"registration_policy\":{\"allow_individual\":true,\"allow_team\":true},
    \"limits\":{\"team_size_max\":5}
  }" | jq -r '.hackathonId')

echo "✅ Hackathon created: $HACKATHON_ID"

# Wait for owner role to be assigned (async event)
echo "⏳ Waiting for owner role assignment..."
sleep 3

# Add task
curl -s -X PUT $API_BASE_URL/v1/hackathons/$HACKATHON_ID/task \
  -H "Authorization: Bearer $OWNER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"task":"Build something innovative"}' > /dev/null

echo "✅ Task added"

# Publish hackathon
curl -s -X POST $API_BASE_URL/v1/hackathons/$HACKATHON_ID/publish \
  -H "Authorization: Bearer $OWNER_TOKEN" \
  -d '{}' > /dev/null

echo "✅ Hackathon published"
sleep 2

# ============================================================================
# Step 3: Invite mentor (via staff invitation API)
# ============================================================================
echo "🎓 Step 3: Inviting mentor..."

MENTOR_INVITATION_ID=$(curl -s -X POST $API_BASE_URL/v1/hackathons/$HACKATHON_ID/staff-invitations \
  -H "Authorization: Bearer $OWNER_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"target_user_id\":\"$MENTOR_ID\",
    \"requested_role\":\"HX_ROLE_MENTOR\",
    \"message\":\"We would love to have you as a mentor!\"
  }" | jq -r '.invitationId')

echo "✅ Mentor invitation created: $MENTOR_INVITATION_ID"

# Mentor accepts invitation
curl -s -X POST $API_BASE_URL/v1/users/me/staff-invitations/$MENTOR_INVITATION_ID/accept \
  -H "Authorization: Bearer $MENTOR_TOKEN" \
  -d '{}' > /dev/null

echo "✅ Mentor accepted invitation"
sleep 2

# ============================================================================
# Step 4: Transition to REGISTRATION stage (via DB, like in tests)
# ============================================================================
echo "⏳ Step 4: Transitioning to REGISTRATION stage..."

REG_OPENED=$(date -u -v-1d +"%Y-%m-%d %H:%M:%S")

psql "$DB_DSN" <<EOF
UPDATE hackathon.hackathons 
SET registration_opens_at = '$REG_OPENED',
    stage = 'registration'
WHERE id = '$HACKATHON_ID';
EOF

sleep 1
echo "✅ Hackathon in REGISTRATION stage"
echo ""

# ============================================================================
# Step 5: Register participant
# ============================================================================
echo "👤 Step 5: Registering participant..."

curl -s -X POST $API_BASE_URL/v1/hackathons/$HACKATHON_ID/participations/register \
  -H "Authorization: Bearer $PARTICIPANT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"desired_status":"PART_INDIVIDUAL","motivation_text":"Testing support system"}' > /dev/null

echo "✅ Participant registered"
sleep 2

# ============================================================================
# Step 6: Transition to RUNNING stage (via DB, like in tests)
# ============================================================================
echo "⏳ Step 6: Transitioning to RUNNING stage..."

STARTED=$(date -u -v-1H +"%Y-%m-%d %H:%M:%S")

psql "$DB_DSN" <<EOF
UPDATE hackathon.hackathons 
SET starts_at = '$STARTED',
    stage = 'running'
WHERE id = '$HACKATHON_ID';
EOF

sleep 1
echo "✅ Hackathon in RUNNING stage"
echo ""

# ============================================================================
# Step 7: Get WebSocket tokens
# ============================================================================
echo "🔑 Step 7: Getting WebSocket tokens..."

PARTICIPANT_WS_TOKEN=$(curl -s $API_BASE_URL/v1/hackathons/$HACKATHON_ID/support/realtime-token \
  -H "Authorization: Bearer $PARTICIPANT_TOKEN" | jq -r '.token')

MENTOR_WS_TOKEN=$(curl -s $API_BASE_URL/v1/hackathons/$HACKATHON_ID/support/realtime-token \
  -H "Authorization: Bearer $MENTOR_TOKEN" | jq -r '.token')

OWNER_WS_TOKEN=$(curl -s $API_BASE_URL/v1/hackathons/$HACKATHON_ID/support/realtime-token \
  -H "Authorization: Bearer $OWNER_TOKEN" | jq -r '.token')

echo "✅ WebSocket tokens obtained"
echo ""

# ============================================================================
# Summary
# ============================================================================
echo "=========================================="
echo "✨ Setup complete!"
echo "=========================================="
echo ""
echo "HACKATHON_ID=$HACKATHON_ID"
echo "PARTICIPANT_ID=$PARTICIPANT_ID"
echo "MENTOR_ID=$MENTOR_ID"
echo "OWNER_ID=$OWNER_ID"
echo ""
echo "=========================================="
echo "🌐 Open 3 terminals and run:"
echo "=========================================="
echo ""
echo "# Terminal 1 (Participant):"
echo "websocat '$WS_URL/connection/websocket?cf_ws_frame_ping_pong=true'"
echo ""
echo "# Then paste these lines one by one:"
echo "{\"id\":1,\"connect\":{\"token\":\"$PARTICIPANT_WS_TOKEN\"}}"
echo "{\"id\":2,\"subscribe\":{\"channel\":\"support:feed#$PARTICIPANT_ID\"}}"
echo ""
echo "=========================================="
echo ""
echo "# Terminal 2 (Mentor):"
echo "websocat '$WS_URL/connection/websocket?cf_ws_frame_ping_pong=true'"
echo ""
echo "# Then paste these lines one by one:"
echo "{\"id\":1,\"connect\":{\"token\":\"$MENTOR_WS_TOKEN\"}}"
echo "{\"id\":2,\"subscribe\":{\"channel\":\"support:feed#$MENTOR_ID\"}}"
echo ""
echo "=========================================="
echo ""
echo "# Terminal 3 (Owner):"
echo "websocat '$WS_URL/connection/websocket?cf_ws_frame_ping_pong=true'"
echo ""
echo "# Then paste these lines one by one:"
echo "{\"id\":1,\"connect\":{\"token\":\"$OWNER_WS_TOKEN\"}}"
echo "{\"id\":2,\"subscribe\":{\"channel\":\"support:feed#$OWNER_ID\"}}"
echo ""
echo "=========================================="
echo "📨 HTTP commands to test:"
echo "=========================================="
echo ""
echo "# 1. Participant sends first message (creates ticket automatically):"
echo "TICKET_ID=\$(curl -s -X POST $API_BASE_URL/v1/hackathons/$HACKATHON_ID/support/messages -H \"Authorization: Bearer $PARTICIPANT_TOKEN\" -H \"Content-Type: application/json\" -d '{\"text\":\"Помогите с API!\"}' | jq -r '.ticketId')"
echo ""
echo "echo \"Ticket ID: \$TICKET_ID\""
echo ""
echo "# 2. Participant views their chat (all messages from all tickets):"
echo "curl -s \"$API_BASE_URL/v1/hackathons/$HACKATHON_ID/support/my-messages?query.limit=50&query.offset=0\" -H \"Authorization: Bearer $PARTICIPANT_TOKEN\" | jq"
echo ""
echo "# 3. Mentor lists all tickets (to see new ticket):"
echo "curl -s -X POST $API_BASE_URL/v1/hackathons/$HACKATHON_ID/support/tickets/all/list -H \"Authorization: Bearer $MENTOR_TOKEN\" -H \"Content-Type: application/json\" -d '{\"limit\":10,\"offset\":0}' | jq"
echo ""
echo "# 4. Mentor claims ticket (system message: 'Mentor joined the chat'):"
echo "curl -s -X POST $API_BASE_URL/v1/hackathons/$HACKATHON_ID/support/tickets/\$TICKET_ID/claim -H \"Authorization: Bearer $MENTOR_TOKEN\" -d '{}' | jq"
echo ""
echo "# 5. Mentor replies in ticket:"
echo "curl -s -X POST $API_BASE_URL/v1/hackathons/$HACKATHON_ID/support/tickets/\$TICKET_ID/reply -H \"Authorization: Bearer $MENTOR_TOKEN\" -H \"Content-Type: application/json\" -d '{\"text\":\"Вот документация по API!\"}' | jq"
echo ""
echo "# 6. Participant sends another message (continues chat):"
echo "curl -s -X POST $API_BASE_URL/v1/hackathons/$HACKATHON_ID/support/messages -H \"Authorization: Bearer $PARTICIPANT_TOKEN\" -H \"Content-Type: application/json\" -d '{\"text\":\"Спасибо! А как работает аутентификация?\"}' | jq"
echo ""
echo "# 7. Mentor views all messages in this ticket:"
echo "curl -s \"$API_BASE_URL/v1/hackathons/$HACKATHON_ID/support/tickets/\$TICKET_ID/messages?query.limit=50&query.offset=0\" -H \"Authorization: Bearer $MENTOR_TOKEN\" | jq"
echo ""
echo "# 8. Participant views their full chat history (all messages):"
echo "curl -s \"$API_BASE_URL/v1/hackathons/$HACKATHON_ID/support/my-messages?query.limit=50&query.offset=0\" -H \"Authorization: Bearer $PARTICIPANT_TOKEN\" | jq"
echo ""
echo "# 9. Mentor closes ticket (system message: 'Ticket closed'):"
echo "curl -s -X POST $API_BASE_URL/v1/hackathons/$HACKATHON_ID/support/tickets/\$TICKET_ID/close -H \"Authorization: Bearer $MENTOR_TOKEN\" -d '{}' | jq"
echo ""
echo "# 10. Participant sends message after close (creates NEW ticket):"
echo "NEW_TICKET_ID=\$(curl -s -X POST $API_BASE_URL/v1/hackathons/$HACKATHON_ID/support/messages -H \"Authorization: Bearer $PARTICIPANT_TOKEN\" -H \"Content-Type: application/json\" -d '{\"text\":\"У меня новый вопрос!\"}' | jq -r '.ticketId')"
echo ""
echo "echo \"New Ticket ID: \$NEW_TICKET_ID\""
echo ""
echo "# 11. Owner (read-only) lists all tickets:"
echo "curl -s -X POST $API_BASE_URL/v1/hackathons/$HACKATHON_ID/support/tickets/all/list -H \"Authorization: Bearer $OWNER_TOKEN\" -H \"Content-Type: application/json\" -d '{\"limit\":10,\"offset\":0}' | jq"
echo ""
echo "# 12. Owner (read-only) views messages in a ticket:"
echo "curl -s \"$API_BASE_URL/v1/hackathons/$HACKATHON_ID/support/tickets/\$TICKET_ID/messages?query.limit=50&query.offset=0\" -H \"Authorization: Bearer $OWNER_TOKEN\" | jq"
echo ""
echo "=========================================="
echo "💡 Key concepts:"
echo "=========================================="
echo ""
echo "✅ Participants:"
echo "  - Use SendMessage (/support/messages) for all communication"
echo "  - View their chat via GetMyChatMessages (/support/my-messages)"
echo "  - Don't see 'tickets' - just a single chat interface"
echo "  - Sending after ticket close creates a NEW ticket automatically"
echo ""
echo "✅ Mentors:"
echo "  - See support as multiple tickets"
echo "  - ListAllTickets to find unassigned tickets"
echo "  - ClaimTicket to assign themselves"
echo "  - ReplyInTicket to respond"
echo "  - GetTicketMessages to view ticket history"
echo "  - CloseTicket when done"
echo ""
echo "✅ Organizers (Owners):"
echo "  - Read-only access to tickets and messages"
echo "  - Can ListAllTickets and GetTicketMessages"
echo "  - CANNOT ReplyInTicket or CloseTicket"
echo "  - Can AssignTicket to a specific mentor"
echo ""
echo "✅ System Messages:"
echo "  - 'Mentor joined the chat' when ticket is claimed/assigned"
echo "  - 'Ticket closed' when mentor closes ticket"
echo "  - author_user_id is empty string for system messages"
echo ""
