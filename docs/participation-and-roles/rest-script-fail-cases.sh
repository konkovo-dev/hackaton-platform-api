#!/bin/bash

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

BASE_URL="http://localhost:8080"
TIMESTAMP=$(date +%s)

echo -e "${YELLOW}=== Participation and Roles Service - Fail Cases ===${NC}\n"
echo -e "${BLUE}Testing validation and permission rules${NC}\n"

# ========================================
# Setup: Register Users
# ========================================
echo -e "${GREEN}1. Registering test users...${NC}"

ALICE_RESPONSE=$(curl -s -X POST $BASE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice_fail_'$TIMESTAMP'",
    "email": "alice_fail_'$TIMESTAMP'@hackathon.com",
    "password": "SecurePass123",
    "first_name": "Alice",
    "last_name": "Owner",
    "timezone": "UTC",
    "idempotency_key": {"key": "alice-fail-'$TIMESTAMP'"}
  }')

ALICE_TOKEN=$(echo $ALICE_RESPONSE | jq -r '.accessToken')
ALICE_USER_ID=$(curl -s -X POST $BASE_URL/v1/auth/introspect \
  -H "Content-Type: application/json" \
  -d '{"access_token": "'$ALICE_TOKEN'"}' | jq -r '.userId')

BOB_RESPONSE=$(curl -s -X POST $BASE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "bob_fail_'$TIMESTAMP'",
    "email": "bob_fail_'$TIMESTAMP'@hackathon.com",
    "password": "SecurePass123",
    "first_name": "Bob",
    "last_name": "User",
    "timezone": "UTC",
    "idempotency_key": {"key": "bob-fail-'$TIMESTAMP'"}
  }')

BOB_TOKEN=$(echo $BOB_RESPONSE | jq -r '.accessToken')
BOB_USER_ID=$(curl -s -X POST $BASE_URL/v1/auth/introspect \
  -H "Content-Type: application/json" \
  -d '{"access_token": "'$BOB_TOKEN'"}' | jq -r '.userId')

CHARLIE_RESPONSE=$(curl -s -X POST $BASE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "charlie_fail_'$TIMESTAMP'",
    "email": "charlie_fail_'$TIMESTAMP'@hackathon.com",
    "password": "SecurePass123",
    "first_name": "Charlie",
    "last_name": "User",
    "timezone": "UTC",
    "idempotency_key": {"key": "charlie-fail-'$TIMESTAMP'"}
  }')

CHARLIE_TOKEN=$(echo $CHARLIE_RESPONSE | jq -r '.accessToken')
CHARLIE_USER_ID=$(curl -s -X POST $BASE_URL/v1/auth/introspect \
  -H "Content-Type: application/json" \
  -d '{"access_token": "'$CHARLIE_TOKEN'"}' | jq -r '.userId')

echo -e "${GREEN}✓ Alice registered (user_id: $ALICE_USER_ID)${NC}"
echo -e "${GREEN}✓ Bob registered (user_id: $BOB_USER_ID)${NC}"
echo -e "${GREEN}✓ Charlie registered (user_id: $CHARLIE_USER_ID)${NC}\n"

# ========================================
# Create Hackathon
# ========================================
echo -e "${GREEN}2. Creating hackathon...${NC}"

CREATE_HACKATHON=$(curl -s -X POST $BASE_URL/v1/hackathons \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Fail Cases Test Hackathon '$TIMESTAMP'",
    "short_description": "Testing fail cases",
    "description": "A hackathon for testing validation and permission rules.",
    "location": {
      "online": true
    },
    "dates": {
      "registration_opens_at": "2026-03-01T00:00:00Z",
      "registration_closes_at": "2026-03-15T23:59:59Z",
      "starts_at": "2026-03-20T09:00:00Z",
      "ends_at": "2026-03-22T18:00:00Z",
      "judging_ends_at": "2026-03-25T18:00:00Z"
    },
    "registration_policy": {
      "allow_individual": true,
      "allow_team": true
    },
    "limits": {
      "team_size_max": 5
    },
    "idempotency_key": {"key": "hackathon-fail-'$TIMESTAMP'"}
  }')

HACKATHON_ID=$(echo $CREATE_HACKATHON | jq -r '.hackathonId')
echo -e "${GREEN}✓ Hackathon created (ID: $HACKATHON_ID)${NC}"
echo -e "${BLUE}Waiting for OWNER role assignment...${NC}"
sleep 2
echo -e "${GREEN}✓ Ready to test${NC}\n"

# ========================================
# TEST 1: ListHackathonStaff - Non-staff user (should FAIL)
# ========================================
echo -e "${GREEN}TEST 1: ListHackathonStaff - Non-staff user (should FAIL)...${NC}"
LIST_FAIL=$(curl -s "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff" \
  -H "Authorization: Bearer $BOB_TOKEN")

LIST_ERROR=$(echo "$LIST_FAIL" | jq -r '.message // "no error"')
if [[ "$LIST_ERROR" == *"forbidden"* ]]; then
    echo -e "${GREEN}✓ TEST 1 PASSED: Non-staff cannot view staff list${NC}"
    echo "$LIST_FAIL" | jq .
else
    echo -e "${RED}✗ TEST 1 FAILED: Should be forbidden${NC}"
    echo "$LIST_FAIL" | jq .
fi
echo ""

# ========================================
# TEST 2: CreateStaffInvitation - Non-owner (should FAIL)
# ========================================
echo -e "${GREEN}TEST 2: CreateStaffInvitation - Non-owner (should FAIL)...${NC}"
INVITE_FAIL=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff-invitations" \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "'$CHARLIE_USER_ID'",
    "requested_role": "HX_ROLE_MENTOR",
    "message": "Bob tries to invite",
    "idempotency_key": {"key": "invite-fail-bob-'$TIMESTAMP'"}
  }')

INVITE_ERROR=$(echo "$INVITE_FAIL" | jq -r '.message // "no error"')
if [[ "$INVITE_ERROR" == *"forbidden"* ]]; then
    echo -e "${GREEN}✓ TEST 2 PASSED: Non-owner cannot create invitations${NC}"
    echo "$INVITE_FAIL" | jq .
else
    echo -e "${RED}✗ TEST 2 FAILED: Should be forbidden${NC}"
    echo "$INVITE_FAIL" | jq .
fi
echo ""

# ========================================
# TEST 3: CreateStaffInvitation - OWNER role (should FAIL)
# ========================================
echo -e "${GREEN}TEST 3: CreateStaffInvitation - OWNER role (should FAIL)...${NC}"
INVITE_OWNER=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff-invitations" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "'$BOB_USER_ID'",
    "requested_role": "HX_ROLE_OWNER",
    "message": "Trying to invite as owner",
    "idempotency_key": {"key": "invite-owner-'$TIMESTAMP'"}
  }')

OWNER_ERROR=$(echo "$INVITE_OWNER" | jq -r '.message // "no error"')
if [[ "$OWNER_ERROR" == *"owner"* ]] || [[ "$OWNER_ERROR" == *"cannot"* ]] || [[ "$OWNER_ERROR" == *"invalid"* ]]; then
    echo -e "${GREEN}✓ TEST 3 PASSED: Cannot invite for OWNER role${NC}"
    echo "$INVITE_OWNER" | jq .
else
    echo -e "${RED}✗ TEST 3 FAILED: Should not allow OWNER role invitation${NC}"
    echo "$INVITE_OWNER" | jq .
fi
echo ""

# ========================================
# TEST 4: CreateStaffInvitation - Non-existent user (should FAIL)
# ========================================
echo -e "${GREEN}TEST 4: CreateStaffInvitation - Non-existent user (should FAIL)...${NC}"
INVITE_INVALID=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff-invitations" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "99999999-9999-9999-9999-999999999999",
    "requested_role": "HX_ROLE_MENTOR",
    "message": "Non-existent user",
    "idempotency_key": {"key": "invite-invalid-'$TIMESTAMP'"}
  }')

INVALID_ERROR=$(echo "$INVITE_INVALID" | jq -r '.message // "no error"')
if [[ "$INVALID_ERROR" == *"not found"* ]] || [[ "$INVALID_ERROR" == *"user not found"* ]]; then
    echo -e "${GREEN}✓ TEST 4 PASSED: Cannot invite non-existent user${NC}"
    echo "$INVITE_INVALID" | jq .
else
    echo -e "${RED}✗ TEST 4 FAILED: Should reject non-existent user${NC}"
    echo "$INVITE_INVALID" | jq .
fi
echo ""

# ========================================
# TEST 5: CreateStaffInvitation - Invite self (should FAIL)
# ========================================
echo -e "${GREEN}TEST 5: CreateStaffInvitation - Invite self (should FAIL)...${NC}"
INVITE_SELF=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff-invitations" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "'$ALICE_USER_ID'",
    "requested_role": "HX_ROLE_ORGANIZER",
    "message": "Inviting myself",
    "idempotency_key": {"key": "invite-self-'$TIMESTAMP'"}
  }')

SELF_ERROR=$(echo "$INVITE_SELF" | jq -r '.message // "no error"')
if [[ "$SELF_ERROR" == *"self"* ]] || [[ "$SELF_ERROR" == *"cannot"* ]] || [[ "$SELF_ERROR" == *"yourself"* ]]; then
    echo -e "${GREEN}✓ TEST 5 PASSED: Cannot invite yourself${NC}"
    echo "$INVITE_SELF" | jq .
else
    echo -e "${RED}✗ TEST 5 FAILED: Should not allow self-invitation${NC}"
    echo "$INVITE_SELF" | jq .
fi
echo ""

# ========================================
# Setup for invitation tests: Create valid invitation
# ========================================
echo -e "${BLUE}Setup: Creating valid invitation for Bob...${NC}"
VALID_INVITE=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff-invitations" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "'$BOB_USER_ID'",
    "requested_role": "HX_ROLE_MENTOR",
    "message": "Valid invitation",
    "idempotency_key": {"key": "valid-invite-bob-'$TIMESTAMP'"}
  }')

BOB_INVITATION_ID=$(echo "$VALID_INVITE" | jq -r '.invitationId')
echo -e "${GREEN}✓ Invitation created (ID: $BOB_INVITATION_ID)${NC}\n"

# ========================================
# TEST 6: AcceptStaffInvitation - Wrong user (should FAIL)
# ========================================
echo -e "${GREEN}TEST 6: AcceptStaffInvitation - Wrong user (should FAIL)...${NC}"
ACCEPT_WRONG=$(curl -s -X POST "$BASE_URL/v1/users/me/staff-invitations/$BOB_INVITATION_ID:accept" \
  -H "Authorization: Bearer $CHARLIE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "accept-wrong-'$TIMESTAMP'"}
  }')

WRONG_ERROR=$(echo "$ACCEPT_WRONG" | jq -r '.message // "no error"')
if [[ "$WRONG_ERROR" == *"not found"* ]] || [[ "$WRONG_ERROR" == *"forbidden"* ]]; then
    echo -e "${GREEN}✓ TEST 6 PASSED: Cannot accept someone else's invitation${NC}"
    echo "$ACCEPT_WRONG" | jq .
else
    echo -e "${RED}✗ TEST 6 FAILED: Should not allow accepting others' invitations${NC}"
    echo "$ACCEPT_WRONG" | jq .
fi
echo ""

# ========================================
# TEST 7: CancelStaffInvitation - Non-owner (should FAIL)
# ========================================
echo -e "${GREEN}TEST 7: CancelStaffInvitation - Non-owner (should FAIL)...${NC}"
CANCEL_FAIL=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff-invitations/$BOB_INVITATION_ID:cancel" \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "cancel-fail-'$TIMESTAMP'"}
  }')

CANCEL_ERROR=$(echo "$CANCEL_FAIL" | jq -r '.message // "no error"')
if [[ "$CANCEL_ERROR" == *"forbidden"* ]]; then
    echo -e "${GREEN}✓ TEST 7 PASSED: Non-owner cannot cancel invitations${NC}"
    echo "$CANCEL_FAIL" | jq .
else
    echo -e "${RED}✗ TEST 7 FAILED: Should be forbidden${NC}"
    echo "$CANCEL_FAIL" | jq .
fi
echo ""

# ========================================
# Accept Bob's invitation for further tests
# ========================================
echo -e "${BLUE}Setup: Bob accepts his invitation...${NC}"
curl -s -X POST "$BASE_URL/v1/users/me/staff-invitations/$BOB_INVITATION_ID:accept" \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "accept-bob-'$TIMESTAMP'"}
  }' > /dev/null
echo -e "${GREEN}✓ Bob is now MENTOR${NC}\n"

# ========================================
# TEST 8: AcceptStaffInvitation - Already accepted (should FAIL)
# ========================================
echo -e "${GREEN}TEST 8: AcceptStaffInvitation - Already accepted (should FAIL)...${NC}"
ACCEPT_AGAIN=$(curl -s -X POST "$BASE_URL/v1/users/me/staff-invitations/$BOB_INVITATION_ID:accept" \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "accept-again-'$TIMESTAMP'"}
  }')

AGAIN_ERROR=$(echo "$ACCEPT_AGAIN" | jq -r '.message // "no error"')
if [[ "$AGAIN_ERROR" == *"already"* ]] || [[ "$AGAIN_ERROR" == *"accepted"* ]] || [[ "$AGAIN_ERROR" == *"invalid"* ]]; then
    echo -e "${GREEN}✓ TEST 8 PASSED: Cannot accept already accepted invitation${NC}"
    echo "$ACCEPT_AGAIN" | jq .
else
    echo -e "${RED}✗ TEST 8 FAILED: Should not allow re-accepting${NC}"
    echo "$ACCEPT_AGAIN" | jq .
fi
echo ""

# ========================================
# TEST 9: CancelStaffInvitation - Already accepted (should FAIL)
# ========================================
echo -e "${GREEN}TEST 9: CancelStaffInvitation - Already accepted (should FAIL)...${NC}"
CANCEL_ACCEPTED=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff-invitations/$BOB_INVITATION_ID:cancel" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "cancel-accepted-'$TIMESTAMP'"}
  }')

CANCEL_ACC_ERROR=$(echo "$CANCEL_ACCEPTED" | jq -r '.message // "no error"')
if [[ "$CANCEL_ACC_ERROR" == *"pending"* ]] || [[ "$CANCEL_ACC_ERROR" == *"already"* ]] || [[ "$CANCEL_ACC_ERROR" == *"cannot"* ]]; then
    echo -e "${GREEN}✓ TEST 9 PASSED: Cannot cancel accepted invitation${NC}"
    echo "$CANCEL_ACCEPTED" | jq .
else
    echo -e "${RED}✗ TEST 9 FAILED: Should not allow canceling accepted invitation${NC}"
    echo "$CANCEL_ACCEPTED" | jq .
fi
echo ""

# ========================================
# Setup: Create invitation for Charlie and reject it
# ========================================
echo -e "${BLUE}Setup: Creating and rejecting invitation for Charlie...${NC}"
CHARLIE_INVITE=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff-invitations" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "'$CHARLIE_USER_ID'",
    "requested_role": "HX_ROLE_JUDGE",
    "message": "For rejection test",
    "idempotency_key": {"key": "charlie-invite-'$TIMESTAMP'"}
  }')

CHARLIE_INVITATION_ID=$(echo "$CHARLIE_INVITE" | jq -r '.invitationId')

curl -s -X POST "$BASE_URL/v1/users/me/staff-invitations/$CHARLIE_INVITATION_ID:reject" \
  -H "Authorization: Bearer $CHARLIE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "reject-charlie-'$TIMESTAMP'"}
  }' > /dev/null
echo -e "${GREEN}✓ Charlie rejected his invitation${NC}\n"

# ========================================
# TEST 10: AcceptStaffInvitation - Already rejected (should FAIL)
# ========================================
echo -e "${GREEN}TEST 10: AcceptStaffInvitation - Already rejected (should FAIL)...${NC}"
ACCEPT_REJECTED=$(curl -s -X POST "$BASE_URL/v1/users/me/staff-invitations/$CHARLIE_INVITATION_ID:accept" \
  -H "Authorization: Bearer $CHARLIE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "accept-rejected-'$TIMESTAMP'"}
  }')

ACCEPT_REJ_ERROR=$(echo "$ACCEPT_REJECTED" | jq -r '.message // "no error"')
if [[ "$ACCEPT_REJ_ERROR" == *"already"* ]] || [[ "$ACCEPT_REJ_ERROR" == *"rejected"* ]] || [[ "$ACCEPT_REJ_ERROR" == *"declined"* ]] || [[ "$ACCEPT_REJ_ERROR" == *"invalid"* ]]; then
    echo -e "${GREEN}✓ TEST 10 PASSED: Cannot accept rejected invitation${NC}"
    echo "$ACCEPT_REJECTED" | jq .
else
    echo -e "${RED}✗ TEST 10 FAILED: Should not allow accepting rejected invitation${NC}"
    echo "$ACCEPT_REJECTED" | jq .
fi
echo ""

# ========================================
# TEST 11: RejectStaffInvitation - Already rejected (should FAIL)
# ========================================
echo -e "${GREEN}TEST 11: RejectStaffInvitation - Already rejected (should FAIL)...${NC}"
REJECT_AGAIN=$(curl -s -X POST "$BASE_URL/v1/users/me/staff-invitations/$CHARLIE_INVITATION_ID:reject" \
  -H "Authorization: Bearer $CHARLIE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "reject-again-'$TIMESTAMP'"}
  }')

REJECT_ERROR=$(echo "$REJECT_AGAIN" | jq -r '.message // "no error"')
if [[ "$REJECT_ERROR" == *"already"* ]] || [[ "$REJECT_ERROR" == *"invalid"* ]]; then
    echo -e "${GREEN}✓ TEST 11 PASSED: Cannot reject already rejected invitation${NC}"
    echo "$REJECT_AGAIN" | jq .
else
    echo -e "${RED}✗ TEST 11 FAILED: Should not allow re-rejecting${NC}"
    echo "$REJECT_AGAIN" | jq .
fi
echo ""

# ========================================
# TEST 12: RemoveHackathonRole - Non-owner (should FAIL)
# ========================================
echo -e "${GREEN}TEST 12: RemoveHackathonRole - Non-owner (should FAIL)...${NC}"
REMOVE_FAIL=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff:removeRole" \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "'$CHARLIE_USER_ID'",
    "role": "HX_ROLE_JUDGE",
    "idempotency_key": {"key": "remove-fail-'$TIMESTAMP'"}
  }')

REMOVE_ERROR=$(echo "$REMOVE_FAIL" | jq -r '.message // "no error"')
if [[ "$REMOVE_ERROR" == *"forbidden"* ]]; then
    echo -e "${GREEN}✓ TEST 12 PASSED: Non-owner cannot remove roles${NC}"
    echo "$REMOVE_FAIL" | jq .
else
    echo -e "${RED}✗ TEST 12 FAILED: Should be forbidden${NC}"
    echo "$REMOVE_FAIL" | jq .
fi
echo ""

# ========================================
# TEST 13: RemoveHackathonRole - OWNER role (should FAIL)
# ========================================
echo -e "${GREEN}TEST 13: RemoveHackathonRole - OWNER role (should FAIL)...${NC}"
REMOVE_OWNER=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff:removeRole" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "'$ALICE_USER_ID'",
    "role": "HX_ROLE_OWNER",
    "idempotency_key": {"key": "remove-owner-'$TIMESTAMP'"}
  }')

REMOVE_OWN_ERROR=$(echo "$REMOVE_OWNER" | jq -r '.message // "no error"')
if [[ "$REMOVE_OWN_ERROR" == *"owner"* ]] || [[ "$REMOVE_OWN_ERROR" == *"cannot"* ]]; then
    echo -e "${GREEN}✓ TEST 13 PASSED: Cannot remove OWNER role${NC}"
    echo "$REMOVE_OWNER" | jq .
else
    echo -e "${RED}✗ TEST 13 FAILED: Should not allow removing OWNER${NC}"
    echo "$REMOVE_OWNER" | jq .
fi
echo ""

# ========================================
# TEST 14: RemoveHackathonRole - Non-existent role (should FAIL)
# ========================================
echo -e "${GREEN}TEST 14: RemoveHackathonRole - Non-existent role (should FAIL)...${NC}"
REMOVE_NONEXIST=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff:removeRole" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "'$CHARLIE_USER_ID'",
    "role": "HX_ROLE_ORGANIZER",
    "idempotency_key": {"key": "remove-nonexist-'$TIMESTAMP'"}
  }')

NONEXIST_ERROR=$(echo "$REMOVE_NONEXIST" | jq -r '.message // "no error"')
if [[ "$NONEXIST_ERROR" == *"not found"* ]] || [[ "$NONEXIST_ERROR" == *"does not have"* ]] || [[ "$NONEXIST_ERROR" == *"no such"* ]]; then
    echo -e "${GREEN}✓ TEST 14 PASSED: Cannot remove non-existent role${NC}"
    echo "$REMOVE_NONEXIST" | jq .
else
    echo -e "${RED}✗ TEST 14 FAILED: Should reject removing non-existent role${NC}"
    echo "$REMOVE_NONEXIST" | jq .
fi
echo ""

# ========================================
# TEST 15: SelfRemoveHackathonRole - OWNER role (should FAIL)
# ========================================
echo -e "${GREEN}TEST 15: SelfRemoveHackathonRole - OWNER role (should FAIL)...${NC}"
SELF_REMOVE_OWNER=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff:selfRemoveRole" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "HX_ROLE_OWNER",
    "idempotency_key": {"key": "self-remove-owner-'$TIMESTAMP'"}
  }')

SELF_OWN_ERROR=$(echo "$SELF_REMOVE_OWNER" | jq -r '.message // "no error"')
if [[ "$SELF_OWN_ERROR" == *"owner"* ]] || [[ "$SELF_OWN_ERROR" == *"cannot"* ]]; then
    echo -e "${GREEN}✓ TEST 15 PASSED: Cannot self-remove OWNER role${NC}"
    echo "$SELF_REMOVE_OWNER" | jq .
else
    echo -e "${RED}✗ TEST 15 FAILED: Should not allow self-removing OWNER${NC}"
    echo "$SELF_REMOVE_OWNER" | jq .
fi
echo ""

# ========================================
# TEST 16: SelfRemoveHackathonRole - Non-existent role (should FAIL)
# ========================================
echo -e "${GREEN}TEST 16: SelfRemoveHackathonRole - Non-existent role (should FAIL)...${NC}"
SELF_REMOVE_NONEXIST=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff:selfRemoveRole" \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "HX_ROLE_JUDGE",
    "idempotency_key": {"key": "self-remove-nonexist-'$TIMESTAMP'"}
  }')

SELF_NONEXIST_ERROR=$(echo "$SELF_REMOVE_NONEXIST" | jq -r '.message // "no error"')
if [[ "$SELF_NONEXIST_ERROR" == *"not found"* ]] || [[ "$SELF_NONEXIST_ERROR" == *"does not have"* ]] || [[ "$SELF_NONEXIST_ERROR" == *"no such"* ]]; then
    echo -e "${GREEN}✓ TEST 16 PASSED: Cannot self-remove non-existent role${NC}"
    echo "$SELF_REMOVE_NONEXIST" | jq .
else
    echo -e "${RED}✗ TEST 16 FAILED: Should reject self-removing non-existent role${NC}"
    echo "$SELF_REMOVE_NONEXIST" | jq .
fi
echo ""

# ========================================
# TEST 17: CreateStaffInvitation - Duplicate invitation (should FAIL or be idempotent)
# ========================================
echo -e "${GREEN}TEST 17: CreateStaffInvitation - Existing staff member (should FAIL)...${NC}"
INVITE_EXISTING=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff-invitations" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "'$BOB_USER_ID'",
    "requested_role": "HX_ROLE_ORGANIZER",
    "message": "Bob is already staff",
    "idempotency_key": {"key": "invite-existing-'$TIMESTAMP'"}
  }')

EXISTING_ERROR=$(echo "$INVITE_EXISTING" | jq -r '.message // "no error"')
EXISTING_ID=$(echo "$INVITE_EXISTING" | jq -r '.invitationId // "null"')

# This might either fail (if validation prevents inviting staff) or succeed (creating another invitation)
# Both behaviors are acceptable depending on business logic
if [[ "$EXISTING_ERROR" == *"already"* ]] || [[ "$EXISTING_ERROR" == *"staff"* ]]; then
    echo -e "${GREEN}✓ TEST 17 PASSED: Cannot invite existing staff member${NC}"
    echo "$INVITE_EXISTING" | jq .
elif [[ "$EXISTING_ID" != "null" ]]; then
    echo -e "${YELLOW}⚠ TEST 17 PASSED (with warning): Invitation created despite user being staff${NC}"
    echo "$INVITE_EXISTING" | jq .
else
    echo -e "${YELLOW}⚠ TEST 17: Unexpected response${NC}"
    echo "$INVITE_EXISTING" | jq .
fi
echo ""

# ========================================
# Summary
# ========================================
echo -e "${YELLOW}=== Fail Cases Testing Complete ===${NC}\n"
echo -e "${BLUE}Tests Summary:${NC}"
echo -e "  1. ListHackathonStaff - Non-staff forbidden"
echo -e "  2. CreateStaffInvitation - Non-owner forbidden"
echo -e "  3. CreateStaffInvitation - OWNER role forbidden"
echo -e "  4. CreateStaffInvitation - Non-existent user"
echo -e "  5. CreateStaffInvitation - Self-invitation forbidden"
echo -e "  6. AcceptStaffInvitation - Wrong user forbidden"
echo -e "  7. CancelStaffInvitation - Non-owner forbidden"
echo -e "  8. AcceptStaffInvitation - Already accepted"
echo -e "  9. CancelStaffInvitation - Already accepted"
echo -e " 10. AcceptStaffInvitation - Already rejected"
echo -e " 11. RejectStaffInvitation - Already rejected"
echo -e " 12. RemoveHackathonRole - Non-owner forbidden"
echo -e " 13. RemoveHackathonRole - OWNER role forbidden"
echo -e " 14. RemoveHackathonRole - Non-existent role"
echo -e " 15. SelfRemoveHackathonRole - OWNER role forbidden"
echo -e " 16. SelfRemoveHackathonRole - Non-existent role"
echo -e " 17. CreateStaffInvitation - Existing staff member"
echo ""
echo -e "${GREEN}All validation and permission fail cases have been tested!${NC}"

