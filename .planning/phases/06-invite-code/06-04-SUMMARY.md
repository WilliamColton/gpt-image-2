---
phase: 06-invite-code
plan: 04
subsystem: frontend-auth-api
tags: [frontend, auth, api-client, migration-modal]
dependency-graph:
  requires: [06-03]
  provides: [06-05, 06-06, 06-07]
  affects:
    - src/lib/backendApi.ts
    - src/admin/adminApi.ts
    - src/store.ts
    - src/App.tsx
tech-stack:
  added: []
  patterns:
    - request<T>() Bearer token auto-injection (existing pattern reused)
    - adminRequest<T>() admin token management (existing pattern reused)
    - Zustand type propagation from import
key-files:
  created:
    - src/components/MigrationModal.tsx
    - src/lib/backendApi.test.ts
    - src/admin/adminApi-invite.test.ts
    - src/App-test.test.tsx
  modified:
    - src/lib/backendApi.ts
    - src/admin/adminApi.ts
    - src/App.tsx
decisions:
  - backendApi.ts extends AuthUser with optional username and needsMigration
  - loginWithPassword and register automatically store token (auto-login)
  - MigrationModal as stub component (Plan 05 replaces with full implementation)
  - Test files: backendApi.test.ts, adminApi-invite.test.ts, App-test.test.tsx
metrics:
  duration: ~6min
  completed: 2026-05-23T11:56:30Z
---

# Phase 6 Plan 4: 前端API客户端扩展 + MigrationModal占位组件 Summary

**One-liner:** Extended backendApi.ts with 6 auth functions (password login, register, migrate, password change, invite code CRUD), adminApi.ts with 4 management functions (reset password, invite config, invite list), and wired MigrationModal conditional rendering in App.tsx.

## Tasks Executed

### Task 5: 扩展前端API客户端 — backendApi.ts和adminApi.ts (TDD)

**RED (168f211):** Created 16 failing tests covering AuthUser type extensions and all 10 new API functions.

**GREEN (e9a307b):** Implemented all 10 functions:
- **backendApi.ts:** `loginWithPassword` (POST /api/auth/login-password), `register` (POST /api/auth/register), `migrate` (POST /api/auth/migrate), `changePassword` (POST /api/auth/change-password), `setInviteCode` (PUT /api/auth/invite-code), `getInviteCode` (GET /api/auth/invite-code)
- **adminApi.ts:** `adminResetPassword` (PUT /api/admin/users/:id/password), `adminGetInviteConfig` (GET /api/admin/invite-config), `adminUpdateInviteConfig` (PUT /api/admin/invite-config), `adminListInvites` (GET /api/admin/invites)
- **AuthUser interface** extended with `username?: string` and `needsMigration?: boolean`

### Task 6: 更新Zustand store支持needsMigration + App.tsx条件渲染MigrationModal + 占位组件 (TDD)

**RED (f66d424):** Created 3 source-check tests verifying MigrationModal import, needsMigration condition, and existing LoginModal behavior.

**GREEN (5eab697):** Created MigrationModal stub component (`return null`), added import and conditional render in App.tsx (`{authUser?.needsMigration && <MigrationModal />}`). AuthUser type propagation from backendApi.ts to store.ts confirmed via TypeScript compilation.

## Verification

- **tsc --noEmit:** Passed cleanly (0 errors)
- **vitest run:** 189 tests passed (0 failures), all 16 test files green
- **Acceptance criteria:** All 18 plan acceptance criteria met

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

| File | Line | Reason |
|------|------|--------|
| src/components/MigrationModal.tsx | 1-3 | Stub component (`return null`) — Plan 05 will replace with full implementation (non-dismissable modal, username/password/confirm fields, migrate API call) |

## Threat Flags

None — no new threat surface beyond existing request/response patterns. JWT token storage and Bearer token injection follow existing patterns in backendApi.ts and adminApi.ts.

## TDD Gate Compliance

| Gate | Commit | Status |
|------|--------|--------|
| RED (Task 5) | 168f211 | PASS |
| GREEN (Task 5) | e9a307b | PASS |
| RED (Task 6) | f66d424 | PASS |
| GREEN (Task 6) | 5eab697 | PASS |

No REFACTOR phase needed — implementation is minimal and follows existing patterns exactly.

## Self-Check: PASSED

All 5 created files exist. All 4 commits verified in git log.
