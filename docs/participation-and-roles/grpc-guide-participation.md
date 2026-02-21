# ParticipationService gRPC Test Cases

## Предусловия
- Запущен в докере с помощью [docker-setup.md](../docker-setup.md)
- Установлен `grpcurl`: `brew install grpcurl` (macOS) или [github.com/fullstorydev/grpcurl](https://github.com/fullstorydev/grpcurl)
- Пользователь зарегистрирован в auth-service
- Есть валидный access_token
- Тестовые данные загружены: `test-data-participation.sql`

## gRPC Endpoint

```
localhost:50055
```

## Получение access_token

Зарегистрируйтесь через auth-service:

```bash
RESPONSE=$(grpcurl -plaintext -d '{
  "username": "test_participant",
  "email": "test_participant@example.com",
  "password": "SecurePass123",
  "first_name": "Test",
  "last_name": "Participant",
  "timezone": "UTC"
}' localhost:50051 auth.v1.AuthService/Register)

ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.accessToken')
echo "Access Token: $ACCESS_TOKEN"
```

> Подробнее о auth-service: [../auth/grpc-guide.md](../auth/grpc-guide.md)

---

## Тест-сценарии

### 1. ListTeamRoles (Получить список ролей)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{}' \
  localhost:50055 participationandroles.v1.ParticipationService/ListTeamRoles
```

**Response:**
```json
{
  "teamRoles": [
    {"id": "00000000-0000-0000-0000-000000000001", "name": "Any"},
    {"id": "00000000-0000-0000-0000-000000000002", "name": "Backend"},
    {"id": "00000000-0000-0000-0000-000000000003", "name": "Frontend"},
    {"id": "00000000-0000-0000-0000-000000000004", "name": "Fullstack"},
    {"id": "00000000-0000-0000-0000-000000000005", "name": "Mobile"},
    {"id": "00000000-0000-0000-0000-000000000006", "name": "Designer"},
    {"id": "00000000-0000-0000-0000-000000000007", "name": "Product Manager"},
    {"id": "00000000-0000-0000-0000-000000000008", "name": "Data Scientist"},
    {"id": "00000000-0000-0000-0000-000000000009", "name": "DevOps"},
    {"id": "00000000-0000-0000-0000-00000000000a", "name": "QA"}
  ]
}
```

### 2. RegisterForHackathon (Регистрация как INDIVIDUAL_ACTIVE)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555",
    "desired_status": "PART_INDIVIDUAL",
    "wished_role_ids": ["00000000-0000-0000-0000-000000000003"],
    "motivation_text": "I love frontend development!",
    "idempotency_key": {"key": "user-register-individual-1"}
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/RegisterForHackathon
```

**Response:**
```json
{
  "participation": {
    "hackathonId": "55555555-5555-5555-5555-555555555555",
    "userId": "generated-uuid",
    "status": "PART_INDIVIDUAL",
    "teamId": "",
    "profile": {
      "wishedRoles": [
        {"id": "00000000-0000-0000-0000-000000000003", "name": "Frontend"}
      ],
      "motivationText": "I love frontend development!"
    },
    "registeredAt": "2026-02-10T12:00:00Z",
    "updatedAt": "2026-02-10T12:00:00Z"
  }
}
```

### 3. RegisterForHackathon (Регистрация как LOOKING_FOR_TEAM)

```bash
# Зарегистрируем другого пользователя
ALICE_RESPONSE=$(grpcurl -plaintext -d '{
  "username": "alice_looking",
  "email": "alice_looking@example.com",
  "password": "SecurePass123",
  "first_name": "Alice",
  "last_name": "Seeker",
  "timezone": "UTC"
}' localhost:50051 auth.v1.AuthService/Register)

ALICE_TOKEN=$(echo "$ALICE_RESPONSE" | jq -r '.accessToken')

grpcurl -plaintext \
  -H "authorization: Bearer $ALICE_TOKEN" \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555",
    "desired_status": "PART_LOOKING_FOR_TEAM",
    "wished_role_ids": [
      "00000000-0000-0000-0000-000000000002",
      "00000000-0000-0000-0000-000000000004"
    ],
    "motivation_text": "Looking for backend/fullstack team",
    "idempotency_key": {"key": "alice-register-team-1"}
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/RegisterForHackathon
```

**Response:**
```json
{
  "participation": {
    "hackathonId": "55555555-5555-5555-5555-555555555555",
    "userId": "alice-uuid",
    "status": "PART_LOOKING_FOR_TEAM",
    "teamId": "",
    "profile": {
      "wishedRoles": [
        {"id": "00000000-0000-0000-0000-000000000002", "name": "Backend"},
        {"id": "00000000-0000-0000-0000-000000000004", "name": "Fullstack"}
      ],
      "motivationText": "Looking for backend/fullstack team"
    },
    "registeredAt": "2026-02-10T12:01:00Z",
    "updatedAt": "2026-02-10T12:01:00Z"
  }
}
```

### 4. GetMyParticipation (Получить свою регистрацию)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555"
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/GetMyParticipation
```

**Response:**
```json
{
  "participation": {
    "hackathonId": "55555555-5555-5555-5555-555555555555",
    "userId": "generated-uuid",
    "status": "PART_INDIVIDUAL",
    "teamId": "",
    "profile": {
      "wishedRoles": [
        {"id": "00000000-0000-0000-0000-000000000003", "name": "Frontend"}
      ],
      "motivationText": "I love frontend development!"
    },
    "registeredAt": "2026-02-10T12:00:00Z",
    "updatedAt": "2026-02-10T12:00:00Z"
  }
}
```

### 5. UpdateMyParticipation (Обновить профиль)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555",
    "wished_role_ids": [
      "00000000-0000-0000-0000-000000000003",
      "00000000-0000-0000-0000-000000000006"
    ],
    "motivation_text": "I love frontend and design!",
    "idempotency_key": {"key": "user-update-profile-1"}
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/UpdateMyParticipation
```

**Response:**
```json
{
  "participation": {
    "hackathonId": "55555555-5555-5555-5555-555555555555",
    "userId": "generated-uuid",
    "status": "PART_INDIVIDUAL",
    "teamId": "",
    "profile": {
      "wishedRoles": [
        {"id": "00000000-0000-0000-0000-000000000003", "name": "Frontend"},
        {"id": "00000000-0000-0000-0000-000000000006", "name": "Designer"}
      ],
      "motivationText": "I love frontend and design!"
    },
    "registeredAt": "2026-02-10T12:00:00Z",
    "updatedAt": "2026-02-10T12:05:00Z"
  }
}
```

### 6. SwitchParticipationMode (Переключить режим)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555",
    "new_status": "PART_LOOKING_FOR_TEAM",
    "idempotency_key": {"key": "user-switch-mode-1"}
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/SwitchParticipationMode
```

**Response:**
```json
{
  "participation": {
    "hackathonId": "55555555-5555-5555-5555-555555555555",
    "userId": "generated-uuid",
    "status": "PART_LOOKING_FOR_TEAM",
    "teamId": "",
    "profile": {
      "wishedRoles": [
        {"id": "00000000-0000-0000-0000-000000000003", "name": "Frontend"},
        {"id": "00000000-0000-0000-0000-000000000006", "name": "Designer"}
      ],
      "motivationText": "I love frontend and design!"
    },
    "registeredAt": "2026-02-10T12:00:00Z",
    "updatedAt": "2026-02-10T12:10:00Z"
  }
}
```

### 7. GetUserParticipation (Staff просматривает участие)

```bash
# Получаем токен staff пользователя (Alice)
# Сначала нужно зарегистрировать Alice через Register
STAFF_RESPONSE=$(grpcurl -plaintext -d '{
  "username": "alice_staff_grpc",
  "email": "alice_staff_grpc@example.com",
  "password": "SecurePass123",
  "first_name": "Alice",
  "last_name": "Staff",
  "timezone": "UTC"
}' localhost:50051 auth.v1.AuthService/Register)

STAFF_TOKEN=$(echo "$STAFF_RESPONSE" | jq -r '.accessToken')

# Просмотр участия другого пользователя
grpcurl -plaintext \
  -H "authorization: Bearer $STAFF_TOKEN" \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555",
    "user_id": "b0b00000-0000-0000-0000-000000000002"
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/GetUserParticipation
```

**Response:**
```json
{
  "participation": {
    "hackathonId": "55555555-5555-5555-5555-555555555555",
    "userId": "b0b00000-0000-0000-0000-000000000002",
    "status": "PART_INDIVIDUAL",
    "teamId": "",
    "profile": {
      "wishedRoles": [],
      "motivationText": "I love to code and build!"
    },
    "registeredAt": "2026-02-10T11:00:00Z",
    "updatedAt": "2026-02-10T11:00:00Z"
  }
}
```

### 8. ListHackathonParticipants (Участник просматривает всех участников)

```bash
# Любой зарегистрированный участник может просматривать список
grpcurl -plaintext \
  -H "authorization: Bearer $BOB_TOKEN" \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555"
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/ListHackathonParticipants
```

**Response:**
```json
{
  "participants": [
    {
      "hackathonId": "55555555-5555-5555-5555-555555555555",
      "userId": "c4a41e00-0000-0000-0000-000000000003",
      "status": "PART_LOOKING_FOR_TEAM",
      "teamId": "",
      "profile": {
        "wishedRoles": [
          {"id": "00000000-0000-0000-0000-000000000002", "name": "Backend"},
          {"id": "00000000-0000-0000-0000-000000000004", "name": "Fullstack"}
        ],
        "motivationText": "Want to find a great team to collaborate with"
      },
      "registeredAt": "2026-02-10T11:00:00Z",
      "updatedAt": "2026-02-10T11:00:00Z"
    },
    {
      "hackathonId": "55555555-5555-5555-5555-555555555555",
      "userId": "b0b00000-0000-0000-0000-000000000002",
      "status": "PART_INDIVIDUAL",
      "teamId": "",
      "profile": {
        "wishedRoles": [],
        "motivationText": "I love to code and build!"
      },
      "registeredAt": "2026-02-10T10:00:00Z",
      "updatedAt": "2026-02-10T10:00:00Z"
    }
  ],
  "page": {
    "nextPageToken": ""
  }
}
```

### 9. ListHackathonParticipants с фильтрацией (по статусу)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $BOB_TOKEN" \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555",
    "status_filter": {
      "statuses": ["PART_LOOKING_FOR_TEAM"]
    }
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/ListHackathonParticipants
```

**Response:**
```json
{
  "participants": [
    {
      "hackathonId": "55555555-5555-5555-5555-555555555555",
      "userId": "c4a41e00-0000-0000-0000-000000000003",
      "status": "PART_LOOKING_FOR_TEAM",
      "teamId": "",
      "profile": {
        "wishedRoles": [
          {"id": "00000000-0000-0000-0000-000000000002", "name": "Backend"},
          {"id": "00000000-0000-0000-0000-000000000004", "name": "Fullstack"}
        ],
        "motivationText": "Want to find a great team to collaborate with"
      },
      "registeredAt": "2026-02-10T11:00:00Z",
      "updatedAt": "2026-02-10T11:00:00Z"
    }
  ],
  "page": {
    "nextPageToken": ""
  }
}
```

### 10. ListHackathonParticipants с фильтрацией (по ролям)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $BOB_TOKEN" \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555",
    "wished_role_ids_filter": ["00000000-0000-0000-0000-000000000002"]
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/ListHackathonParticipants
```

**Response:**
```json
{
  "participants": [
    {
      "hackathonId": "55555555-5555-5555-5555-555555555555",
      "userId": "c4a41e00-0000-0000-0000-000000000003",
      "status": "PART_LOOKING_FOR_TEAM",
      "teamId": "",
      "profile": {
        "wishedRoles": [
          {"id": "00000000-0000-0000-0000-000000000002", "name": "Backend"},
          {"id": "00000000-0000-0000-0000-000000000004", "name": "Fullstack"}
        ],
        "motivationText": "Want to find a great team to collaborate with"
      },
      "registeredAt": "2026-02-10T11:00:00Z",
      "updatedAt": "2026-02-10T11:00:00Z"
    }
  ],
  "page": {
    "nextPageToken": ""
  }
}
```

> **Note**: Возвращаются только участники, у которых есть хотя бы одна из указанных ролей.

### 11. ListHackathonParticipants с пагинацией

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $BOB_TOKEN" \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555",
    "query": {
      "page": {
        "page_size": 10
      }
    }
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/ListHackathonParticipants
```

**Response:**
```json
{
  "participants": [...],
  "page": {
    "nextPageToken": ""
  }
}
```

> **Note**: Default page_size = 20, max = 100

### 12. UnregisterFromHackathon (Отмена регистрации)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555",
    "idempotency_key": {"key": "user-unregister-1"}
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/UnregisterFromHackathon
```

**Response:**
```json
{}
```

> **Ожидаемо**: Пустой ответ означает успешное удаление регистрации.

### 13. GetMyParticipation после отмены (должен вернуть ошибку)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555"
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/GetMyParticipation
```

**Response:**
```json
{
  "code": 5,
  "message": "not found",
  "details": []
}
```

---

## Service-to-Service Methods

> **Важно**: Эти методы вызываются только из других сервисов (Team Service).

### 14. ConvertToTeamParticipation (Team Service only)

```bash
# Этот метод вызывается Team Service при создании команды
# Для тестирования можно вызвать напрямую из grpcurl

# Сначала зарегистрируем Bob
BOB_RESPONSE=$(grpcurl -plaintext -d '{
  "username": "bob_convert",
  "email": "bob_convert@example.com",
  "password": "SecurePass123",
  "first_name": "Bob",
  "last_name": "Convert",
  "timezone": "UTC"
}' localhost:50051 auth.v1.AuthService/Register)

BOB_TOKEN=$(echo "$BOB_RESPONSE" | jq -r '.accessToken')
BOB_USER_ID=$(grpcurl -plaintext -d '{"access_token": "'$BOB_TOKEN'"}' \
  localhost:50051 auth.v1.AuthService/Introspect | jq -r '.userId')

# Зарегистрируем Bob
grpcurl -plaintext \
  -H "authorization: Bearer $BOB_TOKEN" \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555",
    "desired_status": "PART_INDIVIDUAL",
    "wished_role_ids": [],
    "motivation_text": "Testing conversion"
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/RegisterForHackathon

# Теперь конвертируем в team member (обычно вызывается Team Service)
grpcurl -plaintext \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555",
    "user_id": "'$BOB_USER_ID'",
    "team_id": "77777777-7777-7777-7777-777777777777",
    "is_captain": false,
    "idempotency_key": {"key": "convert-bob-to-team-1"}
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/ConvertToTeamParticipation
```

**Response:**
```json
{
  "participation": {
    "hackathonId": "55555555-5555-5555-5555-555555555555",
    "userId": "bob-uuid",
    "status": "PART_TEAM_MEMBER",
    "teamId": "77777777-7777-7777-7777-777777777777"
  }
}
```

### 15. ConvertFromTeamParticipation (Team Service only)

```bash
# Конвертируем Bob обратно из команды
grpcurl -plaintext \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555",
    "user_id": "'$BOB_USER_ID'",
    "idempotency_key": {"key": "convert-bob-from-team-1"}
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/ConvertFromTeamParticipation
```

**Response:**
```json
{
  "participation": {
    "hackathonId": "55555555-5555-5555-5555-555555555555",
    "userId": "bob-uuid",
    "status": "PART_LOOKING_FOR_TEAM",
    "teamId": ""
  }
}
```

---

## Негативные сценарии

### Fail 1: Register twice (409 Conflict)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555",
    "desired_status": "PART_LOOKING_FOR_TEAM",
    "wished_role_ids": [],
    "motivation_text": "Second registration"
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/RegisterForHackathon
```

**Response:**
```json
{
  "code": 9,
  "message": "conflict: already registered",
  "details": []
}
```

### Fail 2: Non-staff lists participants (403 Forbidden)

```bash
# Используем токен обычного участника (не staff)
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555"
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/ListHackathonParticipants
```

**Response:**
```json
{
  "code": 7,
  "message": "forbidden: only staff or participants can list participants",
  "details": []
}
```

> **Note**: Теперь это fail case для пользователя, который вообще не зарегистрирован на хакатоне.

### Fail 3: Switch to same status (403 Forbidden)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $ALICE_TOKEN" \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555",
    "new_status": "PART_LOOKING_FOR_TEAM"
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/SwitchParticipationMode
```

**Response:**
```json
{
  "code": 7,
  "message": "policy violation: new status must be different from current status",
  "details": []
}
```

### Fail 4: Get participation without registration (404 Not Found)

```bash
# Новый пользователь без регистрации
NEW_RESPONSE=$(grpcurl -plaintext -d '{
  "username": "new_user_'$RANDOM'",
  "email": "new_user_'$RANDOM'@example.com",
  "password": "SecurePass123",
  "first_name": "New",
  "last_name": "User",
  "timezone": "UTC"
}' localhost:50051 auth.v1.AuthService/Register)

NEW_TOKEN=$(echo "$NEW_RESPONSE" | jq -r '.accessToken')

grpcurl -plaintext \
  -H "authorization: Bearer $NEW_TOKEN" \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555"
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/GetMyParticipation
```

**Response:**
```json
{
  "code": 5,
  "message": "not found",
  "details": []
}
```

### Fail 5: Register with invalid role IDs (400 Invalid Argument)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $NEW_TOKEN" \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555",
    "desired_status": "PART_INDIVIDUAL",
    "wished_role_ids": ["99999999-9999-9999-9999-999999999999"],
    "motivation_text": "Invalid role"
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/RegisterForHackathon
```

**Response:**
```json
{
  "code": 3,
  "message": "invalid input: some team role IDs are invalid",
  "details": []
}
```

### Fail 6: Switch to TEAM_MEMBER (400 Invalid Argument)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $ALICE_TOKEN" \
  -d '{
    "hackathon_id": "55555555-5555-5555-5555-555555555555",
    "new_status": "PART_TEAM_MEMBER"
  }' \
  localhost:50055 participationandroles.v1.ParticipationService/SwitchParticipationMode
```

**Response:**
```json
{
  "code": 3,
  "message": "new_status must be INDIVIDUAL_ACTIVE or LOOKING_FOR_TEAM",
  "details": []
}
```

---

## Troubleshooting

### Connection refused

**Причина**: Сервис не запущен или работает на другом порту.

**Решение**:
```bash
# Проверить статус сервисов
docker-compose -f deployments/docker-compose.yml ps

# Проверить логи
docker-compose -f deployments/docker-compose.yml logs participation-and-roles-service
```

### Unauthenticated (code 16)

**Причина**: Не передан или невалидный access_token.

**Решение**:
```bash
# Получите новый токен
grpcurl -plaintext -d '{
  "username": "test_participant",
  "password": "SecurePass123"
}' localhost:50051 auth.v1.AuthService/Login | jq -r '.accessToken'
```

### Not found (code 5)

**Причина**: Участие не существует.

**Решение**: Сначала зарегистрируйтесь через RegisterForHackathon.

### Forbidden (code 7)

**Причина**: Недостаточно прав.

**Решение**: 
- Для GetUserParticipation / ListHackathonParticipants: нужно быть либо staff, либо участником хакатона
- Для UpdateMyParticipation / SwitchMode: проверьте текущий статус

---

## Дополнительно

- gRPC endpoint: `localhost:50055`
- HTTP gateway: `localhost:8080`
- REST guide: [rest-guide-participation.md](./rest-guide-participation.md)
- Proto definition: `/proto/participationandroles/v1/participation_service.proto`

## Useful commands

### List all methods

```bash
grpcurl -plaintext localhost:50055 list participationandroles.v1.ParticipationService
```

### Describe method

```bash
grpcurl -plaintext localhost:50055 describe participationandroles.v1.ParticipationService.RegisterForHackathon
```

### List all services

```bash
grpcurl -plaintext localhost:50055 list
```
