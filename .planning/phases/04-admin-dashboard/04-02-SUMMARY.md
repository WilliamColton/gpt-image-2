---
phase: 04-admin-dashboard
plan: 02
subsystem: frontend
tags: [admin, dashboard, quota, user-management, frontend]
dependency_graph:
  requires: [admin-api, quota-enforcement, admin-middleware]
  provides: [admin-dashboard-ui]
  affects: [main-entry, admin-auth-flow]
tech_stack:
  added: []
  patterns: [lazy-admin-routing, separate-admin-token, quota-management-table]
key_files:
  created:
    - src/admin/adminApi.ts
    - src/admin/AdminPage.tsx
    - src/admin/AdminLogin.tsx
    - src/admin/AdminDashboard.tsx
  modified:
    - src/main.tsx
decisions:
  - "Admin uses separate localStorage token key (gpt-image-playground-admin-token) to avoid conflicts with user session"
  - "Admin route uses lazy() import so admin code is only loaded when visiting /admin, keeping main app bundle size unchanged"
  - "No react-router-dom dependency; simple pathname check for /admin routing"
metrics:
  duration: ~5m
  completed: "2026-05-08"
  tasks_completed: 2
  tasks_total: 2
---

# Phase 4 Plan 2: Admin Frontend Dashboard Summary

Frontend admin dashboard with login page, user table with quota management, and user status toggle. Lazy-loaded on /admin route to avoid impacting main app bundle.

## Tasks Completed

### Task 1: Admin API client and routing setup

**Commit:** 26beca3

**Changes:**
- Created `src/admin/adminApi.ts` with admin API client functions: `adminLogin`, `adminListUsers`, `adminUpdateQuota`, `adminToggleStatus`, `isAdminLoggedIn`, `clearAdminToken`
- Uses separate localStorage key `gpt-image-playground-admin-token` to avoid conflicts with user session
- Follows same `request<T>` pattern as `backendApi.ts`
- Updated `src/main.tsx` to check `window.location.pathname` for `/admin` and lazy-load `AdminPage` via `React.lazy()`
- Created `src/admin/AdminPage.tsx` shell that switches between `AdminLogin` and `AdminDashboard` based on login state

### Task 2: Admin login and dashboard UI components

**Commit:** cf544cc

**Changes:**
- Created `src/admin/AdminLogin.tsx` with admin login form:
  - Password input for adminApikey
  - Error display, loading state
  - "Return to home" link
  - Visual style matches LoginModal (glass morphism, Tailwind/Zinc colors)
- Created `src/admin/AdminDashboard.tsx` with user management table:
  - User list with: label, role, registration time, quota (used/total or used/unlimited), status badge
  - Quota operations: input box + "increase"/"decrease"/"reset" buttons per user
  - Status toggle: switches user between active/disabled
  - Header with "refresh" and "logout" buttons
  - Quota exceeded state highlighted in red
  - Dark theme consistent with main app (gray-950 background, white/10 borders)

## Deviations from Plan

None - plan executed exactly as written.

## Decisions Made

1. **Separate admin token storage** - Admin uses `gpt-image-playground-admin-token` in localStorage, completely separate from user session token. This prevents admin auth from interfering with normal user operations.
2. **Lazy-loaded admin module** - `React.lazy()` ensures admin code is only fetched when visiting `/admin`. Main app bundle size is unaffected.
3. **No router dependency** - Simple `window.location.pathname` check instead of adding react-router-dom. The admin page is a separate entry point, not a route within the SPA.

## Known Stubs

None - all components are fully wired to real API endpoints.

## Threat Flags

No new security surface. Admin endpoints are protected by backend AdminMiddleware (from Plan 04-01). Admin token is stored separately from user token.

## Self-Check

- [x] src/admin/adminApi.ts exists with adminLogin, adminListUsers, adminUpdateQuota, adminToggleStatus, isAdminLoggedIn, clearAdminToken
- [x] src/admin/adminApi.ts uses separate localStorage key 'gpt-image-playground-admin-token'
- [x] src/admin/AdminPage.tsx exists with conditional login/dashboard rendering
- [x] src/admin/AdminLogin.tsx exists with login form, error handling, loading state
- [x] src/admin/AdminLogin.tsx contains "管理后台登录" heading
- [x] src/admin/AdminDashboard.tsx exists with user table, quota controls, status toggle
- [x] src/admin/AdminDashboard.tsx contains "增加", "减少", "重置" buttons
- [x] src/admin/AdminDashboard.tsx contains status toggle ("禁用"/"启用")
- [x] src/main.tsx checks window.location.pathname for /admin and lazy-loads AdminPage
- [x] TypeScript compiles without errors (npx tsc --noEmit exits 0)
- [x] Vite production build succeeds (npm run build)
- [x] Commits 26beca3 and cf544cc exist

## Self-Check: PASSED
