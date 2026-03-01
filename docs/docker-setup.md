# Hackathon Platform - Docker Setup

Полная инструкция по запуску всей платформы (Gateway + Auth + Identity + Hackathon + Participation & Roles + Team + Mentors + Matchmaking + Submission + NATS + Centrifugo + MinIO)

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
make team-service-sqlc-generate
make mentors-service-sqlc-generate
make matchmaking-service-sqlc-generate
make submission-service-sqlc-generate
```

### 4. Генерация RSA ключей для auth-service

```bash
make auth-service-generate-keys
```

### 5. Запуск инфраструктуры (Postgres, NATS, Centrifugo, MinIO)

```bash
docker-compose -f deployments/docker-compose.yml up -d postgres nats centrifugo minio
make ps
```

Дождитесь, пока все сервисы станут healthy (около 5-10 секунд).

**Что запускается:**
- **Postgres** (порт 5432) — основная БД
- **NATS** (порты 4222, 8222) — message broker для event streaming
- **Centrifugo** (порт 8000) — WebSocket сервер для real-time уведомлений
- **MinIO** (порты 9000, 9001) — S3-совместимое хранилище для файлов посылок

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
make team-service-migrate-up
make team-service-migrate-status
make mentors-service-migrate-up
make mentors-service-migrate-status
make matchmaking-service-migrate-up
make matchmaking-service-migrate-status
make submission-service-migrate-up
make submission-service-migrate-status
```

### 7. Запуск сервисов

#### Identity Service

```bash
docker-compose -f deployments/docker-compose.yml build identity-service
docker-compose -f deployments/docker-compose.yml up -d identity-service
docker-compose -f deployments/docker-compose.yml logs --tail=20 identity-service
```
docker-compose -f deployments/docker-compose.yml stop identity-service 
docker-compose -f deployments/docker-compose.yml rm -f identity-service                   
docker rmi $(docker images -q deployments-identity-service)

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

docker-compose -f deployments/docker-compose.yml stop hackaton-service 
docker-compose -f deployments/docker-compose.yml rm -f hackaton-service                   
docker rmi $(docker images -q deployments-hackaton-service)

#### Participation and Roles Service

```bash
docker-compose -f deployments/docker-compose.yml build participation-and-roles-service
docker-compose -f deployments/docker-compose.yml up -d participation-and-roles-service
docker-compose -f deployments/docker-compose.yml logs --tail=20 participation-and-roles-service
```

#### Team Service

```bash
docker-compose -f deployments/docker-compose.yml build team-service
docker-compose -f deployments/docker-compose.yml up -d team-service
docker-compose -f deployments/docker-compose.yml logs --tail=20 team-service
```

#### Mentors Service

```bash
docker-compose -f deployments/docker-compose.yml build mentors-service
docker-compose -f deployments/docker-compose.yml up -d mentors-service
docker-compose -f deployments/docker-compose.yml logs --tail=20 mentors-service
```

docker-compose -f deployments/docker-compose.yml stop mentors-service 
docker-compose -f deployments/docker-compose.yml rm -f mentors-service                   
docker rmi $(docker images -q deployments-mentors-service)

**Важно**: Mentors Service требует NATS и Centrifugo для работы outbox и real-time уведомлений.

#### Matchmaking Service

```bash
docker-compose -f deployments/docker-compose.yml build matchmaking-service
docker-compose -f deployments/docker-compose.yml up -d matchmaking-service
docker-compose -f deployments/docker-compose.yml logs --tail=20 matchmaking-service
```

docker-compose -f deployments/docker-compose.yml stop matchmaking-service 
docker-compose -f deployments/docker-compose.yml rm -f matchmaking-service                   
docker rmi $(docker images -q deployments-matchmaking-service)

**Важно**: Matchmaking Service требует NATS для синхронизации read-model из других сервисов.

#### Submission Service

```bash
docker-compose -f deployments/docker-compose.yml build submission-service
docker-compose -f deployments/docker-compose.yml up -d submission-service
docker-compose -f deployments/docker-compose.yml logs --tail=20 submission-service
```

docker-compose -f deployments/docker-compose.yml stop submission-service 
docker-compose -f deployments/docker-compose.yml rm -f submission-service                   
docker rmi $(docker images -q deployments-submission-service)

**Важно**: Submission Service требует MinIO для хранения файлов посылок.

#### Gateway

```bash
docker-compose -f deployments/docker-compose.yml build gateway
docker-compose -f deployments/docker-compose.yml up -d --no-deps gateway
docker-compose -f deployments/docker-compose.yml logs --tail=20 gateway
```

docker-compose -f deployments/docker-compose.yml stop gateway
docker-compose -f deployments/docker-compose.yml rm -f gateway                   
docker rmi $(docker images -q deployments-gateway)

### 8. Проверка

```bash
make ps
```

Должны быть запущены и healthy:
- `hackathon-postgres`
- `hackathon-nats`
- `hackathon-centrifugo`
- `hackathon-minio`
- `hackathon-identity-service`
- `hackathon-auth-service`
- `hackathon-hackaton-service`
- `hackathon-participation-and-roles-service`
- `hackathon-team-service`
- `hackathon-mentors-service`
- `hackathon-matchmaking-service`
- `hackathon-submission-service`
- `hackathon-gateway`

## Endpoints

### HTTP/REST
- **Gateway HTTP**: `http://localhost:8080` (REST API для всех сервисов)
- **Swagger UI**: `http://localhost:8080/swagger/` (документация API)

### gRPC Services
- **Identity gRPC**: `localhost:50051`
- **Hackathon gRPC**: `localhost:50052`
- **Team gRPC**: `localhost:50053`
- **Participation & Roles gRPC**: `localhost:50055`
- **Mentors gRPC**: `localhost:50056`
- **Auth gRPC**: `localhost:50057`
- **Submission gRPC**: `localhost:50058`
- **Matchmaking gRPC**: `localhost:50059`

### Infrastructure
- **Postgres**: `localhost:5432`
- **NATS**: `localhost:4222` (client), `localhost:8222` (monitoring)
- **Centrifugo**: `localhost:8000` (WebSocket + HTTP API)
- **MinIO**: `localhost:9000` (S3 API), `localhost:9001` (Console UI)

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

# Отправить сообщение в support (mentors-service)
curl -X POST http://localhost:8080/v1/hackathons/<HACKATHON_ID>/support/messages \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Нужна помощь с проектом",
    "client_message_id": "msg-001"
  }' | jq .

# Получить токен для WebSocket (Centrifugo)
curl http://localhost:8080/v1/hackathons/<HACKATHON_ID>/support/realtime-token \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .

# Получить рекомендации команд для участника (matchmaking-service)
curl http://localhost:8080/v1/hackathons/<HACKATHON_ID>/matchmaking/teams?limit=10 \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .

# Получить рекомендации кандидатов для капитана команды (matchmaking-service)
curl "http://localhost:8080/v1/hackathons/<HACKATHON_ID>/matchmaking/candidates?team_id=<TEAM_ID>&vacancy_id=<VACANCY_ID>&limit=10" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
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

# Ping mentors
grpcurl -plaintext localhost:50056 grpc.health.v1.Health/Check

# List mentors methods
grpcurl -plaintext localhost:50056 list mentors.v1.MentorsService

# Ping matchmaking
grpcurl -plaintext localhost:50059 grpc.health.v1.Health/Check

# List matchmaking methods
grpcurl -plaintext localhost:50059 list matchmaking.v1.MatchmakingService
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

# Team
docker-compose -f deployments/docker-compose.yml restart team-service

# Mentors
docker-compose -f deployments/docker-compose.yml restart mentors-service

# Matchmaking
docker-compose -f deployments/docker-compose.yml restart matchmaking-service

# Gateway
docker-compose -f deployments/docker-compose.yml restart gateway

# Infrastructure
docker-compose -f deployments/docker-compose.yml restart nats
docker-compose -f deployments/docker-compose.yml restart centrifugo
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

# Team
docker-compose -f deployments/docker-compose.yml stop team-service
docker-compose -f deployments/docker-compose.yml rm -f team-service
docker-compose -f deployments/docker-compose.yml build team-service
docker-compose -f deployments/docker-compose.yml up -d team-service

# Mentors
docker-compose -f deployments/docker-compose.yml stop mentors-service
docker-compose -f deployments/docker-compose.yml rm -f mentors-service
docker-compose -f deployments/docker-compose.yml build mentors-service
docker-compose -f deployments/docker-compose.yml up -d mentors-service

# Matchmaking
docker-compose -f deployments/docker-compose.yml stop matchmaking-service
docker-compose -f deployments/docker-compose.yml rm -f matchmaking-service
docker-compose -f deployments/docker-compose.yml build matchmaking-service
docker-compose -f deployments/docker-compose.yml up -d matchmaking-service

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
docker-compose -f deployments/docker-compose.yml logs -f team-service
docker-compose -f deployments/docker-compose.yml logs -f mentors-service
docker-compose -f deployments/docker-compose.yml logs -f matchmaking-service
docker-compose -f deployments/docker-compose.yml logs -f gateway

# Infrastructure
docker-compose -f deployments/docker-compose.yml logs -f nats
docker-compose -f deployments/docker-compose.yml logs -f centrifugo
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
make team-service-migrate-status
make mentors-service-migrate-status
make matchmaking-service-migrate-status

# Примените заново
make auth-service-migrate-up
make identity-service-migrate-up
make hackaton-service-migrate-up
make participation-and-roles-service-migrate-up
make team-service-migrate-up
make mentors-service-migrate-up
make matchmaking-service-migrate-up
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

### Mentors service не может отправить события

**Симптомы**: "failed to publish" в логах mentors-service

**Решение**:
```bash
# Проверьте что NATS и Centrifugo запущены
docker-compose -f deployments/docker-compose.yml ps nats centrifugo

# Проверьте логи mentors-service
docker-compose -f deployments/docker-compose.yml logs mentors-service | grep -i "outbox\|nats\|centrifugo"

# Проверьте connectivity
docker-compose -f deployments/docker-compose.yml exec mentors-service wget -O- http://centrifugo:8000/health
docker-compose -f deployments/docker-compose.yml exec mentors-service wget -O- http://nats:8222/healthz

# Перезапустите mentors-service
docker-compose -f deployments/docker-compose.yml restart mentors-service
```

### WebSocket не подключается (Centrifugo)

**Симптомы**: Клиент не может подключиться к `ws://localhost:8000/connection/websocket`

**Решение**:
```bash
# Проверьте что Centrifugo запущен
docker-compose -f deployments/docker-compose.yml ps centrifugo

# Проверьте health
curl http://localhost:8000/health

# Проверьте логи
docker-compose -f deployments/docker-compose.yml logs centrifugo

# Получите токен через API
curl http://localhost:8080/v1/hackathons/<HACKATHON_ID>/support/realtime-token \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq -r .token
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
- **Team Service**: [README](./team/README.md)
- **Mentors Service**: Support/Helpdesk с real-time уведомлениями через WebSocket

### Automated Scripts

- **Auth**: [grpc-script.sh](./auth/grpc-script.sh)
- **Hackathon**: [rest-script.sh](./hackathon/rest-script.sh) | [grpc-script.sh](./hackathon/grpc-script.sh)
- **Participation & Roles**: [rest-script.sh](./participation-and-roles/rest-script.sh) | [grpc-script.sh](./participation-and-roles/grpc-script.sh)

### Business Rules

- [Hackathon Rules](./rules/hackathon.md)
- [Identity Rules](./rules/identity.md)
- [Team Rules](./rules/team.md)

### Architecture Components

#### NATS (Message Broker)
- **Назначение**: Event streaming для audit log и межсервисной коммуникации
- **Используется**: mentors-service для публикации событий (message.created, ticket.claimed, ticket.closed)
- **Monitoring**: `http://localhost:8222` (NATS monitoring endpoint)

#### Centrifugo (WebSocket Server)
- **Назначение**: Real-time уведомления для клиентов
- **Используется**: mentors-service для push-уведомлений о новых сообщениях в support чате
- **Channels**: `support:feed#<user_id>` (user-limited channels)
- **API**: `http://localhost:8000/api` (HTTP API для publish)
- **WebSocket**: `ws://localhost:8000/connection/websocket`

#### Mentors Service (Support/Helpdesk)
- **Назначение**: Тикет-система для поддержки участников менторами
- **Особенности**:
  - Claim-модель назначения тикетов
  - Idempotency для всех мутирующих операций
  - Outbox pattern для гарантированной доставки событий
  - Real-time через Centrifugo WebSocket
  - Audit log через NATS
- **Доступность**: Только на стадии RUNNING хакатона

#### Matchmaking Service
- **Назначение**: Рекомендательная система для подбора команд и участников
- **Особенности**:
  - Read-model синхронизируется через NATS события от identity, team, participation сервисов
  - Scoring algorithm: 63% skills + 27% roles + 10% text matching
  - PostgreSQL Full-Text Search для мотивационных текстов и описаний (english + russian)
  - Explainable recommendations с детальным breakdown скоринга
  - Два режима: User→Teams (участник ищет команду), Vacancy→Candidates (капитан ищет участников)
- **Доступность**: REGISTRATION и RUNNING стадии хакатона

