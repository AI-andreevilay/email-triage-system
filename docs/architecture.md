# Email Triage System - Architecture

## 1. Overview

This system classifies Gmail emails and applies labels.
The current implementation focuses on a local, event-driven backend with an explainable rule engine.

---

## 2. Goals

- Automatically classify emails
- Support dry-run and apply modes
- Ensure idempotent processing
- Keep user data private (no email content stored)

---

## 3. Non-Goals (MVP)

- No authentication
- No multi-user support
- No UI
- No LLM-based classification
- No production deployment setup in this public repository
- No heavy observability stack

---

## 4. High-Level Architecture

Main components:

- API Server
- Email Reader (mock or Gmail)
- Broker (RabbitMQ)
- Storage (PostgreSQL)
- Classifier
- Workers

Current flow:

Client -> API -> Reader -> Broker (email.raw) -> Classifier Worker -> PostgreSQL -> Broker (email.classified) -> Label Worker -> PostgreSQL

Future flow:

Client -> API -> Reader -> Broker -> Classifier Worker -> PostgreSQL -> Label Worker -> Gmail -> Log Pipeline -> AI Log Analytics

---

## 5. Components

### 5.1 API Server
- Handles HTTP requests
- Starts scan process
- Publishes `email.raw` events to broker

### 5.2 Email Reader
- Fetches emails from:
  - mock source (default)
  - Gmail API source (optional, OAuth token based)
- For Gmail source, scan uses paginated fetch in batches (`maxResults` per page)
- Normalizes data

### 5.3 Broker (RabbitMQ)
- Receives raw email events from API
- Topic exchange: `email.events`
- Routing key and queue: `email.raw`
- Routing key and queue: `email.classified`

### 5.4 Classifier
- Explainable rule-based classification (MVP)
- Categories: Job, Transactions, Security, Promo, Social, Unknown
- Global rules and user-specific rules are loaded from PostgreSQL for each message
- Supported operators (MVP): `equals`, `contains`
- Rule selection:
  1. Match all applicable user-specific rules for the current user
  2. If any user-specific rule matches, choose the best one by priority and specificity
  3. Otherwise, match all applicable global rules
  4. Choose the best global rule by priority and specificity
  5. Fallback label is `Unknown` when no rule matches
- Custom rule match inputs:
  - sender email
  - sender domain
  - subject keywords
  - optional body keywords (in memory only)
- Example rules:
  - sender_domain equals linkedin.com -> Job
  - subject contains interview -> Job
  - sender_email equals no-reply@accounts.google.com -> Security
  - subject contains receipt -> Transactions
- Each classification stores a short reason for dry-run review and debugging
- Body keywords can be checked in memory during classification, but body content is never persisted

### 5.5 Storage
- PostgreSQL
- Stores email metadata and classification results
- Uses SQL migrations for schema management

### 5.6 Workers
- Classifier worker (current)
- Label applier worker (current, Gmail API apply)

### 5.7 AI Log Analytics (future)
- Consumes structured operational events from API/workers (errors, retries, latency, label-apply failures)
- Runs anomaly detection and trend analysis for system health and delivery quality
- Produces actionable insights:
  - failure clusters (for example by sender domain or Gmail error code)
  - retry hot spots and queue lag warnings
  - candidate rule suggestions for recurring unknown or misclassified messages
- Stores summarized insights only (no raw email body)

---

## 6. Data Model

### email_messages
- id
- gmail_message_id
- user_id
- predicted_label
- applied_label
- confidence
- reason
- status
- processed_at
- created_at
- unique(user_id, gmail_message_id)

### scan_runs
- id
- user_id
- mode (dry_run / apply)
- status
- started_at
- finished_at
- total_found
- total_processed
- total_failed

### user_rules
- id
- user_id (`NULL` for global rules)
- rule_type
- operator
- rule_value
- target_label
- enabled
- priority
- created_at
- updated_at

---

## 7. Data Flow

Current:

1. User triggers scan
2. System fetches emails from configured reader source (mock or Gmail)
3. Gmail source is read page-by-page (batch size default: 100)
4. API publishes one `email.raw` event per message to RabbitMQ
5. Classifier worker consumes `email.raw`
6. Worker classifies using global + user-specific rules
7. Worker stores metadata and classification result in PostgreSQL
8. For `apply` mode, classifier worker publishes `email.classified`
9. Label worker consumes `email.classified`
10. Label worker applies Gmail label via Gmail API, removes `INBOX`, and updates `applied_label`, `status=applied`

Possible future analytics flow:

1. Reader publishes events
2. Classifier worker consumes events
3. Results stored in PostgreSQL
4. Label worker applies labels
5. API/workers publish structured operational logs/events
6. AI analytics job/worker processes logs and writes insight summaries

---

## 8. Idempotency

- Unique constraint: user_id + gmail_message_id
- Reprocessing same email should not create duplicates
- Safe to rerun scans

---

## 9. Privacy

- Email body may be used for in-memory rule matching only (for optional body keyword checks)
- Email body is NOT stored
- Only metadata and classification results are persisted

---

## 10. Key Decisions

### Decision: Introduce RabbitMQ before workers
Reason:
- Move from synchronous scan flow to event flow incrementally
- Keep classifier and label applying in dedicated worker iterations

### Decision: Persist classification in classifier worker
Reason:
- Keep API focused on request orchestration and event publishing
- Centralize idempotent write path in one consumer

### Decision: Introduce real Gmail reader before real label apply
Reason:
- Validate Gmail OAuth and mailbox read path in dry-run mode first
- Reduce integration risk by changing one external dependency at a time

### Decision: Apply labels in dedicated label worker through Gmail API
Reason:
- Preserve async separation between classification and side effects
- Keep Gmail API failures isolated from classifier path

### Decision: Use paginated Gmail scan batches by default
Reason:
- Supports iterative full Inbox scan without loading everything in memory
- Keeps API calls and queue publishing bounded per page

### Decision: Start with single process API foundation
Reason:
- Establish runnable baseline before storage and scan logic

### Decision: SQL-first storage in MVP
Reason:
- Keep persistence explicit and simple
- Enforce idempotency at DB level with unique constraint

### Decision: Use explainable rule-based classifier in MVP
Reason:
- Deterministic behavior is easier to validate in a local MVP
- Classification reason is visible for dry-run review and debugging
- Supports custom user rules without introducing LLM complexity yet

### Decision: Evaluate custom user rules before any future LLM classifier
Reason:
- User intent should override generic model behavior when explicit rules exist
- Keeps classification predictable and easy to troubleshoot

### Decision: Update initial migration directly while still local-only
Reason:
- No production/staging database exists yet
- Faster iteration during early schema shaping

### Decision: Mock email reader before real integration
Reason:
- Validate pipeline incrementally without OAuth and external API dependencies

### Decision: Do not introduce dedicated subagent config files in early MVP
Reason:
- Current scope is small enough for a single-agent workflow with iterative changes
- Avoid extra maintenance overhead from agent-role config before complexity justifies it
- Revisit when changes regularly span multiple modules and require repeatable specialized review roles (for example code-reviewer, migration-reviewer)

---

## 11. Deployment Strategy

### 11.1 Local development
- Use Docker Compose for local dependencies.
- Keep local workflow fast: run API + PostgreSQL, apply migrations, test scan/classification flow.

### 11.2 Production-like deployment
- Deployment manifests are not part of this public repository.
- If this project is deployed, runtime credentials should be provided by the target platform's secret manager.
- PostgreSQL and RabbitMQ credentials should be scoped to this application.
- Backups are required before any real mailbox data is processed outside a local environment.

---

## 12. Open Questions

- How to handle Gmail rate limits?
- How to improve classification accuracy?
- When to introduce LLM?

---

## 13. Future Improvements

- Add observability (Prometheus)
- Add LLM-based classification
- Add a small rules management UI
- Add rule suggestion reports for recurring `Unknown` messages

## Ideas

This section captures ideas that are not in committed scope yet.

### 1 AI Log Analytics
- Status: `idea`
- Why: detect failure clusters and suggest rule improvements from operational events
- MVP: daily/weekly anomaly summary from structured logs only
- Risks: noisy signals without enough event volume
- Success criteria: useful, actionable insights with low false-positive rate

### 2 Fresh Email Digest (Rule-Based, non-AI)
- Status: `idea`
- Why: quick visibility into incoming mail without opening Gmail
- MVP:
  - Periodic Telegram digest
  - "You received X new emails in last N hours"
  - Breakdown by category (Job / Transactions / Security / Promo / Social / Unknown)
  - Optional top senders and failed-to-classify count (`Unknown`)
- Risks: digest noise and too frequent notifications
- Success criteria: user can understand mailbox changes in <30 seconds from one message

### 3 Fresh Email Digest (AI-Assisted)
- Status: `idea`
- Why: compress content of fresh emails into actionable summary
- MVP:
  - AI-generated short summary for new messages
  - Optional extracted action items and priority hints
  - Same Telegram delivery channel as rule-based digest
- Risks:
  - privacy constraints for email content processing
  - model cost and latency
  - hallucinated or overconfident summaries
- Success criteria: summaries save review time while staying factually reliable

### Suggested Rollout Order
1. Implement 2 first (predictable and low cost).
2. Add 3 behind a feature flag and explicit opt-in.
3. Keep fallback to 2 when AI is unavailable.
