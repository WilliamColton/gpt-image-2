---
phase: 06-invite-code
plan: 01
subsystem: backend-data-layer
tags: [database, config, models, bcrypt, GORM, AutoMigrate]
dependency_graph:
  requires: []
  provides: [User-model-extension, Config-invite-fields, AuthUser-DTO-extension, bcrypt-direct-dep]
  affects: [backend-go/database, backend-go/config, backend-go/service, backend-go/go.mod]
tech_stack:
  added: [golang.org/x/crypto v0.52.0 (bcrypt sub-package)]
  patterns: [Config persistMu read-modify-write, GORM AutoMigrate nullable columns, DTO json:"-" for secrets]
key_files:
  created:
    - backend-go/service/models_test.go
  modified:
    - backend-go/database/models.go
    - backend-go/config/config.go
    - backend-go/database/database.go
    - backend-go/service/models.go
    - backend-go/go.mod
    - backend-go/go.sum
    - backend-go/config/config_test.go
    - backend-go/database/models_test.go
decisions:
  - "Invite config stored in config.json sharing persistMu with pricing config"
  - "bcrypt.DefaultCost (10) selected for password hashing"
  - "All new User columns are nullable (*string/*int64) ensuring existing user safety"
  - "username and invite_code use GORM uniqueIndex; SQLite treats NULLs as distinct"
  - "PasswordHash uses json:\"-\" to prevent serialization; InviteCodeSetAt also hidden"
  - "dbUserToAuthUser ImageCount defaults to 0 (filled by AuthMe later)"
metrics:
  duration: "~7min"
  completed: "2026-05-23T10:54:02Z"
---

# Phase 06 Plan 01: 邀请码数据层基础

**One-liner:** Extend User model with 5 nullable invite/password columns, add invite reward config with persistMu-safe persistence, extend AuthUser/User/AdminUser DTOs with username/needsMigration, and promote bcrypt to direct Go dependency.

## Execution Summary

Single-task TDD execution: RED (tests) -> GREEN (implementation). All 3 test files pass, `go build` and `go vet` clean, `golang.org/x/crypto` is now a direct dependency.

### Task 1: 扩展User数据库模型 + 配置扩展 + DTO更新 + bcrypt依赖

**TDD Gates:**
- RED: `34d36f0` — 3 test files with 22 total test cases targeting new fields/functions
- GREEN: `e726d51` — All 5 source files modified, all tests pass

**Changes:**

1. **database/models.go** — User struct extended with 5 nullable fields:
   - `PasswordHash *string` (text, nullable)
   - `Username *string` (text, uniqueIndex — NULLs treated distinct by SQLite)
   - `InviteCode *string` (text, uniqueIndex — NULLs treated distinct by SQLite)
   - `InviteCodeSetAt *int64` (nullable)
   - `InvitedBy *string` (text, index — stores inviter's invite code text)

2. **config/config.go** — Config struct extended with 3 int fields:
   - `InviteInviterReward`, `InviteInviteeReward`, `InviteDefaultQuota` (all `json` tags, no omitempty)
   - 3 getter functions: `GetInviteInviterReward()`, `GetInviteInviteeReward()` (return 0 if <=0), `GetInviteDefaultQuota()`
   - `persistInviteConfig(inviterReward, inviteeReward, defaultQuota int)` — acquires `persistMu`, reads config.json into `map[string]json.RawMessage`, marshals each int value, writes atomically
   - `SetInviteConfig(inviterReward, inviteeReward, defaultQuota int)` — runtime update + persistence

3. **service/models.go** — DTOs extended:
   - `AuthUser`: +Username, +NeedsMigration fields (both omitempty)
   - `User`: +Username, +PasswordHash (json:"-"), +InviteCode, +InviteCodeSetAt (json:"-")
   - `AdminUser`: +Username (omitempty)
   - New function `dbUserToAuthUser(u *database.User) *AuthUser` — converts DB model, derives NeedsMigration from PasswordHash == nil, defaults ImageCount to 0

4. **go.mod** — `golang.org/x/crypto` promoted from indirect to direct (v0.52.0), blank import in service/models.go

5. **database/database.go** — No changes needed; AutoMigrate already includes `&User{}`

## Deviations from Plan

None — plan executed exactly as written.

## TDD Gate Compliance

- RED gate: `34d36f0` (test commit with 3 test files, 22 tests) — verified tests failed before implementation
- GREEN gate: `e726d51` (feat commit with implementation) — all 22 tests pass, `go build` + `go vet` clean

## Test Results

```
ok  	gpt-image-playground/backend/config	0.210s (9 config tests)
ok  	gpt-image-playground/backend/database	0.157s (2 new + existing)
ok  	gpt-image-playground/backend/handler	4.073s (unchanged)
ok  	gpt-image-playground/backend/service	6.350s (11 new + existing)
```

## Known Stubs

| File | Line | Reason |
|------|------|--------|
| backend-go/service/models.go | dbUserToAuthUser | ImageCount=0 intentionally — filled by AuthMe/GetMe handler in Task 2 |

## Threat Flags

None — all security surface covered by plan's threat model (T-06-01-01: persistMu for config concurrency, T-06-01-02: PasswordHash json:"-", T-06-01-03: defaults to 0).

## Self-Check: PASSED

- [x] Files exist: database/models.go (modified), config/config.go (modified), service/models.go (modified)
- [x] Commits exist: 34d36f0 (RED), e726d51 (GREEN)
- [x] `go build ./...` exit code 0
- [x] `go vet ./...` exit code 0
- [x] `go test ./...` all packages pass
- [x] `golang.org/x/crypto` is direct dependency in go.mod
