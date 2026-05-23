---
phase: 06-invite-code
plan: 05
subsystem: frontend-auth-ui
tags: [frontend, auth, ui, login, register, migration, settings, header]
dependency-graph:
  requires: [06-04]
  provides: [06-07]
  affects:
    - src/components/LoginModal.tsx
    - src/components/RegisterModal.tsx
    - src/components/MigrationModal.tsx
    - src/components/SettingsModal.tsx
    - src/components/Header.tsx
tech-stack:
  added: []
  patterns:
    - Radix Tabs (existing ui/tabs.tsx reused)
    - Glass morphism modal pattern (existing convention)
    - Zustand useStore selectors for authUser
    - client-side validation with Array.from for Chinese character counting
key-files:
  created:
    - src/components/RegisterModal.tsx
  modified:
    - src/components/LoginModal.tsx
    - src/components/MigrationModal.tsx
    - src/components/SettingsModal.tsx
    - src/components/Header.tsx
decisions:
  - LoginModal uses Radix Tabs for 兑换码/密码登录 tab switching (default: code)
  - RegisterModal validates username 3-20 chars via Array.from and password >= 8 chars client-side
  - MigrationModal uses raw glass morphism markup — no useCloseOnEscape, no backdrop onClick, no X button
  - SettingsModal invite code section fetches code on modal open via useEffect
  - SettingsModal change password validates newPassword >= 8 and match with confirmNewPassword client-side
  - Header username display uses authUser.username || authUser.label || '用户' fallback chain
  - All copywriting follows 06-UI-SPEC.md locked text exactly
metrics:
  duration: ~4min
  completed: 2026-05-23T12:41:00Z
---

# Phase 6 Plan 5: LoginModal/RegisterModal 改造 + MigrationModal/SettingsModal/Header 扩展 Summary

**One-liner:** Redesigned LoginModal with Tab switching (兑换码/密码登录), created RegisterModal (3-field registration with auto-login), built unclosable MigrationModal (raw glass morphism, z-[110]), extended SettingsModal with invite code and change password sections, and added username display to Header.

## Tasks Executed

### Task 7: 改造LoginModal + 新建RegisterModal (TDD)

**RED (306685d):** Created failing tests for LoginModal Tab switching (兑换码/密码登录 tabs), RegisterModal creation with three fields, and bottom register link behavior.

**GREEN (835aded):** Implemented:
- **LoginModal.tsx:** Added `<Tabs defaultValue="code">` with 兑换码 and 密码登录 tabs. Code tab preserves existing `loginWithCode` + `bootstrapBackendSession` flow. Password tab has username + password `<Input>` fields, calls `loginWithPassword` on submit. Bottom link "没有邀请码？注册" opens RegisterModal.
- **RegisterModal.tsx (NEW):** Three fields — 邀请码(optional) + 用户名 + 密码. Client-side validation: username 3-20 chars (Array.from for Chinese), password >= 8. Calls `register()` then `bootstrapBackendSession()` on success. Glass morphism styling at z-[100].

### Task 8: 新建MigrationModal + 扩展SettingsModal + 扩展Header用户名显示 (TDD)

**RED (46eccf3):** Created 25 failing source-check tests verifying MigrationModal unclosability (no useCloseOnEscape, no backdrop onClick, no X button), SettingsModal extensions (邀请码/修改密码 sections with correct imports), and Header username display.

**GREEN (631ba82):** Implemented:
- **MigrationModal.tsx:** Raw glass morphism modal at z-[110]. No useCloseOnEscape import, no backdrop onClick handler, no X close button. Three fields (用户名 + 密码 + 确认密码) with client-side validation. Calls `migrate()` then `getMe()` on success, updates authUser to clear needsMigration.
- **SettingsModal.tsx:** Added two new sections between redeem and logout:
  - 邀请码 section: displays current code (mono font) or "未设置", copy button (clipboard API), modify button with inline form (input + 确认/取消)
  - 修改密码 section: three password Inputs (旧密码/新密码/确认密码), client-side validation (newPassword >= 8, match check), calls `changePassword`
- **Header.tsx:** Added `useStore((s) => s.authUser)` selector, inserts `{authUser.username || authUser.label || '用户'}` span beside h1 title.
- **Test fix:** MigrationModal submit button uses ternary `{loading ? '设置中...' : '完成设置'}`, adjusted test regex to `['"]完成设置['"]` instead of `>完成设置<`.

## Verification

- **tsc --noEmit:** Passed cleanly (0 errors)
- **vitest run:** 267 tests passed (0 failures), all 18 test files green
- **Acceptance criteria:** All 25 plan acceptance criteria met (13 for Task 7 + 12 for Task 8)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed MigrationModal test regex for ternary-rendered button text**
- **Found during:** Task 8 GREEN verification
- **Issue:** Test expected `>完成设置<` in raw source, but implementation uses ternary `{loading ? '设置中...' : '完成设置'}`
- **Fix:** Changed assertion from `toContain('>完成设置<')` to `toMatch(/['"]完成设置['"]/)`
- **Files modified:** `src/components/MigrationModal.test.tsx`
- **Commit:** 631ba82

## Threat Flags

None — all threat mitigations from the plan's threat model are correctly implemented:
- T-06-05-01 (MigrationModal EoP): No Escape handler, no backdrop onClick, no X button — user must complete migration
- T-06-05-02 (password disclosure): All password Inputs use type="password" with no visibility toggle
- T-06-05-03 (invite code spoofing): No client-side uniqueness check — conflicts caught by backend unique constraint

## Known Stubs

None — all components are fully wired with real API calls, state management, and error handling.

## TDD Gate Compliance

| Gate | Commit | Status |
|------|--------|--------|
| RED (Task 7) | 306685d | PASS |
| GREEN (Task 7) | 835aded | PASS |
| RED (Task 8) | 46eccf3 | PASS |
| GREEN (Task 8) | 631ba82 | PASS |

No REFACTOR phase needed — implementations are clean, minimal, and follow existing patterns exactly.

## Self-Check: PASSED

All 5 key files exist. All 4 commits verified in git log.
