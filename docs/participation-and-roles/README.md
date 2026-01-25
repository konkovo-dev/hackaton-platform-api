# Participation and Roles Service Documentation

## Описание

Participation and Roles Service управляет staff-ролями и участием в хакатонах. Сервис реализует:

- **Staff Management**: Управление staff ролями (OWNER, ORGANIZER, MENTOR, JUDGE)
- **Staff Invitations**: Приглашение пользователей в staff команду
- **Access Control**: Политики доступа основанные на ролях

## Роли

- **OWNER** - владелец хакатона, единственный, полные права
- **ORGANIZER** - организатор, помогает в организации
- **MENTOR** - ментор, помогает участникам
- **JUDGE** - судья, оценивает проекты

## Методы

### Public Methods (требуют авторизации)

1. **ListHackathonStaff** - просмотр staff команды (только для staff членов)
2. **CreateStaffInvitation** - создание приглашения (только OWNER)
3. **CancelStaffInvitation** - отмена приглашения (только OWNER)
4. **ListMyStaffInvitations** - список приглашений текущего пользователя
5. **AcceptStaffInvitation** - принятие приглашения
6. **RejectStaffInvitation** - отклонение приглашения
7. **RemoveHackathonRole** - удаление роли у пользователя (только OWNER, нельзя удалить OWNER)
8. **SelfRemoveHackathonRole** - самоудаление роли (нельзя удалить OWNER)

### Internal Methods (service-to-service)

1. **GetHackathonContext** - получение контекста пользователя (роли, участие)
2. **AssignHackathonRole** - назначение роли (используется HackathonService при создании)

## Endpoints

- **gRPC**: `localhost:50055`
- **REST (через gateway)**: `http://localhost:8080/v1/hackathons/{hackathon_id}/staff*`

## Документация

- [REST API Guide](./rest-guide.md) - руководство по REST API с примерами
- [gRPC API Guide](./grpc-guide.md) - руководство по gRPC API с примерами
- [Test Data SQL](./test-data.sql) - SQL скрипт для загрузки тестовых данных

## Тестирование

### Автоматическое тестирование

```bash
# REST API
chmod +x docs/participation-and-roles/rest-script.sh
./docs/participation-and-roles/rest-script.sh

# gRPC API
chmod +x docs/participation-and-roles/grpc-script.sh
./docs/participation-and-roles/grpc-script.sh
```

### Подготовка тестовых данных

```bash
# Загрузить тестовые данные в БД
docker-compose -f deployments/docker-compose.yml exec postgres \
  psql -U hackathon -d hackathon -f /path/to/test-data.sql

# Или напрямую через psql
docker-compose -f deployments/docker-compose.yml exec -T postgres \
  psql -U hackathon -d hackathon < docs/participation-and-roles/test-data.sql
```

Тестовые данные включают:
- 5 пользователей (alice_owner, bob_organizer, charlie_mentor, diana_judge, eve_user)
- 1 опубликованный хакатон
- 2 начальные staff роли (Alice = OWNER, Bob = ORGANIZER)

## Бизнес-правила

### Staff Invitations

1. **Только OWNER может создавать приглашения**
2. **Нельзя пригласить на роль OWNER**
3. **Нельзя пригласить пользователя, который уже имеет staff роль**
4. **Нельзя пригласить участника хакатона** (staff и participation взаимоисключающие)
5. **Только получатель может принять/отклонить приглашение**
6. **Только OWNER может отменить pending приглашение**

### Role Management

1. **OWNER уникален** - только один OWNER на хакатон
2. **OWNER нельзя удалить** - ни через RemoveHackathonRole, ни через SelfRemoveHackathonRole
3. **Staff и Participation взаимоисключающие** - нельзя быть и staff, и участником одновременно
4. **Только OWNER может удалять роли других пользователей**
5. **Пользователь может самоудалиться** (кроме OWNER)

## Архитектура

Сервис следует Clean Architecture:

```
transport/grpc/          - gRPC handlers
usecase/role/            - Business logic
  - policies/            - Access control policies
  - validators/          - Domain validation
repository/postgres/     - Database layer
domain/                  - Domain entities & enums
```

## Миграции

```bash
# Применить миграции
cd cmd/participation-and-roles-service
make migrate-up

# Откатить миграции
make migrate-down

# Создать новую миграцию
make migrate-create NAME=your_migration_name
```

## Зависимости

- **auth-service** (port 50057) - авторизация и JWT валидация
- **PostgreSQL** - хранилище данных
- Schema: `participation_and_roles`
- Tables: `staff_roles`, `staff_invitations`, `participations`, `idempotency_keys`

## Переменные окружения

```env
PARTICIPATION_AND_ROLES_SERVICE_GRPC_PORT=50055
DB_DSN=postgres://hackathon:password@localhost:5432/hackathon?sslmode=disable
AUTH_CLIENT_SERVICE_URL=auth-service:50057
IDEMPOTENCY_TTL=24h
SERVICE_AUTH_TOKEN=your-service-token
```

## Примеры использования

### Создание staff invitation

```bash
curl -X POST http://localhost:8080/v1/hackathons/{hackathon_id}/staff-invitations \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "user-uuid",
    "requested_role": "HACKATHON_ROLE_MENTOR",
    "message": "Join our team!",
    "idempotency_key": {"key": "unique-key"}
  }'
```

### Просмотр staff команды

```bash
curl http://localhost:8080/v1/hackathons/{hackathon_id}/staff \
  -H "Authorization: Bearer $TOKEN"
```

### Принятие приглашения

```bash
curl -X POST http://localhost:8080/v1/users/me/staff-invitations/{invitation_id}:accept \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": {"key": "unique-key"}
  }'
```

## Troubleshooting

### 401 Unauthenticated
- Проверьте наличие и валидность `Authorization: Bearer` токена
- Обновите токен через `/v1/auth/login`

### 403 Forbidden / 7 PermissionDenied
- Пользователь не имеет необходимых прав (например, не OWNER)
- Проверьте роли пользователя в БД

### 404 Not Found
- Хакатон, инвайт или пользователь не существует
- Проверьте корректность UUID

### 409 Conflict / AlreadyExists
- Уже существует pending invitation для этого пользователя и роли
- Отмените или дождитесь истечения существующего приглашения

## Дополнительная информация

- Business Rules: [../../docs/rules/hackathon-staff.md](../rules/hackathon-staff.md)
- Proto Definitions: [../../proto/participationandroles/v1/](../../proto/participationandroles/v1/)
- Docker Setup: [../docker-setup.md](../docker-setup.md)

