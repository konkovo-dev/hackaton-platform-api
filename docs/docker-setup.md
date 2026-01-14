# Hackathon Platform - Docker Setup

Полная инструкция по запуску всей платформы (Gateway + Auth + Identity)

## Предусловия

- Docker и Docker Compose установлены
- Make установлен
- Go 1.22+ установлен (для генерации)

## Быстрый старт

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

### 4. Генерация RSA ключей для auth-service

```bash
make auth-service-generate-keys
```

### 5. Запуск Postgres

```bash
docker-compose -f deployments/docker-compose.yml up -d postgres
make ps
```

Дождитесь, пока postgres станет healthy (около 5-10 секунд).

### 6. Миграции

#### Auth Service

```bash
export DB_DSN="postgres://hackathon:hackathon_dev_password@localhost:5432/hackathon?sslmode=disable"
make goose-install
make auth-service-migrate-up
make auth-service-migrate-status
```

#### Identity Service

```bash
make identity-service-migrate-up
make identity-service-migrate-status
```

### 7. Запуск сервисов

#### Identity Service

```bash
docker-compose -f deployments/docker-compose.yml build identity-service
docker-compose -f deployments/docker-compose.yml up -d identity-service
docker-compose -f deployments/docker-compose.yml logs --tail=20 identity-service
```

#### Auth Service

```bash
docker-compose -f deployments/docker-compose.yml build auth-service
docker-compose -f deployments/docker-compose.yml up -d auth-service
docker-compose -f deployments/docker-compose.yml logs --tail=20 auth-service
```

#### Gateway

```bash
docker-compose -f deployments/docker-compose.yml build gateway
docker-compose -f deployments/docker-compose.yml up -d --no-deps gateway
docker-compose -f deployments/docker-compose.yml logs --tail=20 gateway
```

### 8. Проверка

```bash
make ps
```

Должны быть запущены и healthy:
- `hackathon-postgres`
- `hackathon-identity-service`
- `hackathon-auth-service`
- `hackathon-gateway`

## Endpoints

- **Gateway HTTP**: `http://localhost:8080` (REST API для всех сервисов)
- **Identity gRPC**: `localhost:50051` (прямой gRPC)
- **Auth gRPC**: `localhost:50057` (прямой gRPC)
- **Postgres**: `localhost:5432`

## Быстрая проверка работоспособности

### Через Gateway (REST)

```bash
# Health check
curl http://localhost:8080/v1/ping | jq .

# Регистрация пользователя
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "SecurePass123",
    "first_name": "Test",
    "last_name": "User",
    "timezone": "UTC"
  }' | jq .

# Сохраните access_token из ответа
ACCESS_TOKEN="<your-access-token>"

# Получить свой профиль
curl http://localhost:8080/v1/users/me \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

### Через gRPC напрямую

```bash
# Ping identity
grpcurl -plaintext localhost:50051 identity.v1.PingService/Ping

# Ping auth
grpcurl -plaintext localhost:50057 auth.v1.PingService/Ping
```

## Управление сервисами

### Перезапуск отдельного сервиса

```bash
# Identity
docker-compose -f deployments/docker-compose.yml restart identity-service

# Auth
docker-compose -f deployments/docker-compose.yml restart auth-service

# Gateway
docker-compose -f deployments/docker-compose.yml restart gateway
```

### Пересборка и перезапуск

```bash
# Identity
docker-compose -f deployments/docker-compose.yml stop identity-service
docker-compose -f deployments/docker-compose.yml rm -f identity-service
docker-compose -f deployments/docker-compose.yml build identity-service
docker-compose -f deployments/docker-compose.yml up -d identity-service

# Auth
docker-compose -f deployments/docker-compose.yml stop auth-service
docker-compose -f deployments/docker-compose.yml rm -f auth-service
docker-compose -f deployments/docker-compose.yml build auth-service
docker-compose -f deployments/docker-compose.yml up -d auth-service

# Gateway
docker-compose -f deployments/docker-compose.yml stop gateway
docker-compose -f deployments/docker-compose.yml rm -f gateway
docker-compose -f deployments/docker-compose.yml build gateway
docker-compose -f deployments/docker-compose.yml up -d --no-deps gateway
```

### Логи

```bash
# Все сервисы
docker-compose -f deployments/docker-compose.yml logs -f

# Конкретный сервис
docker-compose -f deployments/docker-compose.yml logs -f identity-service
docker-compose -f deployments/docker-compose.yml logs -f auth-service
docker-compose -f deployments/docker-compose.yml logs -f gateway
```

### Остановка всех сервисов

```bash
docker-compose -f deployments/docker-compose.yml down
```

### Полная очистка (включая volumes)

```bash
docker-compose -f deployments/docker-compose.yml down -v
```

## Troubleshooting

### Gateway не подключается к сервисам

**Симптомы**: `grpc: the client connection is closing` в логах или при запросах

**Решение**:
```bash
# Убедитесь, что identity и auth запущены
docker-compose -f deployments/docker-compose.yml ps

# Перезапустите gateway
docker-compose -f deployments/docker-compose.yml restart gateway
```

### Миграции не применились

**Решение**:
```bash
# Проверьте статус
make auth-service-migrate-status
make identity-service-migrate-status

# Примените заново
make auth-service-migrate-up
make identity-service-migrate-up
```

### Postgres не стартует

**Решение**:
```bash
# Проверьте логи
docker-compose -f deployments/docker-compose.yml logs postgres

# Пересоздайте контейнер
docker-compose -f deployments/docker-compose.yml stop postgres
docker-compose -f deployments/docker-compose.yml rm -f postgres
docker-compose -f deployments/docker-compose.yml up -d postgres
```

### Port already in use

**Решение**:
```bash
# Найдите процесс на порту (например, 8080)
lsof -ti:8080 | xargs kill -9

# Или измените порт в docker-compose.yml
```

## Дополнительная информация

- [Auth Service REST Guide](./auth/rest-guide.md)
- [Auth Service gRPC Guide](./auth/grpc-guide.md)
- [Identity Service REST Guide](./identity/rest-guide.md)
- [Identity Service gRPC Guide](./identity/grpc-guide.md)

