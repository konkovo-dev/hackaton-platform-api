# Auth + Identity Integration - Docker Setup

Этот документ описывает запуск интеграции `auth-service` и `identity-service` через Transactional Outbox Pattern.

## Архитектура

```
┌──────────────────────────────────────────────────────────────┐
│  Client → Register (auth-service)                            │
│     ↓                                                         │
│  1. Create user in auth.users                                │
│  2. Create outbox event (user.registered)                    │
│  3. Commit transaction                                       │
│  4. Return tokens to client                                  │
└──────────────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────────────┐
│  Outbox Processor (background, poll_interval=1s)            │
│     ↓                                                         │
│  1. Fetch pending events from auth.outbox_events            │
│  2. Call UserRegisteredHandler                               │
│  3. identityClient.CreateUser(gRPC)                          │
│  4. Mark event as 'processed'                                │
└──────────────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────────────┐
│  Identity Service → CreateMe (with idempotency)             │
│     ↓                                                         │
│  1. Check idempotency key                                    │
│  2. Create user in identity.users                            │
│  3. Save idempotency response                                │
│  4. Return success                                           │
└──────────────────────────────────────────────────────────────┘
```

## Предварительные требования

- Docker и Docker Compose
- `make` (для команд)
- `goose` (для миграций)
- `grpcurl` (для тестирования gRPC)

## Установка

### 1. Клонирование

```bash
git clone <repo-url>
cd hackaton-platform-api
```

### 2. Генерация proto

```bash
make buf-generate
```

### 3. Генерация sqlc

```bash
make auth-service-sqlc-generate
make identity-service-sqlc-generate
```

### 4. Генерация RSA ключей (для auth-service)

```bash
make auth-service-generate-keys
```

## Запуск сервисов

### 1. Запуск Postgres

```bash
docker-compose -f deployments/docker-compose.yml up -d postgres
```

Проверка:
```bash
docker-compose -f deployments/docker-compose.yml ps postgres
```

### 2. Применение миграций

#### Auth Service

```bash
export DB_DSN="postgres://hackathon:hackathon_dev_password@localhost:5432/hackathon?sslmode=disable"
make goose-install
make auth-service-migrate-up
```

Проверка:
```bash
make auth-service-migrate-status
```

#### Identity Service

```bash
cd internal/identity-service
goose -dir migrations postgres "$DB_DSN" up
goose -dir migrations postgres "$DB_DSN" status
cd ../..
```

Или через SQL напрямую:
```bash
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon -c "\dt auth.*"
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon -c "\dt identity.*"
```

Ожидаемые таблицы:
- `auth.users`
- `auth.credentials`
- `auth.refresh_tokens`
- `auth.idempotency_keys`
- `auth.outbox_events` ← **новая таблица для outbox**
- `identity.users` ← **таблица профилей**
- `identity.idempotency_keys`

### 3. Запуск Identity Service

```bash
docker-compose -f deployments/docker-compose.yml build identity-service
docker-compose -f deployments/docker-compose.yml up -d identity-service
```

Проверка логов:
```bash
docker-compose -f deployments/docker-compose.yml logs --tail=20 identity-service
```

Ожидаемый вывод:
```
identity database connection pool initialized
starting grpc server addr="[::]:50051"
[Fx] RUNNING
```

### 4. Запуск Auth Service (с Outbox Processor)

```bash
docker-compose -f deployments/docker-compose.yml build auth-service
docker-compose -f deployments/docker-compose.yml up -d auth-service
```

Проверка логов:
```bash
docker-compose -f deployments/docker-compose.yml logs --tail=30 auth-service
```

Ожидаемый вывод:
```
handler registered event_type="user.registered"
database connection pool initialized
starting grpc server addr="[::]:50057"
outbox processor starting
outbox processor started polling_interval=1s batch_size=100 max_attempts=3
[Fx] RUNNING
```

### 5. Проверка статуса

```bash
docker-compose -f deployments/docker-compose.yml ps
```

Ожидаемый вывод:
```
NAME                           STATUS
hackathon-postgres             Up (healthy)
hackathon-auth-service         Up
hackathon-identity-service     Up
```

## Переменные окружения

### Auth Service

```bash
# gRPC
AUTH_SERVICE_GRPC_PORT=50057

# Database
DB_DSN=postgres://hackathon:hackathon_dev_password@postgres:5432/hackathon?sslmode=disable

# JWT
RS256_PRIVATE_KEY_PATH=/app/keys/private_key.pem
JWT_KEY_ID=key-1
JWT_ISSUER=hackaton-platform
JWT_AUDIENCE=hackaton-platform
ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=720h

# Identity Client (для outbox)
IDENTITY_SERVICE_URL=identity-service:50051

# Outbox Configuration
OUTBOX_POLL_INTERVAL=1s       # частота опроса pending событий
OUTBOX_MAX_ATTEMPTS=3         # макс попыток обработки события
IDEMPOTENCY_TTL=24h
```

### Identity Service

```bash
# gRPC
IDENTITY_SERVICE_GRPC_PORT=50051

# Database
DB_DSN=postgres://hackathon:hackathon_dev_password@postgres:5432/hackathon?sslmode=disable

# Auth Client (для introspection)
AUTH_CLIENT_SERVICE_URL=auth-service:50057
AUTH_CLIENT_MAX_CACHE_TTL=60s
AUTH_CLIENT_CACHE_CLEANUP_INTERVAL=1m

# Idempotency
IDEMPOTENCY_TTL=24h
```

## Мониторинг Outbox

### Проверка статуса событий

```bash
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon -c \
  "SELECT event_type, status, attempt_count, created_at FROM auth.outbox_events ORDER BY created_at DESC LIMIT 10;"
```

### Статусы событий

- `pending` - ожидает обработки
- `processing` - в процессе обработки
- `processed` - успешно обработано
- `failed` - превышено максимальное количество попыток (max_attempts)

### Проверка ошибок

```bash
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon -c \
  "SELECT event_type, status, attempt_count, last_error FROM auth.outbox_events WHERE status = 'failed';"
```

### Повторная обработка failed событий

```bash
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon -c \
  "UPDATE auth.outbox_events SET status = 'pending', attempt_count = 0, last_error = '' WHERE status = 'failed';"
```

## Очистка

### Остановить сервисы

```bash
docker-compose -f deployments/docker-compose.yml stop auth-service identity-service
```

### Удалить данные

```bash
docker-compose -f deployments/docker-compose.yml down -v
```

### Откатить миграции

```bash
export DB_DSN="postgres://hackathon:hackathon_dev_password@localhost:5432/hackathon?sslmode=disable"
make auth-service-migrate-down

cd internal/identity-service
goose -dir migrations postgres "$DB_DSN" down
cd ../..
```

## Troubleshooting

### Ошибка: "relation identity.users does not exist"

**Причина:** Миграции identity-service не применены.

**Решение:**
```bash
export DB_DSN="postgres://hackathon:hackathon_dev_password@localhost:5432/hackathon?sslmode=disable"
cd internal/identity-service
goose -dir migrations postgres "$DB_DSN" up
cd ../..
```

### Ошибка: "handler not found for event type: user.registered"

**Причина:** Outbox processor не зарегистрировал handler.

**Решение:** Проверить логи auth-service при старте:
```bash
docker-compose -f deployments/docker-compose.yml logs auth-service | grep "handler registered"
```

Ожидаемый вывод: `handler registered event_type="user.registered"`

### Ошибка: "failed to create user in identity service"

**Причина:** identity-service недоступен или возвращает ошибку.

**Решение:**
1. Проверить логи identity-service:
   ```bash
   docker-compose -f deployments/docker-compose.yml logs --tail=50 identity-service
   ```

2. Проверить что identity-service запущен:
   ```bash
   docker-compose -f deployments/docker-compose.yml ps identity-service
   ```

3. Проверить подключение из auth-service к identity-service:
   ```bash
   docker-compose -f deployments/docker-compose.yml exec auth-service ping -c 1 identity-service
   ```

### События застряли в статусе "processing"

**Причина:** Outbox processor был остановлен во время обработки.

**Решение:**
```bash
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon -c \
  "UPDATE auth.outbox_events SET status = 'pending' WHERE status = 'processing';"
```

### Проверить связь между auth.users и identity.users

```bash
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon -c \
  "SELECT a.id, a.username, a.email, i.first_name, i.last_name, i.timezone 
   FROM auth.users a 
   LEFT JOIN identity.users i ON a.id = i.id 
   ORDER BY a.created_at;"
```

## Полезные команды

### Логи

```bash
# Все логи auth-service
docker-compose -f deployments/docker-compose.yml logs -f auth-service

# Только outbox processor
docker-compose -f deployments/docker-compose.yml logs auth-service | grep outbox

# Все логи identity-service
docker-compose -f deployments/docker-compose.yml logs -f identity-service
```

### Подключение к БД

```bash
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon
```

### Список таблиц

```bash
# Auth schema
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon -c "\dt auth.*"

# Identity schema
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon -c "\dt identity.*"
```

### Очистка outbox событий

```bash
# Удалить обработанные события старше 7 дней
docker-compose -f deployments/docker-compose.yml exec postgres psql -U hackathon -d hackathon -c \
  "DELETE FROM auth.outbox_events WHERE status = 'processed' AND created_at < NOW() - INTERVAL '7 days';"
```

