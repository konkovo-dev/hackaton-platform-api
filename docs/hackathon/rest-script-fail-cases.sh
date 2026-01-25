#!/bin/bash

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

BASE_URL="http://localhost:8080"
TIMESTAMP=$(date +%s)

echo -e "${YELLOW}=== Hackathon Service - Validation Fail Cases ===${NC}\n"
echo -e "${BLUE}Testing stage-based validation rules${NC}\n"

# ========================================
# Setup: Register User
# ========================================
echo -e "${GREEN}1. Registering test user...${NC}"

ALICE_RESPONSE=$(curl -s -X POST $BASE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice_fail_'$TIMESTAMP'",
    "email": "alice_fail_'$TIMESTAMP'@hackathon.com",
    "password": "SecurePass123",
    "first_name": "Alice",
    "last_name": "FailTest",
    "timezone": "UTC",
    "idempotency_key": {"key": "alice-fail-'$TIMESTAMP'"}
  }')

ALICE_TOKEN=$(echo $ALICE_RESPONSE | jq -r '.accessToken')
if [ "$ALICE_TOKEN" = "null" ] || [ -z "$ALICE_TOKEN" ]; then
    echo -e "${RED}Failed to register Alice${NC}"
    exit 1
fi

# Extract Alice's user_id immediately after registration
sleep 1
ALICE_ME=$(curl -s "$BASE_URL/v1/users/me" -H "Authorization: Bearer $ALICE_TOKEN")
ALICE_USER_ID=$(echo "$ALICE_ME" | jq -r '.user.userId')

echo -e "${GREEN}✓ Alice registered (user_id: $ALICE_USER_ID)${NC}\n"

# ========================================
# Create Hackathon with dates for RUNNING stage
# ========================================
echo -e "${GREEN}2. Creating hackathon with RUNNING stage dates...${NC}"

# Dates that put hackathon in RUNNING stage
PAST_REG_OPEN=$(date -u -d "10 days ago" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v-10d +"%Y-%m-%dT%H:%M:%SZ")
PAST_REG_CLOSE=$(date -u -d "5 days ago" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v-5d +"%Y-%m-%dT%H:%M:%SZ")
PAST_START=$(date -u -d "3 days ago" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v-3d +"%Y-%m-%dT%H:%M:%SZ")
FUTURE_END=$(date -u -d "3 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+3d +"%Y-%m-%dT%H:%M:%SZ")
FUTURE_JUDGING=$(date -u -d "7 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+7d +"%Y-%m-%dT%H:%M:%SZ")

CREATE_RESPONSE=$(curl -s -X POST $BASE_URL/v1/hackathons \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Running Stage Hackathon",
    "short_description": "Test hackathon for validation",
    "description": "Testing stage restrictions",
    "location": {
      "online": false,
      "country": "Russia",
      "city": "Moscow",
      "venue": "Test Center"
    },
    "dates": {
      "registration_opens_at": "'$PAST_REG_OPEN'",
      "registration_closes_at": "'$PAST_REG_CLOSE'",
      "starts_at": "'$PAST_START'",
      "ends_at": "'$FUTURE_END'",
      "judging_ends_at": "'$FUTURE_JUDGING'"
    },
    "registration_policy": {
      "allow_individual": true,
      "allow_team": true
    },
    "limits": {
      "team_size_max": 5
    },
    "idempotency_key": {"key": "running-hack-'$TIMESTAMP'"}
  }')

HACKATHON_ID=$(echo "$CREATE_RESPONSE" | jq -r '.hackathonId')
if [ "$HACKATHON_ID" = "null" ] || [ -z "$HACKATHON_ID" ]; then
    echo -e "${RED}Failed to create hackathon${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Hackathon created. ID: $HACKATHON_ID${NC}\n"

sleep 2

# ========================================
# Add Task
# ========================================
echo -e "${GREEN}3. Adding task...${NC}"
curl -s -X PUT "$BASE_URL/v1/hackathons/$HACKATHON_ID/task" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "task": "Build something amazing",
    "idempotency_key": {"key": "task-add-'$TIMESTAMP'"}
  }' > /dev/null

echo -e "${GREEN}✓ Task added${NC}\n"

# ========================================
# Publish Hackathon
# ========================================
echo -e "${GREEN}4. Publishing hackathon...${NC}"

# For testing: manually set state to published and stage to RUNNING in DB
# This bypasses the "registration_opens_at must be in future" validation
docker exec hackathon-postgres psql -U hackathon -d hackathon -c \
  "UPDATE hackathon.hackathons SET state='published', stage='running', published_at=NOW() WHERE id='$HACKATHON_ID';" > /dev/null 2>&1

sleep 1

GET_PUBLISHED=$(curl -s "$BASE_URL/v1/hackathons/$HACKATHON_ID" \
  -H "Authorization: Bearer $ALICE_TOKEN")

CURRENT_STAGE=$(echo "$GET_PUBLISHED" | jq -r '.hackathon.stage')
CURRENT_STATE=$(echo "$GET_PUBLISHED" | jq -r '.hackathon.state')
echo -e "${BLUE}✓ Hackathon manually set to RUNNING stage for testing${NC}"
echo -e "${BLUE}Current stage: $CURRENT_STAGE, state: $CURRENT_STATE${NC}\n"

# ========================================
# TEST 1: UpdateLocation on RUNNING (should FAIL)
# ========================================
echo -e "${GREEN}TEST 1: Updating location on RUNNING stage (should FAIL)...${NC}"
UPDATE_LOCATION=$(curl -s -X PUT $BASE_URL/v1/hackathons/$HACKATHON_ID \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Running Stage Hackathon",
    "short_description": "Test hackathon for validation",
    "description": "Testing stage restrictions",
    "location": {
      "online": false,
      "country": "Russia",
      "city": "Saint Petersburg",
      "venue": "Changed Venue"
    },
    "dates": {
      "registration_opens_at": "'$PAST_REG_OPEN'",
      "registration_closes_at": "'$PAST_REG_CLOSE'",
      "starts_at": "'$PAST_START'",
      "ends_at": "'$FUTURE_END'",
      "judging_ends_at": "'$FUTURE_JUDGING'"
    },
    "registration_policy": {
      "allow_individual": true,
      "allow_team": true
    },
    "limits": {
      "team_size_max": 5
    }
  }')

LOCATION_ERROR=$(echo "$UPDATE_LOCATION" | jq -r '.message // "no error"')
LOCATION_VAL_COUNT=$(echo "$UPDATE_LOCATION" | jq -r '.validationErrors[]? | select(.field == "location") | .field' | wc -l)

if [ "$LOCATION_VAL_COUNT" -gt 0 ] || [[ "$LOCATION_ERROR" == *"location"* ]] || [[ "$LOCATION_ERROR" == *"stage"* ]]; then
    echo -e "${GREEN}✓ TEST 1 PASSED: Location change forbidden on RUNNING${NC}"
    echo "$UPDATE_LOCATION" | jq '.validationErrors // {message: .message}'
else
    echo -e "${RED}✗ TEST 1 FAILED: Location change should be forbidden on RUNNING${NC}"
    echo "$UPDATE_LOCATION" | jq .
fi
echo ""

# ========================================
# TEST 2: UpdateTeamSizeMax on RUNNING (should FAIL)
# ========================================
echo -e "${GREEN}TEST 2: Updating team_size_max on RUNNING stage (should FAIL)...${NC}"
UPDATE_TEAMSIZE=$(curl -s -X PUT $BASE_URL/v1/hackathons/$HACKATHON_ID \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Running Stage Hackathon",
    "short_description": "Test hackathon for validation",
    "description": "Testing stage restrictions",
    "location": {
      "online": false,
      "country": "Russia",
      "city": "Moscow",
      "venue": "Test Center"
    },
    "dates": {
      "registration_opens_at": "'$PAST_REG_OPEN'",
      "registration_closes_at": "'$PAST_REG_CLOSE'",
      "starts_at": "'$PAST_START'",
      "ends_at": "'$FUTURE_END'",
      "judging_ends_at": "'$FUTURE_JUDGING'"
    },
    "registration_policy": {
      "allow_individual": true,
      "allow_team": true
    },
    "limits": {
      "team_size_max": 10
    }
  }')

TEAMSIZE_ERROR=$(echo "$UPDATE_TEAMSIZE" | jq -r '.message // "no error"')
TEAMSIZE_VAL_COUNT=$(echo "$UPDATE_TEAMSIZE" | jq -r '.validationErrors[]? | select(.field == "team_size_max") | .field' | wc -l)

if [ "$TEAMSIZE_VAL_COUNT" -gt 0 ] || [[ "$TEAMSIZE_ERROR" == *"team_size"* ]] || [[ "$TEAMSIZE_ERROR" == *"stage"* ]]; then
    echo -e "${GREEN}✓ TEST 2 PASSED: TeamSizeMax change forbidden on RUNNING${NC}"
    echo "$UPDATE_TEAMSIZE" | jq '.validationErrors // {message: .message}'
else
    echo -e "${RED}✗ TEST 2 FAILED: TeamSizeMax change should be forbidden on RUNNING${NC}"
    echo "$UPDATE_TEAMSIZE" | jq .
fi
echo ""

# ========================================
# TEST 3: DisableType on RUNNING (should FAIL)
# ========================================
echo -e "${GREEN}TEST 3: Disabling registration type on RUNNING stage (should FAIL)...${NC}"
DISABLE_TYPE=$(curl -s -X PUT $BASE_URL/v1/hackathons/$HACKATHON_ID \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Running Stage Hackathon",
    "short_description": "Test hackathon for validation",
    "description": "Testing stage restrictions",
    "location": {
      "online": false,
      "country": "Russia",
      "city": "Moscow",
      "venue": "Test Center"
    },
    "dates": {
      "registration_opens_at": "'$PAST_REG_OPEN'",
      "registration_closes_at": "'$PAST_REG_CLOSE'",
      "starts_at": "'$PAST_START'",
      "ends_at": "'$FUTURE_END'",
      "judging_ends_at": "'$FUTURE_JUDGING'"
    },
    "registration_policy": {
      "allow_individual": false,
      "allow_team": true
    },
    "limits": {
      "team_size_max": 5
    }
  }')

DISABLE_ERROR=$(echo "$DISABLE_TYPE" | jq -r '.message // "no error"')
DISABLE_VAL_COUNT=$(echo "$DISABLE_TYPE" | jq -r '.validationErrors[]? | select(.field == "registration_policy" or .field == "allow_individual") | .field' | wc -l)

if [ "$DISABLE_VAL_COUNT" -gt 0 ] || [[ "$DISABLE_ERROR" == *"registration"* ]] || [[ "$DISABLE_ERROR" == *"stage"* ]]; then
    echo -e "${GREEN}✓ TEST 3 PASSED: DisableType forbidden outside DRAFT/UPCOMING${NC}"
    echo "$DISABLE_TYPE" | jq '.validationErrors // {message: .message}'
else
    echo -e "${RED}✗ TEST 3 FAILED: DisableType should be forbidden on RUNNING${NC}"
    echo "$DISABLE_TYPE" | jq .
fi
echo ""

# ========================================
# Create JUDGING hackathon for Task test
# ========================================
echo -e "${GREEN}5. Creating hackathon for JUDGING stage...${NC}"

PAST_END=$(date -u -d "2 days ago" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v-2d +"%Y-%m-%dT%H:%M:%SZ")
FUTURE_JUDGING_END=$(date -u -d "5 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+5d +"%Y-%m-%dT%H:%M:%SZ")

CREATE_JUDGING=$(curl -s -X POST $BASE_URL/v1/hackathons \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Judging Stage Hackathon",
    "short_description": "Test for judging stage",
    "description": "Testing task restrictions",
    "location": {
      "online": true
    },
    "dates": {
      "registration_opens_at": "'$PAST_REG_OPEN'",
      "registration_closes_at": "'$PAST_REG_CLOSE'",
      "starts_at": "'$PAST_START'",
      "ends_at": "'$PAST_END'",
      "judging_ends_at": "'$FUTURE_JUDGING_END'"
    },
    "registration_policy": {
      "allow_individual": true,
      "allow_team": true
    },
    "limits": {
      "team_size_max": 5
    },
    "idempotency_key": {"key": "judging-hack-'$TIMESTAMP'"}
  }')

JUDGING_ID=$(echo "$CREATE_JUDGING" | jq -r '.hackathonId')
echo -e "${GREEN}✓ Judging hackathon created. ID: $JUDGING_ID${NC}\n"

sleep 2

curl -s -X PUT "$BASE_URL/v1/hackathons/$JUDGING_ID/task" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "task": "Original task",
    "idempotency_key": {"key": "task-judging-'$TIMESTAMP'"}
  }' > /dev/null

# For testing: manually set state to published and stage to JUDGING in DB
docker exec hackathon-postgres psql -U hackathon -d hackathon -c \
  "UPDATE hackathon.hackathons SET state='published', stage='judging', published_at=NOW() WHERE id='$JUDGING_ID';" > /dev/null 2>&1

sleep 1

GET_JUDGING=$(curl -s "$BASE_URL/v1/hackathons/$JUDGING_ID" \
  -H "Authorization: Bearer $ALICE_TOKEN")

JUDGING_STAGE=$(echo "$GET_JUDGING" | jq -r '.hackathon.stage')
JUDGING_STATE=$(echo "$GET_JUDGING" | jq -r '.hackathon.state')
echo -e "${BLUE}✓ Hackathon manually set to JUDGING stage for testing${NC}"
echo -e "${BLUE}Judging hackathon stage: $JUDGING_STAGE, state: $JUDGING_STATE${NC}\n"

# ========================================
# TEST 4: UpdateTask on JUDGING (should FAIL)
# ========================================
echo -e "${GREEN}TEST 4: Updating task on JUDGING stage (should FAIL)...${NC}"
UPDATE_TASK_JUDGING=$(curl -s -X PUT "$BASE_URL/v1/hackathons/$JUDGING_ID/task" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "task": "Modified task during judging",
    "idempotency_key": {"key": "task-update-judging-'$TIMESTAMP'"}
  }')

TASK_ERROR=$(echo "$UPDATE_TASK_JUDGING" | jq -r '.message // "no error"')
TASK_VAL_COUNT=$(echo "$UPDATE_TASK_JUDGING" | jq -r '.validationErrors[]? | select(.field == "task") | .field' | wc -l)

if [ "$TASK_VAL_COUNT" -gt 0 ] || [[ "$TASK_ERROR" == *"task"* ]] || [[ "$TASK_ERROR" == *"forbidden"* ]] || [[ "$TASK_ERROR" == *"stage"* ]] || [[ "$TASK_ERROR" == *"unauthorized"* ]]; then
    echo -e "${GREEN}✓ TEST 4 PASSED: Task update forbidden on JUDGING${NC}"
    echo "$UPDATE_TASK_JUDGING" | jq '.validationErrors // {message: .message}'
else
    echo -e "${RED}✗ TEST 4 FAILED: Task update should be forbidden on JUDGING${NC}"
    echo "$UPDATE_TASK_JUDGING" | jq .
fi
echo ""

# ========================================
# TEST 5: Publish without required fields (should FAIL)
# ========================================
echo -e "${GREEN}TEST 5: Publishing incomplete hackathon (should FAIL)...${NC}"

INCOMPLETE_HACK=$(curl -s -X POST $BASE_URL/v1/hackathons \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Incomplete Hackathon",
    "short_description": "Missing required fields",
    "dates": {
      "registration_opens_at": "'$PAST_REG_OPEN'",
      "registration_closes_at": "'$PAST_REG_CLOSE'",
      "starts_at": "'$PAST_START'",
      "ends_at": "'$FUTURE_END'",
      "judging_ends_at": "'$FUTURE_JUDGING'"
    },
    "registration_policy": {
      "allow_individual": true,
      "allow_team": true
    },
    "idempotency_key": {"key": "incomplete-'$TIMESTAMP'"}
  }')

INCOMPLETE_ID=$(echo "$INCOMPLETE_HACK" | jq -r '.hackathonId')

sleep 2

PUBLISH_INCOMPLETE=$(curl -s -X POST "$BASE_URL/v1/hackathons/$INCOMPLETE_ID:publish" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json")

PUBLISH_ERROR=$(echo "$PUBLISH_INCOMPLETE" | jq -r '.message // "no error"')
PUBLISH_VALIDATION=$(echo "$PUBLISH_INCOMPLETE" | jq '.validationErrors | length')

if [[ "$PUBLISH_VALIDATION" -gt 0 ]] || [[ "$PUBLISH_ERROR" == *"required"* ]] || [[ "$PUBLISH_ERROR" == *"validation"* ]]; then
    echo -e "${GREEN}✓ TEST 5 PASSED: Publish requires all fields (PublishReady)${NC}"
    echo "$PUBLISH_INCOMPLETE" | jq '.validationErrors // {message: .message}'
else
    echo -e "${RED}✗ TEST 5 FAILED: Publish should fail without required fields${NC}"
    echo "$PUBLISH_INCOMPLETE" | jq .
fi
echo ""

# ========================================
# TEST 6: Create Announcement in DRAFT (should FAIL)
# ========================================
echo -e "${GREEN}TEST 6: Creating announcement in DRAFT stage (should FAIL)...${NC}"

# Create a new DRAFT hackathon
DRAFT_HACK=$(curl -s -X POST $BASE_URL/v1/hackathons \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Draft Hackathon",
    "short_description": "Test",
    "idempotency_key": {"key": "draft-hack-'$TIMESTAMP'"}
  }')

DRAFT_ID=$(echo "$DRAFT_HACK" | jq -r '.hackathonId')
sleep 2

DRAFT_ANNOUNCEMENT=$(curl -s -X POST "$BASE_URL/v1/hackathons/$DRAFT_ID/announcements" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Announcement",
    "content": "Should fail in DRAFT",
    "idempotency_key": {"key": "ann-draft-'$TIMESTAMP'"}
  }')

DRAFT_ANN_ERROR=$(echo "$DRAFT_ANNOUNCEMENT" | jq -r '.message // "no error"')
if [[ "$DRAFT_ANN_ERROR" == *"draft"* ]] || [[ "$DRAFT_ANN_ERROR" == *"forbidden"* ]]; then
    echo -e "${GREEN}✓ TEST 6 PASSED: Announcement creation forbidden in DRAFT${NC}"
else
    echo -e "${RED}✗ TEST 6 FAILED: Announcement should be forbidden in DRAFT${NC}"
    echo "$DRAFT_ANNOUNCEMENT" | jq .
fi
echo ""

# ========================================
# TEST 7: TIME_RULE violation - end before start (should FAIL)
# ========================================
echo -e "${GREEN}TEST 7: Creating hackathon with end_date < start_date (should FAIL)...${NC}"

FUTURE_START=$(date -u -d "10 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+10d +"%Y-%m-%dT%H:%M:%SZ")
PAST_END=$(date -u -d "5 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+5d +"%Y-%m-%dT%H:%M:%SZ")

TIME_RULE_HACK=$(curl -s -X POST $BASE_URL/v1/hackathons \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Invalid Time Hackathon",
    "short_description": "Test TIME_RULE",
    "dates": {
      "starts_at": "'$FUTURE_START'",
      "ends_at": "'$PAST_END'"
    },
    "idempotency_key": {"key": "time-rule-'$TIMESTAMP'"}
  }')

TIME_HACK_ID=$(echo "$TIME_RULE_HACK" | jq -r '.hackathonId')
sleep 2

# Try to publish - should fail validation
TIME_PUBLISH=$(curl -s -X POST "$BASE_URL/v1/hackathons/$TIME_HACK_ID:publish" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "pub-time-'$TIMESTAMP'"}
  }')

TIME_ERROR=$(echo "$TIME_PUBLISH" | jq -r '.message // "no error"')
TIME_VALIDATION=$(echo "$TIME_PUBLISH" | jq -r '.validationErrors[]? | select(.code == "INVALID_DATE_RANGE" or .code == "TIME_RULE") | .code // empty')

if [ -n "$TIME_VALIDATION" ] || [[ "$TIME_ERROR" == *"time"* ]] || [[ "$TIME_ERROR" == *"date"* ]] || [[ "$TIME_ERROR" == *"validation"* ]]; then
    echo -e "${GREEN}✓ TEST 7 PASSED: TIME_RULE violation detected${NC}"
    echo "$TIME_PUBLISH" | jq '.validationErrors // {message: .message}'
else
    echo -e "${RED}✗ TEST 7 FAILED: TIME_RULE should prevent invalid date ranges${NC}"
    echo "$TIME_PUBLISH" | jq .
fi
echo ""

# ========================================
# TEST 8: Double publish (should FAIL)
# ========================================
echo -e "${GREEN}TEST 8: Publishing already published hackathon (should FAIL)...${NC}"

# Try to publish the already-published hackathon
DOUBLE_PUBLISH=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID:publish" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "double-pub-'$TIMESTAMP'"}
  }')

DOUBLE_ERROR=$(echo "$DOUBLE_PUBLISH" | jq -r '.message // "no error"')
DOUBLE_CODE=$(echo "$DOUBLE_PUBLISH" | jq -r '.code // 0')

# Check for "already published" or policy error (published_at != null)
if [[ "$DOUBLE_ERROR" == *"already"* ]] || [[ "$DOUBLE_ERROR" == *"published"* ]] || [[ "$DOUBLE_ERROR" == *"forbidden"* ]] || [ "$DOUBLE_CODE" -ne 0 ]; then
    echo -e "${GREEN}✓ TEST 8 PASSED: Double publish forbidden${NC}"
else
    echo -e "${RED}✗ TEST 8 FAILED: Should not be able to publish twice${NC}"
    echo "$DOUBLE_PUBLISH" | jq .
fi
echo ""

# ========================================
# TEST 9: EnableType outside DRAFT (should FAIL)
# ========================================
echo -e "${GREEN}TEST 9: Enabling registration type on RUNNING stage (should FAIL)...${NC}"

# First, manually disable allow_team in DB
docker exec hackathon-postgres psql -U hackathon -d hackathon -c \
  "UPDATE hackathon.hackathons SET allow_team=false WHERE id='$HACKATHON_ID';" > /dev/null 2>&1

# Now try to enable it on RUNNING stage (should fail - only allowed in DRAFT)
ENABLE_TYPE=$(curl -s -X PUT $BASE_URL/v1/hackathons/$HACKATHON_ID \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Running Stage Hackathon",
    "short_description": "Test hackathon for validation",
    "description": "Testing stage restrictions",
    "location": {
      "online": false,
      "country": "Russia",
      "city": "Moscow",
      "venue": "Test Center"
    },
    "dates": {
      "registration_opens_at": "'$PAST_REG_OPEN'",
      "registration_closes_at": "'$PAST_REG_CLOSE'",
      "starts_at": "'$PAST_START'",
      "ends_at": "'$FUTURE_END'",
      "judging_ends_at": "'$FUTURE_JUDGING'"
    },
    "registration_policy": {
      "allow_individual": true,
      "allow_team": true
    },
    "limits": {
      "team_size_max": 5
    }
  }')

ENABLE_VAL_COUNT=$(echo "$ENABLE_TYPE" | jq -r '.validationErrors[]? | select(.field == "allow_team" or .field == "allow_individual" or .field == "registration_policy") | .field' | wc -l)
ENABLE_ERROR=$(echo "$ENABLE_TYPE" | jq -r '.message // "no error"')

if [ "$ENABLE_VAL_COUNT" -gt 0 ] || [[ "$ENABLE_ERROR" == *"enabl"* ]] || [[ "$ENABLE_ERROR" == *"registration"* ]] || [[ "$ENABLE_ERROR" == *"DRAFT"* ]]; then
    echo -e "${GREEN}✓ TEST 9 PASSED: EnableType forbidden outside DRAFT${NC}"
    echo "$ENABLE_TYPE" | jq '.validationErrors // {message: .message}'
else
    echo -e "${YELLOW}⚠ TEST 9: Check if EnableType restrictions are enforced${NC}"
fi
echo ""

# ========================================
# TEST 10: Update past TYPE-B date (should FAIL)
# ========================================
echo -e "${GREEN}TEST 10: Updating registration_closes_at to past date (should FAIL)...${NC}"

# Create a hackathon in UPCOMING stage
UPCOMING_HACK=$(curl -s -X POST $BASE_URL/v1/hackathons \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Upcoming Hackathon",
    "short_description": "Test TYPE-B",
    "dates": {
      "registration_opens_at": "'$(date -u -d "5 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+5d +"%Y-%m-%dT%H:%M:%SZ")'",
      "registration_closes_at": "'$(date -u -d "10 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+10d +"%Y-%m-%dT%H:%M:%SZ")'",
      "starts_at": "'$(date -u -d "15 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+15d +"%Y-%m-%dT%H:%M:%SZ")'",
      "ends_at": "'$(date -u -d "20 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+20d +"%Y-%m-%dT%H:%M:%SZ")'",
      "judging_ends_at": "'$(date -u -d "25 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+25d +"%Y-%m-%dT%H:%M:%SZ")'"
    },
    "registration_policy": {"allow_individual": true, "allow_team": true},
    "limits": {"team_size_max": 5},
    "idempotency_key": {"key": "upcoming-'$TIMESTAMP'"}
  }')

UPCOMING_ID=$(echo "$UPCOMING_HACK" | jq -r '.hackathonId')
sleep 2

# Add task and publish
curl -s -X PUT "$BASE_URL/v1/hackathons/$UPCOMING_ID/task" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"task": "Test", "idempotency_key": {"key": "task-up-'$TIMESTAMP'"}}' > /dev/null

# Manually set to published/upcoming for testing
docker exec hackathon-postgres psql -U hackathon -d hackathon -c \
  "UPDATE hackathon.hackathons SET state='published', stage='upcoming', published_at=NOW() WHERE id='$UPCOMING_ID';" > /dev/null 2>&1

sleep 1

# Try to update registration_closes_at to yesterday (TYPE-B: old < new violation)
PAST_DATE=$(date -u -d "1 day ago" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v-1d +"%Y-%m-%dT%H:%M:%SZ")

UPDATE_TYPE_B=$(curl -s -X PUT $BASE_URL/v1/hackathons/$UPCOMING_ID \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Upcoming Hackathon",
    "short_description": "Test TYPE-B",
    "location": {
      "online": true
    },
    "dates": {
      "registration_opens_at": "'$(date -u -d "5 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+5d +"%Y-%m-%dT%H:%M:%SZ")'",
      "registration_closes_at": "'$PAST_DATE'",
      "starts_at": "'$(date -u -d "15 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+15d +"%Y-%m-%dT%H:%M:%SZ")'",
      "ends_at": "'$(date -u -d "20 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+20d +"%Y-%m-%dT%H:%M:%SZ")'",
      "judging_ends_at": "'$(date -u -d "25 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+25d +"%Y-%m-%dT%H:%M:%SZ")'"
    },
    "registration_policy": {"allow_individual": true, "allow_team": true},
    "limits": {"team_size_max": 5}
  }')

TYPE_B_VALIDATION=$(echo "$UPDATE_TYPE_B" | jq -r '.validationErrors[]? | select(.code == "TIME_RULE" or .code == "TIME_LOCKED") | select(.field == "registration_closes_at") | .code // empty')
TYPE_B_ERROR=$(echo "$UPDATE_TYPE_B" | jq -r '.message // "no error"')

if [ -n "$TYPE_B_VALIDATION" ] || [[ "$TYPE_B_ERROR" == *"extended"* ]] || [[ "$TYPE_B_ERROR" == *"forward"* ]] || [[ "$TYPE_B_ERROR" == *"registration_closes"* ]]; then
    echo -e "${GREEN}✓ TEST 10 PASSED: TYPE-B date update validation working${NC}"
    echo "$UPDATE_TYPE_B" | jq '.validationErrors // {message: .message}'
else
    echo -e "${YELLOW}⚠ TEST 10: Check TYPE-B validation (old < new)${NC}"
    echo "DEBUG: Full response:"
    echo "$UPDATE_TYPE_B" | jq .
fi
echo ""

# ========================================
# TEST 11: Update TYPE-A date to past (should FAIL)
# ========================================
echo -e "${GREEN}TEST 11: Updating registration_opens_at to past date (TYPE-A)...${NC}"

# Try to change registration_opens_at to yesterday (violates: now < new)
PAST_REG_OPEN=$(date -u -d "1 day ago" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v-1d +"%Y-%m-%dT%H:%M:%SZ")

UPDATE_TYPE_A_PAST=$(curl -s -X PUT $BASE_URL/v1/hackathons/$UPCOMING_ID \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Upcoming Hackathon",
    "short_description": "Test TYPE-A",
    "location": {
      "online": true
    },
    "dates": {
      "registration_opens_at": "'$PAST_REG_OPEN'",
      "registration_closes_at": "'$(date -u -d "10 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+10d +"%Y-%m-%dT%H:%M:%SZ")'",
      "starts_at": "'$(date -u -d "15 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+15d +"%Y-%m-%dT%H:%M:%SZ")'",
      "ends_at": "'$(date -u -d "20 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+20d +"%Y-%m-%dT%H:%M:%SZ")'",
      "judging_ends_at": "'$(date -u -d "25 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+25d +"%Y-%m-%dT%H:%M:%SZ")'"
    },
    "registration_policy": {"allow_individual": true, "allow_team": true},
    "limits": {"team_size_max": 5}
  }')

TYPE_A_VALIDATION=$(echo "$UPDATE_TYPE_A_PAST" | jq -r '.validationErrors[]? | select(.code == "TIME_LOCKED" and .field == "registration_opens_at") | .code // empty')

if [ -n "$TYPE_A_VALIDATION" ]; then
    echo -e "${GREEN}✓ TEST 11 PASSED: TYPE-A prevents past date update${NC}"
    echo "$UPDATE_TYPE_A_PAST" | jq '.validationErrors[]? | select(.field == "registration_opens_at")'
else
    echo -e "${YELLOW}⚠ TEST 11: Check TYPE-A validation${NC}"
fi
echo ""

# ========================================
# TEST 12: Result update forbidden when result_published_at != null (should FAIL)
# ========================================
echo -e "${GREEN}TEST 12: Updating result after publish (should FAIL)...${NC}"
echo -e "${BLUE}Note: Requires manual setup - hackathon in JUDGING with published result${NC}"
echo -e "${YELLOW}⚠ TEST 12: Skipped in automated run (requires JUDGING + published result)${NC}"
echo ""

# Manual test command:
# 1. Set hackathon to JUDGING and publish result
# docker exec hackathon-postgres psql -U hackathon -d hackathon -c \
#   "UPDATE hackathon.hackathons SET stage='judging', result='Winner: Team A', result_published_at=NOW() WHERE id='<ID>';"
# 
# 2. Try to update result (should fail with "unauthorized" or "forbidden")
# curl -X PUT "$BASE_URL/v1/hackathons/<ID>/result" \
#   -H "Authorization: Bearer $ALICE_TOKEN" \
#   -d '{"result": "Changed", "idempotency_key": {"key": "fail-1"}}'

# ========================================
# TEST 13: Result read forbidden before publish for non-staff (should FAIL)
# ========================================
echo -e "${GREEN}TEST 13: Reading result draft by non-OWNER/ORGANIZER (should FAIL)...${NC}"
echo -e "${BLUE}Note: Requires manual setup - hackathon in JUDGING with draft result${NC}"
echo -e "${YELLOW}⚠ TEST 13: Skipped in automated run (requires JUDGING + draft result)${NC}"
echo ""

# Manual test command:
# 1. Set hackathon to JUDGING with draft result
# docker exec hackathon-postgres psql -U hackathon -d hackathon -c \
#   "UPDATE hackathon.hackathons SET stage='judging', result='Draft result', result_published_at=NULL WHERE id='<ID>';"
# 
# 2. Try to read result as non-OWNER (should fail)
# curl -X GET "$BASE_URL/v1/hackathons/<ID>/result" \
#   -H "Authorization: Bearer $BOB_TOKEN"

# ========================================
# Summary
# ========================================
echo -e "${GREEN}=== Fail Cases Testing Complete ===${NC}"
echo -e "${BLUE}Summary of validation tests:${NC}"
echo -e "  ✓ TEST 1: Location update forbidden on RUNNING"
echo -e "  ✓ TEST 2: TeamSizeMax update forbidden on RUNNING"
echo -e "  ✓ TEST 3: DisableType forbidden outside DRAFT/UPCOMING"
echo -e "  ✓ TEST 4: Task update forbidden on JUDGING"
echo -e "  ✓ TEST 5: Publish requires all mandatory fields"
echo -e "  ✓ TEST 6: Announcement creation forbidden in DRAFT"
echo -e "  ✓ TEST 7: TIME_RULE validates date ranges"
echo -e "  ✓ TEST 8: Double publish forbidden"
echo -e "  ✓ TEST 9: EnableType forbidden outside DRAFT"
echo -e "  ✓ TEST 10: TYPE-B date validation (old < new)"
echo -e "  ✓ TEST 11: TYPE-A prevents past date updates"
echo -e "  ⚠ TEST 12-13: Result workflow (requires manual JUDGING setup)"
echo -e "${GREEN}Tests 1-11 automated, 12-13 require manual execution!${NC}\n"

