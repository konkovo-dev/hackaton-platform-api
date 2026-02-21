# Hackathon Platform - Docker Setup

Полная инструкция по запуску всей платформы (Gateway + Auth + Identity + Hackathon + Participation & Roles)

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
make hackaton-service-sqlc-generate
make participation-and-roles-service-sqlc-generate
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

```bash
export DB_DSN="postgres://hackathon:hackathon_dev_password@localhost:5432/hackathon?sslmode=disable"
make goose-install
make auth-service-migrate-up
make auth-service-migrate-status
make identity-service-migrate-up
make identity-service-migrate-status
make hackaton-service-migrate-up
make hackaton-service-migrate-status
make participation-and-roles-service-migrate-up
make participation-and-roles-service-migrate-status
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

#### Hackathon Service

```bash
docker-compose -f deployments/docker-compose.yml build hackaton-service
docker-compose -f deployments/docker-compose.yml up -d hackaton-service
docker-compose -f deployments/docker-compose.yml logs --tail=20 hackaton-service
```

#### Participation and Roles Service

```bash
docker-compose -f deployments/docker-compose.yml build participation-and-roles-service
docker-compose -f deployments/docker-compose.yml up -d participation-and-roles-service
docker-compose -f deployments/docker-compose.yml logs --tail=20 participation-and-roles-service
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
- `hackathon-hackaton-service`
- `hackathon-participation-and-roles-service`
- `hackathon-gateway`

## Endpoints

- **Gateway HTTP**: `http://localhost:8080` (REST API для всех сервисов)
- **Identity gRPC**: `localhost:50051` (прямой gRPC)
- **Hackathon gRPC**: `localhost:50052` (прямой gRPC)
- **Participation & Roles gRPC**: `localhost:50055` (прямой gRPC)
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

# Создать хакатон
curl -X POST http://localhost:8080/v1/hackathons \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Hackathon",
    "short_description": "Quick test",
    "description": "Full description here",
    "location": {"online": true},
    "dates": {
      "registration_opens_at": "2026-03-01T00:00:00Z",
      "registration_closes_at": "2026-03-15T23:59:59Z",
      "starts_at": "2026-03-20T09:00:00Z",
      "ends_at": "2026-03-22T18:00:00Z",
      "judging_ends_at": "2026-03-25T18:00:00Z"
    },
    "registration_policy": {"allow_individual": true, "allow_team": true},
    "limits": {"team_size_max": 5}
  }' | jq .

# Список опубликованных хакатонов
curl http://localhost:8080/v1/hackathons?page_size=10 | jq .
```

### Через gRPC напрямую

```bash
# Ping identity
grpcurl -plaintext localhost:50051 identity.v1.PingService/Ping

# Ping auth
grpcurl -plaintext localhost:50057 auth.v1.PingService/Ping

# Ping hackathon
grpcurl -plaintext localhost:50052 grpc.health.v1.Health/Check

# List hackathon methods
grpcurl -plaintext localhost:50052 list hackathon.v1.HackathonService
```

## Управление сервисами

### Перезапуск отдельного сервиса

```bash
# Identity
docker-compose -f deployments/docker-compose.yml restart identity-service

# Auth
docker-compose -f deployments/docker-compose.yml restart auth-service

# Hackathon
docker-compose -f deployments/docker-compose.yml restart hackaton-service

# Participation and Roles
docker-compose -f deployments/docker-compose.yml restart participation-and-roles-service

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

# Hackathon
docker-compose -f deployments/docker-compose.yml stop hackaton-service
docker-compose -f deployments/docker-compose.yml rm -f hackaton-service
docker-compose -f deployments/docker-compose.yml build hackaton-service
docker-compose -f deployments/docker-compose.yml up -d hackaton-service

# Participation and Roles
docker-compose -f deployments/docker-compose.yml stop participation-and-roles-service
docker-compose -f deployments/docker-compose.yml rm -f participation-and-roles-service
docker-compose -f deployments/docker-compose.yml build participation-and-roles-service
docker-compose -f deployments/docker-compose.yml up -d participation-and-roles-service

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
docker-compose -f deployments/docker-compose.yml logs -f hackaton-service
docker-compose -f deployments/docker-compose.yml logs -f participation-and-roles-service
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
make hackaton-service-migrate-status
make participation-and-roles-service-migrate-status

# Примените заново
make auth-service-migrate-up
make identity-service-migrate-up
make hackaton-service-migrate-up
make participation-and-roles-service-migrate-up
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

### Hackathon service не видит роли

**Симптомы**: "access denied" для owner хакатона

**Решение**:
```bash
# Проверьте что participation-and-roles-service запущен
docker-compose -f deployments/docker-compose.yml ps participation-and-roles-service

# Проверьте логи hackathon-service
docker-compose -f deployments/docker-compose.yml logs hackaton-service | grep -i "participation"

# Проверьте что порт правильный (50055)
docker-compose -f deployments/docker-compose.yml exec hackaton-service env | grep PARTICIPATION
```

## Automated Testing

### Запуск полных тестов

```bash
# REST тесты для Hackathon Service
cd docs/hackathon
./rest-script.sh

# gRPC тесты для Hackathon Service
cd docs/hackathon
./grpc-script.sh
```

### Требования для тестов

- `jq` — для парсинга JSON
- `grpcurl` — для gRPC тестов

```bash
# macOS
brew install jq grpcurl

# Ubuntu/Debian
sudo apt-get install jq
# grpcurl: см. https://github.com/fullstorydev/grpcurl
```

## Дополнительная информация

### Guides по сервисам

- **Auth Service**: [REST](./auth/rest-guide.md) | [gRPC](./auth/grpc-guide.md)
- **Identity Service**: [README](./identity/README.md) | [REST Guides](./identity/)
- **Hackathon Service**: [README](./hackathon/README.md) | [REST](./hackathon/rest-guide.md) | [gRPC](./hackathon/grpc-guide.md)
- **Participation & Roles Service**: [README](./participation-and-roles/README.md) | [REST](./participation-and-roles/rest-guide.md) | [gRPC](./participation-and-roles/grpc-guide.md)

### Automated Scripts

- **Auth**: [grpc-script.sh](./auth/grpc-script.sh)
- **Hackathon**: [rest-script.sh](./hackathon/rest-script.sh) | [grpc-script.sh](./hackathon/grpc-script.sh)
- **Participation & Roles**: [rest-script.sh](./participation-and-roles/rest-script.sh) | [grpc-script.sh](./participation-and-roles/grpc-script.sh)

### Business Rules

- [Hackathon Rules](./rules/hackathon.md)
- [Identity Rules](./rules/identity.md)
- [Team Rules](./rules/team.md)

