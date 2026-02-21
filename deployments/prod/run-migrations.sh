#!/bin/bash
set -e

echo "Starting migrations..."

# Ждём готовности Postgres
until pg_isready -h "$POSTGRES_HOST" -U "$POSTGRES_USER"; do
  echo "Waiting for postgres..."
  sleep 2
done

echo "Postgres is ready. Running migrations..."

# Auth DB
echo "→ Migrating hackathon_auth..."
GOOSE_DRIVER=postgres \
GOOSE_DBSTRING="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:5432/hackathon_auth?sslmode=disable" \
goose -dir /migrations/auth up

# Identity DB
echo "→ Migrating hackathon_identity..."
GOOSE_DRIVER=postgres \
GOOSE_DBSTRING="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:5432/hackathon_identity?sslmode=disable" \
goose -dir /migrations/identity up

# Hackathon DB
echo "→ Migrating hackathon_hackaton..."
GOOSE_DRIVER=postgres \
GOOSE_DBSTRING="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:5432/hackathon_hackaton?sslmode=disable" \
goose -dir /migrations/hackaton up

# Participation DB
echo "→ Migrating hackathon_participation..."
GOOSE_DRIVER=postgres \
GOOSE_DBSTRING="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:5432/hackathon_participation?sslmode=disable" \
goose -dir /migrations/participation up

echo "✓ All migrations completed successfully."
