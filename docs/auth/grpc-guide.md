# Auth Service Testing Guide

## Предварительные требования

1. **Docker & Docker Compose**
2. **grpcurl** для тестирования gRPC:
   ```bash
   # macOS
   brew install grpcurl
   
   # Linux
   go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
   ```
3. **goose** для миграций:
   ```bash
   make goose-install
   ```

## Локальное тестирование (без Docker)

### 1. Запустить Postgres

```bash
# Только postgres из docker-compose
docker-compose -f deployments/docker-compose.yml up postgres -d

# Проверить что запустился
docker-compose -f deployments/docker-compose.yml ps
```

### 2. Применить миграции

```bash
export DB_DSN="postgres://hackathon:hackathon_dev_password@localhost:5432/hackathon?sslmode=disable"
make migrate-up
make migrate-status
```

### 3. Запустить auth-service локально

```bash
# Из корня проекта
make auth-service-run

# Или из cmd/auth-service
cd cmd/auth-service
make run
```

Сервис запустится на `localhost:50051`

### 4. Тестировать

```bash
# Автоматический тест-скрипт
make test-auth-local

# Или вручную через grpcurl
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check

grpcurl -plaintext -d '{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123",
  "first_name": "Test",
  "last_name": "User",
  "timezone": "UTC",
  "idempotency_key": {"key": "test-1"}
}' localhost:50051 auth.v1.AuthService/Register
```

## Тестирование в Docker

### 1. Запустить всё через docker-compose

```bash
cd deployments
docker-compose up auth-service postgres -d

# Посмотреть логи
docker-compose logs -f auth-service
```

### 2. Дождаться готовности

```bash
# Проверить что сервис работает
docker-compose ps

# Посмотреть health check
grpcurl -plaintext localhost:50057 grpc.health.v1.Health/Check
```

### 3. Применить миграции (в контейнере)

```bash
docker-compose exec auth-service sh -c "cd /app && goose -dir migrations postgres \"postgres://hackathon:hackathon_dev_password@postgres:5432/hackathon?sslmode=disable\" up"
```

**Или** применить миграции с хоста:

```bash
export DB_DSN="postgres://hackathon:hackathon_dev_password@localhost:5432/hackathon?sslmode=disable"
make migrate-up
```

### 4. Тестировать

```bash
# Автоматический тест (порт 50057 для docker)
make test-auth-docker

# Или вручную
grpcurl -plaintext localhost:50057 auth.v1.AuthService/Register -d '{...}'
```

## Тест-сценарии

### Базовый флоу

```bash
# 1. Register
RESPONSE=$(grpcurl -plaintext -d '{
  "username": "alice",
  "email": "alice@example.com",
  "password": "SecurePass123",
  "first_name": "Alice",
  "last_name": "Smith",
  "timezone": "Europe/Moscow",
  "idempotency_key": {"key": "alice-registration"}
}' localhost:50051 auth.v1.AuthService/Register)

# Извлечь токены
ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.accessToken')
REFRESH_TOKEN=$(echo "$RESPONSE" | jq -r '.refreshToken')

# 2. IntrospectToken
grpcurl -plaintext -d "{
  \"access_token\": \"$ACCESS_TOKEN\"
}" localhost:50051 auth.v1.AuthService/IntrospectToken

# 3. Login
grpcurl -plaintext -d '{
  "email": "alice@example.com",
  "password": "SecurePass123"
}' localhost:50051 auth.v1.AuthService/Login

# 4. Refresh
grpcurl -plaintext -d "{
  \"refresh_token\": \"$REFRESH_TOKEN\"
}" localhost:50051 auth.v1.AuthService/Refresh

# 5. Logout
grpcurl -plaintext -d "{
  \"refresh_token\": \"$REFRESH_TOKEN\"
}" localhost:50051 auth.v1.AuthService/Logout
```

### Тест Idempotency

```bash
# Первый запрос
grpcurl -plaintext -d '{
  "username": "bob",
  "email": "bob@example.com",
  "password": "password123",
  "first_name": "Bob",
  "last_name": "Jones",
  "timezone": "UTC",
  "idempotency_key": {"key": "bob-reg-1"}
}' localhost:50051 auth.v1.AuthService/Register

# Повторный запрос с тем же ключом (вернёт кешированный ответ)
grpcurl -plaintext -d '{
  "username": "bob",
  "email": "bob@example.com",
  "password": "password123",
  "first_name": "Bob",
  "last_name": "Jones",
  "timezone": "UTC",
  "idempotency_key": {"key": "bob-reg-1"}
}' localhost:50051 auth.v1.AuthService/Register
```

### Негативные сценарии

```bash
# Невалидный токен
grpcurl -plaintext -d '{
  "access_token": "invalid.jwt.token"
}' localhost:50051 auth.v1.AuthService/IntrospectToken
# Ожидается: active=false

# Неверный пароль
grpcurl -plaintext -d '{
  "email": "alice@example.com",
  "password": "WrongPassword"
}' localhost:50051 auth.v1.AuthService/Login
# Ожидается: error "invalid credentials"

# Пустые поля
grpcurl -plaintext -d '{
  "username": "",
  "email": "test@test.com",
  "password": "pass"
}' localhost:50051 auth.v1.AuthService/Register
# Ожидается: error "username is required"
```

## Очистка

```bash
# Остановить сервисы
docker-compose down

# Удалить данные
docker-compose down -v

# Откатить миграции
export DB_DSN="postgres://hackathon:hackathon_dev_password@localhost:5432/hackathon?sslmode=disable"
make migrate-down
```

## Troubleshooting

### Ошибка "failed to read private key"

```bash
# Проверить что ключи существуют
ls -la .keys/

# Если нет - сгенерировать
cd cmd/auth-service && make generate-keys
```

### Ошибка "failed to ping db"

```bash
# Проверить что Postgres запущен
docker-compose ps postgres

# Проверить подключение
psql "postgres://hackathon:hackathon_dev_password@localhost:5432/hackathon"
```

### Ошибка "user already exists"

```bash
# Очистить данные и применить миграции заново
docker-compose down -v
docker-compose up postgres -d
sleep 5
make migrate-up
```

## Полезные команды

```bash
# Список методов сервиса
grpcurl -plaintext localhost:50051 list auth.v1.AuthService

# Описание метода
grpcurl -plaintext localhost:50051 describe auth.v1.AuthService.Register

# Логи сервиса
docker-compose logs -f auth-service

# Подключиться к БД
docker-compose exec postgres psql -U hackathon -d hackathon

# Посмотреть таблицы
docker-compose exec postgres psql -U hackathon -d hackathon -c "\dt auth.*"

# Посмотреть users
docker-compose exec postgres psql -U hackathon -d hackathon -c "SELECT id, username, email FROM auth.users;"
```

