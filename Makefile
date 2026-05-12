SHELL := /bin/bash

COMPOSE_FILE := deployments/docker-compose.yml
PG_SERVICE := postgres

.DEFAULT_GOAL := help

.PHONY: install
install: ## install local Go dependencies
	go mod download

.PHONY: run-infra
run-infra: ## start local infrastructure (PostgreSQL)
	docker compose -f $(COMPOSE_FILE) up -d $(PG_SERVICE)

.PHONY: stop-infra
stop-infra: ## stop local infrastructure
	docker compose -f $(COMPOSE_FILE) down

.PHONY: migrate
migrate: ## apply SQL migrations
	go run ./cmd/migrator

.PHONY: run-api
run-api: ## run API server
	go run ./cmd/api-server

.PHONY: run-classifier-worker
run-classifier-worker: ## run classifier worker
	go run ./cmd/classifier-worker

.PHONY: run-label-worker
run-label-worker: ## run label worker
	go run ./cmd/label-worker

.PHONY: run
run: run-infra migrate ## start infra, apply migrations and run API server
	go run ./cmd/api-server

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
