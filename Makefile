.PHONY: help run-all build-all clean test lint deps

SERVICES := identity core teams submission mentorship jury

help:
	@echo "Hackathon Platform API - Monorepo"
	@echo ""
	@echo "Global commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -v "Services:" | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Service commands (use: make <service>-<command>):"
	@echo "  identity-run        run identity service"
	@echo "  identity-build      build identity binary"
	@echo "  identity-test       test identity"
	@echo "  identity-health     health check"
	@echo ""
	@echo "  core-run           run core service"
	@echo "  core-build         build core binary"
	@echo "  ..."
	@echo ""
	@echo "use 'make -C cmd/identity help' for details"

run-all:
	@echo "run each service in separate terminal:"
	@echo "  terminal 1: make identity-run"
	@echo "  terminal 2: make core-run"
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

identity-%:
	@$(MAKE) -C cmd/identity $*

core-%:
	@$(MAKE) -C cmd/core $*

teams-%:
	@$(MAKE) -C cmd/teams $*

submission-%:
	@$(MAKE) -C cmd/submission $*

mentorship-%:
	@$(MAKE) -C cmd/mentorship $*

jury-%:
	@$(MAKE) -C cmd/jury $*

gateway-%:
	@$(MAKE) -C cmd/gateway $*

