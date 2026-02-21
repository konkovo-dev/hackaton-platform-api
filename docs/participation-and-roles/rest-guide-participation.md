# Participation Service REST Test Cases

## 📋 Содержание
- [🚀 Быстрый старт](#-быстрый-старт)
- [Подготовка тестовых данных](#подготовка-тестовых-данных)
- [Тест-сценарии (Happy Path)](#тест-сценарии)
- [Негативные сценарии](#негативные-сценарии)
- [Service-to-Service методы](#service-to-service-methods)
- [Troubleshooting](#troubleshooting)

## Предусловия
- Запущен в докере с помощью [docker-setup.md](../docker-setup.md)
- PostgreSQL мигрирован для всех сервисов

## Эндпоинты

Base URL: `http://localhost:8080`

---

## 🚀 Быстрый старт

### Вариант 1: Автоматические тесты (рекомендуется)

Просто запустите автоматический скрипт - он сам создаст пользователей, хакатон и протестирует все endpoints:

```bash
# Сделать скрипт исполняемым
chmod +x docs/participation-and-roles/rest-script-participation.sh

# Запустить тесты
./docs/participation-and-roles/rest-script-participation.sh
```

### Вариант 2: Полная команда для копирования (все шаги сразу)

Скопируйте и вставьте весь блок команд ниже в терминал:

```bash
# Регистрация пользователей
ALICE_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice_'$(date +%s)'","email":"alice_'$(date +%s)'@test.com","password":"SecurePass123","first_name":"Alice","last_name":"Staff","timezone":"UTC"}')
ALICE_TOKEN=$(echo $ALICE_RESPONSE | jq -r '.accessToken')

BOB_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"bob_'$(date +%s)'","email":"bob_'$(date +%s)'@test.com","password":"SecurePass123","first_name":"Bob","last_name":"Participant","timezone":"UTC"}')
BOB_TOKEN=$(echo $BOB_RESPONSE | jq -r '.accessToken')

DIANA_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"diana_'$(date +%s)'","email":"diana_'$(date +%s)'@test.com","password":"SecurePass123","first_name":"Diana","last_name":"New","timezone":"UTC"}')
DIANA_TOKEN=$(echo $DIANA_RESPONSE | jq -r '.accessToken')

# Создание хакатона
CREATE_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/hackathons \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Hackathon","short_description":"Testing","description":"Test","location":{"online":false,"country":"Russia","city":"Moscow","venue":"Test"},"dates":{"registration_opens_at":"2026-03-01T00:00:00Z","registration_closes_at":"2026-03-20T23:59:59Z","starts_at":"2026-03-25T10:00:00Z","ends_at":"2026-03-27T18:00:00Z","judging_ends_at":"2026-03-30T18:00:00Z"},"registration_policy":{"allow_individual":true,"allow_team":true},"limits":{"team_size_max":5}}')
HACKATHON_ID=$(echo $CREATE_RESPONSE | jq -r '.hackathonId')

# Публикация хакатона
curl -s -X POST http://localhost:8080/v1/hackathons/$HACKATHON_ID:publish \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}' > /dev/null

# Получение team roles
ROLES_RESPONSE=$(curl -s "http://localhost:8080/v1/team-roles" -H "Authorization: Bearer $BOB_TOKEN")
FRONTEND_ROLE_ID=$(echo $ROLES_RESPONSE | jq -r '.teamRoles[] | select(.name == "Frontend") | .id')
DESIGNER_ROLE_ID=$(echo $ROLES_RESPONSE | jq -r '.teamRoles[] | select(.name == "Designer") | .id')

# Вывод переменных
echo "=== Переменные для тестирования ==="
echo "ALICE_TOKEN=$ALICE_TOKEN"
echo "BOB_TOKEN=$BOB_TOKEN"
echo "DIANA_TOKEN=$DIANA_TOKEN"
echo "HACKATHON_ID=$HACKATHON_ID"
echo "FRONTEND_ROLE_ID=$FRONTEND_ROLE_ID"
echo "DESIGNER_ROLE_ID=$DESIGNER_ROLE_ID"
echo "=================================="
echo "Теперь можете использовать эти переменные в примерах ниже!"
```

### Вариант 3: Ручное тестирование (шаг за шагом)

Если хотите тестировать вручную с пояснениями, следуйте инструкциям ниже по шагам.

---

---

## Подготовка тестовых данных

Для тестирования participation нужно:
1. Создать базовую структуру (хакатон, team roles, несколько пользователей)
2. Зарегистрировать тестовых пользователей через API

### Вариант 1: Автоматические тесты (рекомендуется)

Просто запустите автоматические тестовые скрипты - они сами создадут всё необходимое:

```bash
# Сделать скрипты исполняемыми
chmod +x docs/participation-and-roles/rest-script-participation.sh

# Запустить тесты (создаст пользователей, хакатон и протестирует все endpoints)
./docs/participation-and-roles/rest-script-participation.sh
```

### Вариант 2: Ручное тестирование

Если хотите тестировать вручную, сначала загрузите базовые данные:

#### Шаг 1: Загрузите тестовые данные в БД

```bash
# Из корня проекта
docker-compose -f deployments/docker-compose.yml exec -T postgres \
  psql -U hackathon -d hackathon < docs/participation-and-roles/test-data-participation.sql
```

Это создаст:
- Хакатон "Spring Hackathon 2026" (ID: `55555555-5555-5555-5555-555555555555`)
- 10 team roles (Backend, Frontend, etc.)
- Структуру для 4 пользователей (без паролей)
- 2 существующих participations для демонстрации

#### Шаг 2: Зарегистрируйте тестовых пользователей

> **Важно**: Тестовые данные НЕ создают пароли. Нужно зарегистрировать пользователей через API.

```sql
-- ============================================
-- Пользователи для тестирования participations
-- ============================================

-- Alice: Owner хакатона (staff member)
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('a11ce000-0000-0000-0000-000000000001', 'alice_staff', 'Alice', 'Staff', 'https://example.com/alice.jpg', 'UTC', NOW(), NOW());

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('a11ce000-0000-0000-0000-000000000001', 'public', 'public');

-- Bob: Participant (individual active)
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('b0b00000-0000-0000-0000-000000000002', 'bob_participant', 'Bob', 'Participant', 'https://example.com/bob.jpg', 'UTC', NOW(), NOW());

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('b0b00000-0000-0000-0000-000000000002', 'public', 'public');

-- Charlie: Participant (looking for team)
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('c4a41e00-0000-0000-0000-000000000003', 'charlie_participant', 'Charlie', 'Seeker', '', 'Europe/London', NOW(), NOW());

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('c4a41e00-0000-0000-0000-000000000003', 'public', 'public');

-- Diana: New participant
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('d1a4a000-0000-0000-0000-000000000004', 'diana_new', 'Diana', 'New', '', 'America/New_York', NOW(), NOW());

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('d1a4a000-0000-0000-0000-000000000004', 'public', 'public');

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
  '55555555-5555-5555-5555-555555555555',
  'Spring Hackathon 2026',
  'Code, Create, Innovate',
  'Join us for an amazing hackathon experience.',
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
  ('55555555-5555-5555-5555-555555555555', 'a11ce000-0000-0000-0000-000000000001', 'owner', NOW());

-- ============================================
-- Team Roles (если еще не созданы)
-- ============================================

INSERT INTO participation_and_roles.team_role_catalog (id, name) VALUES
  ('00000000-0000-0000-0000-000000000001', 'Any'),
  ('00000000-0000-0000-0000-000000000002', 'Backend'),
  ('00000000-0000-0000-0000-000000000003', 'Frontend'),
  ('00000000-0000-0000-0000-000000000004', 'Fullstack'),
  ('00000000-0000-0000-0000-000000000005', 'Mobile'),
  ('00000000-0000-0000-0000-000000000006', 'Designer'),
  ('00000000-0000-0000-0000-000000000007', 'Product Manager'),
  ('00000000-0000-0000-0000-000000000008', 'Data Scientist'),
  ('00000000-0000-0000-0000-000000000009', 'DevOps'),
  ('00000000-0000-0000-0000-00000000000a', 'QA')
ON CONFLICT (id) DO NOTHING;

-- ============================================
-- Existing participations для Bob и Charlie
-- ============================================

INSERT INTO participation_and_roles.participations (
  hackathon_id, user_id, status, team_id, motivation_text, registered_at, created_at, updated_at
)
VALUES 
  ('55555555-5555-5555-5555-555555555555', 'b0b00000-0000-0000-0000-000000000002', 'individual', NULL, 'Love to code!', NOW(), NOW(), NOW()),
  ('55555555-5555-5555-5555-555555555555', 'c4a41e00-0000-0000-0000-000000000003', 'looking_for_team', NULL, 'Want to find a great team', NOW(), NOW(), NOW());

-- Wished roles для Charlie
INSERT INTO participation_and_roles.participation_wished_roles (hackathon_id, user_id, team_role_id)
VALUES 
  ('55555555-5555-5555-5555-555555555555', 'c4a41e00-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000002'),
  ('55555555-5555-5555-5555-555555555555', 'c4a41e00-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000004');
```

### Шаг 3: Проверьте данные

```sql
-- Проверить пользователей
SELECT id, username, first_name, last_name FROM identity.users 
WHERE id IN (
  'a11ce000-0000-0000-0000-000000000001',
  'b0b00000-0000-0000-0000-000000000002',
  'c4a41e00-0000-0000-0000-000000000003',
  'd1a4a000-0000-0000-0000-000000000004'
);

-- Проверить хакатон
SELECT id, name, stage FROM hackaton.hackathons 
WHERE id = '55555555-5555-5555-5555-555555555555';

-- Проверить participations
SELECT hackathon_id, user_id, status FROM participation_and_roles.participations
WHERE hackathon_id = '55555555-5555-5555-5555-555555555555';

-- Проверить team roles
SELECT id, name FROM participation_and_roles.team_role_catalog LIMIT 5;
```

Должно вернуть:
- 4 пользователя
- 1 хакатон в статусе `published`
- 2 participations (Bob = individual, Charlie = looking_for_team)
- 10 team roles

---

## Тест-сценарии

> **Важно**: Все запросы **требуют авторизации**. Используйте `ACCESS_TOKEN` из раздела выше.

### 1. ListTeamRoles (Public endpoint)

```bash
curl "http://localhost:8080/v1/team-roles" \
  -H "Authorization: Bearer $BOB_TOKEN" | jq .
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

### 2. RegisterForHackathon (Diana registers as individual)

```bash
# Получаем один из role IDs (Frontend)
ROLES_RESPONSE=$(curl -s "http://localhost:8080/v1/team-roles" \
  -H "Authorization: Bearer $DIANA_TOKEN")
FRONTEND_ROLE_ID=$(echo $ROLES_RESPONSE | jq -r '.teamRoles[] | select(.name == "Frontend") | .id')

# Diana регистрируется
curl -X POST "http://localhost:8080/v1/hackathons/$HACKATHON_ID/participations:register" \
  -H "Authorization: Bearer $DIANA_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "desired_status": "PART_INDIVIDUAL",
    "wished_role_ids": ["'$FRONTEND_ROLE_ID'"],
    "motivation_text": "I love frontend development!",
    "idempotency_key": {"key": "diana-register-1"}
  }' | jq .
```

**Response:**
```json
{
  "participation": {
    "hackathonId": "55555555-5555-5555-5555-555555555555",
    "userId": "d1a4a000-0000-0000-0000-000000000004",
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

### 3. GetMyParticipation (Diana checks her registration)

```bash
curl "http://localhost:8080/v1/hackathons/$HACKATHON_ID/participations/me" \
  -H "Authorization: Bearer $DIANA_TOKEN" | jq .
```

**Response:**
```json
{
  "participation": {
    "hackathonId": "55555555-5555-5555-5555-555555555555",
    "userId": "d1a4a000-0000-0000-0000-000000000004",
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

### 4. UpdateMyParticipation (Diana updates her profile)

```bash
# Получаем role IDs (Frontend и Designer)
DESIGNER_ROLE_ID=$(echo $ROLES_RESPONSE | jq -r '.teamRoles[] | select(.name == "Designer") | .id')

curl -X PUT "http://localhost:8080/v1/hackathons/$HACKATHON_ID/participations/me" \
  -H "Authorization: Bearer $DIANA_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "wished_role_ids": ["'$FRONTEND_ROLE_ID'", "'$DESIGNER_ROLE_ID'"],
    "motivation_text": "I love frontend and design!",
    "idempotency_key": {"key": "diana-update-1"}
  }' | jq .
```

**Response:**
```json
{
  "participation": {
    "hackathonId": "55555555-5555-5555-5555-555555555555",
    "userId": "d1a4a000-0000-0000-0000-000000000004",
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

### 5. SwitchParticipationMode (Diana switches to looking for team)

```bash
curl -X POST "http://localhost:8080/v1/hackathons/$HACKATHON_ID/participations/me:switchMode" \
  -H "Authorization: Bearer $DIANA_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "new_status": "PART_LOOKING_FOR_TEAM",
    "idempotency_key": {"key": "diana-switch-1"}
  }' | jq .
```

**Response:**
```json
{
  "participation": {
    "hackathonId": "55555555-5555-5555-5555-555555555555",
    "userId": "d1a4a000-0000-0000-0000-000000000004",
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

### 6. GetUserParticipation (Bob views Diana's participation)

```bash
# Сначала зарегистрируем Bob
curl -X POST "http://localhost:8080/v1/hackathons/$HACKATHON_ID/participations:register" \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "desired_status": "PART_INDIVIDUAL",
    "wished_role_ids": [],
    "motivation_text": "I want to participate!"
  }' | jq .

# Получаем user_id Diana
DIANA_USER_ID=$(curl -s -X POST http://localhost:8080/v1/auth/introspect \
  -H "Content-Type: application/json" \
  -d '{"access_token": "'$DIANA_TOKEN'"}' | jq -r '.userId')

# Bob (тоже участник) просматривает участие Diana
curl "http://localhost:8080/v1/hackathons/$HACKATHON_ID/participations/users/$DIANA_USER_ID" \
  -H "Authorization: Bearer $BOB_TOKEN" | jq .
```

**Response:**
```json
{
  "participation": {
    "hackathonId": "55555555-5555-5555-5555-555555555555",
    "userId": "d1a4a000-0000-0000-0000-000000000004",
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

### 7. ListHackathonParticipants (Bob views all participants)

```bash
# Любой зарегистрированный участник может видеть список
curl -X POST "http://localhost:8080/v1/hackathons/$HACKATHON_ID/participations:list" \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}' | jq .
```

**Response:**
```json
{
  "participants": [
    {
      "hackathonId": "<your-hackathon-id>",
      "userId": "<diana-user-id>",
      "status": "PART_LOOKING_FOR_TEAM",
      "teamId": "",
      "profile": {
        "wishedRoles": [
          {"id": "...", "name": "Frontend"},
          {"id": "...", "name": "Designer"}
        ],
        "motivationText": "I love frontend and design!"
      },
      "registeredAt": "2026-02-10T12:00:00Z",
      "updatedAt": "2026-02-10T12:10:00Z"
    },
    {
      "hackathonId": "<your-hackathon-id>",
      "userId": "<bob-user-id>",
      "status": "PART_INDIVIDUAL",
      "teamId": "",
      "profile": {
        "wishedRoles": [],
        "motivationText": "I want to participate!"
      },
      "registeredAt": "2026-02-10T12:05:00Z",
      "updatedAt": "2026-02-10T12:05:00Z"
    }
  ],
  "page": {
    "nextPageToken": ""
  }
}
```

> **Note**: Теперь Bob тоже зарегистрирован, поэтому видны оба участника.

### 8. ListHackathonParticipants with filters (Bob filters by status)

```bash
curl -X POST "http://localhost:8080/v1/hackathons/$HACKATHON_ID/participations:list" \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status_filter": {
      "statuses": ["PART_LOOKING_FOR_TEAM"]
    }
  }' | jq .
```

**Response:**
```json
{
  "participants": [
    {
      "hackathonId": "<your-hackathon-id>",
      "userId": "<diana-user-id>",
      "status": "PART_LOOKING_FOR_TEAM",
      "teamId": "",
      "profile": {
        "wishedRoles": [
          {"id": "...", "name": "Frontend"},
          {"id": "...", "name": "Designer"}
        ],
        "motivationText": "I love frontend and design!"
      },
      "registeredAt": "2026-02-10T12:00:00Z",
      "updatedAt": "2026-02-10T12:10:00Z"
    }
  ],
  "page": {
    "nextPageToken": ""
  }
}
```

> **Note**: Фильтр вернет только участников со статусом LOOKING_FOR_TEAM.

### 9. UnregisterFromHackathon (Diana unregisters)

```bash
curl -X POST "http://localhost:8080/v1/hackathons/$HACKATHON_ID/participations/me:unregister" \
  -H "Authorization: Bearer $DIANA_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "diana-unregister-1"}
  }' | jq .
```

**Response:**
```json
{}
```

> **Ожидаемо**: Пустой ответ означает успешное удаление registration.

### 10. GetMyParticipation after unregister (should fail)

```bash
curl "http://localhost:8080/v1/hackathons/$HACKATHON_ID/participations/me" \
  -H "Authorization: Bearer $DIANA_TOKEN" | jq .
```

**Response:**
```json
{
  "code": 7,
  "message": "forbidden: user is not a participant in this hackathon",
  "details": []
}
```

---

## Негативные сценарии

### Try to register twice (should fail)

```bash
# Сначала Bob регистрируется
curl -X POST "http://localhost:8080/v1/hackathons/$HACKATHON_ID/participations:register" \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "desired_status": "PART_INDIVIDUAL",
    "wished_role_ids": [],
    "motivation_text": "First registration"
  }' | jq .

# Затем пытается зарегистрироваться снова (должно вернуть ошибку)
curl -X POST "http://localhost:8080/v1/hackathons/$HACKATHON_ID/participations:register" \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "desired_status": "PART_LOOKING_FOR_TEAM",
    "wished_role_ids": [],
    "motivation_text": "Second registration"
  }' | jq .
```

**Response:**
```json
{
  "code": 9,
  "message": "conflict: already registered",
  "details": []
}
```

### Try to view participants as non-registered user (should fail)

```bash
# Новый пользователь, не зарегистрированный на хакатон
NEW_USER_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "eve_nonreg_'$(date +%s)'",
    "email": "eve_nonreg_'$(date +%s)'@test.com",
    "password": "SecurePass123",
    "first_name": "Eve",
    "last_name": "NonReg",
    "timezone": "UTC"
  }')

EVE_TOKEN=$(echo $NEW_USER_RESPONSE | jq -r '.accessToken')

curl -X POST "http://localhost:8080/v1/hackathons/$HACKATHON_ID/participations:list" \
  -H "Authorization: Bearer $EVE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}' | jq .
```

**Response:**
```json
{
  "code": 7,
  "message": "forbidden: only staff or participants can list participants",
  "details": []
}
```

### Try to switch to same status (should fail)

```bash
# Charlie регистрируется как LOOKING_FOR_TEAM
BACKEND_ROLE_ID=$(echo $ROLES_RESPONSE | jq -r '.teamRoles[] | select(.name == "Backend") | .id')

curl -X POST "http://localhost:8080/v1/hackathons/$HACKATHON_ID/participations:register" \
  -H "Authorization: Bearer $CHARLIE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "desired_status": "PART_LOOKING_FOR_TEAM",
    "wished_role_ids": ["'$BACKEND_ROLE_ID'"],
    "motivation_text": "Looking for team"
  }' | jq .

# Пытается переключиться на тот же статус (должно вернуть ошибку)
curl -X POST "http://localhost:8080/v1/hackathons/$HACKATHON_ID/participations/me:switchMode" \
  -H "Authorization: Bearer $CHARLIE_TOKEN" \
  -H "Authorization: Bearer $CHARLIE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "new_status": "PART_LOOKING_FOR_TEAM"
  }' | jq .
```

**Response:**
```json
{
  "code": 7,
  "message": "policy violation: new status must be different from current status",
  "details": []
}
```

---

## Service-to-Service Methods

> **Важно**: Эти методы НЕ имеют HTTP/REST mapping. Они доступны только через gRPC.

### ConvertToTeamParticipation (Team Service only)

Вызывается Team Service при создании команды или добавлении участника в команду.

**gRPC метод:** `ConvertToTeamParticipation`

**Пример вызова (только для демонстрации, недоступно через REST):**

```bash
# Это service-to-service метод, вызывается только Team Service через gRPC
# Для тестирования используйте grpcurl (см. grpc-guide-participation.md)

# Пример входных данных:
# {
#   hackathon_id: "$HACKATHON_ID"
#   user_id: "$BOB_USER_ID"
#   team_id: "77777777-7777-7777-7777-777777777777"
#   is_captain: false
# }
```

**Ожидаемое поведение:**
- Participant status: `individual` или `looking_for_team` → `team_member` или `team_captain`
- Participant team_id: `NULL` → `<team_id>`

### ConvertFromTeamParticipation (Team Service only)

Вызывается Team Service при выходе из команды или удалении команды.

**gRPC метод:** `ConvertFromTeamParticipation`

**Пример вызова (только для демонстрации, недоступно через REST):**

```bash
# Это service-to-service метод, вызывается только Team Service через gRPC
# Для тестирования используйте grpcurl (см. grpc-guide-participation.md)

# Пример входных данных:
# {
#   hackathon_id: "$HACKATHON_ID"
#   user_id: "$BOB_USER_ID"
# }
```

**Ожидаемое поведение:**
- Participant status: `team_member` или `team_captain` → `looking_for_team`
- Participant team_id: `<team_id>` → `NULL`

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
    "username": "diana_new",
    "password": "SecurePass123"
  }' | jq -r '.accessToken'
```

### 403 Forbidden

**Причина**: Пользователь не имеет необходимых прав.

**Решение**: Проверьте:
- Для ListHackathonParticipants / GetUserParticipation: нужна staff роль
- Для SwitchMode: нельзя переключаться из/в team статусы
- Для UnregisterFromHackathon: нельзя отменить если вы в команде

### 404 Not Found

**Причина**: Хакатон или participation не существует.

**Решение**: Проверьте корректность UUID и что пользователь зарегистрирован.

### 409 Conflict

**Причина**: Попытка создать дубликат (уже зарегистрирован).

**Решение**: Используйте UpdateMyParticipation для обновления профиля.

---

## Дополнительно

- gRPC endpoint: `localhost:50055`
- HTTP gateway: `localhost:8080`
- All HTTP endpoints доступны под `/v1/hackathons/*` и `/v1/team-roles`

