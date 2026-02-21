#!/bin/bash

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

BASE_URL="http://localhost:8080"
TIMESTAMP=$(date +%s)

echo -e "${YELLOW}=== Participation Service Fail Cases Testing ===${NC}\n"

# Helper function to check for expected failure
check_failure() {
    local response=$1
    local expected_code=$2
    local test_name=$3
    
    local actual_code=$(echo $response | jq -r '.code // empty')
    if [ "$actual_code" = "$expected_code" ]; then
        echo -e "${GREEN}✓ $test_name: correctly failed with code $expected_code${NC}"
        return 0
    else
        echo -e "${RED}✗ $test_name: expected code $expected_code, got $actual_code${NC}"
        echo $response | jq .
        return 1
    fi
}

# ========================================
# Setup: Register Users and Create Hackathon
# ========================================
echo -e "${GREEN}Setup: Creating test environment...${NC}"

# Register staff user
STAFF_RESPONSE=$(curl -s -X POST $BASE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "staff_fail_'$TIMESTAMP'",
    "email": "staff_fail_'$TIMESTAMP'@test.com",
    "password": "SecurePass123",
    "first_name": "Staff",
    "last_name": "User",
    "timezone": "UTC"
  }')
STAFF_TOKEN=$(echo $STAFF_RESPONSE | jq -r '.accessToken')

# Register participant users
BOB_RESPONSE=$(curl -s -X POST $BASE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "bob_fail_'$TIMESTAMP'",
    "email": "bob_fail_'$TIMESTAMP'@test.com",
    "password": "SecurePass123",
    "first_name": "Bob",
    "last_name": "Test",
    "timezone": "UTC"
  }')
BOB_TOKEN=$(echo $BOB_RESPONSE | jq -r '.accessToken')

CHARLIE_RESPONSE=$(curl -s -X POST $BASE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "charlie_fail_'$TIMESTAMP'",
    "email": "charlie_fail_'$TIMESTAMP'@test.com",
    "password": "SecurePass123",
    "first_name": "Charlie",
    "last_name": "Test",
    "timezone": "UTC"
  }')
CHARLIE_TOKEN=$(echo $CHARLIE_RESPONSE | jq -r '.accessToken')

CHARLIE_USER_ID=$(curl -s -X POST $BASE_URL/v1/auth/introspect \
  -H "Content-Type: application/json" \
  -d '{"access_token": "'$CHARLIE_TOKEN'"}' | jq -r '.userId')

# Create hackathon
CREATE_HACK=$(curl -s -X POST $BASE_URL/v1/hackathons \
  -H "Authorization: Bearer $STAFF_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Fail Test Hackathon '$TIMESTAMP'",
    "short_description": "Testing failures",
    "description": "Test",
    "location": {"online": true},
    "dates": {
      "registration_opens_at": "2026-03-01T00:00:00Z",
      "registration_closes_at": "2026-03-20T23:59:59Z",
      "starts_at": "2026-03-25T10:00:00Z",
      "ends_at": "2026-03-27T18:00:00Z",
      "judging_ends_at": "2026-03-30T18:00:00Z"
    },
    "registration_policy": {"allow_individual": true, "allow_team": true},
    "limits": {"team_size_max": 5}
  }')
HACKATHON_ID=$(echo $CREATE_HACK | jq -r '.hackathonId')

# Publish hackathon
curl -s -X POST $BASE_URL/v1/hackathons/$HACKATHON_ID:publish \
  -H "Authorization: Bearer $STAFF_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}' > /dev/null

# Get team roles
ROLES=$(curl -s $BASE_URL/v1/team-roles -H "Authorization: Bearer $BOB_TOKEN")
BACKEND_ROLE=$(echo $ROLES | jq -r '.teamRoles[] | select(.name == "Backend") | .id')

echo -e "${GREEN}✓ Setup complete. Hackathon: $HACKATHON_ID${NC}\n"

# ========================================
# Fail Case 1: Register twice (should fail with 409)
# ========================================
echo -e "${GREEN}Test 1: Try to register twice...${NC}"

# First registration
curl -s -X POST $BASE_URL/v1/hackathons/$HACKATHON_ID/participations:register \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "desired_status": "PART_INDIVIDUAL",
    "wished_role_ids": [],
    "motivation_text": "First registration"
  }' > /dev/null

# Second registration (should fail)
DOUBLE_REG=$(curl -s -X POST $BASE_URL/v1/hackathons/$HACKATHON_ID/participations:register \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "desired_status": "PART_LOOKING_FOR_TEAM",
    "wished_role_ids": [],
    "motivation_text": "Second registration"
  }')

check_failure "$DOUBLE_REG" "9" "Double registration"
echo ""

# ========================================
# Fail Case 2: Non-participant tries to list participants
# ========================================
echo -e "${GREEN}Test 2: Non-participant tries to list participants (should fail with 7)...${NC}"

# Регистрируем новый токен для пользователя, который НЕ зарегистрирован на хакатоне
EVE_REG=$(curl -s -X POST $BASE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "eve_nonpart_'$(date +%s)'",
    "email": "eve_nonpart_'$(date +%s)'@test.com",
    "password": "SecurePass123",
    "first_name": "Eve",
    "last_name": "NonParticipant",
    "timezone": "UTC"
  }')

EVE_TOKEN=$(echo $EVE_REG | jq -r '.accessToken')

NON_PARTICIPANT_LIST=$(curl -s -X POST $BASE_URL/v1/hackathons/$HACKATHON_ID/participations:list \
  -H "Authorization: Bearer $EVE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}')

check_failure "$NON_PARTICIPANT_LIST" "7" "Non-participant list participants"
echo ""

# ========================================
# Fail Case 3: Non-participant tries to view user participation
# ========================================
echo -e "${GREEN}Test 3: Non-participant tries to view user participation (should fail with 7)...${NC}"

NON_PARTICIPANT_VIEW=$(curl -s $BASE_URL/v1/hackathons/$HACKATHON_ID/participations/users/$CHARLIE_USER_ID \
  -H "Authorization: Bearer $EVE_TOKEN")

check_failure "$NON_PARTICIPANT_VIEW" "7" "Non-participant view user participation"
echo ""

# ========================================
# Fail Case 4: Get participation without registration
# ========================================
echo -e "${GREEN}Test 4: Get participation without registration (should fail with 5)...${NC}"

NO_REG_GET=$(curl -s $BASE_URL/v1/hackathons/$HACKATHON_ID/participations/me \
  -H "Authorization: Bearer $CHARLIE_TOKEN")

check_failure "$NO_REG_GET" "5" "Get non-existent participation"
echo ""

# ========================================
# Fail Case 5: Update participation without registration
# ========================================
echo -e "${GREEN}Test 5: Update participation without registration (should fail with 5)...${NC}"

NO_REG_UPDATE=$(curl -s -X PUT $BASE_URL/v1/hackathons/$HACKATHON_ID/participations/me \
  -H "Authorization: Bearer $CHARLIE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "wished_role_ids": ["'$BACKEND_ROLE'"],
    "motivation_text": "Update without registration"
  }')

check_failure "$NO_REG_UPDATE" "5" "Update non-existent participation"
echo ""

# ========================================
# Fail Case 6: Switch mode without registration
# ========================================
echo -e "${GREEN}Test 6: Switch mode without registration (should fail with 5)...${NC}"

NO_REG_SWITCH=$(curl -s -X POST $BASE_URL/v1/hackathons/$HACKATHON_ID/participations/me:switchMode \
  -H "Authorization: Bearer $CHARLIE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "new_status": "PART_LOOKING_FOR_TEAM"
  }')

check_failure "$NO_REG_SWITCH" "5" "Switch mode without registration"
echo ""

# ========================================
# Fail Case 7: Switch to same status
# ========================================
echo -e "${GREEN}Test 7: Switch to same status (should fail with 7)...${NC}"

# Bob is already INDIVIDUAL_ACTIVE
SAME_STATUS_SWITCH=$(curl -s -X POST $BASE_URL/v1/hackathons/$HACKATHON_ID/participations/me:switchMode \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "new_status": "PART_INDIVIDUAL"
  }')

check_failure "$SAME_STATUS_SWITCH" "7" "Switch to same status"
echo ""

# ========================================
# Fail Case 8: Unregister without registration
# ========================================
echo -e "${GREEN}Test 8: Unregister without registration (should fail with 5)...${NC}"

NO_REG_UNREG=$(curl -s -X POST $BASE_URL/v1/hackathons/$HACKATHON_ID/participations/me:unregister \
  -H "Authorization: Bearer $CHARLIE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}')

check_failure "$NO_REG_UNREG" "5" "Unregister without registration"
echo ""

# ========================================
# Fail Case 9: Register with invalid role IDs
# ========================================
echo -e "${GREEN}Test 9: Register with invalid role IDs (should fail with 3)...${NC}"

INVALID_ROLE_REG=$(curl -s -X POST $BASE_URL/v1/hackathons/$HACKATHON_ID/participations:register \
  -H "Authorization: Bearer $CHARLIE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "desired_status": "PART_INDIVIDUAL",
    "wished_role_ids": ["00000000-0000-0000-0000-999999999999"],
    "motivation_text": "Invalid role"
  }')

check_failure "$INVALID_ROLE_REG" "3" "Register with invalid role IDs"
echo ""

# ========================================
# Fail Case 10: Update with invalid role IDs
# ========================================
echo -e "${GREEN}Test 10: Update with invalid role IDs (should fail with 3)...${NC}"

INVALID_ROLE_UPDATE=$(curl -s -X PUT $BASE_URL/v1/hackathons/$HACKATHON_ID/participations/me \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "wished_role_ids": ["00000000-0000-0000-0000-999999999999"],
    "motivation_text": "Invalid role"
  }')

check_failure "$INVALID_ROLE_UPDATE" "3" "Update with invalid role IDs"
echo ""

# ========================================
# Fail Case 11: Switch to invalid status (TEAM_MEMBER)
# ========================================
echo -e "${GREEN}Test 11: Switch to TEAM_MEMBER status (should fail with 3)...${NC}"

INVALID_STATUS_SWITCH=$(curl -s -X POST $BASE_URL/v1/hackathons/$HACKATHON_ID/participations/me:switchMode \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "new_status": "PART_TEAM_MEMBER"
  }')

check_failure "$INVALID_STATUS_SWITCH" "3" "Switch to TEAM_MEMBER"
echo ""

# ========================================
# Fail Case 12: Register without authentication
# ========================================
echo -e "${GREEN}Test 12: Register without authentication (should fail with HTTP 401)...${NC}"

NO_AUTH_REG=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/v1/hackathons/$HACKATHON_ID/participations:register \
  -H "Content-Type: application/json" \
  -d '{
    "desired_status": "PART_INDIVIDUAL",
    "wished_role_ids": [],
    "motivation_text": "No auth"
  }')

HTTP_CODE=$(echo "$NO_AUTH_REG" | tail -n 1)
if [ "$HTTP_CODE" = "401" ]; then
    echo -e "${GREEN}✓ No authentication: correctly returned 401${NC}"
else
    echo -e "${RED}✗ Expected 401, got $HTTP_CODE${NC}"
fi
echo ""

# ========================================
# Summary
# ========================================
echo -e "${YELLOW}=== All Fail Cases Tested ===${NC}"
echo -e "${GREEN}Summary:${NC}"
echo -e "  - ✓ Double registration blocked (409)"
echo -e "  - ✓ Non-staff access control working (403)"
echo -e "  - ✓ Not found errors working (404)"
echo -e "  - ✓ Invalid input validation working (400)"
echo -e "  - ✓ Authentication required (401)"
echo -e "${GREEN}✓ All negative scenarios validated!${NC}"
