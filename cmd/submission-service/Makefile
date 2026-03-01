SERVICE_NAME := submission-service
SERVICE_PORT := 50058
BINARY_NAME := submission-service

ROOT_DIR := $(shell pwd)/../..
BIN_DIR := $(ROOT_DIR)/bin
CMD_DIR := $(ROOT_DIR)/cmd/$(SERVICE_NAME)

.PHONY: help run build test clean

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run:
	@echo "starting $(SERVICE_NAME) on port $(SERVICE_PORT)"
	@ENV=local \
		SUBMISSION_SERVICE_GRPC_PORT=$(SERVICE_PORT) \
		DB_DSN="postgres://hackathon:hackathon_dev_password@localhost:5432/hackathon_submission?sslmode=disable" \
		S3_ENDPOINT=localhost:9000 \
		S3_ACCESS_KEY=minioadmin \
		S3_SECRET_KEY=minioadmin \
		S3_BUCKET_NAME=submissions \
		S3_USE_SSL=false \
		S3_PRESIGNED_URL_EXPIRY=15m \
		HACKATHON_SERVICE_URL=localhost:50052 \
		PARTICIPATION_ROLES_SERVICE_URL=localhost:50055 \
		TEAM_SERVICE_URL=localhost:50053 \
		SERVICE_AUTH_TOKEN=dev-service-token-change-in-production \
		IDEMPOTENCY_TTL=24h \
		MAX_FILE_SIZE_BYTES=52428800 \
		MAX_TOTAL_SIZE_BYTES=209715200 \
		MAX_FILES_PER_SUBMISSION=20 \
		MAX_SUBMISSIONS_PER_OWNER=50 \
		FILE_UPLOAD_TIMEOUT=3m \
		go run $(CMD_DIR)/main.go

build:
	@echo "building $(SERVICE_NAME)"
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go
	@echo "binary: $(BIN_DIR)/$(BINARY_NAME)"

test:
	@echo "running tests for $(SERVICE_NAME)"
	@go test -v -race -cover $(ROOT_DIR)/internal/$(SERVICE_NAME)/...

clean:
	@echo "cleaning $(SERVICE_NAME) binary"
	@rm -f $(BIN_DIR)/$(BINARY_NAME)

lint:
	@echo "linting $(SERVICE_NAME)"
	@go vet $(ROOT_DIR)/internal/$(SERVICE_NAME)/...
	@go vet $(CMD_DIR)/...

healthcheck:
	@grpcurl -plaintext localhost:$(SERVICE_PORT) grpc.health.v1.Health/Check || echo "service not running"

list-services:
	@grpcurl -plaintext localhost:$(SERVICE_PORT) list

migrate-up:
	@echo "running migrations for $(SERVICE_NAME)"
	@goose -table goose_db_version_submission -dir $(ROOT_DIR)/internal/$(SERVICE_NAME)/migrations postgres "$(DB_DSN)" up

migrate-down:
	@echo "rolling back migrations for $(SERVICE_NAME)"
	@goose -table goose_db_version_submission -dir $(ROOT_DIR)/internal/$(SERVICE_NAME)/migrations postgres "$(DB_DSN)" down

migrate-status:
	@echo "checking migration status for $(SERVICE_NAME)"
	@goose -table goose_db_version_submission -dir $(ROOT_DIR)/internal/$(SERVICE_NAME)/migrations postgres "$(DB_DSN)" status

migrate-create:
	@echo "creating new migration: $(NAME)"
	@goose -dir $(ROOT_DIR)/internal/$(SERVICE_NAME)/migrations create $(NAME) sql

sqlc-generate:
	@echo "generating sqlc code for $(SERVICE_NAME)"
	@cd $(ROOT_DIR)/internal/$(SERVICE_NAME) && sqlc generate

DB_DSN ?= postgres://hackathon:hackathon_dev_password@localhost:5432/hackathon_submission?sslmode=disable
