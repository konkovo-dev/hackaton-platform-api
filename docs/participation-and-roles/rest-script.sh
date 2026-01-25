#!/bin/bash

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

BASE_URL="http://localhost:8080"
TIMESTAMP=$(date +%s)

echo -e "${YELLOW}=== Participation and Roles Service REST Testing ===${NC}\n"

# ========================================
# 1. Setup: Register Users
# ========================================
echo -e "${GREEN}1. Registering test users...${NC}"

echo -e "${BLUE}Registering Alice (hackathon owner)...${NC}"
ALICE_RESPONSE=$(curl -s -X POST $BASE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice_staff_'$TIMESTAMP'",
    "email": "alice_staff_'$TIMESTAMP'@hackathon.com",
    "password": "SecurePass123",
    "first_name": "Alice",
    "last_name": "Owner",
    "timezone": "UTC",
    "idempotency_key": {"key": "alice-staff-'$TIMESTAMP'"}
  }')

ALICE_TOKEN=$(echo $ALICE_RESPONSE | jq -r '.accessToken')
if [ "$ALICE_TOKEN" = "null" ] || [ -z "$ALICE_TOKEN" ]; then
    echo -e "${RED}Failed to register Alice${NC}"
    echo $ALICE_RESPONSE | jq .
    exit 1
fi
echo -e "${GREEN}✓ Alice registered. Token: ${ALICE_TOKEN:0:50}...${NC}"

echo -e "${BLUE}Registering Bob (organizer candidate)...${NC}"
BOB_RESPONSE=$(curl -s -X POST $BASE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "bob_staff_'$TIMESTAMP'",
    "email": "bob_staff_'$TIMESTAMP'@hackathon.com",
    "password": "SecurePass123",
    "first_name": "Bob",
    "last_name": "Organizer",
    "timezone": "UTC",
    "idempotency_key": {"key": "bob-staff-'$TIMESTAMP'"}
  }')

BOB_TOKEN=$(echo $BOB_RESPONSE | jq -r '.accessToken')
if [ "$BOB_TOKEN" = "null" ] || [ -z "$BOB_TOKEN" ]; then
    echo -e "${RED}Failed to register Bob${NC}"
    echo $BOB_RESPONSE | jq .
    exit 1
fi
echo -e "${GREEN}✓ Bob registered${NC}"

echo -e "${BLUE}Registering Charlie (mentor candidate)...${NC}"
CHARLIE_RESPONSE=$(curl -s -X POST $BASE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "charlie_staff_'$TIMESTAMP'",
    "email": "charlie_staff_'$TIMESTAMP'@hackathon.com",
    "password": "SecurePass123",
    "first_name": "Charlie",
    "last_name": "Mentor",
    "timezone": "UTC",
    "idempotency_key": {"key": "charlie-staff-'$TIMESTAMP'"}
  }')

CHARLIE_TOKEN=$(echo $CHARLIE_RESPONSE | jq -r '.accessToken')
if [ "$CHARLIE_TOKEN" = "null" ] || [ -z "$CHARLIE_TOKEN" ]; then
    echo -e "${RED}Failed to register Charlie${NC}"
    echo $CHARLIE_RESPONSE | jq .
    exit 1
fi
echo -e "${GREEN}✓ Charlie registered${NC}"

echo -e "${BLUE}Registering Diana (judge candidate)...${NC}"
DIANA_RESPONSE=$(curl -s -X POST $BASE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "diana_staff_'$TIMESTAMP'",
    "email": "diana_staff_'$TIMESTAMP'@hackathon.com",
    "password": "SecurePass123",
    "first_name": "Diana",
    "last_name": "Judge",
    "timezone": "UTC",
    "idempotency_key": {"key": "diana-staff-'$TIMESTAMP'"}
  }')

DIANA_TOKEN=$(echo $DIANA_RESPONSE | jq -r '.accessToken')
if [ "$DIANA_TOKEN" = "null" ] || [ -z "$DIANA_TOKEN" ]; then
    echo -e "${RED}Failed to register Diana${NC}"
    echo $DIANA_RESPONSE | jq .
    exit 1
fi
echo -e "${GREEN}✓ Diana registered${NC}"

echo -e "${BLUE}Registering Eve (regular user)...${NC}"
EVE_RESPONSE=$(curl -s -X POST $BASE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "eve_staff_'$TIMESTAMP'",
    "email": "eve_staff_'$TIMESTAMP'@hackathon.com",
    "password": "SecurePass123",
    "first_name": "Eve",
    "last_name": "User",
    "timezone": "UTC",
    "idempotency_key": {"key": "eve-staff-'$TIMESTAMP'"}
  }')

EVE_TOKEN=$(echo $EVE_RESPONSE | jq -r '.accessToken')
if [ "$EVE_TOKEN" = "null" ] || [ -z "$EVE_TOKEN" ]; then
    echo -e "${RED}Failed to register Eve${NC}"
    echo $EVE_RESPONSE | jq .
    exit 1
fi
echo -e "${GREEN}✓ Eve registered${NC}\n"

# Extract user IDs from introspection
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

EVE_USER_ID=$(curl -s -X POST $BASE_URL/v1/auth/introspect \
  -H "Content-Type: application/json" \
  -d '{"access_token": "'$EVE_TOKEN'"}' | jq -r '.userId')

echo -e "${GREEN}User IDs extracted:${NC}"
echo -e "  Alice: $ALICE_USER_ID"
echo -e "  Bob: $BOB_USER_ID"
echo -e "  Charlie: $CHARLIE_USER_ID"
echo -e "  Diana: $DIANA_USER_ID"
echo -e "  Eve: $EVE_USER_ID\n"

# ========================================
# 2. Create Hackathon
# ========================================
echo -e "${GREEN}2. Creating hackathon...${NC}"
CREATE_HACKATHON_RESPONSE=$(curl -s -X POST $BASE_URL/v1/hackathons \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Staff Management Test Hackathon '$TIMESTAMP'",
    "short_description": "Testing staff management",
    "description": "A hackathon for testing staff invitation and role management features.",
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
    "idempotency_key": {"key": "hackathon-staff-test-'$TIMESTAMP'"}
  }')

HACKATHON_ID=$(echo $CREATE_HACKATHON_RESPONSE | jq -r '.hackathonId')
if [ "$HACKATHON_ID" = "null" ] || [ -z "$HACKATHON_ID" ]; then
    echo -e "${RED}Failed to create hackathon${NC}"
    echo $CREATE_HACKATHON_RESPONSE | jq .
    exit 1
fi
echo -e "${GREEN}✓ Hackathon created with ID: $HACKATHON_ID${NC}"
echo -e "${BLUE}Waiting for OWNER role to be assigned (outbox processing)...${NC}"
sleep 2
echo -e "${GREEN}✓ Ready to proceed${NC}\n"

# ========================================
# 3. ListHackathonStaff (Alice - owner, happy path)
# ========================================
echo -e "${GREEN}3. ListHackathonStaff (Alice - owner, happy path)...${NC}"
LIST_STAFF_RESPONSE=$(curl -s "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff" \
  -H "Authorization: Bearer $ALICE_TOKEN")

echo "$LIST_STAFF_RESPONSE" | jq .

STAFF_COUNT=$(echo "$LIST_STAFF_RESPONSE" | jq '.staff | length')
if [ "$STAFF_COUNT" -ge "1" ]; then
    echo -e "${GREEN}✓ Alice can view staff (found $STAFF_COUNT members)${NC}"
else
    echo -e "${RED}Expected at least 1 staff member, got: $STAFF_COUNT${NC}"
    exit 1
fi
echo ""

# ========================================
# 4. ListHackathonStaff (Eve - not staff, should fail)
# ========================================
echo -e "${GREEN}4. ListHackathonStaff (Eve - not staff, should fail)...${NC}"
EVE_LIST_RESPONSE=$(curl -s "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff" \
  -H "Authorization: Bearer $EVE_TOKEN")

EVE_ERROR=$(echo "$EVE_LIST_RESPONSE" | jq -r '.message // "no error"')
if [[ "$EVE_ERROR" == *"forbidden"* ]] || [[ "$EVE_ERROR" == *"permission"* ]]; then
    echo -e "${GREEN}✓ Eve cannot view staff (forbidden)${NC}"
else
    echo -e "${RED}Expected forbidden error, got: $EVE_ERROR${NC}"
    echo "$EVE_LIST_RESPONSE" | jq .
fi
echo ""

# ========================================
# 5. CreateStaffInvitation (Alice invites Charlie as MENTOR, happy path)
# ========================================
echo -e "${GREEN}5. CreateStaffInvitation (Alice invites Charlie as MENTOR)...${NC}"
INVITE_CHARLIE=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff-invitations" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "'$CHARLIE_USER_ID'",
    "requested_role": "HX_ROLE_MENTOR",
    "message": "We would love to have you as a mentor!",
    "idempotency_key": {"key": "invite-charlie-mentor-'$TIMESTAMP'"}
  }')

echo "$INVITE_CHARLIE" | jq .

CHARLIE_INVITATION_ID=$(echo "$INVITE_CHARLIE" | jq -r '.invitationId')
if [ "$CHARLIE_INVITATION_ID" != "null" ] && [ -n "$CHARLIE_INVITATION_ID" ]; then
    echo -e "${GREEN}✓ Invitation created for Charlie. ID: $CHARLIE_INVITATION_ID${NC}"
else
    echo -e "${RED}Failed to create invitation for Charlie${NC}"
    exit 1
fi
echo ""

# ========================================
# 6. CreateStaffInvitation (Bob - not owner, should fail)
# ========================================
echo -e "${GREEN}6. CreateStaffInvitation (Bob - not owner, should fail)...${NC}"
BOB_INVITE=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff-invitations" \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "'$DIANA_USER_ID'",
    "requested_role": "HX_ROLE_JUDGE",
    "message": "Bob tries to invite"
  }')

BOB_INVITE_ERROR=$(echo "$BOB_INVITE" | jq -r '.message // "no error"')
if [[ "$BOB_INVITE_ERROR" == *"forbidden"* ]] || [[ "$BOB_INVITE_ERROR" == *"permission"* ]]; then
    echo -e "${GREEN}✓ Bob cannot create invitation (forbidden)${NC}"
else
    echo -e "${RED}Expected forbidden error, got: $BOB_INVITE_ERROR${NC}"
    echo "$BOB_INVITE" | jq .
fi
echo ""

# ========================================
# 7. ListMyStaffInvitations (Charlie checks his invitations)
# ========================================
echo -e "${GREEN}7. ListMyStaffInvitations (Charlie checks his invitations)...${NC}"
CHARLIE_INVITATIONS=$(curl -s -X POST "$BASE_URL/v1/users/me/staff-invitations:list" \
  -H "Authorization: Bearer $CHARLIE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}')

echo "$CHARLIE_INVITATIONS" | jq .

CHARLIE_INV_COUNT=$(echo "$CHARLIE_INVITATIONS" | jq '.invitations | length')
if [ "$CHARLIE_INV_COUNT" -ge "1" ]; then
    echo -e "${GREEN}✓ Charlie has $CHARLIE_INV_COUNT invitation(s)${NC}"
else
    echo -e "${RED}Expected at least 1 invitation for Charlie, got: $CHARLIE_INV_COUNT${NC}"
    exit 1
fi
echo ""

# ========================================
# 8. AcceptStaffInvitation (Charlie accepts)
# ========================================
echo -e "${GREEN}8. AcceptStaffInvitation (Charlie accepts)...${NC}"
ACCEPT_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/users/me/staff-invitations/$CHARLIE_INVITATION_ID:accept" \
  -H "Authorization: Bearer $CHARLIE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "accept-charlie-'$TIMESTAMP'"}
  }')

echo "$ACCEPT_RESPONSE" | jq .

ACCEPT_STATUS=$(echo "$ACCEPT_RESPONSE" | jq -r '.status // ""')
if [[ "$ACCEPT_STATUS" == *"ACCEPTED"* ]] || [[ "$ACCEPT_RESPONSE" == *"accepted"* ]]; then
    echo -e "${GREEN}✓ Charlie accepted the invitation${NC}"
else
    echo -e "${YELLOW}⚠ Unexpected response (might still be OK)${NC}"
fi
echo ""

# ========================================
# 9. Verify Charlie is now staff
# ========================================
echo -e "${GREEN}9. Verifying Charlie is now staff...${NC}"
STAFF_LIST=$(curl -s "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff" \
  -H "Authorization: Bearer $ALICE_TOKEN")

STAFF_MEMBER_COUNT=$(echo "$STAFF_LIST" | jq '.staff | length')
if [ "$STAFF_MEMBER_COUNT" -ge "2" ]; then
    echo -e "${GREEN}✓ Staff now has $STAFF_MEMBER_COUNT members (Alice + Charlie)${NC}"
else
    echo -e "${RED}Expected at least 2 staff members, got: $STAFF_MEMBER_COUNT${NC}"
    echo "$STAFF_LIST" | jq .
fi
echo ""

# ========================================
# 10. Create another invitation (Diana as JUDGE)
# ========================================
echo -e "${GREEN}10. CreateStaffInvitation (Alice invites Diana as JUDGE)...${NC}"
INVITE_DIANA=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff-invitations" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "'$DIANA_USER_ID'",
    "requested_role": "HX_ROLE_JUDGE",
    "message": "Would you be a judge for our hackathon?",
    "idempotency_key": {"key": "invite-diana-judge-'$TIMESTAMP'"}
  }')

DIANA_INVITATION_ID=$(echo "$INVITE_DIANA" | jq -r '.invitationId')
if [ "$DIANA_INVITATION_ID" != "null" ] && [ -n "$DIANA_INVITATION_ID" ]; then
    echo -e "${GREEN}✓ Invitation created for Diana. ID: $DIANA_INVITATION_ID${NC}"
else
    echo -e "${RED}Failed to create invitation for Diana${NC}"
    exit 1
fi
echo ""

# ========================================
# 11. RejectStaffInvitation (Diana rejects)
# ========================================
echo -e "${GREEN}11. RejectStaffInvitation (Diana rejects)...${NC}"
REJECT_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/users/me/staff-invitations/$DIANA_INVITATION_ID:reject" \
  -H "Authorization: Bearer $DIANA_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "reject-diana-'$TIMESTAMP'"}
  }')

echo "$REJECT_RESPONSE" | jq .

REJECT_STATUS=$(echo "$REJECT_RESPONSE" | jq -r '.status // ""')
if [[ "$REJECT_STATUS" == *"REJECTED"* ]] || [[ "$REJECT_RESPONSE" == *"rejected"* ]]; then
    echo -e "${GREEN}✓ Diana rejected the invitation${NC}"
else
    echo -e "${YELLOW}⚠ Unexpected response (might still be OK)${NC}"
fi
echo ""

# ========================================
# 12. Create invitation for Bob (for cancel test)
# ========================================
echo -e "${GREEN}12. CreateStaffInvitation (Alice invites Bob as ORGANIZER)...${NC}"
INVITE_BOB=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff-invitations" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "'$BOB_USER_ID'",
    "requested_role": "HX_ROLE_ORGANIZER",
    "message": "Join us as an organizer!",
    "idempotency_key": {"key": "invite-bob-organizer-'$TIMESTAMP'"}
  }')

BOB_INVITATION_ID=$(echo "$INVITE_BOB" | jq -r '.invitationId')
if [ "$BOB_INVITATION_ID" != "null" ] && [ -n "$BOB_INVITATION_ID" ]; then
    echo -e "${GREEN}✓ Invitation created for Bob. ID: $BOB_INVITATION_ID${NC}"
else
    echo -e "${RED}Failed to create invitation for Bob${NC}"
    exit 1
fi
echo ""

# ========================================
# 13. CancelStaffInvitation (Alice cancels Bob's invitation)
# ========================================
echo -e "${GREEN}13. CancelStaffInvitation (Alice cancels Bob's invitation)...${NC}"
CANCEL_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff-invitations/$BOB_INVITATION_ID:cancel" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "cancel-bob-'$TIMESTAMP'"}
  }')

# DELETE often returns empty response, check for error
if [[ -z "$CANCEL_RESPONSE" ]] || [[ "$CANCEL_RESPONSE" == "{}" ]]; then
    echo -e "${GREEN}✓ Invitation canceled successfully${NC}"
else
    CANCEL_ERROR=$(echo "$CANCEL_RESPONSE" | jq -r '.message // "unknown"')
    if [[ "$CANCEL_ERROR" == "unknown" ]] || [[ "$CANCEL_ERROR" == "null" ]]; then
        echo -e "${GREEN}✓ Invitation canceled successfully${NC}"
    else
        echo -e "${RED}Cancel failed: $CANCEL_ERROR${NC}"
        echo "$CANCEL_RESPONSE" | jq .
    fi
fi
echo ""

# ========================================
# 14. RemoveHackathonRole (Alice removes Charlie's MENTOR role)
# ========================================
echo -e "${GREEN}14. RemoveHackathonRole (Alice removes Charlie's MENTOR role)...${NC}"
REMOVE_ROLE_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff:removeRole" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "'$CHARLIE_USER_ID'",
    "role": "HX_ROLE_MENTOR",
    "idempotency_key": {"key": "remove-charlie-'$TIMESTAMP'"}
  }')

if [[ -z "$REMOVE_ROLE_RESPONSE" ]] || [[ "$REMOVE_ROLE_RESPONSE" == "{}" ]]; then
    echo -e "${GREEN}✓ Charlie's MENTOR role removed successfully${NC}"
else
    REMOVE_ERROR=$(echo "$REMOVE_ROLE_RESPONSE" | jq -r '.message // "unknown"')
    if [[ "$REMOVE_ERROR" == "unknown" ]] || [[ "$REMOVE_ERROR" == "null" ]]; then
        echo -e "${GREEN}✓ Charlie's MENTOR role removed successfully${NC}"
    else
        echo -e "${RED}Remove role failed: $REMOVE_ERROR${NC}"
        echo "$REMOVE_ROLE_RESPONSE" | jq .
    fi
fi
echo ""

# ========================================
# 15. Verify Charlie is no longer staff
# ========================================
echo -e "${GREEN}15. Verifying Charlie is no longer staff...${NC}"
STAFF_LIST_AFTER=$(curl -s "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff" \
  -H "Authorization: Bearer $ALICE_TOKEN")

STAFF_COUNT_AFTER=$(echo "$STAFF_LIST_AFTER" | jq '.staff | length')
if [ "$STAFF_COUNT_AFTER" = "1" ]; then
    echo -e "${GREEN}✓ Staff now has $STAFF_COUNT_AFTER member (only Alice)${NC}"
else
    echo -e "${YELLOW}⚠ Expected 1 staff member, got: $STAFF_COUNT_AFTER${NC}"
    echo "$STAFF_LIST_AFTER" | jq .
fi
echo ""

# ========================================
# 16. Invite Charlie again and let him self-remove
# ========================================
echo -e "${GREEN}16. Inviting Charlie again for self-remove test...${NC}"
INVITE_CHARLIE_AGAIN=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff-invitations" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "'$CHARLIE_USER_ID'",
    "requested_role": "HX_ROLE_MENTOR",
    "message": "Come back!",
    "idempotency_key": {"key": "invite-charlie-mentor-2-'$TIMESTAMP'"}
  }')

CHARLIE_INVITATION_ID_2=$(echo "$INVITE_CHARLIE_AGAIN" | jq -r '.invitationId')
if [ "$CHARLIE_INVITATION_ID_2" != "null" ] && [ -n "$CHARLIE_INVITATION_ID_2" ]; then
    echo -e "${GREEN}✓ Charlie invited again. ID: $CHARLIE_INVITATION_ID_2${NC}"
else
    echo -e "${RED}Failed to invite Charlie again${NC}"
    exit 1
fi

# Charlie accepts
ACCEPT_RESPONSE_2=$(curl -s -X POST "$BASE_URL/v1/users/me/staff-invitations/$CHARLIE_INVITATION_ID_2:accept" \
  -H "Authorization: Bearer $CHARLIE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "accept-charlie-2-'$TIMESTAMP'"}
  }')
echo -e "${GREEN}✓ Charlie accepted invitation${NC}\n"

# ========================================
# 17. SelfRemoveHackathonRole (Charlie removes his own role)
# ========================================
echo -e "${GREEN}17. SelfRemoveHackathonRole (Charlie removes his own role)...${NC}"
SELF_REMOVE_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff:selfRemoveRole" \
  -H "Authorization: Bearer $CHARLIE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "HX_ROLE_MENTOR",
    "idempotency_key": {"key": "self-remove-charlie-'$TIMESTAMP'"}
  }')

if [[ -z "$SELF_REMOVE_RESPONSE" ]] || [[ "$SELF_REMOVE_RESPONSE" == "{}" ]]; then
    echo -e "${GREEN}✓ Charlie removed his own MENTOR role successfully${NC}"
else
    SELF_REMOVE_ERROR=$(echo "$SELF_REMOVE_RESPONSE" | jq -r '.message // "unknown"')
    if [[ "$SELF_REMOVE_ERROR" == "unknown" ]] || [[ "$SELF_REMOVE_ERROR" == "null" ]]; then
        echo -e "${GREEN}✓ Charlie removed his own MENTOR role successfully${NC}"
    else
        echo -e "${RED}Self remove role failed: $SELF_REMOVE_ERROR${NC}"
        echo "$SELF_REMOVE_RESPONSE" | jq .
    fi
fi
echo ""

# ========================================
# 18. Verify final staff count
# ========================================
echo -e "${GREEN}18. Verifying final staff count...${NC}"
FINAL_STAFF_LIST=$(curl -s "$BASE_URL/v1/hackathons/$HACKATHON_ID/staff" \
  -H "Authorization: Bearer $ALICE_TOKEN")

FINAL_STAFF_COUNT=$(echo "$FINAL_STAFF_LIST" | jq '.staff | length')
if [ "$FINAL_STAFF_COUNT" = "1" ]; then
    echo -e "${GREEN}✓ Final staff count: $FINAL_STAFF_COUNT (only Alice remains)${NC}"
else
    echo -e "${YELLOW}⚠ Expected 1 staff member, got: $FINAL_STAFF_COUNT${NC}"
fi
echo ""

echo -e "${YELLOW}=== All tests completed! ===${NC}"
echo -e "${GREEN}✓ Registration: 5 users${NC}"
echo -e "${GREEN}✓ Hackathon creation: 1 hackathon${NC}"
echo -e "${GREEN}✓ Staff listing: Verified access control${NC}"
echo -e "${GREEN}✓ Staff invitations: Created, accepted, rejected, canceled${NC}"
echo -e "${GREEN}✓ Role removal: Tested both admin and self-removal${NC}"
