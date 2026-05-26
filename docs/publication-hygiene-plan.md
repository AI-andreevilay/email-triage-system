# Publication Hygiene Plan

Goal: prepare Email Triage System for opening the GitHub repository as a portfolio project.

Current iteration: complete

## Iterations

### 1. Secret Safety - Complete

- [x] Confirm `secrets/`, `.env`, tokens, OAuth credentials, local DB/cache/log files are not tracked in the current snapshot.
- [x] Check git history for previously committed secrets.
- [x] Confirm current runtime credentials are outside this repository.
- [x] Remove historical secret material from git history before publishing.
- [x] Verify `.gitignore` covers local secrets and generated files.

Finding: git history contained deleted `deployments/k8s/app-secret.yaml` with non-placeholder PostgreSQL and RabbitMQ credentials from commit `a2dd7b6`. The local history was rewritten to remove `deployments/k8s` from all commits, then `refs/original` were deleted and git GC was run. The rewritten history was force-pushed to the remote.

Note: current history scans still match local development defaults such as `postgres://postgres:postgres@...` and `amqp://guest:guest@...` in config examples. Those are documented local defaults, not real production credentials.

### 2. Repo Cleanup - Complete

- [x] Remove `.idea/*` from git tracking.
- [x] Add `.idea/` to `.gitignore`.
- [x] Remove scaffold noise such as unused `.gitkeep` files where appropriate.
- [x] Remove or rewrite private/local references such as private deployment repos and local paths from README.
- [x] Remove or rewrite private/local references such as private deployment repos and local paths from docs.

### 3. README Portfolio Rewrite - Complete

- [x] Rewrite README as a portfolio case study instead of an iteration log.
- [x] Include problem statement, architecture, features, local demo flow, privacy model, and limitations.
- [x] Add a Mermaid architecture diagram.
- [x] Add a clear note that this is a personal backend project, not production SaaS.
- [x] Add a short "Why not just Gmail filters?" section.

### 4. Examples - In Progress

- [x] Add `.env.example`.
- [x] Document OAuth setup without real secrets.
- [x] Add example SQL rules.
- [ ] Consider adding fake-message fixtures or a sample dry-run response.

### 5. Tests And CI - Complete

- [x] Run `go test ./...`.
- [x] Verify GitHub Actions are suitable for a public repo.
- [x] Ensure unit tests do not require Postgres, RabbitMQ, or Gmail.
- [x] Leave CI badge out for now; CI is configured, but the repository is not public yet.

Note: tests pass with `GOCACHE=/tmp/ets-gocache go test ./...` in the local sandbox because the default Go cache directory is read-only here. Normal local and GitHub Actions runs can use `go test ./...`.

### 6. Product Surface Cleanup - Complete

- [x] Decide how to present `Transactions`.
- [x] Describe `Unknown` as a fallback bucket.
- [x] Document that labels are configurable through User Rules.

### 7. Architecture Docs Polish - Complete

- [x] Sync `docs/architecture.md` with current behavior.
- [x] Review ADR wording for external readers.
- [x] Move speculative future work into a clear roadmap section.

### 8. Final Public Check - Complete

- [x] Run `git status`.
- [x] Inspect `git ls-files` for repo noise.
- [x] Search for sensitive/private terms before publishing.
- [x] Run `go test ./...`.
- [x] Read README from an interviewer's perspective.
