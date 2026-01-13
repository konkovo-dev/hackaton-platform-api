# Auth + Identity Integration - gRPC Test Cases

## Предусловия

- Сервисы запущены согласно `docker-setup.md`
- Postgres с примененными миграциями для обоих сервисов
- `grpcurl` установлен

## Тест-сценарии

### 1. Базовый End-to-End флоу

#### Шаг 1: Регистрация пользователя

```bash
RESPONSE=$(grpcurl -plaintext -d '{
  "username": "alice",
  "email": "alice@example.com",
  "password": "SecurePass123",
  "first_name": "Alice",
  "last_name": "Smith",
  "timezone": "Europe/Moscow",
  "idempotency_key": {"key": "alice-registration-1"}
}' localhost:50057 auth.v1.AuthService/Register)

echo "$RESPONSE"
```

**Ожидаемый ответ:**
```json
{
  "accessToken": "eyJhbGci...",
  "refreshToken": "V-WGkm...",
  "accessExpiresAt": "2026-01-12T20:53:36Z",
  "refreshExpiresAt": "2026-02-11T20:38:36Z"
}
```

#### Шаг 2: Проверка создания пользователя в auth.users

```bash
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon -c \
  "SELECT id, username, email FROM auth.users WHERE username = 'alice';"
```

**Ожидаемый результат:**
```
                  id                  | username |       email       
--------------------------------------+----------+-------------------
 <uuid>                               | alice    | alice@example.com
```

#### Шаг 3: Проверка создания outbox события

```bash
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon -c \
  "SELECT event_type, status, payload::json->>'username' as username 
   FROM auth.outbox_events 
   WHERE payload::json->>'username' = 'alice';"
```

**Сразу после Register (< 1 сек):**
```
   event_type    | status  | username 
-----------------+---------+----------
 user.registered | pending | alice
```

**Через 1-2 секунды (outbox processor обработал):**
```
   event_type    |  status   | username 
-----------------+-----------+----------
 user.registered | processed | alice
```

#### Шаг 4: Проверка создания профиля в identity.users

```bash
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon -c \
  "SELECT id, username, first_name, last_name, timezone FROM identity.users WHERE username = 'alice';"
```

**Ожидаемый результат:**
```
                  id                  | username | first_name | last_name |    timezone    
--------------------------------------+----------+------------+-----------+----------------
 <uuid>                               | alice    | Alice      | Smith     | Europe/Moscow
```

#### Шаг 5: Проверка связи auth ↔ identity

```bash
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon -c \
  "SELECT a.id, a.username, a.email, i.first_name, i.last_name, i.timezone 
   FROM auth.users a 
   JOIN identity.users i ON a.id = i.id 
   WHERE a.username = 'alice';"
```

**Ожидаемый результат:**
```
                  id                  | username |       email       | first_name | last_name |    timezone    
--------------------------------------+----------+-------------------+------------+-----------+----------------
 <uuid>                               | alice    | alice@example.com | Alice      | Smith     | Europe/Moscow
```

### 2. Тест идемпотентности Register

#### Повторная регистрация с тем же idempotency_key

```bash
# Первый запрос
RESPONSE1=$(grpcurl -plaintext -d '{
  "username": "bob",
  "email": "bob@example.com",
  "password": "password123",
  "first_name": "Bob",
  "last_name": "Jones",
  "timezone": "UTC",
  "idempotency_key": {"key": "bob-reg-unique"}
}' localhost:50057 auth.v1.AuthService/Register)

echo "First request:"
echo "$RESPONSE1" | jq .

# Повторный запрос с тем же ключом (вернёт кешированный ответ)
RESPONSE2=$(grpcurl -plaintext -d '{
  "username": "bob",
  "email": "bob@example.com",
  "password": "password123",
  "first_name": "Bob",
  "last_name": "Jones",
  "timezone": "UTC",
  "idempotency_key": {"key": "bob-reg-unique"}
}' localhost:50057 auth.v1.AuthService/Register)

echo "Second request (should return same tokens):"
echo "$RESPONSE2" | jq .
```

**Проверка:**
- Оба запроса должны вернуть одинаковые токены
- В БД должна быть только одна запись для bob
- В outbox_events должно быть только одно событие

```bash
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon -c \
  "SELECT COUNT(*) FROM auth.users WHERE username = 'bob';"
# Ожидается: 1

docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon -c \
  "SELECT COUNT(*) FROM auth.outbox_events WHERE payload::json->>'username' = 'bob';"
# Ожидается: 1

docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon -c \
  "SELECT COUNT(*) FROM identity.users WHERE username = 'bob';"
# Ожидается: 1
```

### 3. Тест обработки ошибок

#### Сценарий: Identity Service недоступен

```bash
# Остановить identity-service
docker-compose -f deployments/docker-compose.yml stop identity-service

# Попробовать зарегистрироваться
grpcurl -plaintext -d '{
  "username": "charlie",
  "email": "charlie@example.com",
  "password": "pass1235678",
  "first_name": "Charlie",
  "last_name": "Brown",
  "timezone": "UTC",
  "idempotency_key": {"key": "charlie-reg"}
}' localhost:50057 auth.v1.AuthService/Register
```

**Что произойдет:**
1. Пользователь создастся в `auth.users`
2. Токены вернутся клиенту ✓
3. Событие запишется в `auth.outbox_events` со статусом `pending` ✓
4. Outbox processor попытается обработать событие
5. После 3 неудачных попыток (каждая ~3-4 сек) событие перейдет в статус `failed`

**Проверка статуса:**
```bash
# Подождать ~10-12 секунд (3 попытки * ~3-4 сек process_timeout + polling)
sleep 12

docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon -c \
  "SELECT event_type, status, attempt_count, substring(last_error, 1, 100) as last_error 
   FROM auth.outbox_events 
   WHERE payload::json->>'username' = 'charlie';"
```

**Ожидаемый результат:**
```
   event_type    | status | attempt_count |                         last_error
-----------------+--------+---------------+------------------------------------------------------------
 user.registered | failed |             3 | failed to create user in identity service: failed to crea...
(1 row)
```

**Восстановление:**
```bash
# 1. Запустить identity-service обратно
docker-compose -f deployments/docker-compose.yml start identity-service

# 2. Повторно обработать failed события
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon << 'EOF'
UPDATE auth.outbox_events 
SET status = 'pending', attempt_count = 0, last_error = '', updated_at = NOW()
WHERE status = 'failed' AND payload::json->>'username' = 'charlie';

SELECT event_type, status, attempt_count FROM auth.outbox_events WHERE payload::json->>'username' = 'charlie';
EOF

# 3. Подождать обработки (~2-3 секунды)
sleep 3

# 4. Проверить финальный результат
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon << 'EOF'
-- Проверить что событие обработано
SELECT event_type, status, attempt_count FROM auth.outbox_events WHERE payload::json->>'username' = 'charlie';

-- Проверить что профиль создан
SELECT id, username, first_name, last_name FROM identity.users WHERE username = 'charlie';
EOF
```

**Ожидаемый результат:**
```
   event_type    |  status   | attempt_count 
-----------------+-----------+---------------
 user.registered | processed |             1

                  id                  | username | first_name | last_name 
--------------------------------------+----------+------------+-----------
 3e38a3c9-2e2c-423f-8e88-dc51ada2fdea | charlie  | Charlie    | Brown
```

### 4. Негативные сценарии

#### Дублирующийся username

```bash
# Первая регистрация
grpcurl -plaintext -d '{
  "username": "duplicate",
  "email": "dup1@example.com",
  "password": "pass123",
  "first_name": "First",
  "last_name": "User",
  "timezone": "UTC"
}' localhost:50057 auth.v1.AuthService/Register

# Попытка регистрации с тем же username
grpcurl -plaintext -d '{
  "username": "duplicate",
  "email": "dup2@example.com",
  "password": "pass456",
  "first_name": "Second",
  "last_name": "User",
  "timezone": "UTC"
}' localhost:50057 auth.v1.AuthService/Register
```

**Ожидаемая ошибка:**
```
ERROR:
  Code: AlreadyExists
  Message: username already exists
```

#### Невалидные данные

```bash
# Пустой username
grpcurl -plaintext -d '{
  "username": "",
  "email": "test@test.com",
  "password": "pass",
  "first_name": "Test",
  "last_name": "User",
  "timezone": "UTC"
}' localhost:50057 auth.v1.AuthService/Register
# Ожидается: InvalidArgument "username is required"

# Пустой email
grpcurl -plaintext -d '{
  "username": "testuser",
  "email": "",
  "password": "pass",
  "first_name": "Test",
  "last_name": "User",
  "timezone": "UTC"
}' localhost:50057 auth.v1.AuthService/Register
# Ожидается: InvalidArgument "email is required"

# Пустой first_name
grpcurl -plaintext -d '{
  "username": "testuser",
  "email": "test@test.com",
  "password": "pass",
  "first_name": "",
  "last_name": "User",
  "timezone": "UTC"
}' localhost:50057 auth.v1.AuthService/Register
# Ожидается: InvalidArgument "first_name is required"
```

### 5. Проверка IntrospectToken с новым пользователем

```bash
# Регистрация
RESPONSE=$(grpcurl -plaintext -d '{
  "username": "testauth",
  "email": "testauth@example.com",
  "password": "SecurePass123",
  "first_name": "Test",
  "last_name": "Auth",
  "timezone": "UTC"
}' localhost:50057 auth.v1.AuthService/Register)

# Извлечь токен
ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.accessToken')

# IntrospectToken
grpcurl -plaintext -d "{
  \"access_token\": \"$ACCESS_TOKEN\"
}" localhost:50057 auth.v1.AuthService/IntrospectToken
```

**Ожидаемый ответ:**
```json
{
  "active": true,
  "userId": "<uuid>",
  "expiresAt": "2026-01-12T20:59:19Z"
}
```

### 6. Прямой вызов Identity CreateMe (для отладки)

**Примечание:** CreateMe — публичный метод (без auth), так как вызывается из auth-service.

```bash
grpcurl -plaintext -d '{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "directtest",
  "first_name": "Direct",
  "last_name": "Test",
  "timezone": "UTC",
  "idempotency_key": {"key": "direct-test-1"}
}' localhost:50051 identity.v1.MeService/CreateMe
```

**Ожидаемый ответ:**
```json
{
  "user": {
    "userId": "550e8400-e29b-41d4-a716-446655440000",
    "username": "directtest",
    "firstName": "Direct",
    "lastName": "Test",
    "avatarUrl": "",
    "timezone": "UTC"
  }
}
```

**Повторный вызов с тем же idempotency_key:**
```bash
grpcurl -plaintext -d '{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "directtest",
  "first_name": "Direct",
  "last_name": "Test",
  "timezone": "UTC",
  "idempotency_key": {"key": "direct-test-1"}
}' localhost:50051 identity.v1.MeService/CreateMe
```

**Ожидается:** Тот же ответ (идемпотентный)

## Мониторинг в реальном времени

### Логи Outbox Processor

```bash
# Следить за обработкой событий
docker-compose -f deployments/docker-compose.yml logs -f auth-service | grep -E "(outbox|user.registered)"
```

**Ожидаемый вывод:**
```
handler registered event_type="user.registered"
outbox processor started polling_interval=1s
event processed event_id="..." event_type="user.registered" aggregate_id="..."
```

### Логи Identity Service

```bash
# Следить за созданием профилей
docker-compose -f deployments/docker-compose.yml logs -f identity-service | grep -E "(user created|CreateMe)"
```

**Ожидаемый вывод:**
```
user created user_id="..."
```

### Мониторинг БД в реальном времени

```bash
# Терминал 1: Следить за новыми записями в auth.users
watch -n 1 'docker-compose -f deployments/docker-compose.yml exec -T postgres psql -U hackathon -d hackathon -c "SELECT COUNT(*) FROM auth.users;"'

# Терминал 2: Следить за outbox_events
watch -n 1 'docker-compose -f deployments/docker-compose.yml exec -T postgres psql -U hackathon -d hackathon -c "SELECT status, COUNT(*) FROM auth.outbox_events GROUP BY status;"'

# Терминал 3: Следить за identity.users
watch -n 1 'docker-compose -f deployments/docker-compose.yml exec -T postgres psql -U hackathon -d hackathon -c "SELECT COUNT(*) FROM identity.users;"'
```

## Полный тестовый скрипт

Создайте файл `test-integration.sh`:

```bash
#!/bin/bash

set -e

echo "=== Auth + Identity Integration Test ==="
echo ""

# 1. Регистрация пользователя
echo "[1/5] Registering user 'testuser'..."
RESPONSE=$(grpcurl -plaintext -d '{
  "username": "testuser",
  "email": "test@example.com",
  "password": "SecurePass123",
  "first_name": "Test",
  "last_name": "User",
  "timezone": "Europe/Moscow",
  "idempotency_key": {"key": "test-user-integration"}
}' localhost:50057 auth.v1.AuthService/Register)

USER_ID=$(echo "$RESPONSE" | jq -r '.accessToken' | cut -d'.' -f2 | base64 -d 2>/dev/null | jq -r '.user_id')
echo "✓ User registered, ID: $USER_ID"
echo ""

# 2. Проверка auth.users
echo "[2/5] Checking auth.users..."
AUTH_COUNT=$(docker-compose -f deployments/docker-compose.yml exec -T postgres psql -U hackathon -d hackathon -t -c \
  "SELECT COUNT(*) FROM auth.users WHERE username = 'testuser';")

if [ "$AUTH_COUNT" -eq 1 ]; then
  echo "✓ User found in auth.users"
else
  echo "✗ User NOT found in auth.users"
  exit 1
fi
echo ""

# 3. Ждем обработки outbox
echo "[3/5] Waiting for outbox processing (3 seconds)..."
sleep 3
echo ""

# 4. Проверка outbox_events
echo "[4/5] Checking outbox_events..."
OUTBOX_STATUS=$(docker-compose -f deployments/docker-compose.yml exec -T postgres psql -U hackathon -d hackathon -t -c \
  "SELECT status FROM auth.outbox_events WHERE payload::json->>'username' = 'testuser';")

if [[ "$OUTBOX_STATUS" == *"processed"* ]]; then
  echo "✓ Outbox event processed"
else
  echo "✗ Outbox event NOT processed (status: $OUTBOX_STATUS)"
  exit 1
fi
echo ""

# 5. Проверка identity.users
echo "[5/5] Checking identity.users..."
IDENTITY_COUNT=$(docker-compose -f deployments/docker-compose.yml exec -T postgres psql -U hackathon -d hackathon -t -c \
  "SELECT COUNT(*) FROM identity.users WHERE username = 'testuser';")

if [ "$IDENTITY_COUNT" -eq 1 ]; then
  echo "✓ Profile found in identity.users"
else
  echo "✗ Profile NOT found in identity.users"
  exit 1
fi
echo ""

echo "=== ✓ ALL TESTS PASSED ==="
```

Запуск:
```bash
chmod +x test-integration.sh
./test-integration.sh
```

## Troubleshooting

### События не обрабатываются

**Проверка 1:** Outbox processor запущен?
```bash
docker-compose -f deployments/docker-compose.yml logs auth-service | grep "outbox processor started"
```

**Проверка 2:** Handler зарегистрирован?
```bash
docker-compose -f deployments/docker-compose.yml logs auth-service | grep "handler registered"
```

**Проверка 3:** Есть ли pending события?
```bash
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon -c \
  "SELECT COUNT(*) FROM auth.outbox_events WHERE status = 'pending';"
```

### Профили не создаются в identity.users

**Проверка:** Логи identity-service
```bash
docker-compose -f deployments/docker-compose.yml logs --tail=50 identity-service
```

Частые ошибки:
- `relation "identity.users" does not exist` → Не применены миграции
- `relation "identity.idempotency_keys" does not exist` → Не применены миграции
- `null value in column "avatar_url"` → Исправлено в коде

## Полезные команды

### Список методов

```bash
# Auth Service
grpcurl -plaintext localhost:50057 list auth.v1.AuthService

# Identity Service
grpcurl -plaintext localhost:50051 list identity.v1.MeService
```

### Описание метода

```bash
grpcurl -plaintext localhost:50057 describe auth.v1.AuthService.Register
grpcurl -plaintext localhost:50051 describe identity.v1.MeService.CreateMe
```

### Очистка тестовых данных

```bash
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon << 'EOF'
DELETE FROM auth.refresh_tokens;
DELETE FROM auth.credentials;
DELETE FROM identity.users;
DELETE FROM auth.users;
DELETE FROM auth.outbox_events;
DELETE FROM auth.idempotency_keys;
DELETE FROM identity.idempotency_keys;
EOF
```

