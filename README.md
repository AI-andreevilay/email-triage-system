# Email Triage System

Backend pet project for automatic Gmail email triage and labeling.

## Current Scope (Iteration 4)

- Project skeleton for API/workers/migrator
- PostgreSQL in Docker Compose
- Environment-based config loading
- Healthcheck endpoint
- SQL migrations
- Minimal PostgreSQL storage layer
- Mock Gmail reader
- Rule-based classifier with categories:
  - Job
  - Transactions
  - Security
  - Promo
  - Social
  - Unknown
- User rules support in classifier (priority + rule type match)

## Tech Stack

- Go
- PostgreSQL
- Docker Compose

## Run Locally

1. Start PostgreSQL:
   ```bash
   docker compose -f deployments/docker-compose.yml up -d postgres
   ```
2. Apply migrations:
   ```bash
   go run ./cmd/migrator
   ```
3. Run API server:
   ```bash
   go run ./cmd/api-server
   ```
4. Check health:
   ```bash
   curl http://localhost:8080/healthz
   ```

## Architecture (MVP Direction)

Client -> API -> Reader -> Classifier -> DB

Detailed notes: `docs/architecture.md`.
