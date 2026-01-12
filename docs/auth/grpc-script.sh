#!/bin/bash

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

AUTH_SERVICE_URL="localhost:50057"

echo -e "${YELLOW}=== Auth Service Testing ===${NC}\n"

echo -e "${GREEN}1. Health Check${NC}"
grpcurl -plaintext $AUTH_SERVICE_URL grpc.health.v1.Health/Check
echo ""

echo -e "${GREEN}2. Register User${NC}"
REGISTER_RESPONSE=$(grpcurl -plaintext -d '{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123",
  "first_name": "Test",
  "last_name": "User",
  "timezone": "UTC",
  "idempotency_key": {"key": "test-key-'$(date +%s)'"}
}' $AUTH_SERVICE_URL auth.v1.AuthService/Register)

echo "$REGISTER_RESPONSE"

ACCESS_TOKEN=$(echo "$REGISTER_RESPONSE" | grep -o '"accessToken": "[^"]*"' | cut -d'"' -f4)
REFRESH_TOKEN=$(echo "$REGISTER_RESPONSE" | grep -o '"refreshToken": "[^"]*"' | cut -d'"' -f4)

if [ -z "$ACCESS_TOKEN" ]; then
    echo -e "${RED}Failed to get access token${NC}"
    exit 1
fi

echo -e "\n${GREEN}Access Token: ${NC}${ACCESS_TOKEN:0:50}..."
echo -e "${GREEN}Refresh Token: ${NC}${REFRESH_TOKEN:0:50}...\n"

echo -e "${GREEN}3. Introspect Token${NC}"
grpcurl -plaintext -d "{
  \"access_token\": \"$ACCESS_TOKEN\"
}" $AUTH_SERVICE_URL auth.v1.AuthService/IntrospectToken
echo ""

echo -e "${GREEN}4. Login${NC}"
LOGIN_RESPONSE=$(grpcurl -plaintext -d '{
  "username": "testuser",
  "password": "password123"
}' $AUTH_SERVICE_URL auth.v1.AuthService/Login)

echo "$LOGIN_RESPONSE"
echo ""

NEW_ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"accessToken": "[^"]*"' | cut -d'"' -f4)
NEW_REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"refreshToken": "[^"]*"' | cut -d'"' -f4)

echo -e "${GREEN}5. Refresh Token${NC}"
REFRESH_RESPONSE=$(grpcurl -plaintext -d "{
  \"refresh_token\": \"$NEW_REFRESH_TOKEN\"
}" $AUTH_SERVICE_URL auth.v1.AuthService/Refresh)

echo "$REFRESH_RESPONSE"
echo ""

REFRESHED_ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | grep -o '"accessToken": "[^"]*"' | cut -d'"' -f4)
REFRESHED_REFRESH_TOKEN=$(echo "$REFRESH_RESPONSE" | grep -o '"refreshToken": "[^"]*"' | cut -d'"' -f4)

echo -e "${GREEN}6. Introspect New Token${NC}"
grpcurl -plaintext -d "{
  \"access_token\": \"$REFRESHED_ACCESS_TOKEN\"
}" $AUTH_SERVICE_URL auth.v1.AuthService/IntrospectToken
echo ""

echo -e "${GREEN}7. Logout${NC}"
grpcurl -plaintext -d "{
  \"refresh_token\": \"$REFRESHED_REFRESH_TOKEN\"
}" $AUTH_SERVICE_URL auth.v1.AuthService/Logout
echo ""

echo -e "${GREEN}8. Try to use revoked refresh token (should fail)${NC}"
grpcurl -plaintext -d "{
  \"refresh_token\": \"$REFRESHED_REFRESH_TOKEN\"
}" $AUTH_SERVICE_URL auth.v1.AuthService/Refresh || echo -e "${YELLOW}Expected: token revoked${NC}"
echo ""

echo -e "${GREEN}9. Test Idempotency (same key, same request)${NC}"
IDEMPOTENCY_KEY="test-idempotency-$(date +%s)"
echo "First request with key: $IDEMPOTENCY_KEY"
grpcurl -plaintext -d "{
  \"username\": \"idempotent_user\",
  \"email\": \"idempotent@example.com\",
  \"password\": \"password123\",
  \"first_name\": \"Idempotent\",
  \"last_name\": \"User\",
  \"timezone\": \"UTC\",
  \"idempotency_key\": {\"key\": \"$IDEMPOTENCY_KEY\"}
}" $AUTH_SERVICE_URL auth.v1.AuthService/Register
echo ""

echo "Second request with same key (should return cached response):"
grpcurl -plaintext -d "{
  \"username\": \"idempotent_user\",
  \"email\": \"idempotent@example.com\",
  \"password\": \"password123\",
  \"first_name\": \"Idempotent\",
  \"last_name\": \"User\",
  \"timezone\": \"UTC\",
  \"idempotency_key\": {\"key\": \"$IDEMPOTENCY_KEY\"}
}" $AUTH_SERVICE_URL auth.v1.AuthService/Register
echo ""

echo -e "${GREEN}10. Test Invalid Token${NC}"
grpcurl -plaintext -d '{
  "access_token": "invalid.token.here"
}' $AUTH_SERVICE_URL auth.v1.AuthService/IntrospectToken
echo ""

echo -e "${GREEN}=== All Tests Completed ===${NC}"

