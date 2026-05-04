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

---

## 4. High-Level Architecture

Main components:

- API Server
- Gmail Reader (mock for MVP)
- Classifier
- Storage (PostgreSQL)
- (Later) Broker (Redpanda)
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

### 5.2 Gmail Reader
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

### Decision: Mock Gmail reader before real integration
Reason:
- Validate pipeline incrementally without OAuth and external API dependencies

---

## 11. Open Questions

- How to handle Gmail rate limits?
- How to improve classification accuracy?
- When to introduce LLM?

---

## 12. Future Improvements

- Add Redpanda for async processing
- Add Gmail API integration
- Add Kubernetes deployment
- Add observability (Prometheus)
- Add LLM-based classification
