# Participation and Roles Service - Testing Summary

## Обзор

Полная end-to-end документация и тесты для Participation and Roles Service.

## Структура документации

```
docs/participation-and-roles/
├── README.md              # Основная документация
├── rest-guide.md          # REST API руководство с примерами
├── grpc-guide.md          # gRPC API руководство с примерами
├── rest-script.sh         # Автоматический REST тест (executable)
├── grpc-script.sh         # Автоматический gRPC тест (executable)
├── test-data.sql          # SQL скрипт для загрузки тестовых данных
└── TESTING_SUMMARY.md     # Этот файл
```

## Быстрый старт

### 1. Подготовка окружения

```bash
# Запустить все сервисы в Docker
cd deployments
docker-compose up -d

# Дождаться готовности всех сервисов
docker-compose ps

# Применить миграции (если не применены автоматически)
cd ../cmd/participation-and-roles-service
make migrate-up
```

### 2. Загрузить тестовые данные

```bash
# Из корня проекта
docker-compose -f deployments/docker-compose.yml exec -T postgres \
  psql -U hackathon -d hackathon < docs/participation-and-roles/test-data.sql
```

### 3. Запустить тесты

```bash
# REST API тесты
./docs/participation-and-roles/rest-script.sh

# gRPC API тесты
./docs/participation-and-roles/grpc-script.sh
```

## Тестовые данные

### Пользователи

| Username        | Role (initial) | Password       | Description                    |
|-----------------|----------------|----------------|--------------------------------|
| alice_owner     | OWNER          | SecurePass123  | Владелец хакатона              |
| bob_organizer   | ORGANIZER      | SecurePass123  | Организатор (удаляется в тестах)|
| charlie_mentor  | -              | SecurePass123  | Кандидат в MENTOR              |
| diana_judge     | -              | SecurePass123  | Кандидат в JUDGE               |
| eve_user        | -              | SecurePass123  | Обычный пользователь без ролей |

### Хакатон

- **ID**: `44444444-4444-4444-4444-444444444444`
- **Name**: AI Innovation Hackathon 2026
- **Stage**: published
- **Owner**: alice_owner

## Тестовые сценарии

### Happy Path (успешные сценарии)

1. ✅ **ListHackathonStaff** - Alice (OWNER) видит всех staff членов
2. ✅ **CreateStaffInvitation** - Alice приглашает Charlie как MENTOR
3. ✅ **ListMyStaffInvitations** - Charlie видит своё приглашение
4. ✅ **AcceptStaffInvitation** - Charlie принимает приглашение
5. ✅ **ListHackathonStaff** - Charlie теперь в списке staff с ролью MENTOR
6. ✅ **CreateStaffInvitation** - Alice приглашает Diana как JUDGE
7. ✅ **RejectStaffInvitation** - Diana отклоняет приглашение
8. ✅ **CreateStaffInvitation** - Alice снова приглашает Diana
9. ✅ **CancelStaffInvitation** - Alice отменяет приглашение
10. ✅ **RemoveHackathonRole** - Alice удаляет роль ORGANIZER у Bob
11. ✅ **SelfRemoveHackathonRole** - Charlie самоудаляется как MENTOR
12. ✅ **ListHackathonStaff** - Только Alice (OWNER) остается

### Fail Cases (негативные сценарии)

1. ❌ **ListHackathonStaff** - Eve (не staff) не может просматривать staff → 403 Forbidden
2. ❌ **CreateStaffInvitation** - Bob (не OWNER) не может создавать приглашения → 403 Forbidden
3. ❌ **CreateStaffInvitation** - Попытка пригласить на роль OWNER → 400 InvalidArgument
4. ❌ **RemoveHackathonRole** - Попытка удалить роль OWNER → 403 Forbidden
5. ❌ **SelfRemoveHackathonRole** - Попытка самоудалить роль OWNER → 403 Forbidden

## Покрытие методов

### Public Methods

| Метод                       | Happy Path | Fail Cases | Status |
|-----------------------------|------------|------------|--------|
| ListHackathonStaff          | ✅         | ✅         | ✅     |
| CreateStaffInvitation       | ✅         | ✅         | ✅     |
| CancelStaffInvitation       | ✅         | ❌         | ✅     |
| ListMyStaffInvitations      | ✅         | ❌         | ✅     |
| AcceptStaffInvitation       | ✅         | ❌         | ✅     |
| RejectStaffInvitation       | ✅         | ❌         | ✅     |
| RemoveHackathonRole         | ✅         | ✅         | ✅     |
| SelfRemoveHackathonRole     | ✅         | ✅         | ✅     |

### Internal Methods

| Метод                  | Tested | Status |
|------------------------|--------|--------|
| GetHackathonContext    | ❌     | 🔄 Manual testing required |
| AssignHackathonRole    | ❌     | 🔄 Tested via HackathonService |

## Проверяемые бизнес-правила

### Staff Invitations

- ✅ Только OWNER может создавать приглашения
- ✅ Нельзя пригласить на роль OWNER
- ✅ Только получатель может принять/отклонить приглашение
- ✅ Только OWNER может отменить pending приглашение
- ✅ После принятия приглашения роль назначается автоматически

### Role Management

- ✅ OWNER уникален на хакатон
- ✅ OWNER нельзя удалить через RemoveHackathonRole
- ✅ OWNER нельзя самоудалить через SelfRemoveHackathonRole
- ✅ Только OWNER может удалять роли других пользователей
- ✅ Пользователь может самоудалиться (кроме OWNER)

### Access Control

- ✅ Только staff члены могут просматривать список staff
- ✅ Только OWNER может управлять приглашениями
- ✅ Любой залогиненный пользователь может видеть свои приглашения

## Идемпотентность

Все мутирующие операции поддерживают идемпотентность через `IdempotencyKey`:

- ✅ CreateStaffInvitation
- ✅ CancelStaffInvitation
- ✅ AcceptStaffInvitation
- ✅ RejectStaffInvitation
- ✅ RemoveHackathonRole
- ✅ SelfRemoveHackathonRole

## Транзакционность

### Методы с транзакциями (Unit of Work)

- ✅ **AcceptStaffInvitation**: атомарное обновление статуса приглашения и создание staff роли

### Методы без транзакций (одна операция)

- CreateStaffInvitation - одна INSERT операция
- CancelStaffInvitation - одна UPDATE операция
- RejectStaffInvitation - одна UPDATE операция
- RemoveHackathonRole - одна DELETE операция
- SelfRemoveHackathonRole - одна DELETE операция

## Ожидаемый output тестов

```
=== Participation and Roles Service [REST/gRPC] Testing ===

0. Loading test data into database...
✓ Test data loaded

1. Logging in test users...
✓ Alice logged in. Token: eyJhbGci...
✓ Bob logged in
✓ Charlie logged in
✓ Diana logged in
✓ Eve logged in

2. ListHackathonStaff (Alice - owner, happy path)...
✓ Alice can view staff (found 2 members)

3. ListHackathonStaff (Eve - not staff, should fail)...
✓ Eve cannot view staff (forbidden)

[... остальные тесты ...]

19. SelfRemoveHackathonRole (Try to self-remove OWNER, should fail)...
✓ Cannot self-remove OWNER role (forbidden)

=== All Tests Completed Successfully ===
Summary:
  - ✓ ListHackathonStaff: access control works
  - ✓ CreateStaffInvitation: only OWNER can invite
  - ✓ ListMyStaffInvitations: users see their invitations
  - ✓ AcceptStaffInvitation: role assigned after acceptance
  - ✓ RejectStaffInvitation: invitation declined successfully
  - ✓ CancelStaffInvitation: OWNER can cancel pending invitations
  - ✓ RemoveHackathonRole: OWNER can remove roles (except OWNER)
  - ✓ SelfRemoveHackathonRole: users can leave (except OWNER)
  - ✓ All negative scenarios validated
✓ All tests passed!
```

## Troubleshooting

### Тесты не запускаются

```bash
# Проверить что скрипты исполняемые
chmod +x docs/participation-and-roles/*.sh

# Проверить что все сервисы запущены
docker-compose -f deployments/docker-compose.yml ps
```

### Ошибки в тестах

```bash
# Очистить БД и перезагрузить тестовые данные
docker-compose -f deployments/docker-compose.yml exec postgres \
  psql -U hackathon -d hackathon -c "
  DELETE FROM participation_and_roles.staff_invitations WHERE hackathon_id = '44444444-4444-4444-4444-444444444444';
  DELETE FROM participation_and_roles.staff_roles WHERE hackathon_id = '44444444-4444-4444-4444-444444444444';
  DELETE FROM hackaton.hackathons WHERE id = '44444444-4444-4444-4444-444444444444';
  DELETE FROM identity.users WHERE id IN (
    'a11ce000-0000-0000-0000-000000000001',
    'b0b00000-0000-0000-0000-000000000002',
    'c4a41e00-0000-0000-0000-000000000003',
    'd1a4a000-0000-0000-0000-000000000004',
    'e5e00000-0000-0000-0000-000000000005'
  );
"

# Перезагрузить тестовые данные
docker-compose -f deployments/docker-compose.yml exec -T postgres \
  psql -U hackathon -d hackathon < docs/participation-and-roles/test-data.sql

# Запустить тесты снова
./docs/participation-and-roles/rest-script.sh
```

### Сервис не отвечает

```bash
# Проверить логи
docker-compose -f deployments/docker-compose.yml logs participation-and-roles-service

# Перезапустить сервис
docker-compose -f deployments/docker-compose.yml restart participation-and-roles-service
```

## Ссылки

- [REST API Guide](./rest-guide.md)
- [gRPC API Guide](./grpc-guide.md)
- [README](./README.md)
- [Business Rules](../rules/hackathon-staff.md)
- [Docker Setup](../docker-setup.md)

