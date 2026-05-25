# Requirements: GPT Image Playground Stability Program

**Defined:** 2026-05-25
**Core Value:** Users and admins can rely on image generation, authentication, task lifecycle, quota, and billing workflows to behave correctly without manual recovery or avoidable support cost.

## v1 Requirements

Requirements for this stability milestone. Each maps to roadmap phases.

### Authentication and Access

- [ ] **AUTH-01**: Concurrent first-use login attempts with the same unused redemption code result in exactly one successful user/code claim.
- [ ] **AUTH-02**: A failed concurrent redemption-code claim does not create an extra usable user, issue a valid token, or grant quota.
- [ ] **AUTH-03**: Admin API requests validate the current persisted admin/account state before granting admin access.
- [ ] **AUTH-04**: A validly signed but no-longer-authorized admin-role JWT cannot access protected admin APIs.
- [ ] **AUTH-05**: The `needsMigration` field follows one explicit API contract that is reflected by backend service tests.

### Task Lifecycle and Generation Reliability

- [ ] **TASK-01**: User deletion of a queued or running task prevents the task from reappearing after backend generation work completes.
- [ ] **TASK-02**: Deleted queued or running tasks do not create successful output state, consume quota, create billing rows, or leave untracked generated output as if the task still existed.
- [ ] **TASK-03**: Backend task finalization checks task existence/cancellation state before persisting completion data.
- [ ] **TASK-04**: Task deletion/cancellation behavior has regression coverage for queued/running completion races.

### Invite and Quota Accounting

- [ ] **INVT-01**: Invited-user registration and inviter reward updates occur in one atomic transaction.
- [ ] **INVT-02**: If inviter quota reward persistence fails, invited-user creation fails instead of silently succeeding with missing reward accounting.
- [ ] **INVT-03**: Invite reward behavior has regression coverage for success and failure paths.

### Test and Release Confidence

- [ ] **TEST-01**: `npm test` can start Vitest successfully in this checkout after documented dependency repair or dependency state correction.
- [ ] **TEST-02**: Frontend test setup documents or enforces the dependency install state needed for Rollup/Vitest optional native packages.
- [ ] **TEST-03**: Backend tests covering the fixed known bugs pass with `go test ./...` from `backend-go/`.
- [ ] **TEST-04**: Each fixed known bug has targeted regression coverage at the service, handler, or frontend test layer where the defect occurred.

## v2 Requirements

Deferred to future stability/production hardening milestones. Tracked but not in current roadmap.

### Security Hardening

- **SECU-01**: Backend rejects startup with missing, malformed, or placeholder production secrets.
- **SECU-02**: Public and admin auth move away from long-lived localStorage tokens toward a safer session design.
- **SECU-03**: Image access no longer uses bearer tokens in query strings.
- **SECU-04**: API routes apply deployment-appropriate CORS, rate limiting, lockout, and trusted proxy policies.
- **SECU-05**: Admin endpoint API keys are write-only/redacted in API responses and stored outside plain admin-editable JSON where practical.

### Operations and Data Reliability

- **OPER-01**: Backend has graceful shutdown and restart recovery for in-flight generation jobs.
- **OPER-02**: Generation requests have server-side validation for prompt length, image count, `n`, enum values, compression range, image IDs, and dimensions.
- **OPER-03**: Uploaded/generated image storage has retention rules, cleanup jobs, and per-user storage controls.
- **OPER-04**: Admin actions are recorded in durable audit logs.
- **OPER-05**: Task, image, billing, and quota finalization is idempotent and reconcilable after partial failures.

### Scalability and Architecture

- **SCAL-01**: User/admin list endpoints and task bootstrap paths support pagination.
- **SCAL-02**: SSE task updates avoid per-connection SQLite polling where possible.
- **SCAL-03**: Production deployments can use network database/object storage without local SQLite/filesystem constraints.
- **SCAL-04**: Runtime configuration separates secrets from mutable admin settings.
- **SCAL-05**: Large frontend modules are decomposed only when required by ongoing feature or stability work.

## Out of Scope

Explicitly excluded from v1 to prevent scope creep.

| Feature | Reason |
|---------|--------|
| CI/CD pipeline creation | Useful for release confidence, but the selected scope is known bug fixes, not infrastructure setup. |
| Docker, compose, Kubernetes, Vercel, Netlify, or Cloudflare deployment manifests | Deployment productionization is deferred beyond this bug-fix milestone. |
| Replacing SQLite or local filesystem storage | Current known bugs can be fixed within existing persistence/storage architecture. |
| Full authentication redesign with cookies/refresh tokens | Important hardening, but broader than the admin middleware known bug. |
| Global rate limiting, CORS redesign, signed image URLs, and upload content scanning | Security hardening is tracked as v2 unless directly required by a v1 known bug fix. |
| Large UI redesign or admin dashboard decomposition | The milestone prioritizes correctness and stability, not presentation or component architecture cleanup. |
| New customer-facing image generation features | The goal is to make existing workflows reliable. |
| Full observability stack such as Sentry, OpenTelemetry, Prometheus, or dashboards | Operationally valuable but not part of the chosen known-bug list. |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| AUTH-01 | Phase 2 | Pending |
| AUTH-02 | Phase 2 | Pending |
| AUTH-03 | Phase 2 | Pending |
| AUTH-04 | Phase 2 | Pending |
| AUTH-05 | Phase 2 | Pending |
| TASK-01 | Phase 3 | Pending |
| TASK-02 | Phase 3 | Pending |
| TASK-03 | Phase 3 | Pending |
| TASK-04 | Phase 3 | Pending |
| INVT-01 | Phase 4 | Pending |
| INVT-02 | Phase 4 | Pending |
| INVT-03 | Phase 4 | Pending |
| TEST-01 | Phase 1 | Pending |
| TEST-02 | Phase 1 | Pending |
| TEST-03 | Phase 5 | Pending |
| TEST-04 | Phase 5 | Pending |

**Coverage:**
- v1 requirements: 16 total
- Mapped to phases: 16
- Unmapped: 0
- Coverage: 100%

---
*Requirements defined: 2026-05-25*
*Last updated: 2026-05-25 after roadmap creation*
