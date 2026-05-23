---
phase: 06-invite-code
plan: 03
subsystem: backend-auth-handlers
tags: [auth, handler, admin, password-login, registration, invite-code, routing]
requires:
  - 06-02 (service functions)
provides:
  - POST /api/auth/login-password handler
  - POST /api/auth/register handler
  - POST /api/auth/migrate handler
  - POST /api/auth/change-password handler
  - PUT /api/auth/invite-code handler
  - GET /api/auth/invite-code handler
  - PUT /api/admin/users/:id/password handler
  - GET /api/admin/invite-config handler
  - PUT /api/admin/invite-config handler
  - GET /api/admin/invites handler
affects:
  - backend-go/handler/auth.go
  - backend-go/handler/admin.go
  - backend-go/main.go
  - backend-go/handler/auth_handler_test.go
  - backend-go/handler/admin_handler_test.go
  - backend-go/service/auth.go
tech-stack:
  added: []
  patterns: [gin.HandlerFunc, ShouldBindJSON, gin.H, AuthMiddleware, AdminMiddleware, GetAuthUser]
key-files:
  created:
    - backend-go/handler/auth_handler_test.go
    - backend-go/handler/admin_handler_test.go
  modified:
    - backend-go/handler/auth.go
    - backend-go/handler/admin.go
    - backend-go/main.go
    - backend-go/service/auth.go
decisions:
  - AuthLoginPassword returns 401 for invalid credentials, 400 for missing fields
  - AuthRegister does server-side validation (username 3-20 runes, password min 8) and returns 400
  - AuthMigrate and AuthChangePassword use middleware.GetAuthUser for current user ID
  - AuthLogin (code login) modified to include needsMigration in response
  - AuthMe populates username and needsMigration from FindUserByID result
  - AdminResetPassword validates min 8 chars password, uses c.Param("id") for target user
  - AdminUpdateInviteConfig validates non-negative rewards, calls config.SetInviteConfig for persistence
  - AdminListInvites returns {invites: []InviteRow} with username, inviteCode, usageCount
  - FindUserByID extended to return Username and PasswordHash for handler consumption
  - Auth routes: login-password and register are public; migrate, change-password, invite-code require AuthMiddleware
  - Admin routes: all 4 new admin routes are behind AdminMiddleware
metrics:
  duration: ~15min
  tasks: 2
  files_created: 2
  files_modified: 4
  tests_added: 22
  tests_passing: 22
  completed: "2026-05-23T11:45Z"
---

# Phase 06 Plan 03: Auth and Admin Handler Endpoints Summary

All 10 new HTTP endpoints implemented: 6 auth handlers (LoginPassword, Register, Migrate, ChangePassword, SetInviteCode, GetInviteCode) and 4 admin handlers (ResetPassword, GetInviteConfig, UpdateInviteConfig, ListInvites), plus modifications to AuthLogin and AuthMe for the needsMigration flag.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed FindUserByID not returning Username and PasswordHash**
- **Found during:** Task 3 (AuthMe needed Username + NeedsMigration)
- **Issue:** `service.FindUserByID` returned a `User` DTO missing `Username` and `PasswordHash` fields, preventing AuthMe from populating these values for the response.
- **Fix:** Extended `FindUserByID` to dereference `*string` fields (`Username`, `PasswordHash`) and include them in the returned DTO.
- **Files modified:** backend-go/service/auth.go
- **Commit:** d647565

**2. [Rule 1 - Bug] Fixed test helper slice bounds panic**
- **Found during:** Task 3 RED
- **Issue:** `createAuthTestUser` used `userID[:8]` which panicked when userID was shorter than 8 characters (e.g., "user-1" has 6 characters).
- **Fix:** Added a bounds check — only takes prefix if userID length >= 8.
- **Files modified:** backend-go/handler/auth_handler_test.go
- **Commit:** 5dfa1d0

### Stub Detection

No stubs detected in production code. All handler functions are fully implemented with complete validation, service delegation, and error responses.

### Threat Flags

| Flag | File | Description |
|------|------|-------------|
| threat_flag: new_public_endpoint | backend-go/handler/auth.go | POST /api/auth/login-password and POST /api/auth/register are public endpoints without authentication |
| threat_flag: auth_required_endpoint | backend-go/handler/auth.go | POST /api/auth/migrate, POST /api/auth/change-password, PUT/GET /api/auth/invite-code require AuthMiddleware |
| threat_flag: admin_only_endpoint | backend-go/handler/admin.go | 4 new admin endpoints protected by AdminMiddleware, including password reset capability |

## Task Summary

| # | Name | Status | Commit | Tests |
|---|------|--------|--------|-------|
| 3 | Auth handler新端点 (6 functions) | Complete | d647565 | 15 pass |
| 4 | Admin handler (4 functions) + routing | Complete | 9efebd0 | 7 pass |

## Commits

| Hash | Type | Message |
|------|------|---------|
| 5dfa1d0 | test | test(06-invite-code): add failing handler tests for auth endpoints (Task 3+4 RED) |
| d647565 | feat | feat(06-invite-code): implement 6 auth handler endpoints (Task 3 GREEN) |
| b8535e9 | test | test(06-invite-code): add failing admin handler tests (Task 4 RED) |
| 9efebd0 | feat | feat(06-invite-code): implement admin handlers and register all new routes (Task 4 GREEN) |

## TDD Gate Compliance

| Gate | Commit | Status |
|------|--------|--------|
| RED (Task 3) | 5dfa1d0 | PASS |
| GREEN (Task 3) | d647565 | PASS |
| RED (Task 4) | b8535e9 | PASS |
| GREEN (Task 4) | 9efebd0 | PASS |

## Verification

- **go vet ./...**: PASS (clean)
- **go build ./...**: PASS (clean)
- **go test ./...**: PASS (all packages, no regressions)
- **Acceptance criteria**: All 10 handler functions implemented, all 10 routes registered, correct middleware applied, needsMigration in login + me endpoints

## Self-Check: PASSED

- [x] `backend-go/handler/auth.go` — 6 new exported handler functions exist
- [x] `backend-go/handler/admin.go` — 4 new exported handler functions exist
- [x] `backend-go/main.go` — 10 new routes registered
- [x] `backend-go/service/auth.go` — FindUserByID extended with Username and PasswordHash
- [x] All 4 commits verified in git log
- [x] Full test suite passes (config, database, handler, service)
- [x] go vet clean
- [x] go build clean
