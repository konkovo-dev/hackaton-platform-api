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

# Team DB
echo "→ Migrating hackathon_team..."
GOOSE_DRIVER=postgres \
GOOSE_DBSTRING="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:5432/hackathon_team?sslmode=disable" \
goose -dir /migrations/team up

# Matchmaking DB
echo "→ Migrating hackathon_matchmaking..."
GOOSE_DRIVER=postgres \
GOOSE_DBSTRING="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:5432/hackathon_matchmaking?sslmode=disable" \
goose -dir /migrations/matchmaking up

echo "✓ All migrations completed successfully."

# Mentors DB
echo "→ Migrating hackathon_mentors..."
GOOSE_DRIVER=postgres \
GOOSE_DBSTRING="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:5432/hackathon_mentors?sslmode=disable" \
goose -dir /migrations/mentors up

echo "✓ All migrations completed successfully."

# Submission DB
echo "→ Migrating hackathon_submission..."
GOOSE_DRIVER=postgres \
GOOSE_DBSTRING="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:5432/hackathon_submission?sslmode=disable" \
goose -dir /migrations/submission up

echo "✓ All migrations completed successfully."
