# Submission Service

Микросервис для управления посылками (submissions) участников хакатона.

## Функциональность

### Основные операции
- **CreateSubmission** - создание новой посылки (только на этапе RUNNING)
- **UpdateSubmission** - обновление описания посылки (только создателем)
- **ListSubmissions** - список посылок с фильтрацией по владельцу
- **SelectFinalSubmission** - выбор финальной посылки (капитан команды или индивидуальный участник)
- **GetSubmission** - получение информации о посылке с файлами
- **GetFinalSubmission** - получение финальной посылки владельца

### Операции с файлами
- **CreateSubmissionUpload** - инициация загрузки файла (возвращает pre-signed PUT URL)
- **CompleteSubmissionUpload** - завершение загрузки с верификацией в S3
- **GetSubmissionFileDownloadURL** - получение ссылки для скачивания (pre-signed GET URL)

## Архитектура

### Слои
- **Domain** - сущности (Submission, SubmissionFile) и константы
- **Repository** - работа с PostgreSQL через SQLC
- **Client** - gRPC клиенты для hackathon, participation-roles, team сервисов
- **Policy** - политики авторизации для каждой операции
- **Usecase** - бизнес-логика с валидацией и транзакциями
- **Transport** - gRPC API (два сервиса: SubmissionService, SubmissionFilesService)

### Внешние зависимости
- **PostgreSQL** - хранение метаданных посылок и файлов
- **MinIO (S3)** - хранение файлов
- **Hackathon Service** - получение стадии хакатона
- **Participation-Roles Service** - проверка ролей и статуса участия
- **Team Service** - получение капитана и членов команды

## Лимиты (конфигурируемые)

- Максимальный размер файла: 50 MB
- Максимальный суммарный размер посылки: 200 MB
- Максимальное количество файлов в посылке: 20
- Максимальное количество посылок на владельца: 50
- Разрешенные форматы: `.pdf`, `.zip`, `.png`, `.jpg`, `.jpeg`, `.txt`, `.md`, `.csv`

## Авторизация

### Создание и изменение посылок
- Только активные участники (RUNNING stage)
- Только создатель может обновлять описание
- Только владелец (или капитан команды) может выбирать финальную посылку

### Чтение посылок и файлов
- Владелец посылки (пользователь или команда)
- Организаторы, менторы, судьи (staff)
- Доступно с этапа RUNNING и далее

## Особенности реализации

### Версионирование
- Участник может создать несколько посылок (версий)
- Только одна посылка может быть помечена как финальная (`is_final`)
- Первая созданная посылка автоматически становится финальной
- Финальную посылку можно менять через `SelectFinalSubmission`

### Работа с файлами
- Используются pre-signed URLs для прямой загрузки/скачивания клиентом
- Статусы файлов: `PENDING` → `COMPLETED` или `FAILED`
- Верификация загрузки через HEAD request к S3
- Автоматическая очистка: файлы в статусе `PENDING` старше 3 минут помечаются как `FAILED`

### Идемпотентность
- Поддержка idempotency keys для безопасных повторных запросов
- TTL ключей конфигурируется через `IDEMPOTENCY_TTL`

## База данных

### Таблицы
- `submission.submissions` - метаданные посылок
- `submission.submission_files` - метаданные файлов
- `submission.idempotency_keys` - ключи идемпотентности

### Индексы
- Уникальный индекс на финальную посылку по владельцу
- Индексы для быстрого поиска по hackathon_id + owner
- Индекс для cleanup job (pending файлы по created_at)

## Переменные окружения

### База данных
- `SUBMISSION_DB_HOST` - хост PostgreSQL (default: localhost)
- `SUBMISSION_DB_PORT` - порт PostgreSQL (default: 5432)
- `SUBMISSION_DB_USER` - пользователь БД (default: postgres)
- `SUBMISSION_DB_PASSWORD` - пароль БД (default: postgres)
- `SUBMISSION_DB_NAME` - имя БД (default: hackathon_submission)
- `SUBMISSION_DB_SSLMODE` - SSL режим (default: disable)
- `SUBMISSION_DB_MAX_CONNS` - максимум соединений (default: 25)
- `SUBMISSION_DB_MIN_CONNS` - минимум соединений (default: 5)

### S3/MinIO
- `S3_ENDPOINT` - адрес S3 (default: localhost:9000)
- `S3_REGION` - регион (default: us-east-1)
- `S3_ACCESS_KEY_ID` - access key (default: minioadmin)
- `S3_SECRET_ACCESS_KEY` - secret key (default: minioadmin)
- `S3_SUBMISSIONS_BUCKET` - имя bucket (default: submissions)
- `S3_USE_SSL` - использовать SSL (default: false)

### Лимиты
- `SUBMISSION_MAX_FILE_SIZE_BYTES` - макс размер файла (default: 52428800 = 50MB)
- `SUBMISSION_MAX_TOTAL_SIZE_BYTES` - макс суммарный размер (default: 209715200 = 200MB)
- `SUBMISSION_MAX_FILES_PER_SUBMISSION` - макс файлов (default: 20)
- `SUBMISSION_MAX_SUBMISSIONS_PER_OWNER` - макс посылок (default: 50)
- `SUBMISSION_PRESIGNED_URL_EXPIRY_MINS` - TTL pre-signed URLs (default: 15)
- `SUBMISSION_ALLOWED_EXTENSIONS` - разрешенные расширения (default: .pdf,.zip,.png,.jpg,.jpeg,.txt,.md,.csv)
- `SUBMISSION_ALLOWED_CONTENT_TYPES` - разрешенные MIME-типы

### Внешние сервисы
- `AUTH_SERVICE_URL` - адрес auth-service (default: localhost:50057)
- `HACKATON_SERVICE_URL` - адрес hackaton-service (default: localhost:50052)
- `PARTICIPATION_ROLES_SERVICE_URL` - адрес participation-roles-service (default: localhost:50055)
- `TEAM_SERVICE_URL` - адрес team-service (default: localhost:50053)
- `SERVICE_AUTH_TOKEN` - токен для межсервисного взаимодействия

### Общие
- `GRPC_PORT` - порт gRPC сервера (default: 50058)
- `IDEMPOTENCY_TTL` - TTL ключей идемпотентности (default: 24h)

## Запуск

### Локальная разработка

1. Запустить зависимости (PostgreSQL, MinIO):
```bash
docker-compose -f deployments/dev/docker-compose.yml up -d postgres minio
```

2. Применить миграции:
```bash
cd internal/submission-service
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5432/hackathon_submission?sslmode=disable" up
```

3. Сгенерировать код:
```bash
# В корне проекта
buf generate

# В директории submission-service
cd internal/submission-service
sqlc generate

# В корне проекта
go mod tidy
```

4. Запустить сервис:
```bash
go run cmd/submission-service/main.go
```

### Production

```bash
docker-compose -f deployments/prod/docker-compose.yml up -d submission-service
```

## Тестирование

```bash
go test ./tests/integration -run TestSubmission
```
