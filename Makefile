.PHONY: help run-all build-all clean test lint deps

SERVICES := identity-service hackaton-service team-service submission-service participation-and-roles-service mentors-service matchmaking-service auth-service

help:
	@echo "Hackathon Platform API - Monorepo"
	@echo ""
	@echo "Global commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -v "Services:" | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Service commands (use: make <service>-<command>):"
	@echo "  identity-service-run        run identity service"
	@echo "  identity-service-build      build identity binary"
	@echo "  identity-service-test       test identity"
	@echo "  identity-service-health     health check"
	@echo ""
	@echo "  hackaton-service-run        run hackaton service"
	@echo "  hackaton-service-build      build hackaton binary"
	@echo "  ..."
	@echo ""
	@echo "use 'make -C cmd/identity-service help' for details"

run-all:
	@echo "run each service in separate terminal:"
	@echo "  terminal 1: make identity-service-run"
	@echo "  terminal 2: make hackaton-service-run"
	@echo "  ..."

build-all:
	@echo "building all services"
	@for service in $(SERVICES); do \
		if [ -f cmd/$$service/Makefile ]; then \
			echo "building $$service"; \
			$(MAKE) -C cmd/$$service build; \
		fi; \
	done
	@echo "done"

clean:
	@echo "cleaning all binaries"
	@rm -rf bin/
	@echo "done"

test:
	@echo "running tests"
	@go test -v -race -cover ./...

lint:
	@echo "linting"
	@go vet ./...
	@go fmt ./...

deps:
	@echo "installing dependencies"
	@go mod download
	@go mod tidy

buf-generate:
	@echo "generating protobuf code"
	@buf generate

openapiv2-install:
	@echo "installing protoc-gen-openapiv2"
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

sqlc-install:
	@echo "installing sqlc"
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

goose-install:
	@echo "installing goose"
	@go install github.com/pressly/goose/v3/cmd/goose@latest

# docker
up:
	@echo "starting services with docker-compose"
	@docker-compose -f deployments/docker-compose.yml up -d

down:
	@echo "stopping services"
	@docker-compose -f deployments/docker-compose.yml down

logs:
	@docker-compose -f deployments/docker-compose.yml logs -f

restart:
	@echo "restarting services"
	@docker-compose -f deployments/docker-compose.yml restart

ps:
	@docker-compose -f deployments/docker-compose.yml ps

identity-service-%:
	@$(MAKE) -C cmd/identity-service $*

hackaton-service-%:
	@$(MAKE) -C cmd/hackaton-service $*

team-service-%:
	@$(MAKE) -C cmd/team-service $*

submission-service-%:
	@$(MAKE) -C cmd/submission-service $*

participation-and-roles-service-%:
	@$(MAKE) -C cmd/participation-and-roles-service $*

mentors-service-%:
	@$(MAKE) -C cmd/mentors-service $*

matchmaking-service-%:
	@$(MAKE) -C cmd/matchmaking-service $*

auth-service-%:
	@$(MAKE) -C cmd/auth-service $*

gateway-%:
	@$(MAKE) -C cmd/gateway $*
