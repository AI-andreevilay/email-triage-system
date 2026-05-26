SHELL := /bin/bash

COMPOSE_FILE := deployments/docker-compose.yml
INFRA_SERVICES := postgres rabbitmq
APP_SERVICES := migrator api-server classifier-worker label-worker

.DEFAULT_GOAL := help

.PHONY: install
install: ## install local Go dependencies
	go mod download

.PHONY: run-infra
run-infra: ## start local infrastructure (PostgreSQL + RabbitMQ)
	docker compose -f $(COMPOSE_FILE) up -d $(INFRA_SERVICES)

.PHONY: stop-infra
stop-infra: ## stop local infrastructure
	docker compose -f $(COMPOSE_FILE) down

.PHONY: stop
stop: ## stop local Docker Compose stack
	docker compose -f $(COMPOSE_FILE) down

.PHONY: logs
logs: ## follow local Docker Compose logs
	docker compose -f $(COMPOSE_FILE) logs -f

.PHONY: migrate
migrate: ## apply SQL migrations
	docker compose -f $(COMPOSE_FILE) run --rm migrator

.PHONY: run-api
run-api: ## run API server
	go run ./cmd/api-server

.PHONY: run-classifier-worker
run-classifier-worker: ## run classifier worker
	go run ./cmd/classifier-worker

.PHONY: run-label-worker
run-label-worker: ## run label worker
	go run ./cmd/label-worker

.PHONY: gmail-auth
gmail-auth: ## run Gmail OAuth flow and save token file
	go run ./cmd/gmail-auth

.PHONY: run
run: ## start full local Docker Compose stack
	docker compose -f $(COMPOSE_FILE) up --build $(APP_SERVICES)

.PHONY: test
test: ## run tests
	go test ./...

.PHONY: healthz
healthz: ## call health endpoint
	curl http://localhost:8080/healthz

.PHONY: scan-dry-run
scan-dry-run: ## trigger scan in dry_run mode
	curl -X POST http://localhost:8080/scans \
		-H "Content-Type: application/json" \
		-d '{"mode":"dry_run"}'

.PHONY: scan-apply
scan-apply: ## trigger scan in apply mode
	curl -X POST http://localhost:8080/scans \
		-H "Content-Type: application/json" \
		-d '{"mode":"apply"}'

.PHONY: help
help: ## show available targets
	@echo "Available targets:"
	@grep -E '^[a-zA-Z0-9_.-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2}'
