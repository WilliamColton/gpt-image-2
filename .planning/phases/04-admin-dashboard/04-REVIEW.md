---
phase: 04-admin-dashboard
reviewed: 2026-05-08T12:00:00Z
depth: standard
files_reviewed: 12
files_reviewed_list:
  - backend-go/database/database.go
  - backend-go/handler/admin.go
  - backend-go/handler/generate.go
  - backend-go/main.go
  - backend-go/middleware/middleware.go
  - backend-go/service/auth.go
  - backend-go/service/models.go
  - src/admin/AdminDashboard.tsx
  - src/admin/AdminLogin.tsx
  - src/admin/AdminPage.tsx
  - src/admin/adminApi.ts
  - src/main.tsx
findings:
  critical: 0
  warning: 8
  info: 4
  total: 12
status: issues_found
---

# Phase 4: Code Review Report

**Reviewed:** 2026-05-08T12:00:00Z
**Depth:** standard
**Files Reviewed:** 12
**Status:** issues_found

## Summary

Reviewed the admin dashboard feature spanning backend Go handlers, middleware, database schema migration, service layer auth/quota logic, and the React admin frontend. The implementation is structurally sound -- route protection is correctly layered, the admin JWT flow is separated from user auth, and parameterized SQL queries prevent injection. However, 8 warnings were found covering a Fetch API body-consumption bug that loses error messages, a JWT signing-method validation gap, overly permissive CORS, a quota enforcement race condition, and several error-handling deficiencies in the database init and login paths.

## Warnings

### WR-01: Response body consumed twice on error -- server error messages silently lost

**File:** `src/admin/adminApi.ts:40-48`
**Issue:** When the server returns a non-JSON error response (e.g., a reverse proxy 502 with HTML body), `response.json()` consumes the body stream and throws a SyntaxError. The catch block then calls `response.text()`, but the body is already consumed per the Fetch API spec, so this throws a second error. The actual server error message is lost and the user sees an unhandled exception instead of a meaningful error.
**Fix:**
```typescript
if (!response.ok) {
  let message = `HTTP ${response.status}`
  const cloned = response.clone()
  try {
    const payload = await response.json()
    message = payload.error || payload.message || message
  } catch {
    try {
      message = await cloned.text()
    } catch {
      // fallback to default message
    }
  }
  throw new Error(message)
}
```

### WR-02: JWT signing method not explicitly validated in VerifyToken

**File:** `backend-go/service/auth.go:26-28`
**Issue:** The `jwt.Parse` keyfunc returns `[]byte(jwtSecret)` without checking `t.Method`. While golang-jwt v5 rejects `alg: none` by default and only accepts HMAC methods when the key is `[]byte`, the signing method is not explicitly validated against `HS256`. An attacker cannot forge tokens (they still need the secret), but best practice is to pin the algorithm to prevent any ambiguity.
**Fix:**
```go
token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
    }
    return []byte(jwtSecret), nil
})
```

### WR-03: JWT accepted via query string parameter

**File:** `backend-go/middleware/middleware.go:21-22`
**Issue:** `AuthMiddleware` falls back to `c.Query("token")` when the Authorization header is absent. Query-string tokens are recorded in server access logs, browser history, Referer headers, and CDN/proxy logs. This leaks authentication tokens to any logging infrastructure in the request path.
**Fix:** Remove the query-string fallback. Require all authenticated requests to use the `Authorization: Bearer <token>` header. If some clients (e.g., SSE/EventSource) cannot set custom headers, use a short-lived cookie or a separate token-exchange endpoint instead.

### WR-04: CORS allows all origins

**File:** `backend-go/main.go:31-36`
**Issue:** `AllowAllOrigins: true` permits any website to send credentialed-looking requests to the API. Although `AllowCredentials` is false (so cookies are not sent), the permissive CORS policy means any origin can probe the admin endpoints, attempt login brute-force, or enumerate users if they guess or steal a token from another vector.
**Fix:**
```go
r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{config.App.FrontendOrigin}, // add this config field
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: false,
}))
```

### WR-05: Quota check has TOCTOU race condition

**File:** `backend-go/service/auth.go:154-163` and `backend-go/handler/generate.go:48-51`
**Issue:** `CheckQuota` reads `used_count` and compares it to `quota`, then the handler proceeds to generate images. `IncrementUsedCount` runs after generation completes. Between the check and the increment, concurrent requests from the same user all pass the quota check, allowing the user to exceed their quota by the number of concurrent requests.
**Fix:** Atomically check and increment in a single SQL statement, or use a transaction with a row-level lock:
```go
func CheckAndReserveQuota(userID string, count int) error {
    result, err := database.DB.Exec(
        "UPDATE users SET used_count = used_count + ? WHERE id = ? AND (quota = 0 OR used_count + ? <= quota)",
        count, userID, count,
    )
    if err != nil {
        return err
    }
    affected, _ := result.RowsAffected()
    if affected == 0 {
        return fmt.Errorf("配额已用完")
    }
    return nil
}
```

### WR-06: migrateSchema silently discards all errors

**File:** `backend-go/database/database.go:87-93`
**Issue:** Each `DB.Exec` call in `migrateSchema()` discards its error return. This is intentional for idempotent "ALTER TABLE ADD COLUMN" statements (SQLite returns an error if the column already exists), but it also silently swallows real failures such as disk full, database corruption, or permission errors. A genuine migration failure would go unnoticed.
**Fix:**
```go
func migrateSchema() {
    migrations := []string{
        "ALTER TABLE tasks ADD COLUMN api_mode TEXT",
        "ALTER TABLE tasks ADD COLUMN codex_cli INTEGER NOT NULL DEFAULT 0",
        "ALTER TABLE users ADD COLUMN quota INTEGER NOT NULL DEFAULT 0",
        "ALTER TABLE users ADD COLUMN used_count INTEGER NOT NULL DEFAULT 0",
    }
    for _, m := range migrations {
        _, err := DB.Exec(m)
        if err != nil && !strings.Contains(err.Error(), "duplicate column") {
            log.Printf("migration warning: %v", err)
        }
    }
}
```

### WR-07: initAdmin treats all QueryRow errors as "no rows"

**File:** `backend-go/database/database.go:96-108`
**Issue:** `initAdmin()` checks `if err == nil { return nil }` -- any error from `QueryRow.Scan` (including real database errors like corruption or I/O failure) falls through to the INSERT path. If the database is corrupted, the SELECT fails for a real reason, and then the INSERT also fails (possibly with a UNIQUE constraint violation if the row actually exists but was unreadable), producing a confusing error message.
**Fix:**
```go
func initAdmin() error {
    adminHash := util.HashApikey(config.App.AdminApikey)
    var existingID string
    err := DB.QueryRow("SELECT id FROM users WHERE role = ? LIMIT 1", "admin").Scan(&existingID)
    if err == nil {
        return nil // admin exists
    }
    if err != sql.ErrNoRows {
        return fmt.Errorf("查询管理员用户失败: %w", err)
    }
    // ... proceed with INSERT ...
}
```

### WR-08: Error from FindUserByID swallowed after last_login_at update

**File:** `backend-go/service/auth.go:87-88`
**Issue:** After updating `last_login_at`, the code calls `u, _ = FindUserByID(u.ID)` and discards the error. If `FindUserByID` fails (e.g., the user was deleted between the initial query and this re-read), `u` becomes nil. The nil check on line 91 catches this and returns a generic "登录失败", but the actual error is lost, making debugging difficult.
**Fix:**
```go
if err := database.DB.Exec("UPDATE users SET last_login_at = ? WHERE id = ?", now, u.ID); err != nil {
    log.Printf("更新登录时间失败: %v", err)
}
u, err = FindUserByID(u.ID)
if err != nil {
    return "", nil, fmt.Errorf("登录失败: %w", err)
}
```

## Info

### IN-01: Redundant dynamic import in AdminPage

**File:** `src/admin/AdminPage.tsx:14`
**Issue:** The `onLogout` handler uses `import('./adminApi').then(m => m.clearAdminToken())` but `adminApi` is already statically imported on line 2 (`import { isAdminLoggedIn } from './adminApi'`). The dynamic import adds unnecessary async complexity for a module that is already loaded.
**Fix:**
```typescript
import { isAdminLoggedIn, clearAdminToken } from './adminApi'
// ...
return <AdminDashboard onLogout={() => {
  clearAdminToken()
  setLoggedIn(false)
}} />
```

### IN-02: No confirmation dialog for destructive admin actions

**File:** `src/admin/AdminDashboard.tsx:139-174`
**Issue:** Clicking "disable" or "reset" immediately executes the action with no confirmation prompt. An accidental click could disable a user or zero their quota.
**Fix:** Add a `window.confirm()` or a custom confirmation modal before calling `handleToggleStatus` and `handleReset`.

### IN-03: Admin can modify their own account in the dashboard

**File:** `src/admin/AdminDashboard.tsx:108-177` and `backend-go/handler/admin.go:42-77`
**Issue:** The admin user row (created by `initAdmin()`) appears in the users table. The admin can disable their own account or zero their own quota. While this does not lock the admin out (the `AdminMiddleware` checks JWT role, not database user status), it creates confusing state in the database.
**Fix:** Filter out the admin user from the users list in `AdminListUsers`, or disable the toggle/quota controls for the admin's own row in the frontend.

### IN-04: Admin token stored in localStorage

**File:** `src/admin/adminApi.ts:2`
**Issue:** The admin JWT is stored in `localStorage`, which is accessible to any JavaScript running on the page. If an XSS vulnerability exists (now or introduced later), the admin token can be exfiltrated. This is a common pattern but worth documenting as a known risk.
**Fix:** For higher security, consider storing the token in an httpOnly cookie set by the server, or use a short-lived token with refresh rotation. For this project's threat model, localStorage may be acceptable.

---

_Reviewed: 2026-05-08T12:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
