# Email Triage System

Backend pet project for automatic Gmail email triage and labeling.

## Current Scope (Iteration 10)

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
  - For Gmail source, scan is paginated and processed in batches
- RabbitMQ in local Docker Compose
- Classifier worker consumes `email.raw`, classifies emails, and stores metadata in PostgreSQL
- Classifier worker publishes `email.classified` for apply mode
- Label worker consumes `email.classified` and applies labels via Gmail API
- Label worker removes message from Inbox after successful apply
- Label worker updates `applied_label` and `status=applied` in PostgreSQL on success
- Real Gmail reader (optional via config) for scan source
- OAuth CLI command to connect your Gmail account and save token
- Docker image build via root `Dockerfile`
- Kubernetes manifests live in private `k3s-deploy` repo (see `../k3s-deploy/k3s/email-triage-system` in this workspace)
- Kubernetes migrator `Job` for applying SQL migrations in cluster
- Label worker deployment is included, but scaled to `0` by default
- Infra services (`postgres`, `rabbitmq`) run in dedicated namespace `infra`
- App uses project-scoped infra credentials from Kubernetes `Secret` (`email-triage-secrets`)

## Tech Stack

- Go
- PostgreSQL
- RabbitMQ
- Docker Compose
- Kubernetes

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

Client -> API -> Reader -> RabbitMQ (`email.raw`) -> Classifier Worker -> PostgreSQL -> RabbitMQ (`email.classified`) -> Label Worker (Gmail apply) -> PostgreSQL

Detailed notes: `docs/architecture.md`.

## Run on Kubernetes (Iteration 10)

Kubernetes manifests were moved to the private `k3s-deploy` repo: see `../k3s-deploy/README.md`.

## Gmail Connection (for real reader)

1. In Google Cloud Console:
   - Enable Gmail API
   - Create OAuth Client ID (`Desktop app`)
   - Download credentials JSON
2. Save it as:
   ```bash
   secrets/gmail_credentials.json
   ```
3. Run OAuth flow:
   ```bash
   make gmail-auth
   ```
   Command starts local callback on `http://localhost:8090/oauth2/callback`.
   After Google consent, token is saved automatically.
4. Start API with Gmail source:
   ```bash
    EMAIL_SOURCE=gmail \
    GMAIL_CREDENTIALS_FILE=secrets/gmail_credentials.json \
    GMAIL_TOKEN_FILE=secrets/gmail_token.json \
    GMAIL_READ_MAX_RESULTS=100 \
    GMAIL_READ_QUERY='in:inbox -in:trash' \
    go run ./cmd/api-server
    ```
5. Trigger dry-run scan:
   ```bash
   curl -X POST http://localhost:8080/scans \
     -H "Content-Type: application/json" \
     -d '{"mode":"dry_run"}'
   ```
