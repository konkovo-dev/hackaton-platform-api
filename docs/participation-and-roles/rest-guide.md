# Participation and Roles Service REST Test Cases

## Предусловия
- Запущен в докере с помощью [docker-setup.md](../docker-setup.md)
- **Тестовые данные добавлены в БД** (см. раздел ниже)

## Эндпоинты

Base URL: `http://localhost:8080`

## Получение access_token

> **Важно**: ParticipationAndRolesService **требует авторизации**. Все методы доступны только залогиненным пользователям.

Зарегистрируйтесь через auth-service:

```bash
RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "test-staff-user",
    "email": "test-staff@example.com",
    "password": "SecurePass123",
    "first_name": "Test",
    "last_name": "Staff",
    "timezone": "UTC"
  }')

ACCESS_TOKEN=$(echo $RESPONSE | jq -r '.accessToken')
echo "Access Token: $ACCESS_TOKEN"
```

> Подробнее о auth-service: [../auth/rest-guide.md](../auth/rest-guide.md)

---

## Подготовка тестовых данных

Для тестирования staff management нужно:
1. Создать пользователей (owner, invited users)
2. Создать и опубликовать хакатон
3. Назначить OWNER роль создателю через AssignHackathonRole (service-to-service)

### Шаг 1: Подключитесь к БД

```bash
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon
```

### Шаг 2: Добавьте тестовых пользователей и хакатон

```sql
-- ============================================
-- Пользователи для тестирования staff management
-- ============================================

-- Alice: Owner хакатона
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('a11ce000-0000-0000-0000-000000000001', 'alice_owner', 'Alice', 'Owner', 'https://example.com/alice.jpg', 'UTC', NOW(), NOW());

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('a11ce000-0000-0000-0000-000000000001', 'public', 'public');

-- Bob: будет приглашен как ORGANIZER
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('b0b00000-0000-0000-0000-000000000002', 'bob_organizer', 'Bob', 'Organizer', 'https://example.com/bob.jpg', 'UTC', NOW(), NOW());

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('b0b00000-0000-0000-0000-000000000002', 'public', 'public');

-- Charlie: будет приглашен как MENTOR
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('c4a41e00-0000-0000-0000-000000000003', 'charlie_mentor', 'Charlie', 'Mentor', '', 'Europe/London', NOW(), NOW());

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('c4a41e00-0000-0000-0000-000000000003', 'public', 'public');

-- Diana: будет приглашена как JUDGE
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('d1a4a000-0000-0000-0000-000000000004', 'diana_judge', 'Diana', 'Judge', '', 'America/New_York', NOW(), NOW());

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('d1a4a000-0000-0000-0000-000000000004', 'public', 'public');

-- Eve: простой пользователь без staff ролей
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('e5e00000-0000-0000-0000-000000000005', 'eve_user', 'Eve', 'User', '', 'UTC', NOW(), NOW());

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('e5e00000-0000-0000-0000-000000000005', 'public', 'public');

-- ============================================
-- Хакатон для тестирования
-- ============================================

INSERT INTO hackaton.hackathons (
  id, 
  name, 
  short_description, 
  description, 
  stage,
  online, 
  country, 
  city, 
  venue,
  registration_opens_at,
  registration_closes_at,
  starts_at,
  ends_at,
  judging_ends_at,
  allow_individual,
  allow_team,
  team_size_max,
  created_at, 
  updated_at,
  published_at
)
VALUES (
  '44444444-4444-4444-4444-444444444444',
  'AI Innovation Hackathon 2026',
  'Build the future with AI',
  'Join us for an exciting hackathon focused on AI and ML innovations.',
  'published',
  false,
  'Russia',
  'Moscow',
  'Skolkovo',
  '2026-03-01T00:00:00Z',
  '2026-03-20T23:59:59Z',
  '2026-03-25T10:00:00Z',
  '2026-03-27T18:00:00Z',
  '2026-03-30T18:00:00Z',
  true,
  true,
  5,
  NOW(),
  NOW(),
  NOW()
);

-- ============================================
-- Alice = OWNER этого хакатона
-- ============================================

INSERT INTO participation_and_roles.staff_roles (hackathon_id, user_id, role, created_at)
VALUES 
  ('44444444-4444-4444-4444-444444444444', 'a11ce000-0000-0000-0000-000000000001', 'owner', NOW());

-- ============================================
-- Bob уже является ORGANIZER (для тестирования RemoveHackathonRole)
-- ============================================

INSERT INTO participation_and_roles.staff_roles (hackathon_id, user_id, role, created_at)
VALUES 
  ('44444444-4444-4444-4444-444444444444', 'b0b00000-0000-0000-0000-000000000002', 'organizer', NOW());
```

### Шаг 3: Проверьте данные

```sql
-- Проверить пользователей
SELECT id, username, first_name, last_name FROM identity.users 
WHERE id IN (
  'a11ce000-0000-0000-0000-000000000001',
  'b0b00000-0000-0000-0000-000000000002',
  'c4a41e00-0000-0000-0000-000000000003',
  'd1a4a000-0000-0000-0000-000000000004',
  'e5e00000-0000-0000-0000-000000000005'
);

-- Проверить хакатон
SELECT id, name, stage FROM hackaton.hackathons 
WHERE id = '44444444-4444-4444-4444-444444444444';

-- Проверить staff роли
SELECT hackathon_id, user_id, role FROM participation_and_roles.staff_roles
WHERE hackathon_id = '44444444-4444-4444-4444-444444444444';
```

Должно вернуть:
- 5 пользователей
- 1 хакатон в статусе `published`
- 2 staff роли (Alice = owner, Bob = organizer)

---

## Тест-сценарии

> **Важно**: Все запросы **требуют авторизации**. Используйте `ACCESS_TOKEN` из раздела выше.

### 1. ListHackathonStaff (Alice - owner, should see all staff)

```bash
# Alice - owner хакатона
ALICE_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice_owner",
    "password": "SecurePass123"
  }')

ALICE_TOKEN=$(echo $ALICE_RESPONSE | jq -r '.accessToken')

curl "http://localhost:8080/v1/hackathons/44444444-4444-4444-4444-444444444444/staff" \
  -H "Authorization: Bearer $ALICE_TOKEN" | jq .
```

**Response:**
```json
{
  "staff": [
    {
      "userId": "a11ce000-0000-0000-0000-000000000001",
      "roles": ["HACKATHON_ROLE_OWNER"]
    },
    {
      "userId": "b0b00000-0000-0000-0000-000000000002",
      "roles": ["HACKATHON_ROLE_ORGANIZER"]
    }
  ],
  "page": {}
}
```

### 2. ListHackathonStaff (Eve - not staff, should fail)

```bash
# Eve - обычный пользователь без staff ролей
EVE_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "eve_user",
    "password": "SecurePass123"
  }')

EVE_TOKEN=$(echo $EVE_RESPONSE | jq -r '.accessToken')

curl "http://localhost:8080/v1/hackathons/44444444-4444-4444-4444-444444444444/staff" \
  -H "Authorization: Bearer $EVE_TOKEN" | jq .
```

**Response:**
```json
{
  "code": 7,
  "message": "forbidden",
  "details": []
}
```

> **Ожидаемо**: Только staff члены могут просматривать список staff.

### 3. CreateStaffInvitation (Alice invites Charlie as MENTOR)

```bash
curl -X POST "http://localhost:8080/v1/hackathons/44444444-4444-4444-4444-444444444444/staff-invitations" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "c4a41e00-0000-0000-0000-000000000003",
    "requested_role": "HACKATHON_ROLE_MENTOR",
    "message": "We would love to have you as a mentor!",
    "idempotency_key": {"key": "invite-charlie-mentor-1"}
  }' | jq .
```

**Response:**
```json
{
  "invitationId": "00000000-0000-0000-0000-000000000001"
}
```

### 4. CreateStaffInvitation (Bob - not owner, should fail)

```bash
# Bob - organizer, но не owner
BOB_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "bob_organizer",
    "password": "SecurePass123"
  }')

BOB_TOKEN=$(echo $BOB_RESPONSE | jq -r '.accessToken')

curl -X POST "http://localhost:8080/v1/hackathons/44444444-4444-4444-4444-444444444444/staff-invitations" \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "d1a4a000-0000-0000-0000-000000000004",
    "requested_role": "HACKATHON_ROLE_JUDGE",
    "message": "Let us invite Diana!"
  }' | jq .
```

**Response:**
```json
{
  "code": 7,
  "message": "forbidden",
  "details": []
}
```

> **Ожидаемо**: Только OWNER может создавать staff invitations.

### 5. ListMyStaffInvitations (Charlie sees invitation)

```bash
# Charlie - получатель инвайта
CHARLIE_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "charlie_mentor",
    "password": "SecurePass123"
  }')

CHARLIE_TOKEN=$(echo $CHARLIE_RESPONSE | jq -r '.accessToken')

curl -X POST "http://localhost:8080/v1/users/me/staff-invitations:list" \
  -H "Authorization: Bearer $CHARLIE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}' | jq .
```

**Response:**
```json
{
  "invitations": [
    {
      "invitationId": "00000000-0000-0000-0000-000000000001",
      "hackathonId": "44444444-4444-4444-4444-444444444444",
      "targetUserId": "c4a41e00-0000-0000-0000-000000000003",
      "requestedRole": "HACKATHON_ROLE_MENTOR",
      "createdByUserId": "a11ce000-0000-0000-0000-000000000001",
      "message": "We would love to have you as a mentor!",
      "status": "STAFF_INVITATION_STATUS_PENDING",
      "createdAt": "2026-01-25T12:00:00Z",
      "updatedAt": "2026-01-25T12:00:00Z",
      "expiresAt": "2026-02-01T12:00:00Z"
    }
  ],
  "page": {}
}
```

### 6. AcceptStaffInvitation (Charlie accepts)

```bash
# Сохраняем INVITATION_ID из предыдущего шага
INVITATION_ID="00000000-0000-0000-0000-000000000001"

curl -X POST "http://localhost:8080/v1/users/me/staff-invitations/$INVITATION_ID:accept" \
  -H "Authorization: Bearer $CHARLIE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "charlie-accept-invitation-1"}
  }' | jq .
```

**Response:**
```json
{}
```

> **Ожидаемо**: Пустой ответ означает успешное принятие. Charlie теперь MENTOR.

### 7. ListHackathonStaff (verify Charlie is now MENTOR)

```bash
curl "http://localhost:8080/v1/hackathons/44444444-4444-4444-4444-444444444444/staff" \
  -H "Authorization: Bearer $ALICE_TOKEN" | jq .
```

**Response:**
```json
{
  "staff": [
    {
      "userId": "a11ce000-0000-0000-0000-000000000001",
      "roles": ["HACKATHON_ROLE_OWNER"]
    },
    {
      "userId": "b0b00000-0000-0000-0000-000000000002",
      "roles": ["HACKATHON_ROLE_ORGANIZER"]
    },
    {
      "userId": "c4a41e00-0000-0000-0000-000000000003",
      "roles": ["HACKATHON_ROLE_MENTOR"]
    }
  ],
  "page": {}
}
```

### 8. CreateStaffInvitation for Diana as JUDGE

```bash
curl -X POST "http://localhost:8080/v1/hackathons/44444444-4444-4444-4444-444444444444/staff-invitations" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "d1a4a000-0000-0000-0000-000000000004",
    "requested_role": "HACKATHON_ROLE_JUDGE",
    "message": "We need your expertise!",
    "idempotency_key": {"key": "invite-diana-judge-1"}
  }' | jq .
```

**Response:**
```json
{
  "invitationId": "00000000-0000-0000-0000-000000000002"
}
```

### 9. RejectStaffInvitation (Diana rejects)

```bash
# Diana - получатель второго инвайта
DIANA_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "diana_judge",
    "password": "SecurePass123"
  }')

DIANA_TOKEN=$(echo $DIANA_RESPONSE | jq -r '.accessToken')

INVITATION_ID_2="00000000-0000-0000-0000-000000000002"

curl -X POST "http://localhost:8080/v1/users/me/staff-invitations/$INVITATION_ID_2:reject" \
  -H "Authorization: Bearer $DIANA_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "diana-reject-invitation-1"}
  }' | jq .
```

**Response:**
```json
{}
```

### 10. ListMyStaffInvitations (Diana sees declined invitation)

```bash
curl -X POST "http://localhost:8080/v1/users/me/staff-invitations:list" \
  -H "Authorization: Bearer $DIANA_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}' | jq .
```

**Response:**
```json
{
  "invitations": [
    {
      "invitationId": "00000000-0000-0000-0000-000000000002",
      "hackathonId": "44444444-4444-4444-4444-444444444444",
      "targetUserId": "d1a4a000-0000-0000-0000-000000000004",
      "requestedRole": "HACKATHON_ROLE_JUDGE",
      "createdByUserId": "a11ce000-0000-0000-0000-000000000001",
      "message": "We need your expertise!",
      "status": "STAFF_INVITATION_STATUS_DECLINED",
      "createdAt": "2026-01-25T12:05:00Z",
      "updatedAt": "2026-01-25T12:06:00Z",
      "expiresAt": "2026-02-01T12:05:00Z"
    }
  ],
  "page": {}
}
```

### 11. CreateStaffInvitation for Diana again

```bash
curl -X POST "http://localhost:8080/v1/hackathons/44444444-4444-4444-4444-444444444444/staff-invitations" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "d1a4a000-0000-0000-0000-000000000004",
    "requested_role": "HACKATHON_ROLE_JUDGE",
    "message": "Please reconsider!",
    "idempotency_key": {"key": "invite-diana-judge-2"}
  }' | jq .
```

**Response:**
```json
{
  "invitationId": "00000000-0000-0000-0000-000000000003"
}
```

### 12. CancelStaffInvitation (Alice cancels)

```bash
INVITATION_ID_3="00000000-0000-0000-0000-000000000003"

curl -X POST "http://localhost:8080/v1/hackathons/44444444-4444-4444-4444-444444444444/staff-invitations/$INVITATION_ID_3:cancel" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "alice-cancel-invitation-1"}
  }' | jq .
```

**Response:**
```json
{}
```

### 13. RemoveHackathonRole (Alice removes Bob's ORGANIZER role)

```bash
curl -X POST "http://localhost:8080/v1/hackathons/44444444-4444-4444-4444-444444444444/staff:removeRole" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "b0b00000-0000-0000-0000-000000000002",
    "role": "HACKATHON_ROLE_ORGANIZER",
    "idempotency_key": {"key": "alice-remove-bob-organizer-1"}
  }' | jq .
```

**Response:**
```json
{}
```

### 14. ListHackathonStaff (verify Bob is removed)

```bash
curl "http://localhost:8080/v1/hackathons/44444444-4444-4444-4444-444444444444/staff" \
  -H "Authorization: Bearer $ALICE_TOKEN" | jq .
```

**Response:**
```json
{
  "staff": [
    {
      "userId": "a11ce000-0000-0000-0000-000000000001",
      "roles": ["HACKATHON_ROLE_OWNER"]
    },
    {
      "userId": "c4a41e00-0000-0000-0000-000000000003",
      "roles": ["HACKATHON_ROLE_MENTOR"]
    }
  ],
  "page": {}
}
```

### 15. SelfRemoveHackathonRole (Charlie removes himself as MENTOR)

```bash
curl -X POST "http://localhost:8080/v1/hackathons/44444444-4444-4444-4444-444444444444/staff:selfRemoveRole" \
  -H "Authorization: Bearer $CHARLIE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "HACKATHON_ROLE_MENTOR",
    "idempotency_key": {"key": "charlie-self-remove-mentor-1"}
  }' | jq .
```

**Response:**
```json
{}
```

### 16. ListHackathonStaff (verify Charlie is removed)

```bash
curl "http://localhost:8080/v1/hackathons/44444444-4444-4444-4444-444444444444/staff" \
  -H "Authorization: Bearer $ALICE_TOKEN" | jq .
```

**Response:**
```json
{
  "staff": [
    {
      "userId": "a11ce000-0000-0000-0000-000000000001",
      "roles": ["HACKATHON_ROLE_OWNER"]
    }
  ],
  "page": {}
}
```

> **Ожидаемо**: Только Alice (OWNER) остается.

---

## Негативные сценарии

### Try to remove OWNER role (should fail)

```bash
curl -X POST "http://localhost:8080/v1/hackathons/44444444-4444-4444-4444-444444444444/staff:removeRole" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "a11ce000-0000-0000-0000-000000000001",
    "role": "HACKATHON_ROLE_OWNER"
  }' | jq .
```

**Response:**
```json
{
  "code": 7,
  "message": "forbidden",
  "details": []
}
```

### Try to self-remove OWNER role (should fail)

```bash
curl -X POST "http://localhost:8080/v1/hackathons/44444444-4444-4444-4444-444444444444/staff:selfRemoveRole" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "HACKATHON_ROLE_OWNER"
  }' | jq .
```

**Response:**
```json
{
  "code": 7,
  "message": "forbidden",
  "details": []
}
```

### Try to invite to OWNER role (should fail)

```bash
curl -X POST "http://localhost:8080/v1/hackathons/44444444-4444-4444-4444-444444444444/staff-invitations" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "e5e00000-0000-0000-0000-000000000005",
    "requested_role": "HACKATHON_ROLE_OWNER",
    "message": "Become owner!"
  }' | jq .
```

**Response:**
```json
{
  "code": 3,
  "message": "invalid input",
  "details": []
}
```

### Try to accept invitation not addressed to you (should fail)

```bash
# Bob пытается принять инвайт, адресованный Charlie
curl -X POST "http://localhost:8080/v1/users/me/staff-invitations/00000000-0000-0000-0000-000000000001:accept" \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}' | jq .
```

**Response:**
```json
{
  "code": 7,
  "message": "forbidden",
  "details": []
}
```

---

## Troubleshooting

### 401 Unauthenticated

**Причина**: Не передан или невалидный access_token.

**Решение**:
```bash
# Получите новый токен
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice_owner",
    "password": "SecurePass123"
  }' | jq -r '.accessToken'
```

### 403 Forbidden

**Причина**: Пользователь не имеет необходимых прав (например, не OWNER).

**Решение**: Проверьте роль пользователя в БД и используйте токен пользователя с правами OWNER.

### 404 Not Found

**Причина**: Хакатон, инвайт или пользователь не существует.

**Решение**: Проверьте корректность UUID.

### 409 Conflict

**Причина**: Попытка создать дубликат (например, pending invitation already exists).

**Решение**: Отмените или дождитесь истечения существующего инвайта.

---

## Дополнительно

- gRPC endpoint: `localhost:50055`
- HTTP gateway: `localhost:8080`
- gRPC guide: [grpc-guide.md](./grpc-guide.md)
- All HTTP endpoints доступны под `/v1/hackathons/*` и `/v1/users/me/*`

