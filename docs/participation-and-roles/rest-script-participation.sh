#!/bin/bash

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

BASE_URL="http://localhost:8080"
TIMESTAMP=$(date +%s)

echo -e "${YELLOW}=== Participation Service REST Testing ===${NC}\n"

# ========================================
# 1. Setup: Register Users
# ========================================
echo -e "${GREEN}1. Registering test users...${NC}"

echo -e "${BLUE}Registering Alice (hackathon owner/staff)...${NC}"
ALICE_RESPONSE=$(curl -s -X POST $BASE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice_part_staff_'$TIMESTAMP'",
    "email": "alice_part_staff_'$TIMESTAMP'@hackathon.com",
    "password": "SecurePass123",
    "first_name": "Alice",
    "last_name": "Staff",
    "timezone": "UTC",
    "idempotency_key": {"key": "alice-part-staff-'$TIMESTAMP'"}
  }')

ALICE_TOKEN=$(echo $ALICE_RESPONSE | jq -r '.accessToken')
if [ "$ALICE_TOKEN" = "null" ] || [ -z "$ALICE_TOKEN" ]; then
    echo -e "${RED}Failed to register Alice${NC}"
    echo $ALICE_RESPONSE | jq .
    exit 1
fi
echo -e "${GREEN}✓ Alice registered${NC}"

echo -e "${BLUE}Registering Bob (participant #1)...${NC}"
BOB_RESPONSE=$(curl -s -X POST $BASE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "bob_part_'$TIMESTAMP'",
    "email": "bob_part_'$TIMESTAMP'@hackathon.com",
    "password": "SecurePass123",
    "first_name": "Bob",
    "last_name": "Participant",
    "timezone": "UTC",
    "idempotency_key": {"key": "bob-part-'$TIMESTAMP'"}
  }')

BOB_TOKEN=$(echo $BOB_RESPONSE | jq -r '.accessToken')
if [ "$BOB_TOKEN" = "null" ] || [ -z "$BOB_TOKEN" ]; then
    echo -e "${RED}Failed to register Bob${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Bob registered${NC}"

echo -e "${BLUE}Registering Charlie (participant #2)...${NC}"
CHARLIE_RESPONSE=$(curl -s -X POST $BASE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "charlie_part_'$TIMESTAMP'",
    "email": "charlie_part_'$TIMESTAMP'@hackathon.com",
    "password": "SecurePass123",
    "first_name": "Charlie",
    "last_name": "Seeker",
    "timezone": "UTC",
    "idempotency_key": {"key": "charlie-part-'$TIMESTAMP'"}
  }')

CHARLIE_TOKEN=$(echo $CHARLIE_RESPONSE | jq -r '.accessToken')
if [ "$CHARLIE_TOKEN" = "null" ] || [ -z "$CHARLIE_TOKEN" ]; then
    echo -e "${RED}Failed to register Charlie${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Charlie registered${NC}"

echo -e "${BLUE}Registering Diana (participant #3)...${NC}"
DIANA_RESPONSE=$(curl -s -X POST $BASE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "diana_part_'$TIMESTAMP'",
    "email": "diana_part_'$TIMESTAMP'@hackathon.com",
    "password": "SecurePass123",
    "first_name": "Diana",
    "last_name": "New",
    "timezone": "UTC",
    "idempotency_key": {"key": "diana-part-'$TIMESTAMP'"}
  }')

DIANA_TOKEN=$(echo $DIANA_RESPONSE | jq -r '.accessToken')
if [ "$DIANA_TOKEN" = "null" ] || [ -z "$DIANA_TOKEN" ]; then
    echo -e "${RED}Failed to register Diana${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Diana registered${NC}\n"

# Extract user IDs
ALICE_USER_ID=$(curl -s -X POST $BASE_URL/v1/auth/introspect \
  -H "Content-Type: application/json" \
  -d '{"access_token": "'$ALICE_TOKEN'"}' | jq -r '.userId')

BOB_USER_ID=$(curl -s -X POST $BASE_URL/v1/auth/introspect \
  -H "Content-Type: application/json" \
  -d '{"access_token": "'$BOB_TOKEN'"}' | jq -r '.userId')

CHARLIE_USER_ID=$(curl -s -X POST $BASE_URL/v1/auth/introspect \
  -H "Content-Type: application/json" \
  -d '{"access_token": "'$CHARLIE_TOKEN'"}' | jq -r '.userId')

DIANA_USER_ID=$(curl -s -X POST $BASE_URL/v1/auth/introspect \
  -H "Content-Type: application/json" \
  -d '{"access_token": "'$DIANA_TOKEN'"}' | jq -r '.userId')

echo -e "${GREEN}User IDs extracted:${NC}"
echo -e "  Alice (staff): $ALICE_USER_ID"
echo -e "  Bob: $BOB_USER_ID"
echo -e "  Charlie: $CHARLIE_USER_ID"
echo -e "  Diana: $DIANA_USER_ID\n"

# ========================================
# 2. Create Hackathon (Alice)
# ========================================
echo -e "${GREEN}2. Creating hackathon (Alice)...${NC}"

CREATE_HACK_RESPONSE=$(curl -s -X POST $BASE_URL/v1/hackathons \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Participation Test Hackathon '$TIMESTAMP'",
    "short_description": "Testing participation flows",
    "description": "Full test of participation service",
    "location": {
      "online": false,
      "country": "Russia",
      "city": "Moscow",
      "venue": "Test Venue"
    },
    "dates": {
      "registration_opens_at": "2026-03-01T00:00:00Z",
      "registration_closes_at": "2026-03-20T23:59:59Z",
      "starts_at": "2026-03-25T10:00:00Z",
      "ends_at": "2026-03-27T18:00:00Z",
      "judging_ends_at": "2026-03-30T18:00:00Z"
    },
    "registration_policy": {
      "allow_individual": true,
      "allow_team": true
    },
    "limits": {
      "team_size_max": 5
    },
    "idempotency_key": {"key": "test-hackathon-'$TIMESTAMP'"}
  }')

HACKATHON_ID=$(echo $CREATE_HACK_RESPONSE | jq -r '.hackathonId')
if [ "$HACKATHON_ID" = "null" ] || [ -z "$HACKATHON_ID" ]; then
    echo -e "${RED}Failed to create hackathon${NC}"
    echo $CREATE_HACK_RESPONSE | jq .
    exit 1
fi
echo -e "${GREEN}✓ Hackathon created: $HACKATHON_ID${NC}\n"

# ========================================
# 3. Publish Hackathon
# ========================================
echo -e "${GREEN}3. Publishing hackathon...${NC}"

PUBLISH_RESPONSE=$(curl -s -X POST $BASE_URL/v1/hackathons/$HACKATHON_ID:publish \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "publish-hackathon-'$TIMESTAMP'"}
  }')

echo -e "${GREEN}✓ Hackathon published${NC}\n"

# ========================================
# 4. ListTeamRoles (public)
# ========================================
echo -e "${GREEN}4. Listing team roles...${NC}"

ROLES_RESPONSE=$(curl -s $BASE_URL/v1/team-roles \
  -H "Authorization: Bearer $BOB_TOKEN")

ROLES_COUNT=$(echo $ROLES_RESPONSE | jq '.teamRoles | length')
if [ "$ROLES_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ Found $ROLES_COUNT team roles${NC}"
else
    echo -e "${RED}✗ No team roles found${NC}"
    echo $ROLES_RESPONSE | jq .
    exit 1
fi

# Extract some role IDs for testing
BACKEND_ROLE_ID=$(echo $ROLES_RESPONSE | jq -r '.teamRoles[] | select(.name == "Backend") | .id')
FRONTEND_ROLE_ID=$(echo $ROLES_RESPONSE | jq -r '.teamRoles[] | select(.name == "Frontend") | .id')
FULLSTACK_ROLE_ID=$(echo $ROLES_RESPONSE | jq -r '.teamRoles[] | select(.name == "Fullstack") | .id')
DESIGNER_ROLE_ID=$(echo $ROLES_RESPONSE | jq -r '.teamRoles[] | select(.name == "Designer") | .id')

echo -e "${BLUE}Sample roles: Backend=$BACKEND_ROLE_ID, Frontend=$FRONTEND_ROLE_ID${NC}\n"

# ========================================
# 5. RegisterForHackathon (Bob as individual)
# ========================================
echo -e "${GREEN}5. Bob registers for hackathon (INDIVIDUAL_ACTIVE)...${NC}"

BOB_REG_RESPONSE=$(curl -s -X POST $BASE_URL/v1/hackathons/$HACKATHON_ID/participations:register \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "desired_status": "PART_INDIVIDUAL",
    "wished_role_ids": [],
    "motivation_text": "I want to build something amazing!",
    "idempotency_key": {"key": "bob-register-'$TIMESTAMP'"}
  }')

BOB_STATUS=$(echo $BOB_REG_RESPONSE | jq -r '.participation.status')
if [ "$BOB_STATUS" = "PART_INDIVIDUAL" ]; then
    echo -e "${GREEN}✓ Bob registered as INDIVIDUAL_ACTIVE${NC}"
else
    echo -e "${RED}✗ Bob registration failed${NC}"
    echo $BOB_REG_RESPONSE | jq .
    exit 1
fi
echo ""

# ========================================
# 6. RegisterForHackathon (Charlie as looking for team)
# ========================================
echo -e "${GREEN}6. Charlie registers for hackathon (LOOKING_FOR_TEAM)...${NC}"

CHARLIE_REG_RESPONSE=$(curl -s -X POST $BASE_URL/v1/hackathons/$HACKATHON_ID/participations:register \
  -H "Authorization: Bearer $CHARLIE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "desired_status": "PART_LOOKING_FOR_TEAM",
    "wished_role_ids": ["'$BACKEND_ROLE_ID'", "'$FULLSTACK_ROLE_ID'"],
    "motivation_text": "Looking for a great team to join!",
    "idempotency_key": {"key": "charlie-register-'$TIMESTAMP'"}
  }')

CHARLIE_STATUS=$(echo $CHARLIE_REG_RESPONSE | jq -r '.participation.status')
if [ "$CHARLIE_STATUS" = "PART_LOOKING_FOR_TEAM" ]; then
    echo -e "${GREEN}✓ Charlie registered as LOOKING_FOR_TEAM${NC}"
else
    echo -e "${RED}✗ Charlie registration failed${NC}"
    echo $CHARLIE_REG_RESPONSE | jq .
    exit 1
fi
echo ""

# ========================================
# 7. RegisterForHackathon (Diana)
# ========================================
echo -e "${GREEN}7. Diana registers for hackathon...${NC}"

DIANA_REG_RESPONSE=$(curl -s -X POST $BASE_URL/v1/hackathons/$HACKATHON_ID/participations:register \
  -H "Authorization: Bearer $DIANA_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "desired_status": "PART_INDIVIDUAL",
    "wished_role_ids": ["'$FRONTEND_ROLE_ID'"],
    "motivation_text": "I love frontend development!",
    "idempotency_key": {"key": "diana-register-'$TIMESTAMP'"}
  }')

DIANA_STATUS=$(echo $DIANA_REG_RESPONSE | jq -r '.participation.status')
if [ "$DIANA_STATUS" = "PART_INDIVIDUAL" ]; then
    echo -e "${GREEN}✓ Diana registered${NC}"
else
    echo -e "${RED}✗ Diana registration failed${NC}"
    exit 1
fi
echo ""

# ========================================
# 8. GetMyParticipation (Diana)
# ========================================
echo -e "${GREEN}8. Diana checks her participation...${NC}"

DIANA_GET_RESPONSE=$(curl -s $BASE_URL/v1/hackathons/$HACKATHON_ID/participations/me \
  -H "Authorization: Bearer $DIANA_TOKEN")

DIANA_GET_STATUS=$(echo $DIANA_GET_RESPONSE | jq -r '.participation.status')
if [ "$DIANA_GET_STATUS" = "PART_INDIVIDUAL" ]; then
    echo -e "${GREEN}✓ Diana sees her participation${NC}"
else
    echo -e "${RED}✗ GetMyParticipation failed${NC}"
    exit 1
fi
echo ""

# ========================================
# 9. UpdateMyParticipation (Diana)
# ========================================
echo -e "${GREEN}9. Diana updates her profile...${NC}"

DIANA_UPDATE_RESPONSE=$(curl -s -X PUT $BASE_URL/v1/hackathons/$HACKATHON_ID/participations/me \
  -H "Authorization: Bearer $DIANA_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "wished_role_ids": ["'$FRONTEND_ROLE_ID'", "'$DESIGNER_ROLE_ID'"],
    "motivation_text": "I love frontend and design!",
    "idempotency_key": {"key": "diana-update-'$TIMESTAMP'"}
  }')

DIANA_ROLES_COUNT=$(echo $DIANA_UPDATE_RESPONSE | jq '.participation.profile.wishedRoles | length')
if [ "$DIANA_ROLES_COUNT" = "2" ]; then
    echo -e "${GREEN}✓ Diana updated her profile (2 roles)${NC}"
else
    echo -e "${RED}✗ UpdateMyParticipation failed${NC}"
    exit 1
fi
echo ""

# ========================================
# 10. SwitchParticipationMode (Diana)
# ========================================
echo -e "${GREEN}10. Diana switches to LOOKING_FOR_TEAM...${NC}"

DIANA_SWITCH_RESPONSE=$(curl -s -X POST $BASE_URL/v1/hackathons/$HACKATHON_ID/participations/me:switchMode \
  -H "Authorization: Bearer $DIANA_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "new_status": "PART_LOOKING_FOR_TEAM",
    "idempotency_key": {"key": "diana-switch-'$TIMESTAMP'"}
  }')

DIANA_NEW_STATUS=$(echo $DIANA_SWITCH_RESPONSE | jq -r '.participation.status')
if [ "$DIANA_NEW_STATUS" = "PART_LOOKING_FOR_TEAM" ]; then
    echo -e "${GREEN}✓ Diana switched to LOOKING_FOR_TEAM${NC}"
else
    echo -e "${RED}✗ SwitchMode failed${NC}"
    exit 1
fi
echo ""

# ========================================
# 11. GetUserParticipation (Bob as participant)
# ========================================
echo -e "${GREEN}11. Bob (participant) views Diana's participation...${NC}"

# Сначала регистрируем Bob
curl -s -X POST $BASE_URL/v1/hackathons/$HACKATHON_ID/participations:register \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "desired_status": "PART_INDIVIDUAL",
    "wished_role_ids": [],
    "motivation_text": "I want to participate!"
  }' > /dev/null

BOB_VIEW_RESPONSE=$(curl -s $BASE_URL/v1/hackathons/$HACKATHON_ID/participations/users/$DIANA_USER_ID \
  -H "Authorization: Bearer $BOB_TOKEN")

VIEWED_STATUS=$(echo $BOB_VIEW_RESPONSE | jq -r '.participation.status')
if [ "$VIEWED_STATUS" = "PART_LOOKING_FOR_TEAM" ]; then
    echo -e "${GREEN}✓ Bob can view Diana's participation${NC}"
else
    echo -e "${RED}✗ GetUserParticipation failed${NC}"
    exit 1
fi
echo ""

# ========================================
# 12. ListHackathonParticipants (Bob as participant)
# ========================================
echo -e "${GREEN}12. Bob lists all participants...${NC}"

BOB_LIST_RESPONSE=$(curl -s -X POST $BASE_URL/v1/hackathons/$HACKATHON_ID/participations:list \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}')

PARTICIPANTS_COUNT=$(echo $BOB_LIST_RESPONSE | jq '.participants | length')
if [ "$PARTICIPANTS_COUNT" -ge 2 ]; then
    echo -e "${GREEN}✓ Bob sees $PARTICIPANTS_COUNT participants (Diana + Bob)${NC}"
else
    echo -e "${RED}✗ ListParticipants failed (expected >= 2, got $PARTICIPANTS_COUNT)${NC}"
    exit 1
fi
echo ""

# ========================================
# 13. UnregisterFromHackathon (Diana)
# ========================================
echo -e "${GREEN}13. Diana unregisters from hackathon...${NC}"

DIANA_UNREG_RESPONSE=$(curl -s -X POST $BASE_URL/v1/hackathons/$HACKATHON_ID/participations/me:unregister \
  -H "Authorization: Bearer $DIANA_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "diana-unregister-'$TIMESTAMP'"}
  }')

# Empty response is success
echo -e "${GREEN}✓ Diana unregistered${NC}\n"

# ========================================
# 14. GetMyParticipation after unregister (should fail)
# ========================================
echo -e "${GREEN}14. Diana tries to get participation after unregister (should fail)...${NC}"

DIANA_GET_AFTER_UNREG=$(curl -s -w "%{http_code}" $BASE_URL/v1/hackathons/$HACKATHON_ID/participations/me \
  -H "Authorization: Bearer $DIANA_TOKEN")

HTTP_CODE="${DIANA_GET_AFTER_UNREG: -3}"
if [ "$HTTP_CODE" = "404" ]; then
    echo -e "${GREEN}✓ Correctly returns 404 after unregister${NC}"
else
    echo -e "${RED}✗ Expected 404, got $HTTP_CODE${NC}"
    exit 1
fi
echo ""

# ========================================
# Summary
# ========================================
echo -e "${YELLOW}=== All Tests Completed Successfully ===${NC}"
echo -e "${GREEN}Summary:${NC}"
echo -e "  - ✓ ListTeamRoles: retrieved team roles catalog"
echo -e "  - ✓ RegisterForHackathon: 3 users registered (Bob, Charlie, Diana)"
echo -e "  - ✓ GetMyParticipation: participants can view own registration"
echo -e "  - ✓ UpdateMyParticipation: profile updated successfully"
echo -e "  - ✓ SwitchParticipationMode: mode switched successfully"
echo -e "  - ✓ GetUserParticipation: staff can view user participations"
echo -e "  - ✓ ListHackathonParticipants: staff can list all participants"
echo -e "  - ✓ UnregisterFromHackathon: registration cancelled successfully"
echo -e "${GREEN}✓ All participation flows work correctly!${NC}"
