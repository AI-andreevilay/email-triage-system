# Email Triage System

Backend pet project for automatic Gmail email triage and labeling.

## Current Scope (Iteration 6)

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
- User rules support in classifier (`rule_type` + `operator` + priority-based scoring)
- Manual full scan endpoint: `POST /scans`
  - Publishes raw email events to RabbitMQ queue `email.raw`
- RabbitMQ in local Docker Compose

## Tech Stack

- Go
- PostgreSQL
- RabbitMQ
- Docker Compose

## Run Locally

1. Show available commands:
   ```bash
   make help
   ```
2. Install local dependencies:
   ```bash
   make install
   ```
3. Start PostgreSQL:
   ```bash
   make run-infra
   ```
4. Apply migrations:
   ```bash
   make migrate
   ```
5. Run API server:
   ```bash
   make run-api
   ```
6. Check health:
   ```bash
   make healthz
   ```
7. Trigger scan:
   ```bash
   make scan-dry-run
   ```

One-command flow:

```bash
make run
```

## User Rules (MVP)

`user_rules` fields used by classifier:

- `rule_type`: `sender_email` | `sender_domain` | `subject` | `body` | `any`
- `operator`: `equals` | `contains`
- `rule_value`: value to match
- `target_label`: `Job` | `Transactions` | `Security` | `Promo` | `Social` | `Unknown`
- `priority`: higher value = higher precedence
- `enabled`: `true`/`false`

Quick SQL examples:

```bash
psql "postgres://postgres:postgres@localhost:5432/email_triage?sslmode=disable" -c "
INSERT INTO user_rules (user_id, rule_type, operator, rule_value, target_label, enabled, priority)
VALUES
  ('user_1','sender_domain','equals','linkedin.com','Job',true,300),
  ('user_1','subject','contains','receipt','Transactions',true,250),
  ('user_1','sender_email','equals','no-reply@accounts.google.com','Security',true,350);
"
```

Then trigger scan again:

```bash
curl -X POST http://localhost:8080/scans \
  -H "Content-Type: application/json" \
  -d '{"mode":"dry_run"}'
```

## Architecture (Current)

Client -> API -> Reader -> RabbitMQ (`email.raw`)

Detailed notes: `docs/architecture.md`.
