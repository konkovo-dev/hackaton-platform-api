# Auth Service - HTTP REST API Quick Start

Этот гайд показывает как работать с `auth-service` через HTTP REST API (grpc-gateway).

## Требования

- Docker и Docker Compose
- `curl` и `jq` для тестирования

## Запуск сервисов

```bash
# 1. Запускаем postgres
docker-compose -f deployments/docker-compose.yml up -d postgres

# 2. Применяем миграции (если еще не применены)
export DB_DSN="postgres://hackathon:hackathon_dev_password@localhost:5432/hackathon?sslmode=disable"
make auth-service-migrate-up

# 3. Запускаем auth-service
docker-compose -f deployments/docker-compose.yml up -d auth-service

# 4. Запускаем gateway
docker-compose -f deployments/docker-compose.yml up -d --no-deps gateway

# Проверяем статус
docker-compose -f deployments/docker-compose.yml ps
```

**Важно:** Gateway нужно запускать **после** auth-service, чтобы он успешно подключился.

## API Endpoints

Base URL: `http://localhost:8080`

### 1. Register (Регистрация пользователя)

```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "email": "alice@example.com",
    "password": "SecurePass123",
    "first_name": "Alice",
    "last_name": "Smith",
    "timezone": "Europe/Moscow",
    "idempotency_key": {"key": "alice-registration-1"}
  }' | jq .
```

**Response:**
```json
{
  "accessToken": "eyJhbGci...",
  "refreshToken": "V-WGkm...",
  "accessExpiresAt": "2026-01-12T20:53:36.896176003Z",
  "refreshExpiresAt": "2026-02-11T20:38:36.897114677Z"
}
```

### 2. Login (Вход)

С username:
```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "password": "SecurePass123"
  }' | jq .
```

С email:
```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "SecurePass123"
  }' | jq .
```

**Response:** такой же как у Register

### 3. IntrospectToken (Проверка токена)

```bash
ACCESS_TOKEN="eyJhbGci..." # токен из Register/Login

curl -X POST http://localhost:8080/v1/auth/introspect \
  -H "Content-Type: application/json" \
  -d '{
    "access_token": "'"$ACCESS_TOKEN"'"
  }' | jq .
```

**Response (валидный токен):**
```json
{
  "active": true,
  "userId": "4d0c9c23-8548-4c66-91a1-283e572a702f",
  "expiresAt": "2026-01-12T20:59:19.302365870Z"
}
```

**Response (невалидный токен):**
```json
{}
```
> **Примечание**: Пустой ответ `{}` означает `active: false` (protobuf3 не сериализует дефолтные значения).

### 4. Refresh (Обновление токена)

```bash
REFRESH_TOKEN="V-WGkm..." # refresh token из Register/Login

curl -X POST http://localhost:8080/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "'"$REFRESH_TOKEN"'"
  }' | jq .
```

**Response:** новая пара токенов (access + refresh)

### 5. Logout (Выход)

```bash
REFRESH_TOKEN="V-WGkm..." # refresh token

curl -X POST http://localhost:8080/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "'"$REFRESH_TOKEN"'"
  }' | jq .
```

**Response:**
```json
{}
```

## Полный тестовый сценарий

```bash
#!/bin/bash

# 1. Регистрация
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "SecurePass123",
    "first_name": "Test",
    "last_name": "User",
    "timezone": "UTC",
    "idempotency_key": {"key": "test-user-reg"}
  }')

echo "Register response:"
echo $REGISTER_RESPONSE | jq .

ACCESS_TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.accessToken')
REFRESH_TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.refreshToken')

# 2. Introspect
echo -e "\nIntrospecting token..."
curl -s -X POST http://localhost:8080/v1/auth/introspect \
  -H "Content-Type: application/json" \
  -d '{"access_token": "'"$ACCESS_TOKEN"'"}' | jq .

# 3. Login
echo -e "\nLogin..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "SecurePass123"
  }')
echo $LOGIN_RESPONSE | jq .

NEW_REFRESH_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.refreshToken')

# 4. Refresh
echo -e "\nRefreshing token..."
REFRESH_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "'"$NEW_REFRESH_TOKEN"'"}')
echo $REFRESH_RESPONSE | jq .

FINAL_REFRESH_TOKEN=$(echo $REFRESH_RESPONSE | jq -r '.refreshToken')

# 5. Logout
echo -e "\nLogging out..."
curl -s -X POST http://localhost:8080/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "'"$FINAL_REFRESH_TOKEN"'"}' | jq .

echo -e "\nAll tests completed!"
```

## Идемпотентность

Register поддерживает идемпотентность через `idempotency_key`. Повторный запрос с тем же ключом и теми же данными вернёт кешированный ответ:

```bash
# Первый запрос
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "bob",
    "email": "bob@example.com",
    "password": "Pass123",
    "first_name": "Bob",
    "last_name": "Test",
    "timezone": "UTC",
    "idempotency_key": {"key": "bob-unique-key"}
  }' | jq .

# Повторный запрос с тем же ключом вернёт тот же ответ
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "bob",
    "email": "bob@example.com",
    "password": "Pass123",
    "first_name": "Bob",
    "last_name": "Test",
    "timezone": "UTC",
    "idempotency_key": {"key": "bob-unique-key"}
  }' | jq .
```

## Troubleshooting

### "grpc: the client connection is closing"

**Причина:** Gateway запущен до auth-service или потерял подключение.

**Решение:**
```bash
# Перезапустить gateway
docker-compose -f deployments/docker-compose.yml restart gateway

# Или полностью пересоздать
docker-compose -f deployments/docker-compose.yml stop gateway
docker-compose -f deployments/docker-compose.yml rm -f gateway
docker-compose -f deployments/docker-compose.yml up -d --no-deps gateway
```

### Пустой ответ `{}`

Это нормально для:
- `IntrospectToken` с невалидным токеном (`active: false`)
- `Logout` (успешный ответ)

Protobuf3 не сериализует поля со значениями по умолчанию.

## Дополнительно

- gRPC endpoint: `localhost:50057` (для прямых gRPC запросов)
- HTTP gateway: `localhost:8080` (REST API)
- Все HTTP endpoints доступны под `/v1/auth/*`

