---
phase: 04-admin-dashboard
plan: 01
subsystem: backend
tags: [admin, quota, middleware, authentication]
dependency_graph:
  requires: []
  provides: [admin-api, quota-enforcement, admin-middleware]
  affects: [generate-handler, auth-service, database-schema]
tech_stack:
  added: []
  patterns: [admin-middleware, quota-check-before-task, used-count-increment]
key_files:
  created:
    - backend-go/handler/admin.go
  modified:
    - backend-go/database/database.go
    - backend-go/service/models.go
    - backend-go/service/auth.go
    - backend-go/middleware/middleware.go
    - backend-go/handler/generate.go
    - backend-go/main.go
decisions:
  - "Admin auth is stateless (JWT role=admin), no admin user record lookup needed for middleware"
  - "quota=0 means unlimited (no limit) per context decision"
  - "UsedCount increments by number of output images, not by 1"
metrics:
  duration: ~5m
  completed: "2026-05-08"
  tasks_completed: 2
  tasks_total: 2
---

# Phase 4 Plan 1: Admin Backend API Summary

Admin backend API with database schema changes, admin authentication, user management endpoints, and quota enforcement for image generation.

## Tasks Completed

### Task 1: Database migration, User model update, and admin middleware

**Commit:** e494dee

**Changes:**
- Added `quota` and `used_count` columns to users table migration in `database.go`
- Extended `User` struct with `Quota` and `UsedCount` fields in `models.go`
- Added `AdminUser` type for admin API JSON responses
- Updated `FindUserByID` and `LoginWithApikey` queries to include new columns
- Added `AdminMiddleware` that validates admin JWT and checks role=admin
- Added admin helper functions to `auth.go`: `ListAllUsers`, `UpdateUserQuota`, `SetUserStatus`, `IncrementUsedCount`, `CheckQuota`

### Task 2: Admin handler and quota enforcement in generate handler

**Commit:** e2541ea

**Changes:**
- Created `backend-go/handler/admin.go` with 4 handlers:
  - `AdminLogin` - validates adminApikey, issues JWT with role=admin
  - `AdminListUsers` - returns all users with quota info
  - `AdminUpdateQuota` - adjusts quota by delta, optionally resets used_count
  - `AdminToggleStatus` - sets user to active/disabled
- Added quota check before task creation in `GenerateImage` (returns 403 with "配额已用完" if exceeded)
- Added `IncrementUsedCount` call after successful image generation in `executeImageGeneration`
- Registered admin routes in `main.go`: login (public), users/quota/status (admin-authenticated)

## Deviations from Plan

None - plan executed exactly as written.

## Decisions Made

1. **Admin auth is stateless** - AdminMiddleware validates JWT and checks role claim directly, no database lookup needed. This is simpler and faster than looking up a user record.
2. **quota=0 means unlimited** - Per context decision, users with quota=0 can always generate images.
3. **UsedCount increments by output count** - When N images are generated in one task, used_count increases by N, not 1.

## Known Stubs

None - all functionality is wired to real database operations.

## Threat Flags

No new security surface beyond what the plan explicitly defines. Admin endpoints are protected by AdminMiddleware.

## Self-Check

- [x] backend-go/database/database.go contains quota/used_count migration
- [x] backend-go/service/models.go has Quota/UsedCount in User struct and AdminUser type
- [x] backend-go/service/auth.go has ListAllUsers, UpdateUserQuota, SetUserStatus, IncrementUsedCount, CheckQuota
- [x] backend-go/middleware/middleware.go has AdminMiddleware
- [x] backend-go/handler/admin.go exists with 4 handlers
- [x] backend-go/handler/generate.go has CheckQuota before task creation
- [x] backend-go/handler/generate.go has IncrementUsedCount on success
- [x] backend-go/main.go has admin routes registered
- [x] Commits e494dee and e2541ea exist

## Self-Check: PASSED
