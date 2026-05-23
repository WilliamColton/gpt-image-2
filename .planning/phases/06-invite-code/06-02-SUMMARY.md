---
phase: 06-invite-code
plan: 02
subsystem: auth
tags: [bcrypt, jwt, gorm, sqlite, password-auth, invite-code]

# Dependency graph
requires:
  - phase: 06-invite-code
    plan: 01
    provides: "Extended User model (PasswordHash, Username, InviteCode, InviteCodeSetAt, InvitedBy), dbUserToAuthUser helper, invite config getters"
provides:
  - "LoginWithPassword: bcrypt password authentication with anti-enumeration"
  - "RegisterUser: username/password registration with optional invite code and quota rewards"
  - "MigrateUser: username+password setup for legacy users"
  - "ChangePassword: old-password-verified password update"
  - "SetInviteCode: atomic invite code assignment with DB unique constraint"
  - "GetInviteCode: query user's current invite code"
  - "ListInvites: aggregate invite usage counts via InvitedBy field"
  - "AdminResetPassword: admin password reset without old password requirement"
  - "Updated LoginWithCode: returns needsMigration flag via dbUserToAuthUser"
affects: [06-03-PLAN.md, 06-04-PLAN.md, 06-05-PLAN.md, 06-06-PLAN.md, 06-07-PLAN.md]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "bcrypt.DefaultCost (10) for all password hashing/verification"
    - "DB unique constraint for invite code conflict detection (never check-then-set)"
    - "gorm.Expr atomic quota updates in registration reward flow"
    - "len([]rune(...)) for Unicode-aware username length validation"
    - "Service-layer validation before DB operations (username 3-20 chars, password 8+ chars)"
    - "slog structured logging: never logs passwords, logs user_id for audit"

key-files:
  created:
    - backend-go/service/auth_test.go
  modified:
    - backend-go/service/auth.go

key-decisions:
  - "bcrypt.DefaultCost (10) selected as plan specified"
  - "LoginWithCode updated to use dbUserToAuthUser for both paths (existing and new user)"
  - "Invite code conflict caught via strings.Contains on UNIQUE constraint error message"
  - "ImageCount left at 0 in dbUserToAuthUser (AuthMe handler fills it separately)"

requirements-completed:
  - D-01
  - D-02
  - D-03
  - D-04
  - D-05
  - D-06
  - D-07
  - D-08
  - D-09
  - D-10
  - D-11
  - D-12
  - D-13
  - D-14
  - D-15
  - D-16
  - D-17
  - D-18

# Metrics
duration: 5min
completed: 2026-05-23
---

# Phase 06 Plan 02: Password auth, registration, invite code, migration service functions

**bcrypt password authentication with anti-enumeration, username/password registration with invite code quota rewards, and forced migration for legacy users**

## Performance

- **Duration:** ~5 min
- **Started:** 2026-05-23T11:16:00Z
- **Completed:** 2026-05-23T11:21:57Z
- **Tasks:** 2 (combined TDD: RED + GREEN for both tasks)
- **Files modified:** 2

## Accomplishments
- All 8 exported service functions implemented: LoginWithPassword, RegisterUser, MigrateUser, ChangePassword, SetInviteCode, GetInviteCode, ListInvites, AdminResetPassword
- LoginWithCode updated to return `needsMigration` via `dbUserToAuthUser` helper
- 27 automated tests covering all success paths, error paths, edge cases, and validation rules
- Full service test suite passes (82 tests total, 0 failures)

## Task Commits

Each task was committed atomically:

1. **Task 2+3 RED: failing tests** - `cbab1f5` (test)
2. **Task 2+3 GREEN: implementation** - `b0751ef` (feat)

**Plan metadata:** pending

## Files Created/Modified
- `backend-go/service/auth.go` - Added LoginWithPassword, RegisterUser, MigrateUser, ChangePassword, SetInviteCode, GetInviteCode, ListInvites, AdminResetPassword; updated LoginWithCode to use dbUserToAuthUser; added checkPassword/hashPassword bcrypt helpers
- `backend-go/service/auth_test.go` - 27 tests covering all service functions

## Decisions Made
- bcrypt.DefaultCost (10) used as specified in plan and research
- LoginWithCode refactored to use `dbUserToAuthUser` for both existing-user and new-user paths, ensuring needsMigration flag is consistently returned
- Invite code uniqueness enforced via `strings.Contains(result.Error.Error(), "UNIQUE constraint")` after direct UPDATE attempt (race-condition-safe)
- ImageCount left at 0 in `dbUserToAuthUser` -- AuthMe handler will fill it later

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed TestLoginWithPassword_NoPasswordHash using wrong lookup field**
- **Found during:** Task 2 (test execution)
- **Issue:** Test passed "LEGACYCODE" (Label value) as username to `LoginWithPassword`, but the function queries by `Username` field which was NULL. Test could never pass.
- **Fix:** Created user with `Username` explicitly set and `PasswordHash` left nil, matching the test's intent.
- **Files modified:** `backend-go/service/auth_test.go`
- **Verification:** Test now passes
- **Committed in:** `b0751ef` (GREEN commit)

**2. [Rule 1 - Bug] Fixed slice bounds panic in test helpers with short user IDs**
- **Found during:** Task 2 (first test run)
- **Issue:** `createTestUserWithPassword` used `userID[:8]` for Label, but test IDs like "user-1" (6 chars) caused runtime panic.
- **Fix:** Added `testLabel()` helper that checks length before slicing.
- **Files modified:** `backend-go/service/auth_test.go`
- **Verification:** All 27 tests pass without panic
- **Committed in:** `b0751ef` (GREEN commit)

---

**Total deviations:** 2 auto-fixed (both Rule 1 bugs)
**Impact on plan:** Both fixes necessary for test correctness. No scope creep.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required. All dependencies already present in go.mod.

## Next Phase Readiness
- All service-layer functions ready for handler integration in Plan 03 (auth handlers) and Plan 05 (admin handlers)
- `InviteRow` type exported for admin handler consumption
- BCrypt helpers (hashPassword, checkPassword) unexported but available for reuse if needed

---
*Phase: 06-invite-code*
*Completed: 2026-05-23*
