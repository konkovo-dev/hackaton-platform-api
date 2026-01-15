# MeService REST Test Cases

## Предусловия
- Запущен в докере с помощью [docker-setup.md](../docker-setup.md)
- Пользователь зарегистрирован в auth-service
- Есть валидный access_token

## Эндпоинты

Base URL: `http://localhost:8080`

### Получение access_token

Зарегистрируйтесь через auth-service:

```bash
# Register
RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice_me",
    "email": "alice_me@example.com",
    "password": "SecurePass123",
    "first_name": "Alice",
    "last_name": "Smith",
    "timezone": "Europe/Moscow"
  }')

ACCESS_TOKEN=$(echo $RESPONSE | jq -r '.accessToken')
echo "Access Token: $ACCESS_TOKEN"
```

> Подробнее о auth-service: [../auth/rest-guide.md](../auth/rest-guide.md)

---

## Тест-сценарии

### 1. GetMe (Получить свой профиль)

```bash
curl http://localhost:8080/v1/users/me \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response:**
```json
{
  "user": {
    "userId": "generated-uuid",
    "username": "alice",
    "firstName": "Alice",
    "lastName": "Smith",
    "avatarUrl": "",
    "timezone": "Europe/Moscow"
  },
  "skills": [],
  "contacts": [],
  "visibility": {
    "skills": "VISIBILITY_LEVEL_PUBLIC",
    "contacts": "VISIBILITY_LEVEL_PUBLIC"
  }
}
```

### 2. UpdateMe (Обновить профиль)

```bash
curl -X PUT http://localhost:8080/v1/users/me \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Alice",
    "last_name": "Johnson",
    "avatar_url": "https://example.com/avatar.jpg",
    "timezone": "UTC"
  }' | jq .
```

**Response:**
```json
{
  "user": {
    "userId": "generated-uuid",
    "username": "alice_me",
    "firstName": "Alice",
    "lastName": "Johnson",
    "avatarUrl": "https://example.com/avatar.jpg",
    "timezone": "UTC"
  }
}
```

> **Примечание**: username изменить нельзя (удалён из proto).

### 3. UpdateMySkills (Обновить навыки)

Replace-all операция: все старые навыки удаляются, новые добавляются.

```bash
curl -X PUT http://localhost:8080/v1/users/me/skills \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "catalog_skill_ids": [],
    "user_skills": ["Go", "PostgreSQL", "Docker"],
    "skills_visibility": "VISIBILITY_LEVEL_PUBLIC"
  }' | jq .
```

**Response:**
```json
{
  "skills": [
    {
      "custom": {
        "name": "Go"
      }
    },
    {
      "custom": {
        "name": "PostgreSQL"
      }
    },
    {
      "custom": {
        "name": "Docker"
      }
    }
  ],
  "visibility": {
    "skills": "VISIBILITY_LEVEL_PUBLIC",
    "contacts": "VISIBILITY_LEVEL_PUBLIC"
  }
}
```

> **Важно**: catalog_skill_ids должны существовать в БД (таблица `identity.skill_catalog`).

### 4. UpdateMyContacts (Обновить контакты)

Replace-all операция: все старые контакты удаляются, новые добавляются.

```bash
curl -X PUT http://localhost:8080/v1/users/me/contacts \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "contacts": [
      {
        "contact": {
          "type": "CONTACT_TYPE_EMAIL",
          "value": "alice@work.com"
        },
        "visibility": "VISIBILITY_LEVEL_PUBLIC"
      },
      {
        "contact": {
          "type": "CONTACT_TYPE_TELEGRAM",
          "value": "@alice_tg"
        },
        "visibility": "VISIBILITY_LEVEL_PRIVATE"
      }
    ],
    "contacts_visibility": "VISIBILITY_LEVEL_PUBLIC"
  }' | jq .
```

**Response:**
```json
{
  "contacts": [
    {
      "contact": {
        "id": "new-uuid-1",
        "type": "CONTACT_TYPE_EMAIL",
        "value": "alice@work.com"
      },
      "visibility": "VISIBILITY_LEVEL_PUBLIC"
    },
    {
      "contact": {
        "id": "new-uuid-2",
        "type": "CONTACT_TYPE_TELEGRAM",
        "value": "@alice_tg"
      },
      "visibility": "VISIBILITY_LEVEL_PRIVATE"
    }
  ],
  "visibility": {
    "skills": "VISIBILITY_LEVEL_PUBLIC",
    "contacts": "VISIBILITY_LEVEL_PUBLIC"
  }
}
```

---

## Полный тестовый сценарий

```bash
#!/bin/bash

# 1. Регистрация
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser_me",
    "email": "testme@example.com",
    "password": "SecurePass123",
    "first_name": "Test",
    "last_name": "User",
    "timezone": "UTC"
  }')

echo "Register response:"
echo $REGISTER_RESPONSE | jq .

ACCESS_TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.accessToken')

# 2. GetMe
echo -e "\n=== GetMe ==="
curl -s http://localhost:8080/v1/users/me \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .

# 3. UpdateMe
echo -e "\n=== UpdateMe ==="
curl -s -X PUT http://localhost:8080/v1/users/me \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Updated",
    "last_name": "Name",
    "avatar_url": "https://example.com/avatar.jpg",
    "timezone": "Europe/Moscow"
  }' | jq .

# 4. UpdateMySkills
echo -e "\n=== UpdateMySkills ==="
curl -s -X PUT http://localhost:8080/v1/users/me/skills \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "catalog_skill_ids": [],
    "user_skills": ["React", "Node.js", "Docker"],
    "skills_visibility": "VISIBILITY_LEVEL_PUBLIC"
  }' | jq .

# 5. UpdateMyContacts
echo -e "\n=== UpdateMyContacts ==="
curl -s -X PUT http://localhost:8080/v1/users/me/contacts \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "contacts": [
      {
        "contact": {
          "type": "CONTACT_TYPE_EMAIL",
          "value": "test@work.com"
        },
        "visibility": "VISIBILITY_LEVEL_PUBLIC"
      },
      {
        "contact": {
          "type": "CONTACT_TYPE_GITHUB",
          "value": "github.com/testuser"
        },
        "visibility": "VISIBILITY_LEVEL_PUBLIC"
      }
    ],
    "contacts_visibility": "VISIBILITY_LEVEL_PUBLIC"
  }' | jq .

# 6. GetMe снова (проверка изменений)
echo -e "\n=== GetMe (after updates) ==="
curl -s http://localhost:8080/v1/users/me \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .

echo -e "\nAll tests completed!"
```

---

## Contact Types

Доступные типы контактов:
- `CONTACT_TYPE_EMAIL` — Email
- `CONTACT_TYPE_TELEGRAM` — Telegram
- `CONTACT_TYPE_GITHUB` — GitHub
- `CONTACT_TYPE_LINKEDIN` — LinkedIn

## Visibility Levels

Доступные уровни видимости:
- `VISIBILITY_LEVEL_PUBLIC` — Публичный (виден всем)
- `VISIBILITY_LEVEL_PRIVATE` — Приватный (виден только мне)

### Глобальная vs Per-Item видимость

- **Глобальная**: `skills_visibility`, `contacts_visibility` в `UpdateMySkills` и `UpdateMyContacts`
- **Per-Item**: `visibility` для каждого контакта отдельно

> В MyService все навыки и контакты всегда возвращаются полностью (это "я"). Фильтрация по видимости будет в UsersService для публичных профилей.

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
    "username": "testuser_me",
    "password": "SecurePass123"
  }' | jq -r '.accessToken'
```

### 400 Invalid catalog_skill_id

**Причина**: UUID навыка не существует в `identity.skill_catalog`.

**Решение**: Используйте только существующие UUID или пустой массив.

### 404 Not Found

**Причина**: Пользователь не был создан consumer'ом после регистрации.

**Решение**:
- Проверьте логи identity-service: `docker-compose -f deployments/docker-compose.yml logs identity-service`
- Убедитесь, что outbox processor работает

### Replace-all удаляет данные

Это ожидаемое поведение для `UpdateMySkills` и `UpdateMyContacts`. Чтобы сохранить данные, передайте полный список (старые + новые).

---

## Дополнительно

- gRPC endpoint: `localhost:50051` (для прямых gRPC запросов)
- HTTP gateway: `localhost:8080` (REST API)
- Все HTTP endpoints доступны под `/v1/users/me*`
- gRPC guide: [me-service-grpc-guide.md](./me-service-grpc-guide.md)

