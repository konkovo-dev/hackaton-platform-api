# Auth Service - Docker Setup

### Клонирование

```bash
git clone <repo-url>
cd hackaton-platform-api
```

### Генерация proto

```bash
make buf-generate
```

### Генерация sqlc

```bash
make auth-service-sqlc-generate
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
make auth-service-migrate-up
make auth-service-migrate-status
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
