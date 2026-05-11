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
- Email Reader (mock for MVP)
- Classifier
- Storage (PostgreSQL)
- (Later) Broker (RabbitMQ)
- (Later) Workers

Flow (MVP):

Client -> API -> Reader -> Classifier -> DB

Future flow:

Client -> API -> Reader -> Broker -> Classifier Worker -> DB -> Label Worker -> Gmail

---

## 5. Components

### 5.1 API Server
- Handles HTTP requests
- Starts scan process
- Returns scan results

### 5.2 Email Reader
- Fetches emails (mock for MVP)
- Normalizes data

### 5.3 Classifier
- Rule-based classification
- Categories: Job, Transactions, Security, Promo, Social, Unknown

### 5.4 Storage
- PostgreSQL
- Stores email metadata and classification results
- Uses SQL migrations for schema management

### 5.5 Workers (future)
- Classifier worker
- Label applier worker

---

## 6. Data Model

### email_messages
- gmail_message_id
- user_id
- predicted_label
- applied_label
- confidence
- status

### scan_runs
- id
- user_id
- mode (dry_run / apply)
- status

### user_rules
- rule_type
- rule_value
- target_label

---

## 7. Data Flow

MVP:

1. User triggers scan
2. System fetches emails from reader
3. Emails are classified
4. Results are stored in DB
5. Labels are applied (only in apply mode)

Future (event-driven):

1. Reader publishes events
2. Classifier worker consumes events
3. Results stored in DB
4. Label worker applies labels

---

## 8. Idempotency

- Unique constraint: user_id + gmail_message_id
- Reprocessing same email should not create duplicates
- Safe to rerun scans

---

## 9. Privacy

- Email content is NOT stored
- Classifier uses content in memory only
- Only metadata is persisted

---

## 10. Key Decisions

### Decision: No broker in MVP
Reason:
- Reduce complexity
- Focus on core logic first

### Decision: Start with single process API foundation
Reason:
- Establish runnable baseline before storage and scan logic

### Decision: SQL-first storage in MVP
Reason:
- Keep persistence explicit and simple
- Enforce idempotency at DB level with unique constraint

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

---

## 11. Deployment Strategy

### 11.1 Local development
- Use Docker Compose for local dependencies.
- Keep local workflow fast: run API + PostgreSQL, apply migrations, test scan/classification flow.

### 11.2 Initial MVP deployment
- First runnable baseline does not require Kubernetes.
- Prioritize correctness of the synchronous MVP flow:
  Client -> API -> Reader -> Classifier -> PostgreSQL.

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
