# UsersService gRPC Test Cases

## Предусловия
- Запущен в докере с помощью [docker-setup.md](../docker-setup.md)
- Установлен `grpcurl`: `brew install grpcurl` (macOS) или [github.com/fullstorydev/grpcurl](https://github.com/fullstorydev/grpcurl)
- **Тестовые данные добавлены в БД** (см. раздел ниже)

## gRPC Endpoint

```
localhost:50051
```

## Получение access_token

> **Важно**: UsersService **требует авторизации**. Все методы доступны только залогиненным пользователям.

Зарегистрируйтесь через auth-service:

```bash
RESPONSE=$(grpcurl -plaintext -d '{
  "username": "testuser2",
  "email": "testuser2@example.com",
  "password": "SecurePass123",
  "first_name": "Test",
  "last_name": "User",
  "timezone": "UTC"
}' localhost:50057 auth.v1.AuthService/Register)
echo "Response $RESPONSE"

ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.accessToken')
echo "Access Token: $ACCESS_TOKEN"
```

> Подробнее о auth-service: [../auth/grpc-guide.md](../auth/grpc-guide.md)

---

## Подготовка тестовых данных

UsersService — публичный API, не требует авторизации. Для тестирования нужно создать несколько пользователей с разными настройками видимости.

### Шаг 1: Подключитесь к БД

```bash
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon
```

### Шаг 2: Добавьте тестовых пользователей

```sql
-- Пользователь 1: bob (все PUBLIC)
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('b0b00000-0000-0000-0000-000000000001', 'bob_public', 'Bob', 'Public', 'https://example.com/bob.jpg', 'UTC', NOW(), NOW());

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('b0b00000-0000-0000-0000-000000000001', 'public', 'public');

INSERT INTO identity.user_custom_skills (id, user_id, name)
VALUES 
  (gen_random_uuid(), 'b0b00000-0000-0000-0000-000000000001', 'Python'),
  (gen_random_uuid(), 'b0b00000-0000-0000-0000-000000000001', 'Django'),
  (gen_random_uuid(), 'b0b00000-0000-0000-0000-000000000001', 'PostgreSQL');

INSERT INTO identity.user_contacts (id, user_id, type, value, visibility)
VALUES 
  (gen_random_uuid(), 'b0b00000-0000-0000-0000-000000000001', 'email', 'bob@example.com', 'public'),
  (gen_random_uuid(), 'b0b00000-0000-0000-0000-000000000001', 'github', 'github.com/bob', 'public'),
  (gen_random_uuid(), 'b0b00000-0000-0000-0000-000000000001', 'telegram', '@bob_tg', 'private');

-- Пользователь 2: charlie (skills PRIVATE, contacts PUBLIC но только PUBLIC контакты)
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('c4a41e00-0000-0000-0000-000000000002', 'charlie_mixed', 'Charlie', 'Mixed', '', 'Europe/London', NOW(), NOW());

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('c4a41e00-0000-0000-0000-000000000002', 'private', 'public');

INSERT INTO identity.user_custom_skills (id, user_id, name)
VALUES 
  (gen_random_uuid(), 'c4a41e00-0000-0000-0000-000000000002', 'Java'),
  (gen_random_uuid(), 'c4a41e00-0000-0000-0000-000000000002', 'Spring');

INSERT INTO identity.user_contacts (id, user_id, type, value, visibility)
VALUES 
  (gen_random_uuid(), 'c4a41e00-0000-0000-0000-000000000002', 'email', 'charlie@work.com', 'public'),
  (gen_random_uuid(), 'c4a41e00-0000-0000-0000-000000000002', 'linkedin', 'linkedin.com/in/charlie', 'private');

-- Пользователь 3: diana (все PRIVATE)
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('d1a4a000-0000-0000-0000-000000000003', 'diana_private', 'Diana', 'Private', '', 'America/New_York', NOW(), NOW());

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('d1a4a000-0000-0000-0000-000000000003', 'private', 'private');

INSERT INTO identity.user_custom_skills (id, user_id, name)
VALUES 
  (gen_random_uuid(), 'd1a4a000-0000-0000-0000-000000000003', 'Rust'),
  (gen_random_uuid(), 'd1a4a000-0000-0000-0000-000000000003', 'WebAssembly');

INSERT INTO identity.user_contacts (id, user_id, type, value, visibility)
VALUES 
  (gen_random_uuid(), 'd1a4a000-0000-0000-0000-000000000003', 'email', 'diana@private.com', 'private');

-- Пользователь 4: eve (для поиска по имени/навыкам)
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('e5e00000-0000-0000-0000-000000000004', 'eve_golang', 'Eve', 'Gopher', 'https://example.com/eve.jpg', 'UTC', NOW(), NOW());

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('e5e00000-0000-0000-0000-000000000004', 'public', 'public');

INSERT INTO identity.user_custom_skills (id, user_id, name)
VALUES 
  (gen_random_uuid(), 'e5e00000-0000-0000-0000-000000000004', 'Go'),
  (gen_random_uuid(), 'e5e00000-0000-0000-0000-000000000004', 'Kubernetes'),
  (gen_random_uuid(), 'e5e00000-0000-0000-0000-000000000004', 'Docker');

INSERT INTO identity.user_contacts (id, user_id, type, value, visibility)
VALUES 
  (gen_random_uuid(), 'e5e00000-0000-0000-0000-000000000004', 'email', 'eve@example.com', 'public'),
  (gen_random_uuid(), 'e5e00000-0000-0000-0000-000000000004', 'github', 'github.com/eve', 'public');
```

### Шаг 3: Проверьте данные

```sql
SELECT id, username, first_name, last_name FROM identity.users 
WHERE username LIKE 'bob_public' OR username LIKE 'charlie_mixed' OR username LIKE 'diana_private' OR username LIKE 'eve_golang';
```

Должно вернуть 4 пользователей.

---

## Тест-сценарии

> **Важно**: Все запросы **требуют авторизации**. Используйте `ACCESS_TOKEN` из раздела выше.

### 1. GetUser (базовый запрос без дополнительных данных)

```bash
# Bob (все PUBLIC)
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "user_id": "b0b00000-0000-0000-0000-000000000001"
  }' \
  localhost:50051 identity.v1.UsersService/GetUser
```

**Response:**
```json
{
  "user": {
    "userId": "b0b00000-0000-0000-0000-000000000001",
    "username": "bob_public",
    "firstName": "Bob",
    "lastName": "Public",
    "avatarUrl": "https://example.com/bob.jpg",
    "timezone": "UTC"
  }
}
```

### 2. GetUser с include_skills и include_contacts (все PUBLIC)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "user_id": "b0b00000-0000-0000-0000-000000000001",
    "include_skills": true,
    "include_contacts": true
  }' \
  localhost:50051 identity.v1.UsersService/GetUser
```

**Response:**
```json
{
  "user": {
    "userId": "b0b00000-0000-0000-0000-000000000001",
    "username": "bob_public",
    "firstName": "Bob",
    "lastName": "Public",
    "avatarUrl": "https://example.com/bob.jpg",
    "timezone": "UTC"
  },
  "skills": [
    {
      "custom": {
        "name": "Python"
      }
    },
    {
      "custom": {
        "name": "Django"
      }
    },
    {
      "custom": {
        "name": "PostgreSQL"
      }
    }
  ],
  "contacts": [
    {
      "type": "CONTACT_TYPE_EMAIL",
      "value": "bob@example.com"
    },
    {
      "type": "CONTACT_TYPE_GITHUB",
      "value": "github.com/bob"
    }
  ]
}
```

> **Обратите внимание**: Контакт `@bob_tg` с `visibility=PRIVATE` не возвращается!

### 3. GetUser с PRIVATE skills (Charlie)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "user_id": "c4a41e00-0000-0000-0000-000000000002",
    "include_skills": true,
    "include_contacts": true
  }' \
  localhost:50051 identity.v1.UsersService/GetUser
```

**Response:**
```json
{
  "user": {
    "userId": "c4a41e00-0000-0000-0000-000000000002",
    "username": "charlie_mixed",
    "firstName": "Charlie",
    "lastName": "Mixed",
    "timezone": "Europe/London"
  },
  "skills": [],
  "contacts": [
    {
      "type": "CONTACT_TYPE_EMAIL",
      "value": "charlie@work.com"
    }
  ]
}
```

> **Обратите внимание**: 
> - `skills = []` потому что `skills_visibility = PRIVATE`
> - LinkedIn с `visibility=PRIVATE` не возвращается

### 4. GetUser с PRIVATE skills и contacts (Diana)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "user_id": "d1a4a000-0000-0000-0000-000000000003",
    "include_skills": true,
    "include_contacts": true
  }' \
  localhost:50051 identity.v1.UsersService/GetUser
```

**Response:**
```json
{
  "user": {
    "userId": "d1a4a000-0000-0000-0000-000000000003",
    "username": "diana_private",
    "firstName": "Diana",
    "lastName": "Private",
    "timezone": "America/New_York"
  },
  "skills": [],
  "contacts": []
}
```

> **Visibility правила работают**: оба массива пустые.

### 5. GetUser (несуществующий пользователь)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "user_id": "00000000-0000-0000-0000-000000000001"
  }' \
  localhost:50051 identity.v1.UsersService/GetUser
```

**Response:**
```
ERROR:
  Code: NotFound
  Message: user not found
```

---

### 6. BatchGetUsers

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "user_ids": [
      "b0b00000-0000-0000-0000-000000000001",
      "c4a41e00-0000-0000-0000-000000000002",
      "d1a4a000-0000-0000-0000-000000000003"
    ],
    "include_skills": true,
    "include_contacts": false
  }' \
  localhost:50051 identity.v1.UsersService/BatchGetUsers
```

**Response:**
```json
{
  "users": [
    {
      "user": {
        "userId": "b0b00000-0000-0000-0000-000000000001",
        "username": "bob_public",
        "firstName": "Bob",
        "lastName": "Public"
      },
      "skills": [
        {
          "custom": {
            "name": "Python"
          }
        },
        {
          "custom": {
            "name": "Django"
          }
        },
        {
          "custom": {
            "name": "PostgreSQL"
          }
        }
      ]
    },
    {
      "user": {
        "userId": "c4a41e00-0000-0000-0000-000000000002",
        "username": "charlie_mixed",
        "firstName": "Charlie",
        "lastName": "Mixed"
      },
      "skills": []
    },
    {
      "user": {
        "userId": "d1a4a000-0000-0000-0000-000000000003",
        "username": "diana_private",
        "firstName": "Diana",
        "lastName": "Private"
      },
      "skills": []
    }
  ]
}
```

> **Visibility применяется индивидуально** для каждого пользователя.

---

### 7. ListUsers (простой поиск по query.q)

```bash
# Поиск по имени "Bob" или "bob"
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "query": {
      "q": "bob",
      "page": {
        "page_size": 10
      }
    },
    "include_skills": false
  }' \
  localhost:50051 identity.v1.UsersService/ListUsers
```

**Response:**
```json
{
  "users": [
    {
      "user": {
        "userId": "b0b00000-0000-0000-0000-000000000001",
        "username": "bob_public",
        "firstName": "Bob",
        "lastName": "Public"
      }
    }
  ],
  "page": {}
}
```

### 8. ListUsers (поиск по username через query.q)

```bash
# Поиск по username "_mi" - query.q ищет также в username
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "query": {
      "q": "_mi",
      "page": {
        "page_size": 10
      }
    },
    "include_skills": false
  }' \
  localhost:50051 identity.v1.UsersService/ListUsers
```

**Response:**
```json
{
  "users": [
    {
      "user": {
        "userId": "c4a41e00-0000-0000-0000-000000000002",
        "username": "charlie_mixed",
        "firstName": "Charlie",
        "lastName": "Mixed"
      }
    }
  ],
  "page": {}
}
```

> **Примечание:** `query.q` выполняет поиск (case-insensitive, ILIKE '%value%') по полям: `username`, `first_name`, `last_name`. В данном случае найдено совпадение в username `charlie_mixed`.

### 9. ListUsers с фильтрацией по username (PREFIX)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "query": {
      "filter_groups": [
        {
          "filters": [
            {
              "field": "username",
              "operation": "FILTER_OPERATION_PREFIX",
              "string_value": "bob"
            }
          ]
        }
      ],
      "sort": [
        {
          "field": "username",
          "direction": "SORT_DIRECTION_ASC"
        }
      ],
      "page": {
        "page_size": 20
      }
    },
    "include_skills": true
  }' \
  localhost:50051 identity.v1.UsersService/ListUsers
```

### 10. ListUsers с фильтрацией по навыкам

```bash
# Найти всех пользователей с навыком "Go"
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "query": {
      "filter_groups": [
        {
          "filters": [
            {
              "field": "skills",
              "operation": "FILTER_OPERATION_CONTAINS",
              "string_value": "go"
            }
          ]
        }
      ],
      "page": {
        "page_size": 20
      }
    },
    "include_skills": true
  }' \
  localhost:50051 identity.v1.UsersService/ListUsers
```

**Response:**
```json
{
  "users": [
    {
      "user": {
        "userId": "e5e00000-0000-0000-0000-000000000004",
        "username": "eve_golang",
        "firstName": "Eve",
        "lastName": "Gopher"
      },
      "skills": [
        {
          "custom": {
            "name": "Go"
          }
        },
        {
          "custom": {
            "name": "Kubernetes"
          }
        },
        {
          "custom": {
            "name": "Docker"
          }
        }
      ]
    }
  ],
  "page": {}
}
```

### 11. ListUsers с фильтрацией по user_id (IN)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "query": {
      "filter_groups": [
        {
          "filters": [
            {
              "field": "user_id",
              "operation": "FILTER_OPERATION_IN",
              "string_list": {
                "values": [
                  "b0b00000-0000-0000-0000-000000000001",
                  "e5e00000-0000-0000-0000-000000000004"
                ]
              }
            }
          ]
        }
      ],
      "page": {
        "page_size": 50
      }
    },
    "include_skills": true,
    "include_contacts": true
  }' \
  localhost:50051 identity.v1.UsersService/ListUsers
```

---

## Полный тестовый скрипт

Сохраните как `test-users-service.sh`:

```bash
#!/bin/bash

set -e

echo "=== UsersService gRPC Test ==="

# 0. Регистрация (для возможности тестировать MeService параллельно)
echo -e "\n0. Registering test user..."
RESPONSE=$(grpcurl -plaintext -d '{
  "username": "testuser_combined",
  "email": "testcombined@example.com",
  "password": "SecurePass123",
  "first_name": "Test",
  "last_name": "Combined",
  "timezone": "UTC"
}' localhost:50051 auth.v1.AuthService/Register)

ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.accessToken')
echo "Got access token: ${ACCESS_TOKEN:0:20}..."

# 1. GetUser (Bob - все PUBLIC)
echo -e "\n1. GetUser (Bob - все PUBLIC)..."
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "user_id": "b0b00000-0000-0000-0000-000000000001",
    "include_skills": true,
    "include_contacts": true
  }' \
  localhost:50051 identity.v1.UsersService/GetUser

# 2. GetUser (Charlie - skills PRIVATE)
echo -e "\n2. GetUser (Charlie - skills PRIVATE)..."
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "user_id": "c4a41e00-0000-0000-0000-000000000002",
    "include_skills": true,
    "include_contacts": true
  }' \
  localhost:50051 identity.v1.UsersService/GetUser

# 3. GetUser (Diana - все PRIVATE)
echo -e "\n3. GetUser (Diana - все PRIVATE)..."
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "user_id": "d1a4a000-0000-0000-0000-000000000003",
    "include_skills": true,
    "include_contacts": true
  }' \
  localhost:50051 identity.v1.UsersService/GetUser

# 4. BatchGetUsers
echo -e "\n4. BatchGetUsers (3 пользователя)..."
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "user_ids": [
      "b0b00000-0000-0000-0000-000000000001",
      "c4a41e00-0000-0000-0000-000000000002",
      "e5e00000-0000-0000-0000-000000000004"
    ],
    "include_skills": true
  }' \
  localhost:50051 identity.v1.UsersService/BatchGetUsers

# 5. ListUsers (поиск по "bob")
echo -e "\n5. ListUsers (поиск по 'bob')..."
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "query": {
      "q": "bob",
      "page": {
        "page_size": 10
      }
    },
    "include_skills": true
  }' \
  localhost:50051 identity.v1.UsersService/ListUsers

# 6. ListUsers (фильтр по навыку "Go")
echo -e "\n6. ListUsers (фильтр по навыку 'Go')..."
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "query": {
      "filter_groups": [
        {
          "filters": [
            {
              "field": "skills",
              "operation": "FILTER_OPERATION_CONTAINS",
              "string_value": "go"
            }
          ]
        }
      ],
      "page": {
        "page_size": 20
      }
    },
    "include_skills": true
  }' \
  localhost:50051 identity.v1.UsersService/ListUsers

echo -e "\n=== All UsersService tests completed! ==="
echo -e "\nТеперь можете протестировать MeService с тем же токеном:"
echo "grpcurl -plaintext -H \"authorization: Bearer \$ACCESS_TOKEN\" -d '{}' localhost:50051 identity.v1.MeService/GetMe"
```

Запуск:

```bash
chmod +x test-users-service.sh
./test-users-service.sh
```

---

## Фильтры и сортировка

### Поддерживаемые поля для фильтрации

- `user_id` (IN)
- `username` (EQ, PREFIX, CONTAINS)
- `first_name` (PREFIX, CONTAINS)
- `last_name` (PREFIX, CONTAINS)
- `skills` (CONTAINS — **точное совпадение** имени навыка, case-insensitive; для частичного поиска используйте PREFIX)

### Поддерживаемые поля для сортировки

- `username` (ASC/DESC, по умолчанию ASC)

### Пагинация

- Cursor-based по `(username, user_id)`
- `page_size` по умолчанию: 50, максимум: 100
- `page_token` возвращается в `page.nextPageToken` для следующей страницы

---

## List всех методов

```bash
# Методы UsersService
grpcurl -plaintext localhost:50051 list identity.v1.UsersService

# Описание методов
grpcurl -plaintext localhost:50051 describe identity.v1.UsersService.GetUser
grpcurl -plaintext localhost:50051 describe identity.v1.UsersService.BatchGetUsers
grpcurl -plaintext localhost:50051 describe identity.v1.UsersService.ListUsers
```

---

## Visibility Rules

### Глобальная видимость (из `user_visibility`)

- `skills_visibility = PRIVATE` → `skills = []` в ответе
- `contacts_visibility = PRIVATE` → `contacts = []` в ответе

### Per-Contact видимость (из `user_contacts.visibility`)

Если `contacts_visibility = PUBLIC`, то возвращаются только контакты с `visibility = PUBLIC`.

### Важно

- В `UsersService` поле `visibility` для каждого контакта **НЕ** возвращается (в отличие от `MyService`)
- UsersService **требует авторизации** — только залогиненные пользователи могут просматривать профили других участников
- Показывается только то, что пользователь хочет сделать публичным (согласно его настройкам visibility)

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

### Code = NotFound

**Причина**: Пользователь с указанным `user_id` не существует.

**Решение**: Проверьте корректность UUID или добавьте тестовые данные (см. раздел "Подготовка тестовых данных").

### Code = InvalidArgument

**Причина**: 
- `user_id` не валидный UUID
- `user_ids` в BatchGetUsers > 100 элементов
- `page_size` > 100
- Неподдерживаемая операция фильтра

**Решение**: Проверьте корректность параметров.

### Code = Internal

**Причина**: Проблема с БД или внутренняя ошибка сервиса.

**Решение**:
- Проверьте логи identity-service: `docker-compose -f deployments/docker-compose.yml logs identity-service`
- Убедитесь, что БД доступна

---

## Дополнительно

- UsersService **требует авторизации** — используйте `-H "authorization: Bearer $ACCESS_TOKEN"`
- REST API доступно через gateway: `http://localhost:8080/v1/users/*`
- Для отладки используйте `-v` флаг: `grpcurl -v -plaintext ...`
- Для pretty-print добавьте `| jq .`
- MeService guide: [me-service-grpc-guide.md](./me-service-grpc-guide.md)

