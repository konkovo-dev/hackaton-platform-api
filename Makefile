.PHONY: help run-all build-all clean test lint deps

SERVICES := identity-service hackaton-service team-service submission-service participation-and-roles-service support-and-judging-service

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

gen:
	@echo "generating protobuf code"
	@buf generate

# docker commands
up:
	@echo "starting services with docker-compose"
	@docker-compose up -d

down:
	@echo "stopping services"
	@docker-compose down

logs:
	@docker-compose logs -f

restart:
	@echo "restarting services"
	@docker-compose restart

ps:
	@docker-compose ps

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

support-and-judging-service-%:
	@$(MAKE) -C cmd/support-and-judging-service $*

gateway-%:
	@$(MAKE) -C cmd/gateway $*

