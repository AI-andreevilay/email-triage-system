# Email Triage System - Architecture

## 1. Overview

This system classifies Gmail emails and applies labels.
Current phase focuses on a minimal backend foundation for incremental delivery.

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
- No Kubernetes (initially)
- No multi-node Kubernetes cluster initially
- No Kubernetes operators initially
- No ArgoCD initially
- No service mesh
- No heavy observability stack at the beginning

---

## 4. High-Level Architecture

Main components:

- API Server
- Email Reader (mock or Gmail)
- Broker (RabbitMQ)
- Storage (PostgreSQL)
- Classifier
- Workers

Flow (MVP):

Client -> API -> Reader -> Broker (email.raw) -> Classifier Worker -> PostgreSQL -> Broker (email.classified) -> Label Worker -> PostgreSQL

Future flow:

Client -> API -> Reader -> Broker -> Classifier Worker -> PostgreSQL -> Label Worker -> Gmail

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
- Both default rules and user-defined rules are evaluated for each message
- Supported operators (MVP): `equals`, `contains`
- Rule selection:
  1. Match all applicable rules
  2. Score each match (`priority + source_bonus + specificity_bonus`)
  3. Highest score wins
  4. User rule wins when priority is equal
  5. More specific rule wins when priority and source are equal
  6. Fallback label is `Unknown` when no rule matches
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
- user_id
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

Current (Iteration 9):

1. User triggers scan
2. System fetches emails from configured reader source (mock or Gmail)
3. Gmail source is read page-by-page (batch size default: 100)
4. API publishes one `email.raw` event per message to RabbitMQ
5. Classifier worker consumes `email.raw`
6. Worker classifies using default + user rules
7. Worker stores metadata and classification result in PostgreSQL
8. For `apply` mode, classifier worker publishes `email.classified`
9. Label worker consumes `email.classified`
10. Label worker applies Gmail label via Gmail API, removes `INBOX`, and updates `applied_label`, `status=applied`

Future (event-driven):

1. Reader publishes events
2. Classifier worker consumes events
3. Results stored in PostgreSQL
4. Label worker applies labels

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

### Decision: Keep label applying mocked in Iteration 8
Reason:
- Introduce label stage and event contract before Gmail OAuth/API complexity
- Validate end-to-end apply flow with DB updates only

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

### Decision: Use k3s later for learning and production-like deployment
Reason:
- Kubernetes is a project learning goal
- Keep the current MVP simple first, then move to k3s when core flow is stable

### Decision: Use single-node Kubernetes initially
Reason:
- Target environment is one Hetzner VPS (for example CX33 class: 4 vCPU, 8 GB RAM)
- Lower operational complexity for a pet project while still practicing real deployment workflows

### Decision: Use one shared PostgreSQL instance for pet projects, separated by databases and users
Reason:
- Reuse one database server while isolating projects by credentials and schema ownership
- Keep this project isolated with dedicated database, dedicated DB user, and dedicated migrations

### Decision: Keep Kubernetes out of the first MVP implementation unless explicitly requested
Reason:
- First runnable baseline should stay focused on product behavior, not infra complexity

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

### 11.2 Initial MVP deployment
- First runnable baseline does not require Kubernetes.
- Prioritize correctness of the current event-driven MVP flow:
  Client -> API -> Reader -> Broker -> Classifier Worker -> PostgreSQL -> Label Worker -> Gmail API.

### 11.3 Future k3s deployment
- Deploy later to a single-node k3s cluster on one Hetzner VPS.
- Keep deployment assets in git (manifests/Helm charts when introduced).
- Publish app images to a registry such as GHCR.
- Manage secrets in a reproducible way from git-controlled inputs and deployment steps, not only as ad-hoc in-cluster manual state.

### 11.4 Namespace strategy
- Place this project in its own namespace.
- Other pet projects may share the same cluster but run in separate namespaces.

### 11.5 Shared PostgreSQL strategy
- A shared PostgreSQL instance can serve multiple pet projects.
- Isolate each project by separate database and separate DB user.
- For this project: dedicated database, dedicated DB user, dedicated migrations.
- Running PostgreSQL inside k3s as StatefulSet + PVC is acceptable for learning in this pet-project context.

### 11.6 Backup and restore principles
- Backups are mandatory if PostgreSQL runs inside the cluster.
- Define scheduled logical backups and keep copies outside the node.
- Keep a documented restore procedure and test restore periodically.
- Treat backup/restore procedure as part of the architecture, not as optional ops work.

---

## 12. Open Questions

- How to handle Gmail rate limits?
- How to improve classification accuracy?
- When to introduce LLM?

---

## 13. Future Improvements

- Add RabbitMQ for async processing
- Add Gmail API integration
- Add Kubernetes deployment
- Add observability (Prometheus)
- Add LLM-based classification
