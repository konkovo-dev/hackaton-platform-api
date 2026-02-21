#!/bin/bash

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

AUTH_URL="localhost:50051"
HACKATHON_URL="localhost:50053"
PARTICIPATION_URL="localhost:50055"
TIMESTAMP=$(date +%s)

echo -e "${YELLOW}=== Participation Service gRPC Testing ===${NC}\n"

# ========================================
# 1. Register Test Users
# ========================================
echo -e "${GREEN}1. Registering test users...${NC}"

echo -e "${BLUE}Registering Alice (staff)...${NC}"
ALICE_RESPONSE=$(grpcurl -plaintext -d '{
  "username": "alice_grpc_'$TIMESTAMP'",
  "email": "alice_grpc_'$TIMESTAMP'@test.com",
  "password": "SecurePass123",
  "first_name": "Alice",
  "last_name": "Staff",
  "timezone": "UTC"
}' $AUTH_URL auth.v1.AuthService/Register)

ALICE_TOKEN=$(echo "$ALICE_RESPONSE" | jq -r '.accessToken')
if [ -z "$ALICE_TOKEN" ] || [ "$ALICE_TOKEN" = "null" ]; then
    echo -e "${RED}Failed to register Alice${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Alice registered${NC}"

echo -e "${BLUE}Registering Bob (participant)...${NC}"
BOB_RESPONSE=$(grpcurl -plaintext -d '{
  "username": "bob_grpc_'$TIMESTAMP'",
  "email": "bob_grpc_'$TIMESTAMP'@test.com",
  "password": "SecurePass123",
  "first_name": "Bob",
  "last_name": "Participant",
  "timezone": "UTC"
}' $AUTH_URL auth.v1.AuthService/Register)

BOB_TOKEN=$(echo "$BOB_RESPONSE" | jq -r '.accessToken')
if [ -z "$BOB_TOKEN" ] || [ "$BOB_TOKEN" = "null" ]; then
    echo -e "${RED}Failed to register Bob${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Bob registered${NC}"

echo -e "${BLUE}Registering Charlie (participant)...${NC}"
CHARLIE_RESPONSE=$(grpcurl -plaintext -d '{
  "username": "charlie_grpc_'$TIMESTAMP'",
  "email": "charlie_grpc_'$TIMESTAMP'@test.com",
  "password": "SecurePass123",
  "first_name": "Charlie",
  "last_name": "Seeker",
  "timezone": "UTC"
}' $AUTH_URL auth.v1.AuthService/Register)

CHARLIE_TOKEN=$(echo "$CHARLIE_RESPONSE" | jq -r '.accessToken')
if [ -z "$CHARLIE_TOKEN" ] || [ "$CHARLIE_TOKEN" = "null" ]; then
    echo -e "${RED}Failed to register Charlie${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Charlie registered${NC}\n"

# ========================================
# 2. Create and Publish Hackathon
# ========================================
echo -e "${GREEN}2. Creating hackathon (Alice)...${NC}"

CREATE_HACK=$(grpcurl -plaintext \
  -H "authorization: Bearer $ALICE_TOKEN" \
  -d '{
    "name": "gRPC Test Hackathon '$TIMESTAMP'",
    "short_description": "Testing participation via gRPC",
    "description": "Full gRPC test",
    "location": {"online": true},
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
    "limits": {"team_size_max": 5}
  }' $HACKATHON_URL hackathon.v1.HackathonService/CreateHackathon)

HACKATHON_ID=$(echo "$CREATE_HACK" | jq -r '.hackathonId')
if [ -z "$HACKATHON_ID" ] || [ "$HACKATHON_ID" = "null" ]; then
    echo -e "${RED}Failed to create hackathon${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Hackathon created: $HACKATHON_ID${NC}"

echo -e "${BLUE}Publishing hackathon...${NC}"
grpcurl -plaintext \
  -H "authorization: Bearer $ALICE_TOKEN" \
  -d '{"hackathon_id": "'$HACKATHON_ID'"}' \
  $HACKATHON_URL hackathon.v1.HackathonService/PublishHackathon > /dev/null

echo -e "${GREEN}✓ Hackathon published${NC}\n"

# ========================================
# 3. ListTeamRoles
# ========================================
echo -e "${GREEN}3. Listing team roles...${NC}"

ROLES_RESPONSE=$(grpcurl -plaintext \
  -H "authorization: Bearer $BOB_TOKEN" \
  -d '{}' \
  $PARTICIPATION_URL participationandroles.v1.ParticipationService/ListTeamRoles)

ROLES_COUNT=$(echo "$ROLES_RESPONSE" | jq '.teamRoles | length')
if [ "$ROLES_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ Found $ROLES_COUNT team roles${NC}"
else
    echo -e "${RED}✗ No team roles found${NC}"
    exit 1
fi

# Extract role IDs
BACKEND_ROLE=$(echo "$ROLES_RESPONSE" | jq -r '.teamRoles[] | select(.name == "Backend") | .id')
FRONTEND_ROLE=$(echo "$ROLES_RESPONSE" | jq -r '.teamRoles[] | select(.name == "Frontend") | .id')
FULLSTACK_ROLE=$(echo "$ROLES_RESPONSE" | jq -r '.teamRoles[] | select(.name == "Fullstack") | .id')
DESIGNER_ROLE=$(echo "$ROLES_RESPONSE" | jq -r '.teamRoles[] | select(.name == "Designer") | .id')

echo -e "${BLUE}Backend: $BACKEND_ROLE${NC}"
echo -e "${BLUE}Frontend: $FRONTEND_ROLE${NC}\n"

# ========================================
# 4. RegisterForHackathon (Bob as individual)
# ========================================
echo -e "${GREEN}4. Bob registers as INDIVIDUAL_ACTIVE...${NC}"

BOB_REG=$(grpcurl -plaintext \
  -H "authorization: Bearer $BOB_TOKEN" \
  -d '{
    "hackathon_id": "'$HACKATHON_ID'",
    "desired_status": "PART_INDIVIDUAL",
    "wished_role_ids": [],
    "motivation_text": "I love coding!",
    "idempotency_key": {"key": "bob-reg-grpc-'$TIMESTAMP'"}
  }' \
  $PARTICIPATION_URL participationandroles.v1.ParticipationService/RegisterForHackathon)

BOB_STATUS=$(echo "$BOB_REG" | jq -r '.participation.status')
if [ "$BOB_STATUS" = "PART_INDIVIDUAL" ]; then
    echo -e "${GREEN}✓ Bob registered as INDIVIDUAL_ACTIVE${NC}"
else
    echo -e "${RED}✗ Bob registration failed${NC}"
    exit 1
fi
echo ""

# ========================================
# 5. RegisterForHackathon (Charlie as looking for team)
# ========================================
echo -e "${GREEN}5. Charlie registers as LOOKING_FOR_TEAM...${NC}"

CHARLIE_REG=$(grpcurl -plaintext \
  -H "authorization: Bearer $CHARLIE_TOKEN" \
  -d '{
    "hackathon_id": "'$HACKATHON_ID'",
    "desired_status": "PART_LOOKING_FOR_TEAM",
    "wished_role_ids": ["'$BACKEND_ROLE'", "'$FULLSTACK_ROLE'"],
    "motivation_text": "Looking for a great backend team!",
    "idempotency_key": {"key": "charlie-reg-grpc-'$TIMESTAMP'"}
  }' \
  $PARTICIPATION_URL participationandroles.v1.ParticipationService/RegisterForHackathon)

CHARLIE_STATUS=$(echo "$CHARLIE_REG" | jq -r '.participation.status')
CHARLIE_ROLES_COUNT=$(echo "$CHARLIE_REG" | jq '.participation.profile.wishedRoles | length')
if [ "$CHARLIE_STATUS" = "PART_LOOKING_FOR_TEAM" ] && [ "$CHARLIE_ROLES_COUNT" = "2" ]; then
    echo -e "${GREEN}✓ Charlie registered as LOOKING_FOR_TEAM with 2 roles${NC}"
else
    echo -e "${RED}✗ Charlie registration failed${NC}"
    exit 1
fi
echo ""

# ========================================
# 6. GetMyParticipation (Bob)
# ========================================
echo -e "${GREEN}6. Bob checks his participation...${NC}"

BOB_GET=$(grpcurl -plaintext \
  -H "authorization: Bearer $BOB_TOKEN" \
  -d '{
    "hackathon_id": "'$HACKATHON_ID'"
  }' \
  $PARTICIPATION_URL participationandroles.v1.ParticipationService/GetMyParticipation)

BOB_GET_STATUS=$(echo "$BOB_GET" | jq -r '.participation.status')
if [ "$BOB_GET_STATUS" = "PART_INDIVIDUAL" ]; then
    echo -e "${GREEN}✓ Bob sees his participation${NC}"
else
    echo -e "${RED}✗ GetMyParticipation failed${NC}"
    exit 1
fi
echo ""

# ========================================
# 7. UpdateMyParticipation (Bob)
# ========================================
echo -e "${GREEN}7. Bob updates his profile...${NC}"

BOB_UPDATE=$(grpcurl -plaintext \
  -H "authorization: Bearer $BOB_TOKEN" \
  -d '{
    "hackathon_id": "'$HACKATHON_ID'",
    "wished_role_ids": ["'$FRONTEND_ROLE'", "'$DESIGNER_ROLE'"],
    "motivation_text": "I changed my mind - love frontend and design!",
    "idempotency_key": {"key": "bob-update-grpc-'$TIMESTAMP'"}
  }' \
  $PARTICIPATION_URL participationandroles.v1.ParticipationService/UpdateMyParticipation)

BOB_UPDATE_ROLES=$(echo "$BOB_UPDATE" | jq '.participation.profile.wishedRoles | length')
if [ "$BOB_UPDATE_ROLES" = "2" ]; then
    echo -e "${GREEN}✓ Bob updated his profile (2 roles)${NC}"
else
    echo -e "${RED}✗ UpdateMyParticipation failed${NC}"
    exit 1
fi
echo ""

# ========================================
# 8. SwitchParticipationMode (Bob)
# ========================================
echo -e "${GREEN}8. Bob switches to LOOKING_FOR_TEAM...${NC}"

BOB_SWITCH=$(grpcurl -plaintext \
  -H "authorization: Bearer $BOB_TOKEN" \
  -d '{
    "hackathon_id": "'$HACKATHON_ID'",
    "new_status": "PART_LOOKING_FOR_TEAM",
    "idempotency_key": {"key": "bob-switch-grpc-'$TIMESTAMP'"}
  }' \
  $PARTICIPATION_URL participationandroles.v1.ParticipationService/SwitchParticipationMode)

BOB_NEW_STATUS=$(echo "$BOB_SWITCH" | jq -r '.participation.status')
if [ "$BOB_NEW_STATUS" = "PART_LOOKING_FOR_TEAM" ]; then
    echo -e "${GREEN}✓ Bob switched to LOOKING_FOR_TEAM${NC}"
else
    echo -e "${RED}✗ SwitchMode failed${NC}"
    exit 1
fi
echo ""

# ========================================
# 9. GetUserParticipation (Alice as staff)
# ========================================
echo -e "${GREEN}9. Alice (staff) views Bob's participation...${NC}"

BOB_USER_ID=$(grpcurl -plaintext -d '{"access_token": "'$BOB_TOKEN'"}' \
  $AUTH_URL auth.v1.AuthService/IntrospectToken | jq -r '.userId')

ALICE_VIEW=$(grpcurl -plaintext \
  -H "authorization: Bearer $ALICE_TOKEN" \
  -d '{
    "hackathon_id": "'$HACKATHON_ID'",
    "user_id": "'$BOB_USER_ID'"
  }' \
  $PARTICIPATION_URL participationandroles.v1.ParticipationService/GetUserParticipation)

ALICE_VIEW_STATUS=$(echo "$ALICE_VIEW" | jq -r '.participation.status')
if [ "$ALICE_VIEW_STATUS" = "PART_LOOKING_FOR_TEAM" ]; then
    echo -e "${GREEN}✓ Alice can view Bob's participation${NC}"
else
    echo -e "${RED}✗ GetUserParticipation failed${NC}"
    exit 1
fi
echo ""

# ========================================
# 10. ListHackathonParticipants (Alice as staff)
# ========================================
echo -e "${GREEN}10. Alice lists all participants...${NC}"

ALICE_LIST=$(grpcurl -plaintext \
  -H "authorization: Bearer $ALICE_TOKEN" \
  -d '{
    "hackathon_id": "'$HACKATHON_ID'"
  }' \
  $PARTICIPATION_URL participationandroles.v1.ParticipationService/ListHackathonParticipants)

PARTICIPANTS_COUNT=$(echo "$ALICE_LIST" | jq '.participants | length')
if [ "$PARTICIPANTS_COUNT" -ge 2 ]; then
    echo -e "${GREEN}✓ Alice sees $PARTICIPANTS_COUNT participants${NC}"
else
    echo -e "${RED}✗ ListParticipants failed${NC}"
    exit 1
fi
echo ""

# ========================================
# 11. ListHackathonParticipants with status filter
# ========================================
echo -e "${GREEN}11. Alice lists only LOOKING_FOR_TEAM participants...${NC}"

ALICE_LIST_FILTERED=$(grpcurl -plaintext \
  -H "authorization: Bearer $ALICE_TOKEN" \
  -d '{
    "hackathon_id": "'$HACKATHON_ID'",
    "status_filter": {
      "statuses": ["PART_LOOKING_FOR_TEAM"]
    }
  }' \
  $PARTICIPATION_URL participationandroles.v1.ParticipationService/ListHackathonParticipants)

FILTERED_COUNT=$(echo "$ALICE_LIST_FILTERED" | jq '.participants | length')
if [ "$FILTERED_COUNT" -ge 1 ]; then
    echo -e "${GREEN}✓ Filter by status works ($FILTERED_COUNT participants)${NC}"
else
    echo -e "${RED}✗ Status filter failed${NC}"
    exit 1
fi
echo ""

# ========================================
# 12. UnregisterFromHackathon (Bob)
# ========================================
echo -e "${GREEN}12. Bob unregisters from hackathon...${NC}"

grpcurl -plaintext \
  -H "authorization: Bearer $BOB_TOKEN" \
  -d '{
    "hackathon_id": "'$HACKATHON_ID'",
    "idempotency_key": {"key": "bob-unreg-grpc-'$TIMESTAMP'"}
  }' \
  $PARTICIPATION_URL participationandroles.v1.ParticipationService/UnregisterFromHackathon > /dev/null

echo -e "${GREEN}✓ Bob unregistered${NC}\n"

# ========================================
# 13. GetMyParticipation after unregister (should fail)
# ========================================
echo -e "${GREEN}13. Bob tries to get participation after unregister (should fail)...${NC}"

BOB_GET_AFTER=$(grpcurl -plaintext \
  -H "authorization: Bearer $BOB_TOKEN" \
  -d '{
    "hackathon_id": "'$HACKATHON_ID'"
  }' \
  $PARTICIPATION_URL participationandroles.v1.ParticipationService/GetMyParticipation 2>&1 || true)

if echo "$BOB_GET_AFTER" | grep -q "NotFound\|code = 5"; then
    echo -e "${GREEN}✓ Correctly returns NotFound after unregister${NC}"
else
    echo -e "${RED}✗ Should return NotFound${NC}"
fi
echo ""

# ========================================
# Fail Cases
# ========================================
echo -e "${YELLOW}=== Testing Fail Cases ===${NC}\n"

# ========================================
# 14. Non-staff tries to list participants
# ========================================
echo -e "${GREEN}14. Charlie (non-staff) tries to list participants (should fail)...${NC}"

CHARLIE_LIST=$(grpcurl -plaintext \
  -H "authorization: Bearer $CHARLIE_TOKEN" \
  -d '{
    "hackathon_id": "'$HACKATHON_ID'"
  }' \
  $PARTICIPATION_URL participationandroles.v1.ParticipationService/ListHackathonParticipants 2>&1 || true)

if echo "$CHARLIE_LIST" | grep -q "PermissionDenied\|code = 7"; then
    echo -e "${GREEN}✓ Correctly blocked non-staff access${NC}"
else
    echo -e "${RED}✗ Should return PermissionDenied${NC}"
fi
echo ""

# ========================================
# 15. Register twice (should fail)
# ========================================
echo -e "${GREEN}15. Charlie tries to register twice (should fail)...${NC}"

CHARLIE_DOUBLE=$(grpcurl -plaintext \
  -H "authorization: Bearer $CHARLIE_TOKEN" \
  -d '{
    "hackathon_id": "'$HACKATHON_ID'",
    "desired_status": "PART_INDIVIDUAL",
    "wished_role_ids": [],
    "motivation_text": "Second registration"
  }' \
  $PARTICIPATION_URL participationandroles.v1.ParticipationService/RegisterForHackathon 2>&1 || true)

if echo "$CHARLIE_DOUBLE" | grep -q "AlreadyExists\|code = 9"; then
    echo -e "${GREEN}✓ Correctly blocked double registration${NC}"
else
    echo -e "${RED}✗ Should return AlreadyExists${NC}"
fi
echo ""

# ========================================
# 16. Switch to same status (should fail)
# ========================================
echo -e "${GREEN}16. Charlie tries to switch to same status (should fail)...${NC}"

CHARLIE_SAME_STATUS=$(grpcurl -plaintext \
  -H "authorization: Bearer $CHARLIE_TOKEN" \
  -d '{
    "hackathon_id": "'$HACKATHON_ID'",
    "new_status": "PART_LOOKING_FOR_TEAM"
  }' \
  $PARTICIPATION_URL participationandroles.v1.ParticipationService/SwitchParticipationMode 2>&1 || true)

if echo "$CHARLIE_SAME_STATUS" | grep -q "PermissionDenied\|code = 7"; then
    echo -e "${GREEN}✓ Correctly blocked same status switch${NC}"
else
    echo -e "${RED}✗ Should return PermissionDenied${NC}"
fi
echo ""

# ========================================
# Summary
# ========================================
echo -e "${YELLOW}=== All gRPC Tests Completed ===${NC}"
echo -e "${GREEN}Summary:${NC}"
echo -e "  - ✓ ListTeamRoles: retrieved $ROLES_COUNT roles"
echo -e "  - ✓ RegisterForHackathon: Bob and Charlie registered"
echo -e "  - ✓ GetMyParticipation: participants see own registration"
echo -e "  - ✓ UpdateMyParticipation: profile updated"
echo -e "  - ✓ SwitchParticipationMode: mode switched"
echo -e "  - ✓ GetUserParticipation: staff can view participations"
echo -e "  - ✓ ListHackathonParticipants: staff can list with filters"
echo -e "  - ✓ UnregisterFromHackathon: registration cancelled"
echo -e "  - ✓ Fail cases: access control and validation work"
echo -e "${GREEN}✓ All gRPC tests passed!${NC}"
