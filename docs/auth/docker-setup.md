# Auth Service - Docker Setup

---

## Вариант 1: Прямые команды

### Клонирование

```bash
git clone <repo-url>
cd hackaton-platform-api
```

### Генерация proto

```bash
buf generate
```

### Генерация RSA ключей

```bash
mkdir -p .keys
openssl genrsa -out .keys/private_key.pem 2048
openssl rsa -in .keys/private_key.pem -pubout -out .keys/public_key.pem
```

### Запуск Postgres

```bash
docker-compose -f deployments/docker-compose.yml up -d postgres
docker-compose -f deployments/docker-compose.yml ps postgres
```

### Миграции

```bash
export DB_DSN="postgres://hackathon:hackathon_dev_password@localhost:5432/hackathon?sslmode=disable"
go install github.com/pressly/goose/v3/cmd/goose@latest
goose -dir internal/auth-service/migrations postgres "$DB_DSN" up
goose -dir internal/auth-service/migrations postgres "$DB_DSN" status
```

### Запуск auth-service

```bash
docker-compose -f deployments/docker-compose.yml build auth-service
docker-compose -f deployments/docker-compose.yml up -d auth-service
docker-compose -f deployments/docker-compose.yml logs --tail=20 auth-service
```

### Запуск gateway

```bash
docker-compose -f deployments/docker-compose.yml build gateway
docker-compose -f deployments/docker-compose.yml up -d --no-deps gateway
docker-compose -f deployments/docker-compose.yml logs --tail=20 gateway
```

### Проверка

```bash
docker-compose -f deployments/docker-compose.yml ps
```

---

## Вариант 2: Makefile

### Клонирование

```bash
git clone <repo-url>
cd hackaton-platform-api
```

### Генерация proto

```bash
make buf-generate
```

### Генерация RSA ключей

```bash
make auth-service-generate-keys
```

### Запуск Postgres

```bash
docker-compose -f deployments/docker-compose.yml up -d postgres
make ps
```

### Миграции

```bash
export DB_DSN="postgres://hackathon:hackathon_dev_password@localhost:5432/hackathon?sslmode=disable"
make goose-install
make migrate-up
make migrate-status
```

### Запуск auth-service

```bash
docker-compose -f deployments/docker-compose.yml build auth-service
docker-compose -f deployments/docker-compose.yml up -d auth-service
docker-compose -f deployments/docker-compose.yml logs --tail=20 auth-service
```

### Запуск gateway

```bash
docker-compose -f deployments/docker-compose.yml build gateway
docker-compose -f deployments/docker-compose.yml up -d --no-deps gateway
docker-compose -f deployments/docker-compose.yml logs --tail=20 gateway
```

### Проверка

```bash
make ps
```
