# Identity Service gRPC Test Cases

## Предусловия
- Запущен в докере с помощью [docker-setup.md](../docker-setup.md)
- Установлен `grpcurl`: `brew install grpcurl` (macOS) или [github.com/fullstorydev/grpcurl](https://github.com/fullstorydev/grpcurl)
- Пользователь зарегистрирован в auth-service
- Есть валидный access_token

## gRPC Endpoint

```
localhost:50051
```

## Получение access_token

Сначала зарегистрируйтесь через auth-service:

```bash
RESPONSE=$(grpcurl -plaintext -d '{
  "username": "alice",
  "email": "alice@example.com",
  "password": "SecurePass123",
  "first_name": "Alice",
  "last_name": "Smith",
  "timezone": "Europe/Moscow"
}' localhost:50051 auth.v1.AuthService/Register)

ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.accessToken')
echo "Access Token: $ACCESS_TOKEN"
```

---

## Тест-сценарии

### 1. GetMe (Получить свой профиль)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{}' \
  localhost:50051 identity.v1.MeService/GetMe
```

**Response:**
```json
{
  "user": {
    "userId": "4d0c9c23-8548-4c66-91a1-283e572a702f",
    "username": "alice",
    "firstName": "Alice",
    "lastName": "Smith",
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
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "first_name": "Alice",
    "last_name": "Johnson",
    "avatar_url": "https://example.com/avatar.jpg",
    "timezone": "UTC"
  }' \
  localhost:50051 identity.v1.MeService/UpdateMe
```

**Response:**
```json
{
  "user": {
    "userId": "4d0c9c23-8548-4c66-91a1-283e572a702f",
    "username": "alice",
    "firstName": "Alice",
    "lastName": "Johnson",
    "avatarUrl": "https://example.com/avatar.jpg",
    "timezone": "UTC"
  }
}
```

### 3. UpdateMySkills (Обновить навыки)

Replace-all операция:

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "catalog_skill_ids": [],
    "user_skills": ["Go", "PostgreSQL", "Docker", "Kubernetes"],
    "skills_visibility": "VISIBILITY_LEVEL_PUBLIC"
  }' \
  localhost:50051 identity.v1.MeService/UpdateMySkills
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
    },
    {
      "custom": {
        "name": "Kubernetes"
      }
    }
  ],
  "visibility": {
    "skills": "VISIBILITY_LEVEL_PUBLIC",
    "contacts": "VISIBILITY_LEVEL_PUBLIC"
  }
}
```

### 4. UpdateMyContacts (Обновить контакты)

Replace-all операция:

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
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
      },
      {
        "contact": {
          "type": "CONTACT_TYPE_GITHUB",
          "value": "github.com/alice"
        },
        "visibility": "VISIBILITY_LEVEL_PUBLIC"
      }
    ],
    "contacts_visibility": "VISIBILITY_LEVEL_PUBLIC"
  }' \
  localhost:50051 identity.v1.MeService/UpdateMyContacts
```

**Response:**
```json
{
  "contacts": [
    {
      "contact": {
        "id": "uuid-1",
        "type": "CONTACT_TYPE_EMAIL",
        "value": "alice@work.com"
      },
      "visibility": "VISIBILITY_LEVEL_PUBLIC"
    },
    {
      "contact": {
        "id": "uuid-2",
        "type": "CONTACT_TYPE_TELEGRAM",
        "value": "@alice_tg"
      },
      "visibility": "VISIBILITY_LEVEL_PRIVATE"
    },
    {
      "contact": {
        "id": "uuid-3",
        "type": "CONTACT_TYPE_GITHUB",
        "value": "github.com/alice"
      },
      "visibility": "VISIBILITY_LEVEL_PUBLIC"
    }
  ],
  "visibility": {
    "skills": "VISIBILITY_LEVEL_PUBLIC",
    "contacts": "VISIBILITY_LEVEL_PUBLIC"
  }
}
```

---

## Полный тестовый скрипт

Сохраните как `test-identity.sh`:

```bash
#!/bin/bash

set -e

echo "=== Identity Service gRPC Test ==="

# 1. Register через auth
echo -e "\n1. Registering user..."
RESPONSE=$(grpcurl -plaintext -d '{
  "username": "testuser",
  "email": "test@example.com",
  "password": "SecurePass123",
  "first_name": "Test",
  "last_name": "User",
  "timezone": "UTC"
}' localhost:50051 auth.v1.AuthService/Register)

ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.accessToken')
echo "Got access token: ${ACCESS_TOKEN:0:20}..."

# 2. GetMe
echo -e "\n2. GetMe (initial state)..."
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{}' \
  localhost:50051 identity.v1.MeService/GetMe

# 3. UpdateMe
echo -e "\n3. UpdateMe..."
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "first_name": "Updated",
    "last_name": "Name",
    "avatar_url": "https://example.com/avatar.jpg",
    "timezone": "Europe/Moscow"
  }' \
  localhost:50051 identity.v1.MeService/UpdateMe

# 4. UpdateMySkills
echo -e "\n4. UpdateMySkills..."
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "user_skills": ["React", "Node.js", "TypeScript", "PostgreSQL"],
    "skills_visibility": "VISIBILITY_LEVEL_PUBLIC"
  }' \
  localhost:50051 identity.v1.MeService/UpdateMySkills

# 5. UpdateMyContacts
echo -e "\n5. UpdateMyContacts..."
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
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
          "type": "CONTACT_TYPE_TELEGRAM",
          "value": "@testuser"
        },
        "visibility": "VISIBILITY_LEVEL_PRIVATE"
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
  }' \
  localhost:50051 identity.v1.MeService/UpdateMyContacts

# 6. GetMe (final state)
echo -e "\n6. GetMe (after all updates)..."
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{}' \
  localhost:50051 identity.v1.MeService/GetMe

echo -e "\n=== All tests completed! ==="
```

Запуск:

```bash
chmod +x test-identity.sh
./test-identity.sh
```

---

## List всех методов

```bash
# Все сервисы
grpcurl -plaintext localhost:50051 list

# Методы MeService
grpcurl -plaintext localhost:50051 list identity.v1.MeService

# Описание метода
grpcurl -plaintext localhost:50051 describe identity.v1.MeService.GetMe
```

---

## Contact Types

- `CONTACT_TYPE_EMAIL`
- `CONTACT_TYPE_TELEGRAM`
- `CONTACT_TYPE_GITHUB`
- `CONTACT_TYPE_LINKEDIN`

## Visibility Levels

- `VISIBILITY_LEVEL_PUBLIC` — публичный
- `VISIBILITY_LEVEL_PRIVATE` — приватный

---

## Replace-All Semantics

`UpdateMySkills` и `UpdateMyContacts` используют replace-all семантику:

### Пример: обновление навыков

```bash
# Добавляем навыки
grpcurl -plaintext -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{"user_skills": ["Go", "Docker"], "skills_visibility": "VISIBILITY_LEVEL_PUBLIC"}' \
  localhost:50051 identity.v1.MeService/UpdateMySkills

# Теперь добавим ещё навыки. Если не указать "Go" и "Docker" снова, они удалятся!
grpcurl -plaintext -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{"user_skills": ["Go", "Docker", "Kubernetes"], "skills_visibility": "VISIBILITY_LEVEL_PUBLIC"}' \
  localhost:50051 identity.v1.MeService/UpdateMySkills
```

---

## Troubleshooting

### Code = Unauthenticated

**Причина**: Не передан или невалидный access_token.

**Решение**:
```bash
# Получите новый токен
RESPONSE=$(grpcurl -plaintext -d '{
  "username": "testuser",
  "password": "SecurePass123"
}' localhost:50051 auth.v1.AuthService/Login)

ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.accessToken')
```

### Code = InvalidArgument

**Причина**: Неверный формат данных или обязательное поле пропущено.

**Решение**: Проверьте JSON payload и убедитесь, что все обязательные поля заполнены.

### Code = NotFound (user not found)

**Причина**: Пользователь не был создан consumer'ом после регистрации.

**Решение**:
- Проверьте логи auth-service: `docker-compose -f deployments/docker-compose.yml logs auth-service`
- Проверьте логи identity-service: `docker-compose -f deployments/docker-compose.yml logs identity-service`
- Убедитесь, что outbox processor работает

---

## Дополнительно

- REST API доступно через gateway: `http://localhost:8080/v1/users/me*`
- Для отладки используйте `-v` флаг: `grpcurl -v -plaintext ...`
- Для pretty-print добавьте `| jq .`

