---
phase: 09
slug: audit-log
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-24
---

# Phase 09 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (backend) + vitest 4.x (frontend) |
| **Config file** | vitest.config.ts (frontend), go test (backend — no config needed) |
| **Quick run command** | `go test ./backend-go/service/ -run TestLog -count=1` |
| **Full suite command** | `cd backend-go && go test ./... -count=1 && cd .. && npx vitest run` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./backend-go/service/ -run TestLog -count=1` (backend) or `npx vitest run related --reporter=verbose` (frontend)
- **After every plan wave:** Run full suite: `cd backend-go && go test ./... -count=1 && cd .. && npx vitest run`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 09-01-01 | 01 | 1 | LOG-01 | T-09-SC | AuditLog model compiles and AutoMigrate registers table | unit | `cd backend-go && go build -o /dev/null .` | ❌ W0 | ⬜ pending |
| 09-01-02 | 01 | 1 | LOG-03, LOG-04 | T-09-01, T-09-02 | LogEvent/QueryLogs/CleanLogs functional, handlers return correct JSON | unit/integration | `cd backend-go && go build -o /dev/null .` | ❌ W0 | ⬜ pending |
| 09-02-01 | 02 | 2 | LOG-02 | T-09-06, T-09-07 | Auth and generation events instrumented, IP captured for goroutine | unit | `cd backend-go && go build -o /dev/null .` | ❌ W0 | ⬜ pending |
| 09-02-02 | 02 | 2 | LOG-02 | T-09-08, T-09-09 | Admin events instrumented, AdminMiddleware sets user context | unit | `cd backend-go && go build -o /dev/null . && go vet ./handler/ ./middleware/` | ❌ W0 | ⬜ pending |
| 09-03-01 | 03 | 2 | LOG-05 | T-09-10 | API client types and functions compile, correct query string building | unit | `npx tsc --noEmit` | ❌ W0 | ⬜ pending |
| 09-03-02 | 03 | 2 | LOG-06, LOG-07 | T-09-11, T-09-12 | Logs tab renders with filter bar, table, cleanup modals; no TS errors | component | `npx tsc --noEmit` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `backend-go/service/log_test.go` — TestLogEvent, TestQueryLogs, TestCleanLogs unit tests
- [ ] `backend-go/handler/log_test.go` — TestAdminListLogs, TestAdminCleanLogs integration tests
- [ ] `src/admin/adminApi.test.ts` — adminListLogs, adminCleanLogs API client tests
- [ ] `src/admin/AdminDashboard.test.tsx` — Logs tab filter bar, table, cleanup UI tests

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Audit log rows appear after business events | LOG-02 | Integration across multiple handler files + goroutine | Perform each business action (login, generate, admin quota change) and verify log entries appear in admin logs tab |
| Cleanup deletes correct rows | LOG-04, LOG-07 | Manual delete + verify count | Generate test logs, use cleanup buttons, verify toast shows correct deleted count |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
