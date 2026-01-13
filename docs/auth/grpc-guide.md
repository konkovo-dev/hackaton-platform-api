# Auth Service gRPC Test Cases

## Предусловия
- запущен в докере с помощью docker-setup.md

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
}" localhost:50057 auth.v1.AuthService/IntrospectToken

# 3. Login
grpcurl -plaintext -d '{
  "email": "testuser3@example.com",
  "password": "SecurePass123"
}' localhost:50057 auth.v1.AuthService/Login

# 4. Refresh
grpcurl -plaintext -d "{
  \"refresh_token\": \"$REFRESH_TOKEN\"
}" localhost:50057 auth.v1.AuthService/Refresh

# 5. Logout
grpcurl -plaintext -d "{
  \"refresh_token\": \"$REFRESH_TOKEN\"
}" localhost:50057 auth.v1.AuthService/Logout
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
}' localhost:50057 auth.v1.AuthService/Register

# Повторный запрос с тем же ключом (вернёт кешированный ответ)
grpcurl -plaintext -d '{
  "username": "bob",
  "email": "bob@example.com",
  "password": "password123",
  "first_name": "Bob",
  "last_name": "Jones",
  "timezone": "UTC",
  "idempotency_key": {"key": "bob-reg-1"}
}' localhost:50057 auth.v1.AuthService/Register
```

### Негативные сценарии

```bash
# Невалидный токен
grpcurl -plaintext -d '{
  "access_token": "invalid.jwt.token"
}' localhost:50057 auth.v1.AuthService/IntrospectToken
# Ожидается: active=false

# Неверный пароль
grpcurl -plaintext -d '{
  "email": "alice@example.com",
  "password": "WrongPassword"
}' localhost:50057 auth.v1.AuthService/Login
# Ожидается: error "invalid credentials"

# Пустые поля
grpcurl -plaintext -d '{
  "username": "",
  "email": "test@test.com",
  "password": "pass"
}' localhost:50057 auth.v1.AuthService/Register
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

