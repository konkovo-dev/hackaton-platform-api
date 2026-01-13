SERVICE_NAME := identity-service
SERVICE_PORT := 50051
BINARY_NAME := identity-service

ROOT_DIR := $(shell pwd)/../..
BIN_DIR := $(ROOT_DIR)/bin
CMD_DIR := $(ROOT_DIR)/cmd/$(SERVICE_NAME)

.PHONY: help run build test clean

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run:
	@echo "starting $(SERVICE_NAME) on port $(SERVICE_PORT)"
	@ENV=local IDENTITY_SERVICE_GRPC_PORT=$(SERVICE_PORT) go run $(CMD_DIR)/main.go

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
	@goose -dir $(ROOT_DIR)/internal/$(SERVICE_NAME)/migrations postgres "$(DB_DSN)" up

migrate-down:
	@echo "rolling back migrations for $(SERVICE_NAME)"
	@goose -dir $(ROOT_DIR)/internal/$(SERVICE_NAME)/migrations postgres "$(DB_DSN)" down

migrate-status:
	@echo "checking migration status for $(SERVICE_NAME)"
	@goose -dir $(ROOT_DIR)/internal/$(SERVICE_NAME)/migrations postgres "$(DB_DSN)" status

migrate-create:
	@echo "creating new migration: $(NAME)"
	@goose -dir $(ROOT_DIR)/internal/$(SERVICE_NAME)/migrations create $(NAME) sql

sqlc-generate:
	@echo "generating sqlc code for $(SERVICE_NAME)"
	@cd $(ROOT_DIR)/internal/$(SERVICE_NAME) && sqlc generate
