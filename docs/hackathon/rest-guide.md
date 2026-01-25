# Hackathon Service REST API Guide

## Предусловия
- Запущен в докере с помощью [docker-setup.md](../docker-setup.md)
- PostgreSQL мигрирован для всех сервисов (auth, identity, hackathon, participation-and-roles)

## Эндпоинты

Base URL: `http://localhost:8080`

---

## Подготовка: Регистрация пользователей

Все методы требуют авторизации.

```bash
# Регистрация владельца хакатона (Alice)
ALICE_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice_hackathon_owner",
    "email": "alice@hackathon.com",
    "password": "SecurePass123",
    "first_name": "Alice",
    "last_name": "Owner",
    "timezone": "UTC",
    "idempotency_key": {"key": "alice-hackathon-owner-reg"}
  }')

ALICE_TOKEN=$(echo $ALICE_RESPONSE | jq -r '.accessToken')
echo "Alice Token: $ALICE_TOKEN"

# Регистрация второго пользователя (Bob) для тестов доступа
BOB_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "bob_hackathon_viewer",
    "email": "bob@hackathon.com",
    "password": "SecurePass123",
    "first_name": "Bob",
    "last_name": "Viewer",
    "timezone": "UTC",
    "idempotency_key": {"key": "bob-hackathon-viewer-reg"}
  }')

BOB_TOKEN=$(echo $BOB_RESPONSE | jq -r '.accessToken')
echo "Bob Token: $BOB_TOKEN"
```

---

## 1. CreateHackathon (Создание хакатона)

**Endpoint:** `POST /v1/hackathons`

**Access:** Authenticated users (creator becomes OWNER)

**Description:** Создает новый хакатон в стадии `DRAFT`.

```bash
CREATE_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/hackathons \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "AI Innovation Hackathon 2026",
    "short_description": "Build the future with AI",
    "description": "Join us for an exciting 48-hour hackathon focused on AI and machine learning innovations.",
    "location": {
      "online": false,
      "country": "Russia",
      "city": "Moscow",
      "venue": "Digital October Center"
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
    "links": [
      {
        "title": "Official Website",
        "url": "https://ai-hackathon.example.com"
      }
    ],
    "idempotency_key": {"key": "ai-hackathon-2026-creation"}
  }')

HACKATHON_ID=$(echo "$CREATE_RESPONSE" | jq -r '.hackathonId')
echo "Hackathon ID: $HACKATHON_ID"
```

**Response:**
- `hackathonId`: UUID хакатона
- `validationErrors`: Список ошибок валидации (может быть пустым в DRAFT)

**Notes:**
- В DRAFT все поля необязательны (мягкая валидация)
- Validation errors возвращаются для информации, но не блокируют создание

---

## 2. GetHackathon (Получение хакатона)

**Endpoint:** `GET /v1/hackathons/{hackathon_id}`

**Access:**
- DRAFT: только OWNER/ORGANIZER
- Published: все авторизованные пользователи

**Query Parameters:**
- `include_description` (bool): включить полное описание
- `include_links` (bool): включить ссылки
- `include_limits` (bool): включить лимиты
- `include_task` (bool): включить задание (если доступно)
- `include_result` (bool): включить результаты (если доступно)

```bash
GET_RESPONSE=$(curl -s "http://localhost:8080/v1/hackathons/$HACKATHON_ID?include_description=true&include_links=true&include_task=true" \
  -H "Authorization: Bearer $ALICE_TOKEN")

echo "$GET_RESPONSE" | jq .
```

**Response:**
```json
{
  "hackathon": {
    "hackathonId": "uuid",
    "name": "...",
    "shortDescription": "...",
    "description": "...",
    "stage": "HACKATHON_STAGE_DRAFT",
    "state": "HACKATHON_STATE_DRAFT",
    "location": {...},
    "dates": {...},
    "registrationPolicy": {...},
    "limits": {...},
    "links": [...],
    "task": "...",
    "result": "...",
    "publishedAt": "...",
    "createdAt": "...",
    "updatedAt": "..."
  }
}
```

---

## 3. UpdateHackathon (Обновление хакатона)

**Endpoint:** `PUT /v1/hackathons/{hackathon_id}`

**Access:** OWNER/ORGANIZER

**Stage-based Restrictions:**

| Field | DRAFT | UPCOMING | REGISTRATION | PRESTART | RUNNING | JUDGING | FINISHED |
|-------|-------|----------|--------------|----------|---------|---------|----------|
| name, description | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| location | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| links | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| team_size_max | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| DisableType | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| EnableType | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| dates | ✅ | ✅* | ✅* | ✅* | ✅* | ✅* | ❌ |

\* Dates follow TYPE-A/TYPE-B rules in published stages

```bash
UPDATE_RESPONSE=$(curl -s -X PUT http://localhost:8080/v1/hackathons/$HACKATHON_ID \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "AI Innovation Hackathon 2026 (Updated)",
    "short_description": "Build the future with AI",
    "description": "Updated description",
    "location": {
      "online": false,
      "country": "Russia",
      "city": "Moscow",
      "venue": "Digital October Center"
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
      "team_size_max": 6
    }
  }')
```

**Response:**
- `validationErrors`: Список ошибок валидации

**Validation Modes:**
- **DRAFT (soft mode)**: Ошибки возвращаются, но сохранение происходит
- **Published (strict mode)**: Критические ошибки блокируют сохранение

---

## 4. ValidateHackathon (Валидация для публикации)

**Endpoint:** `GET /v1/hackathons/{hackathon_id}:validate`

**Access:** OWNER/ORGANIZER

**Description:** Проверяет готовность хакатона к публикации без изменения состояния.

```bash
VALIDATE_RESPONSE=$(curl -s -X GET "http://localhost:8080/v1/hackathons/$HACKATHON_ID:validate" \
  -H "Authorization: Bearer $ALICE_TOKEN")

echo "$VALIDATE_RESPONSE" | jq .
```

**Response:**
```json
{
  "validationErrors": [
    {
      "code": "REQUIRED",
      "field": "task",
      "message": "task is required for publishing"
    }
  ]
}
```

**Required for Publish (PublishReady):**
- `name != ""`
- `location != ""`
- `task` задан и валиден
- Все временные поля заданы
- `TIME_RULE` выполняется
- `at_least_one_true(allow_team, allow_individual)`

---

## 5. PublishHackathon (Публикация хакатона)

**Endpoint:** `POST /v1/hackathons/{hackathon_id}:publish`

**Access:** OWNER only

**Preconditions:**
- `stage == DRAFT`
- `published_at == null`
- `now < registration_opens_at`
- `PublishReady(hackathon)` выполняется

```bash
PUBLISH_RESPONSE=$(curl -s -X POST "http://localhost:8080/v1/hackathons/$HACKATHON_ID:publish" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json")
```

**Side Effects:**
- `published_at` устанавливается в `now`
- `state` меняется на `PUBLISHED`
- `stage` пересчитывается на основе времени

---

## 6. GetHackathonTask (Получение задания)

**Endpoint:** `GET /v1/hackathons/{hackathon_id}:task`

**Access:**
- OWNER/ORGANIZER: всегда
- MENTOR/JURY: после публикации (`stage != DRAFT`)
- Participants (SINGLE/TEAM): только на `stage == RUNNING`
- Others: нет доступа

```bash
TASK_RESPONSE=$(curl -s "http://localhost:8080/v1/hackathons/$HACKATHON_ID:task" \
  -H "Authorization: Bearer $ALICE_TOKEN")

echo "$TASK_RESPONSE" | jq .
```

**Response:**
```json
{
  "task": {
    "body": "Build an innovative AI solution..."
  }
}
```

---

## 7. UpdateHackathonTask (Обновление задания)

**Endpoint:** `PUT /v1/hackathons/{hackathon_id}:task`

**Access:** OWNER/ORGANIZER

**Stage Restrictions:**
- Allowed: `DRAFT`, `UPCOMING`, `REGISTRATION`, `PRESTART`, `RUNNING`
- **Forbidden**: `JUDGING`, `FINISHED`

```bash
UPDATE_TASK=$(curl -s -X PUT "http://localhost:8080/v1/hackathons/$HACKATHON_ID:task" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "task": "Build an innovative AI solution that solves a real-world problem.",
    "idempotency_key": {"key": "task-update-123"}
  }')

echo "$UPDATE_TASK" | jq .
```

**Response:**
```json
{
  "task": {
    "body": "..."
  },
  "validationErrors": []
}
```

**Validation:**
- `task != ""`
- Stage not in `{JUDGING, FINISHED}`

---

## 8. GetHackathonResult (Получение результата)

**Endpoint:** `GET /v1/hackathons/{hackathon_id}:result`

**Access:**
- **FINISHED stage**: все авторизованные пользователи (public)
- **JUDGING stage**: OWNER/ORGANIZER (если `result_published_at == null`)
- Others: нет доступа

```bash
RESULT_RESPONSE=$(curl -s "http://localhost:8080/v1/hackathons/$HACKATHON_ID:result" \
  -H "Authorization: Bearer $ALICE_TOKEN")

echo "$RESULT_RESPONSE" | jq .
```

**Response:**
```json
{
  "result": {
    "body": "Congratulations to Team Alpha for winning..."
  }
}
```

---

## 9. UpdateHackathonResultDraft (Обновление черновика результата)

**Endpoint:** `PUT /v1/hackathons/{hackathon_id}:result`

**Access:** OWNER/ORGANIZER

**Preconditions:**
- `stage == JUDGING`
- `result_published_at == null`

```bash
UPDATE_RESULT=$(curl -s -X PUT "http://localhost:8080/v1/hackathons/$HACKATHON_ID:result" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "result": "Congratulations to Team Alpha for winning the first place!",
    "idempotency_key": {"key": "result-update-123"}
  }')

echo "$UPDATE_RESULT" | jq .
```

**Response:**
```json
{
  "result": {
    "body": "..."
  },
  "validationErrors": []
}
```

---

## 10. PublishHackathonResult (Публикация результата)

**Endpoint:** `POST /v1/hackathons/{hackathon_id}:result:publish`

**Access:** OWNER/ORGANIZER

**Preconditions:**
- `stage == JUDGING`
- `result_published_at == null`
- `ResultReady(hackathon)` (result != "")

```bash
PUBLISH_RESULT=$(curl -s -X POST "http://localhost:8080/v1/hackathons/$HACKATHON_ID:result:publish" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json")

echo "$PUBLISH_RESULT" | jq .
```

**Side Effects:**
- `result_published_at` устанавливается в `now`
- `judging_ends_at` устанавливается в `now`
- `stage` меняется на `FINISHED`

---

## 11. ListHackathons (Список хакатонов)

**Endpoint:** `GET /v1/hackathons`

**Access:** Public (no auth required)

**Query Parameters:**
- `page_size` (int32): количество элементов на странице (default: 10, max: 100)
- `page_token` (string): токен для следующей страницы
- `include_description` (bool): включить полное описание
- `include_links` (bool): включить ссылки
- `include_limits` (bool): включить лимиты

```bash
LIST_RESPONSE=$(curl -s "http://localhost:8080/v1/hackathons?page_size=10&include_links=true")

echo "$LIST_RESPONSE" | jq .
```

**Response:**
```json
{
  "hackathons": [...],
  "page": {
    "nextPageToken": "...",
    "totalCount": 42
  }
}
```

**Notes:**
- Возвращает только опубликованные хакатоны (`state == PUBLISHED`)
- DRAFT хакатоны не включаются в список

---

## 12. CreateHackathonAnnouncement (Создание объявления)

**Endpoint:** `POST /v1/hackathons/{hackathon_id}/announcements`

**Access:** OWNER/ORGANIZER

**Preconditions:**
- `stage != DRAFT` (announcements запрещены в DRAFT)

```bash
ANNOUNCEMENT_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/hackathons/$HACKATHON_ID/announcements \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Registration Opening Soon!",
    "content": "We are excited to announce that registration will open on March 1st, 2026.",
    "idempotency_key": {"key": "announcement-1"}
  }')

ANNOUNCEMENT_ID=$(echo "$ANNOUNCEMENT_RESPONSE" | jq -r '.announcement.id')
```

---

## 13. ListHackathonAnnouncements (Список объявлений)

**Endpoint:** `GET /v1/hackathons/{hackathon_id}/announcements`

**Access:**
- Staff (OWNER/ORGANIZER/MENTOR/JURY)
- Participants (LOOKING_FOR_TEAM/SINGLE/TEAM)

**Preconditions:**
- `stage != DRAFT`

```bash
LIST_ANNOUNCEMENTS=$(curl -s "http://localhost:8080/v1/hackathons/$HACKATHON_ID/announcements?page_size=10" \
  -H "Authorization: Bearer $BOB_TOKEN")

echo "$LIST_ANNOUNCEMENTS" | jq .
```

---

## 14. UpdateHackathonAnnouncement (Обновление объявления)

**Endpoint:** `PUT /v1/hackathons/{hackathon_id}/announcements/{announcement_id}`

**Access:** OWNER/ORGANIZER

**Preconditions:**
- `stage != DRAFT`

```bash
UPDATE_ANNOUNCEMENT=$(curl -s -X PUT http://localhost:8080/v1/hackathons/$HACKATHON_ID/announcements/$ANNOUNCEMENT_ID \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Registration is NOW OPEN!",
    "content": "Registration is officially open! Sign up now."
  }')
```

---

## 15. DeleteHackathonAnnouncement (Удаление объявления)

**Endpoint:** `DELETE /v1/hackathons/{hackathon_id}/announcements/{announcement_id}`

**Access:** OWNER/ORGANIZER

**Preconditions:**
- `stage != DRAFT`

```bash
DELETE_ANNOUNCEMENT=$(curl -s -X DELETE http://localhost:8080/v1/hackathons/$HACKATHON_ID/announcements/$ANNOUNCEMENT_ID \
  -H "Authorization: Bearer $ALICE_TOKEN")
```

---

## Validation Codes

| Code | Description |
|------|-------------|
| `REQUIRED` | Обязательное поле не заполнено |
| `TIME_RULE` | Нарушено правило последовательности времен |
| `TIME_LOCKED` | Изменение времени запрещено (уже прошло) |
| `FORBIDDEN` | Действие запрещено на текущей стадии |
| `POLICY_RULE` | Нарушено бизнес-правило (например, at_least_one_true) |
| `FORMAT` | Неверный формат данных |

---

## Stage Flow

```
DRAFT → UPCOMING → REGISTRATION → PRESTART → RUNNING → JUDGING → FINISHED
  ↑                                                                    ↑
  |                                                                    |
published_at = null                                    result_published_at != null
```

---

## Testing

Для полного тестирования используйте:

1. **Happy Path**: `./rest-script.sh` - основные сценарии
2. **Fail Cases**: `./rest-script-fail-cases.sh` - валидации и ограничения

См. также [TESTING_SUMMARY.md](./TESTING_SUMMARY.md) для полного отчета.
