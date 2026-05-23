---
phase: 06-invite-code
plan: 06
subsystem: admin-dashboard
tags: [admin, invite-config, reset-password, UI]
requires: [06-04]
provides: [06-07]
affects: [admin-dashboard]
tech-stack:
  added: []
  patterns: [dark-only-admin-theme, glass-morphism-modals, TDD-source-tests]
key-files:
  created: []
  modified:
    - src/admin/AdminDashboard.tsx
    - src/admin/AdminDashboard.test.tsx
decisions: []
metrics:
  duration: ~3min
  completed: 2026-05-23
---

# Phase 06 Plan 06: Admin Invites Tab + Reset Password Summary

**One-liner:** Admin dashboard gains "邀请码设置" tab with reward config area and invite usage table, plus per-user password reset modal.

## Plan Outcome

- Added `invites` to the Tab union type
- Created invite config section: three number inputs (inviter reward, invitee reward, default quota) with green "保存配置" button calling `adminUpdateInviteConfig`
- Created invite usage list: table with user/inviteCode/usageCount columns, "暂无邀请码使用记录" empty state
- Added "重置密码" button to each user row in the users table
- Added reset password modal with username display, password input (min 8 chars), cancel/confirm buttons
- All admin areas follow dark-only theme with `border-white/10` conventions

## Tasks Completed

| # | Type | Name | Commit | Files |
|---|------|------|--------|-------|
| 9 | auto (TDD RED) | Add failing tests for invites tab and reset password | 76872b9 | src/admin/AdminDashboard.test.tsx |
| 9 | auto (TDD GREEN) | Implement invites tab and reset password | b7a4dfe | src/admin/AdminDashboard.tsx |

## Test Verification

- **RED:** 28 new tests added, all failing (28 failed, 39 passed)
- **GREEN:** All 67 tests passing after implementation
- **TypeScript:** `tsc --noEmit` clean (no errors)

## TDD Gate Compliance

- RED gate commit: `76872b9` — `test(06-06): add failing tests for invites tab and reset password`
- GREEN gate commit: `b7a4dfe` — `feat(06-06): add invites tab and reset password to admin dashboard`
- REFACTOR gate: not needed (implementation followed plan exactly)

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None. All data sources are wired to actual API functions (`adminGetInviteConfig`, `adminUpdateInviteConfig`, `adminListInvites`, `adminResetPassword`).

## Self-Check: PASSED

- src/admin/AdminDashboard.tsx exists and modified
- src/admin/AdminDashboard.test.tsx exists and modified
- Commit 76872b9 confirmed (RED)
- Commit b7a4dfe confirmed (GREEN)
- All 67 tests pass
- TypeScript compiles clean
