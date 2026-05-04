# Email Triage System

Backend pet project for automatic Gmail email triage and labeling.

## Current Scope (Iteration 1)

- Project skeleton for API/workers/migrator
- PostgreSQL in Docker Compose
- Environment-based config loading
- Healthcheck endpoint

## Tech Stack

- Go
- PostgreSQL
- Docker Compose

## Run Locally

1. Start PostgreSQL:
   ```bash
   docker compose -f deployments/docker-compose.yml up -d postgres
   ```
2. Run API server:
   ```bash
   go run ./cmd/api-server
   ```
3. Check health:
   ```bash
   curl http://localhost:8080/healthz
   ```

## Architecture (MVP Direction)

Client -> API -> Reader -> Classifier -> DB

Detailed notes: `docs/architecture.md`.
