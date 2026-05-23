---
phase: 06
slug: invite-code
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-23
---

# Phase 06 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | vitest v4.1.5 (frontend), Go `testing` + `net/http/httptest` (backend) |
| **Config file** | none — vitest runs without config file (zero-config mode), Go uses `*_test.go` convention |
| **Quick run command** | `npm test` (vitest run), `cd backend-go && go test ./...` |
| **Full suite command** | `npm test && cd backend-go && go test ./...` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `cd backend-go && go test ./... -count=1` (backend), `npx vitest run` (frontend)
- **After every plan wave:** Run `npm test && cd backend-go && go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 06-01-01 | 01 | 1 | D-02 | T-06-01 | bcrypt hash stored, never plaintext | unit | `go test ./service/ -run TestPasswordHash -v` | ❌ W0 | ⬜ pending |
| 06-01-02 | 01 | 1 | D-03 | T-06-02 | POST /api/auth/login-password returns token+user+needsMigration | integration | `go test ./handler/ -run TestPasswordLogin -v` | ❌ W0 | ⬜ pending |
| 06-02-01 | 02 | 1 | D-04 | T-06-03 | Username 3-20 chars, allows Chinese, globally unique | unit | `go test ./service/ -run TestUsernameValidation -v` | ❌ W0 | ⬜ pending |
| 06-02-02 | 02 | 1 | D-05 | T-06-04 | Password min 8 chars enforced | unit | `go test ./service/ -run TestPasswordValidation -v` | ❌ W0 | ⬜ pending |
| 06-03-01 | 03 | 2 | D-06 | T-06-05 | Change password: old pw required; admin reset: no old pw required | integration | `go test ./handler/ -run TestChangePassword -v` | ❌ W0 | ⬜ pending |
| 06-03-02 | 03 | 2 | D-10 | T-06-06 | needsMigration: true when password_hash IS NULL | integration | `go test ./handler/ -run TestPasswordLoginMigration -v` | ❌ W0 | ⬜ pending |
| 06-04-01 | 04 | 1 | D-11 | T-06-07 | POST /api/auth/register returns token+user | integration | `go test ./handler/ -run TestRegister -v` | ❌ W0 | ⬜ pending |
| 06-04-02 | 04 | 1 | D-15 | T-06-08 | Invite code globally unique, user can set own | integration | `go test ./handler/ -run TestSetInviteCode -v` | ❌ W0 | ⬜ pending |
| 06-05-01 | 05 | 2 | D-18 | T-06-09 | Register with invite code awards both inviter and invitee quota | integration | `go test ./handler/ -run TestRegisterInviteReward -v` | ❌ W0 | ⬜ pending |
| 06-05-02 | 05 | 2 | D-29 | T-06-10 | Admin API endpoints return correct data | integration | `go test ./handler/ -run TestAdminInvite -v` | ❌ W0 | ⬜ pending |
| 06-06-01 | 06 | 1 | D-23 | — | LoginModal renders tabs, switches correctly | unit | `npx vitest run src/components/LoginModal.test.tsx` | ❌ W0 | ⬜ pending |
| 06-06-02 | 06 | 1 | D-25 | — | MigrationModal cannot be closed (no Escape, no backdrop click) | unit | `npx vitest run src/components/MigrationModal.test.tsx` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `backend-go/service/auth_test.go` — NEW: covers password hashing, username validation, invite code uniqueness, reward distribution
- [ ] `backend-go/handler/auth_test.go` — NEW: covers password login, register, migrate, change password, invite code CRUD endpoints
- [ ] `backend-go/handler/admin_test.go` — EXTEND: covers admin reset password, invite config GET/PUT, invites list
- [ ] `src/components/LoginModal.test.tsx` — NEW: covers tab switching, form validation, error display
- [ ] `src/components/MigrationModal.test.tsx` — NEW: covers forced non-close behavior, form validation, submit
- [ ] `src/components/RegisterModal.test.tsx` — NEW: covers form validation, invite code optionality
- [ ] `src/lib/backendApi.test.ts` — NEW/EXTEND: covers new API functions (loginWithPassword, register, migrate, changePassword, etc.)
- [ ] `src/admin/adminApi.test.ts` — EXTEND: covers new admin API functions (resetPassword, inviteConfig, listInvites)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Migration Modal visual non-close | D-25 | Escape/backdrop behavior requires real browser interaction | Open modal, press Escape — nothing happens. Click outside modal — nothing happens. Only submit button closes it. |
| SettingsModal section layout | D-26 | Visual layout verification | Open SettingsModal, verify invite code section, change password section, and redemption code section are visually distinct with separators. |
| Admin invite tab table display | D-22 | Table rendering requires browser | Open admin /invites tab, verify reward config inputs render, table shows all users with invite codes and usage counts. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
