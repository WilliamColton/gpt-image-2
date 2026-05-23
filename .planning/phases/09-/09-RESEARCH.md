# Phase 09: 操作日志 (Audit Log) - Research

**Researched:** 2026-05-23
**Domain:** Audit logging / backend persistence + admin UI
**Confidence:** HIGH

## Summary

Phase 9 adds a synchronous, GORM-backed audit log system to the existing Go backend and a new "操作日志" (Audit Log) tab to the React admin dashboard. The work is predominantly an integration task -- inserting log rows at existing business event sites (auth, generation, admin, quota) and building a filtered table view on the frontend. Zero new external dependencies are required; all work uses the existing stack (GORM, Gin, React/Tailwind).

The backend pattern is well-established: define a GORM model in `database/models.go`, register it in `AutoMigrate` in `database/database.go`, write service functions in `backend-go/service/`, add handlers in `backend-go/handler/`, and register routes in `backend-go/main.go`. The frontend pattern reuses the existing `Tab` enum extension, `adminRequest<T>()` API client, and glass-morphism table/card styling.

**Primary recommendation:** Use the existing `service.AnalyticsRange` time-filter pattern for query parameters, extend the `ConfirmDialog` store for cleanup confirmations, and write a single `LogEvent()` helper function that handlers call inline at each business event site. No goroutines, no channels, no new database.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Audit log write (INSERT) | API / Backend | -- | D-12 synchronous INSERT at business event sites |
| Audit log query + filter | API / Backend | -- | GORM WHERE clauses with query params |
| Audit log display table | Browser / Client | -- | React component, no server-side pagination (D-10) |
| Client-side filtering (UI dropdowns/search) | Browser / Client | -- | Filter state lives in React; sends query params to API |
| Audit log cleanup (DELETE) | API / Backend | -- | DELETE FROM audit_logs with optional age filter |
| Audit log persistence | Database / Storage | -- | SQLite audit_logs table via GORM |
| Audit log retention (manual) | Browser / Client | API / Backend | User triggers cleanup from admin UI |

<phase_requirements>

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| LOG-01 | GORM AuditLog model with standard fields (event_type, user_id, user_label, message, severity, ip, details_json, created_at) | Section: Standard Stack > AuditLog GORM Model |
| LOG-02 | Sync INSERT at 4 event categories: auth, generation, admin, quota | Section: Integration Points > Event Instrumentation Map |
| LOG-03 | Admin API: GET /api/admin/logs with query filters (event_type, severity, user, keyword, range) | Section: API Designs > GET /api/admin/logs |
| LOG-04 | Admin API: DELETE /api/admin/logs for cleanup (clear all, or older than N days) | Section: API Designs > DELETE /api/admin/logs |
| LOG-05 | New "操作日志" Tab in AdminDashboard with table display | Section: Frontend > AdminDashboard Integration |
| LOG-06 | Full filter bar: event type dropdown, severity dropdown, user search, keyword search, time range | Section: Frontend > Filter Bar |
| LOG-07 | Manual cleanup UI with confirm dialog and deleted count feedback | Section: Frontend > Cleanup UI |

</phase_requirements>

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- D-01: Only key business events (auth, generation, admin actions, quota events), NOT HTTP request logging
- D-02: Four event categories: auth, generation, admin operations, quota events
- D-03: Standard fields: event_type, user_id, user_label, message, severity (INFO/WARN/ERROR), IP, details_json, created_at
- D-04: SQLite audit_logs table via GORM, no new storage dependency
- D-05: Manual cleanup only (clear all / delete older than N days), no auto-cleanup
- D-06: No cascade delete when admin deletes users/tasks
- D-07: New "操作日志" tab in AdminDashboard via Tab enum extension
- D-08: Table list display matching existing admin table style
- D-09: Full filters: event type dropdown, severity dropdown, user search, keyword search, time range (today/7d/30d/all)
- D-10: No pagination -- load all logs at once
- D-11: INFO/WARN/ERROR all stored
- D-12: Synchronous INSERT -- no channel/goroutine complexity
- D-13: No rate limiting or sampling

### Claude's Discretion

- AuditLog GORM model field definitions
- details_json per-event-type field structures
- Admin UI layout and styling (glass morphism design system)
- Backend handler/service layering
- Log helper function signatures
- Cleanup API return format (deleted count)

### Deferred Ideas (OUT OF SCOPE)

- HTTP 请求日志入库
- 日志导出 (CSV/PDF)
- 自动清理策略
- 实时日志流 (SSE/WebSocket)
- 日志告警/通知
- 按用户差异化日志保留策略

</user_constraints>

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| gorm.io/gorm | v1.30.0 | AuditLog model + CRUD operations | Already in go.mod, all models use this |
| gorm.io/driver/sqlite | v1.6.0 | SQLite driver | Already in go.mod, existing storage |
| github.com/gin-gonic/gin | v1.10.1 | HTTP handlers for logs API | Already in go.mod, all handlers use this |

### Frontend (existing, no new deps)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| react | ^19.1.0 | UI rendering | Already in package.json |
| zustand | ^5.0.5 | State management (toast, confirmDialog) | Already in package.json |
| @radix-ui/react-select | (in node_modules) | Dropdown selects (severity, event_type filters) | Already used in ui/select.tsx |
| tailwindcss | ^3.4.17 | Glass morphism styling | Already in package.json |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| (none) | -- | -- | No new dependencies needed |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Direct GORM INSERT | Channel-based async writer | CONTEXT D-12 forbids; SQLite write overhead is negligible for this scale |
| Paginated API | Cursor/offset pagination | CONTEXT D-10 forbids; log volume limited by quota system |
| New cleanup service | Cron job / timer-based | CONTEXT D-05 forbids auto-cleanup |

**Installation:** No new packages to install. All dependencies already exist.

**Version verification:**
```bash
# All dependencies already confirmed in go.mod / package.json:
# gorm.io/gorm v1.30.0 -- used by all existing models
# gorm.io/driver/sqlite v1.6.0 -- used by database.Init()
# github.com/gin-gonic/gin v1.10.1 -- used by all existing handlers
# react ^19.1.0 -- AdminDashboard.tsx already uses this
# tailwindcss ^3.4.17 -- AdminDashboard.tsx already uses Tailwind classes
```

## Package Legitimacy Audit

**No new packages are installed in this phase.** All dependencies (gorm, gin, react, tailwindcss, @radix-ui/react-select, zustand) are pre-existing in go.mod, package.json, and node_modules. slopcheck audit is not applicable -- no package additions.

| Package | Registry | Age | Downloads | Source Repo | slopcheck | Disposition |
|---------|----------|-----|-----------|-------------|-----------|-------------|
| (none) | -- | -- | -- | -- | N/A | No new packages |

**Packages removed due to slopcheck [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

## Architecture Patterns

### System Architecture Diagram

```
User Action (web UI / admin UI)
       │
       ▼
  ┌─────────────────────────────────────┐
  │  Gin Handler (auth.go, admin.go,    │
  │  generate.go, feedback.go)          │
  │                                     │
  │  After business logic succeeds:     │
  │    service.LogEvent(ctx) ◄────────┐ │
  └──────────┬─────────────────────────┤ │
             │                         │ │
             ▼                         │ │
  ┌──────────────────────────┐         │ │
  │  service/log.go          │         │ │
  │  LogEvent(params) error  │─────────┘ │
  │  QueryLogs(filters) []*AuditLog     │
  │  CleanLogs(before) int64            │
  └──────────┬──────────────────────────┘
             │ GORM INSERT / SELECT / DELETE
             ▼
  ┌──────────────────────────┐
  │  database/models.go      │
  │  AuditLog struct         │
  └──────────┬───────────────┘
             │ AutoMigrate
             ▼
  ┌──────────────────────────┐
  │  SQLite: audit_logs      │
  └──────────────────────────┘

  ═══════════ Admin UI Path ═══════════

  AdminDashboard (logs tab)
       │
       ├── GET /api/admin/logs?event_type=&severity=&user=&keyword=&range=
       │      └── handler.AdminListLogs → service.QueryLogs → GORM Find
       │
       └── DELETE /api/admin/logs?before=<timestamp>
       │      └── handler.AdminCleanLogs → service.CleanLogs → GORM Delete
       │
       ▼
  Filter bar + Table display + ConfirmDialog for cleanup
```

### Recommended Project Structure

```
backend-go/
├── database/
│   ├── models.go          # + AuditLog struct (GORM model)
│   └── database.go        # + &AuditLog{} in AutoMigrate call
├── service/
│   ├── log.go             # NEW: LogEvent(), QueryLogs(), CleanLogs()
│   └── ... (existing services, no changes)
├── handler/
│   ├── log.go             # NEW: AdminListLogs(), AdminCleanLogs()
│   ├── auth.go            # + LogEvent calls at login/register/migrate/change-password
│   ├── admin.go           # + LogEvent calls at quota/status/delete/config/invite changes
│   ├── generate.go        # + LogEvent calls at task submit/success/failure
│   └── feedback.go        # + LogEvent call at feedback creation
└── main.go                # + GET /api/admin/logs, DELETE /api/admin/logs

src/
├── admin/
│   ├── AdminDashboard.tsx # + 'logs' tab, filter state, loadLogs, cleanup handler
│   └── adminApi.ts        # + adminListLogs(), adminCleanLogs()
└── components/
    └── ConfirmDialog.tsx  # (existing, reused for cleanup confirmation)
```

### Pattern 1: GORM Model Following Existing Convention

**What:** Define AuditLog struct matching existing model patterns (text ID primary key, int64 timestamp, TableName() method).
**When to use:** All new GORM models in this project.
**Example (based on existing Feedback model):**

```go
// Source: backend-go/database/models.go (existing Feedback model pattern)
type AuditLog struct {
    ID          string  `gorm:"primaryKey;type:text"`
    EventType   string  `gorm:"type:text;not null;index"`
    UserID      string  `gorm:"type:text;not null;index"`
    UserLabel   string  `gorm:"type:text;not null"`
    Message     string  `gorm:"type:text;not null"`
    Severity    string  `gorm:"type:text;not null;index;default:INFO"`
    IP          string  `gorm:"type:text;not null;default:''"`
    DetailsJSON string  `gorm:"type:text;not null;default:'{}';column:details_json"`
    CreatedAt   int64   `gorm:"not null;index"`
}

func (AuditLog) TableName() string { return "audit_logs" }
```
[VERIFIED: Codebase pattern -- Feedback, BillingRecord, ChangelogEntry all follow text ID + int64 CreatedAt pattern]

### Pattern 2: Service Layer Log Helper

**What:** A single `LogEvent()` function in `service/log.go` that accepts parameters and performs synchronous GORM INSERT.
**When to use:** Every business event site calls this function after the business logic succeeds (or on known failure like auth failure).
**Example:**

```go
// Source: synthesized from existing service patterns (service/billing.go RecordBillingForSuccessfulImages)
package service

import (
    "gpt-image-playground/backend/database"
    "gpt-image-playground/backend/util"
    "time"
)

type LogEventParams struct {
    EventType   string
    UserID      string
    UserLabel   string
    Message     string
    Severity    string // "INFO", "WARN", "ERROR"
    IP          string
    DetailsJSON string // JSON string of event-specific fields
}

func LogEvent(p LogEventParams) {
    if p.Severity == "" {
        p.Severity = "INFO"
    }
    entry := &database.AuditLog{
        ID:          util.GenerateID(),
        EventType:   p.EventType,
        UserID:      p.UserID,
        UserLabel:   p.UserLabel,
        Message:     p.Message,
        Severity:    p.Severity,
        IP:          p.IP,
        DetailsJSON: p.DetailsJSON,
        CreatedAt:   time.Now().UnixMilli(),
    }
    if err := database.DB.Create(entry).Error; err != nil {
        // Log to slog, do NOT return error to caller (D-12: fire-and-forget)
        slog.Error("审计日志写入失败", "event_type", p.EventType, "error", err)
    }
}
```
[VERIFIED: Codebase pattern -- util.GenerateID() at backend-go/util/id.go, database.DB.Create() pattern from database/database.go initAdmin()]

### Pattern 3: Handler Query Pattern with Multiple Optional Filters

**What:** Gin handler reads query parameters and passes them to a service query function that builds a dynamic GORM WHERE clause.
**When to use:** Any list endpoint with optional filters.
**Example (based on AdminBillingSummary handler pattern):**

```go
// Source: backend-go/handler/admin.go AdminBillingSummary (query param reading pattern)
func AdminListLogs(c *gin.Context) {
    eventType := c.Query("event_type")
    severity  := c.Query("severity")
    user      := c.Query("user")
    keyword   := c.Query("keyword")
    rangeVal  := c.Query("range")

    logs, err := service.QueryLogs(service.LogQuery{
        EventType: eventType,
        Severity:  severity,
        User:      user,
        Keyword:   keyword,
        Range:     rangeVal,
    })
    if err != nil {
        slog.Error("查询审计日志失败", "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "查询审计日志失败"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"logs": logs})
}
```
[VERIFIED: Codebase pattern -- AdminBillingSummary reads `range` query param, AdminListFeedbacks reads `status` query param]

### Anti-Patterns to Avoid

- **Goroutine for log writes:** D-12 explicitly forbids channel/goroutine complexity. SQLite single-writer (SetMaxOpenConns(1)) means async writes are pointless and risk data loss on shutdown.
- **Middleware-based audit logging:** D-01 explicitly forbids HTTP request-level logging. Logging must happen at the business event semantic level, not at the HTTP layer.
- **Mixing slog output with audit_logs table:** Existing `middleware/logger.go` continues logging to stdout via slog. The new `audit_logs` table is a separate persistence layer for admin-viewable business events. They serve different purposes.
- **Adding audit writes inside transactions that can roll back:** If `LogEvent()` is called inside a GORM transaction that later fails, the audit entry should still persist (or be written after transaction commit). Write audit logs after the business transaction succeeds.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| ID generation | Custom UUID/sequence | util.GenerateID() | Already exists, used by all models |
| JSON serialization for details_json | Manual string building | encoding/json (json.Marshal) | Standard Go library, avoids injection |
| Time range parsing | Custom time range logic | service.ParseAnalyticsRange() | Already exists in service/analytics.go, supports today/7d/30d/all |
| Modal dialog for cleanup | Custom modal | store.confirmDialog + ConfirmDialog.tsx | Already exists, used throughout AdminDashboard |
| Toast notifications | Custom implementation | useStore((s) => s.showToast) | Already exists, used throughout AdminDashboard |
| SQL escaping / query building | String concatenation | GORM Where() with ? placeholders | GORM handles parameterized queries, prevents SQL injection |

**Key insight:** The entire phase reuses existing infrastructure. The only new code is (1) the AuditLog GORM model, (2) the LogEvent/QueryLogs/CleanLogs service functions, (3) two handler functions, and (4) the frontend logs tab. Every pattern needed already exists in the codebase.

## Common Pitfalls

### Pitfall 1: GORM AutoMigrate Not Registered

**What goes wrong:** AuditLog table is never created because AutoMigrate only migrates models explicitly listed in database/database.go.
**Why it happens:** The AutoMigrate call in database/database.go has an explicit list: `&User{}, &RedemptionCode{}, &Image{}, &Task{}, &Announcement{}, &Feedback{}, &ChangelogEntry{}, &BillingRecord{}`. New models must be appended.
**How to avoid:** Add `&AuditLog{}` to the AutoMigrate call. This is a one-line change.
**Warning signs:** "no such table: audit_logs" error on first insert.

### Pitfall 2: SQLite Single-Writer Contention

**What goes wrong:** SQLite with SetMaxOpenConns(1) means concurrent writes queue up. If a long-running query happens alongside frequent log writes, the INSERT may block.
**Why it happens:** database/database.go sets MaxOpenConns to 1 for SQLite.
**How to avoid:** D-12 synchronous INSERT means each LogEvent call is a single GORM Create. For this app's scale (quota-limited generation, one admin user viewing logs), contention is negligible. Do not add retry logic or buffering -- it would violate D-12.
**Warning signs:** Would only manifest at very high throughput, which the quota system prevents.

### Pitfall 3: Missing LogEvent Calls at All Integration Points

**What goes wrong:** Some business events don't get logged because the LogEvent call was forgotten.
**Why it happens:** Integration requires touching multiple handler files (auth.go, admin.go, generate.go, feedback.go). It is easy to miss one.
**How to avoid:** Use the Event Instrumentation Map below as a checklist. Every event row must have a corresponding LogEvent call.
**Warning signs:** Test by performing each business action and checking the logs tab.

### Pitfall 4: IP Address Access in Async Context

**What goes wrong:** `executeImageGeneration` runs in a goroutine, where `c.ClientIP()` is no longer available.
**Why it happens:** The Gin context `c` is not safe to use after the handler returns, and `executeImageGeneration` is called via `go executeImageGeneration(...)`.
**How to avoid:** Capture `c.ClientIP()` in the handler before launching the goroutine, and pass it as a parameter to the goroutine function or to service functions that write logs. For task completion/failure events (which happen in the goroutine), pass the IP alongside other captured values (userID, userLabel).
**Warning signs:** Panic or empty IP field in goroutine-written logs.

### Pitfall 5: details_json Not Valid JSON

**What goes wrong:** Manually building JSON strings leads to malformed JSON or injection-like issues.
**Why it happens:** String concatenation is fragile.
**How to avoid:** Always use `json.Marshal` on a Go struct or map and convert to string. Example:
```go
details, _ := json.Marshal(map[string]interface{}{
    "task_id": taskID,
    "delta":   delta,
})
LogEvent(LogEventParams{..., DetailsJSON: string(details)})
```
**Warning signs:** Frontend JSON.parse errors when trying to display details_json.

### Pitfall 6: GORM Column Name Mismatch

**What goes wrong:** GORM generates column names from struct field names by default (snake_case), but the existing codebase uses explicit `column:` tags for JSON-typed fields (e.g., `column:params_json`, `column:billing_records`).
**Why it happens:** Unlike the JSON column fields in Task and BillingRecord which needed explicit `column:` tags, standard fields with simple types work fine with GORM's default naming. However, `details_json` should use an explicit `column:details_json` tag for consistency and clarity.
**How to avoid:** Use `gorm:"column:details_json"` on the DetailsJSON field.
**Warning signs:** GORM looking for `details_json` column but creating `details_json` -- actually GORM's default would be `details_json` anyway (JSON -> json, so DetailsJSON -> details_json). The explicit tag prevents future surprises.

## Runtime State Inventory

**Not applicable.** This is a greenfield feature phase -- no existing state to rename or migrate. The `audit_logs` table does not exist yet and will be created by AutoMigrate.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | Backend build/run | Yes | go1.26.2 | -- |
| Node.js | Frontend build/run | Yes | v22.14.0 | -- |
| npm | Frontend deps | Yes | 11.14.0 | -- |
| SQLite (via GORM) | Audit log storage | Yes | via gorm.io/driver/sqlite v1.6.0 | -- |

**Missing dependencies with no fallback:** None -- all required runtimes are available.
**Missing dependencies with fallback:** None.

## Event Instrumentation Map

This is the critical integration checklist. Each row represents a LogEvent call that must be added in a handler or service function.

### Category: auth (认证事件)

| # | Event | Handler File | Where to Add | Severity | details_json Fields |
|---|-------|-------------|--------------|----------|---------------------|
| A1 | Login success (code) | handler/auth.go | AuthLogin, after service.LoginWithCode succeeds | INFO | `{"method":"code"}` |
| A2 | Login success (password) | handler/auth.go | AuthLoginPassword, after service.LoginWithPassword succeeds | INFO | `{"method":"password"}` |
| A3 | Login failure | handler/auth.go | AuthLogin/AuthLoginPassword, on error return | WARN | `{"method":"code\|password","reason":"<err msg>"}` |
| A4 | Register success | handler/auth.go | AuthRegister, after service.RegisterUser succeeds | INFO | `{"username":"<username>"}` |
| A5 | Register failure | handler/auth.go | AuthRegister, on error return | WARN | `{"username":"<username>","reason":"<err msg>"}` |
| A6 | Migrate success | handler/auth.go | AuthMigrate, after service.MigrateUser succeeds | INFO | `{"migrated_from":"code","username":"<username>"}` |
| A7 | Change password success | handler/auth.go | AuthChangePassword, after service.ChangePassword succeeds | INFO | `{}` |
| A8 | Change password failure | handler/auth.go | AuthChangePassword, on error return | WARN | `{"reason":"<err msg>"}` |
| A9 | Redeem code success | handler/auth.go | AuthRedeem, after service.RedeemForUser succeeds | INFO | `{"code":"<code>","quota_gained":<n>}` |
| A10 | Redeem code failure | handler/auth.go | AuthRedeem, on error return | WARN | `{"code":"<code>","reason":"<err msg>"}` |

### Category: generation (生成事件)

| # | Event | Handler File | Where to Add | Severity | details_json Fields |
|---|-------|-------------|--------------|----------|---------------------|
| G1 | Task submitted | handler/generate.go | GenerateImage, after UpsertTask succeeds (before goroutine) | INFO | `{"task_id":"<id>","prompt":"<truncated>","n":<n>}` |
| G2 | Task completed (success) | handler/generate.go | executeImageGeneration, after task status set to "done" | INFO | `{"task_id":"<id>","output_count":<n>,"elapsed_ms":<ms>,"endpoint":"<url>"}` |
| G3 | Task failed | handler/generate.go | failTask(), after status set to "error" | ERROR | `{"task_id":"<id>","reason":"<err msg>"}` |

### Category: admin (管理操作)

| # | Event | Handler File | Where to Add | Severity | details_json Fields |
|---|-------|-------------|--------------|----------|---------------------|
| M1 | Update user quota | handler/admin.go | AdminUpdateQuota, after service call succeeds | INFO | `{"target_user_id":"<id>","target_user_label":"<label>","mode":"set\|delta","delta":<n>}` |
| M2 | Toggle user status | handler/admin.go | AdminToggleStatus, after service.SetUserStatus succeeds | WARN | `{"target_user_id":"<id>","target_user_label":"<label>","new_status":"<status>"}` |
| M3 | Delete single user | handler/admin.go | AdminDeleteUser, after service.DeleteUser succeeds | WARN | `{"target_user_id":"<id>"}` |
| M4 | Batch delete users | handler/admin.go | AdminDeleteUsers, after service.DeleteUsers succeeds | WARN | `{"deleted_count":<n>}` |
| M5 | Create redemption codes | handler/admin.go | AdminCreateCode, after codes created | INFO | `{"count":<n>,"quota_per_code":<n>}` |
| M6 | Delete redemption codes | handler/admin.go | AdminDeleteCodes, after deletion | WARN | `{"deleted_count":<n>}` |
| M7 | Update endpoints config | handler/admin.go | AdminUpdateEndpoints, after config.SetEndpoints succeeds | WARN | `{"endpoint_count":<n>}` |
| M8 | Update pricing config | handler/admin.go | AdminUpdatePricingConfig, after config.SetPricingConfig succeeds | WARN | `{"sale_price_x10000":<n>}` |
| M9 | Update announcement | handler/announcement.go | AdminUpdateAnnouncement, after service.UpdateAnnouncement succeeds | INFO | `{"enabled":<bool>}` |
| M10 | Reset user password | handler/admin.go | AdminResetPassword, after service.AdminResetPassword succeeds | WARN | `{"target_user_id":"<id>","target_user_label":"<label>"}` |
| M11 | Update invite config | handler/admin.go | AdminUpdateInviteConfig, after config.SetInviteConfig succeeds | INFO | `{"inviter_reward":<n>,"invitee_reward":<n>,"default_quota":<n>}` |
| M12 | Set invite code | handler/auth.go | AuthSetInviteCode, after success | INFO | `{"invite_code":"<code>"}` |
| M13 | Update feedback status | handler/feedback.go | AdminUpdateFeedbackStatus, after success | INFO | `{"feedback_id":"<id>","new_status":"<status>"}` |

### Category: quota (配额事件)

| # | Event | Handler/Service File | Where to Add | Severity | details_json Fields |
|---|-------|---------------------|--------------|----------|---------------------|
| Q1 | Quota consumed | handler/generate.go | executeImageGeneration, after IncrementUsedCount succeeds | INFO | `{"task_id":"<id>","delta":<n>,"used_count_after":<n>}` |
| Q2 | Invite reward granted | service/auth.go | In service.MigrateUser or wherever invite reward is applied | INFO | `{"invited_by":"<user_id>","reward":<n>}` |
| Q3 | Redemption quota granted | handler/auth.go | AuthRedeem, after service.RedeemForUser succeeds | INFO | `{"code":"<code>","quota_gained":<n>}` |

### Admin actor identity note

For admin actions (M1-M13), the actor performing the action is the logged-in admin (user ID from middleware.GetAuthUser(c)). The target user is the user being modified. The LogEvent call should use the admin's ID/Label for `UserID`/`UserLabel`, and put the target user info in `Message` and `DetailsJSON`.

## API Designs

### GET /api/admin/logs

**Query Parameters (all optional):**

| Param | Type | Example | Behavior |
|-------|------|---------|----------|
| event_type | string | `auth`, `generation`, `admin`, `quota` | Filter by event category |
| severity | string | `INFO`, `WARN`, `ERROR` | Filter by severity level |
| user | string | `admin`, `testuser` | Case-insensitive LIKE match on user_label |
| keyword | string | `task`, `login` | LIKE match on message field |
| range | string | `today`, `7d`, `30d`, `all` | Time range filter (default: `all`) |

**Response:**
```json
{
  "logs": [
    {
      "id": "abc123",
      "event_type": "auth",
      "user_id": "user-xyz",
      "user_label": "admin",
      "message": "管理员登录成功",
      "severity": "INFO",
      "ip": "127.0.0.1",
      "details_json": "{\"method\":\"password\"}",
      "created_at": 1716451200000
    }
  ]
}
```

**Backend query logic:**
```go
// service/log.go
func QueryLogs(q LogQuery) ([]database.AuditLog, error) {
    db := database.DB.Model(&database.AuditLog{})
    if q.EventType != "" {
        db = db.Where("event_type = ?", q.EventType)
    }
    if q.Severity != "" {
        db = db.Where("severity = ?", q.Severity)
    }
    if q.User != "" {
        db = db.Where("user_label LIKE ?", "%"+q.User+"%")
    }
    if q.Keyword != "" {
        db = db.Where("message LIKE ?", "%"+q.Keyword+"%")
    }
    if q.Range != "" && q.Range != "all" {
        r, err := ParseAnalyticsRange(q.Range, time.Now())
        if err == nil {
            db = db.Where("created_at BETWEEN ? AND ?", r.From, r.To)
        }
    }
    var logs []database.AuditLog
    // Order by newest first (D-09 requires reverse chronological)
    if err := db.Order("created_at DESC").Find(&logs).Error; err != nil {
        return nil, err
    }
    return logs, nil
}
```
[VERIFIED: Codebase pattern -- service/analytics.go GetBillingSummary uses same WHERE clause building pattern]

### DELETE /api/admin/logs

**Query Parameters:**

| Param | Type | Example | Behavior |
|-------|------|---------|----------|
| before | int64 (unix ms) | `1716364800000` | Delete logs older than this timestamp |
| (no params) | -- | -- | Delete ALL logs (clear everything) |

**Response:**
```json
{
  "ok": true,
  "deleted": 42
}
```

**Backend logic:**
```go
// service/log.go
func CleanLogs(before *int64) (int64, error) {
    db := database.DB.Model(&database.AuditLog{})
    if before != nil {
        db = db.Where("created_at < ?", *before)
    }
    result := db.Delete(&database.AuditLog{})
    return result.RowsAffected, result.Error
}
```

## Frontend

### AdminDashboard Integration

The `Tab` type extends from:
```typescript
type Tab = 'users' | 'codes' | 'config' | 'analytics' | 'announcement' | 'feedback' | 'changelog' | 'invites'
```
to:
```typescript
type Tab = 'users' | 'codes' | 'config' | 'analytics' | 'announcement' | 'feedback' | 'changelog' | 'invites' | 'logs'
```

Following the existing pattern:
1. Add `'logs'` to the `Tab` type union
2. Add a tab button: `<button onClick={() => setTab('logs')} className={...}>操作日志</button>`
3. Add `else if (tab === 'logs') loadLogs()` in the `useEffect` switch
4. Add `{tab === 'logs' && (<>...log table UI...</>)}` render block

### New Admin State

```typescript
// AdminDashboard.tsx -- new state variables
const [logs, setLogs] = useState<AuditLog[]>([])
const [logsLoading, setLogsLoading] = useState(false)
const [logFilter, setLogFilter] = useState<LogFilter>({
  event_type: '',
  severity: '',
  user: '',
  keyword: '',
  range: 'all' as AnalyticsRange,
})
```

### API Client Types

```typescript
// src/admin/adminApi.ts

export interface AuditLog {
  id: string
  event_type: string
  user_id: string
  user_label: string
  message: string
  severity: 'INFO' | 'WARN' | 'ERROR'
  ip: string
  details_json: string
  created_at: number
}

export interface LogFilter {
  event_type?: string
  severity?: string
  user?: string
  keyword?: string
  range?: AnalyticsRange
}

export function adminListLogs(filter: LogFilter): Promise<{ logs: AuditLog[] }> {
  const params = new URLSearchParams()
  if (filter.event_type) params.set('event_type', filter.event_type)
  if (filter.severity) params.set('severity', filter.severity)
  if (filter.user) params.set('user', filter.user)
  if (filter.keyword) params.set('keyword', filter.keyword)
  if (filter.range && filter.range !== 'all') params.set('range', filter.range)
  const qs = params.toString()
  return adminRequest(`/api/admin/logs${qs ? '?' + qs : ''}`)
}

export function adminCleanLogs(before?: number): Promise<{ ok: true; deleted: number }> {
  const qs = before !== undefined ? `?before=${before}` : ''
  return adminRequest(`/api/admin/logs${qs}`, { method: 'DELETE' })
}
```

### Filter Bar

Reuse the analytics `AnalyticsRange` type (`'today' | '7d' | '30d' | 'all'`). The filter bar layout:

```
[事件类型 dropdown] [级别 dropdown] [用户搜索 input] [关键词搜索 input] [时间范围 button group]
```

- Event type dropdown: Radix Select with options: 全部, 认证, 生成, 管理, 配额
- Severity dropdown: Radix Select with options: 全部, INFO, WARN, ERROR
- User search: Input component, debounced (or filtered on API call after Enter/blur)
- Keyword search: Input component, same pattern
- Time range: Button group matching analytics tab pattern (今日/7天/30天/全部)

### Table Display

Columns: 事件类型 | 消息 | 用户 | 级别 | 时间 | IP

Use badge component for severity (green/amber/red matching INFO/WARN/ERROR) and event_type (distinct colors per category). Match the existing admin table styling: `rounded-2xl border border-white/10 bg-white/[0.03]` container, `border-b border-white/5 hover:bg-white/[0.02]` rows.

### Cleanup UI

Two buttons in the logs tab header area:
1. "清空全部日志" -- triggers `confirmDialog` action that calls `adminCleanLogs()`
2. "删除 N 天前的日志" -- input + confirm flow, calls `adminCleanLogs(beforeTimestamp)`

Use the existing `useStore((s) => s.setConfirmDialog)` pattern. After cleanup, show `toast(`已删除 ${result.deleted} 条日志`, 'success')`.

## Code Examples

### LogEvent call from auth handler (login success)

```go
// Source: synthesized from handler/auth.go AuthLogin pattern + service/log.go LogEvent
// Inside AuthLogin, after successful login:
service.LogEvent(service.LogEventParams{
    EventType:   "auth",
    UserID:      user.ID,
    UserLabel:   user.Label,
    Message:     "用户登录成功（兑换码）",
    Severity:    "INFO",
    IP:          c.ClientIP(),
    DetailsJSON: `{"method":"code"}`,
})
```

### LogEvent call from admin handler (quota change)

```go
// Source: synthesized from handler/admin.go AdminUpdateQuota pattern
// Inside AdminUpdateQuota, after successful quota update:
actor := middleware.GetAuthUser(c)
details, _ := json.Marshal(map[string]interface{}{
    "target_user_id":    userID,
    "target_user_label": targetLabel,  // need to fetch or pass along
    "mode":              body.Mode,
    "delta":             body.Delta,
})
service.LogEvent(service.LogEventParams{
    EventType:   "admin",
    UserID:      actor.ID,
    UserLabel:   actor.Label,
    Message:     fmt.Sprintf("修改用户配额：%s %s %d", targetLabel, body.Mode, body.Delta),
    Severity:    "INFO",
    IP:          c.ClientIP(),
    DetailsJSON: string(details),
})
```

### Load logs with filters (React)

```typescript
// Source: synthesized from AdminDashboard.tsx loadUsers / loadAnalyticsSummary pattern
const loadLogs = useCallback(async () => {
  setLogsLoading(true)
  try {
    const { logs } = await adminListLogs(logFilter)
    setLogs(logs || [])
  } catch (err) {
    toast(err instanceof Error ? err.message : String(err), 'error')
  } finally {
    setLogsLoading(false)
  }
}, [logFilter, toast])
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| No audit log | GORM sync INSERT into audit_logs | Phase 9 (now) | New capability |

**Deprecated/outdated:** Not applicable -- this is a new feature.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | util.GenerateID() is the correct ID generator for AuditLog entries | Standard Stack > Pattern 2 | Low -- it is used by all existing models. Trivially verifiable by reading backend-go/util/id.go. |
| A2 | service.ParseAnalyticsRange() can be reused for log time filtering | API Designs > GET /api/admin/logs | Low -- the function is public, tested in analytics_test.go, and accepts the same range values (today/7d/30d/all). |
| A3 | The admin actor identity for admin audit events should be the admin performing the action, not the target user | Event Instrumentation Map > Admin actor identity | Low -- this is standard audit log practice and consistent with D-03's user_id field meaning "who performed the action". |
| A4 | details_json stored as JSON text string (not GORM JSON type) is acceptable since SQLite has no native JSON column type | Standard Stack > AuditLog GORM Model | Low -- existing codebase uses `type:text` for JSON fields (params_json, etc.) and handles serialization in Go. |
| A5 | The ConfirmDialog component (via store.confirmDialog) can be reused for cleanup confirmation | Frontend > Cleanup UI | Low -- it is already used for delete user, delete changelog, batch delete users/codes confirmations in AdminDashboard.tsx. |
| A6 | For task completion/failure events in executeImageGeneration goroutine, IP must be captured in the handler and passed as a parameter | Common Pitfalls > Pitfall 4 | Medium -- if IP is not captured before the goroutine, the logs will have empty IP field. The current executeImageGeneration signature does not accept an IP parameter, so the function signature must be extended. |

## Open Questions (RESOLVED)

1. **Should the LogEvent helper silently swallow errors or return them?** RESOLVED: Silently log to slog on failure, do not return errors to callers. Audit logging is best-effort observability, not a transactional requirement. Confirmed by Plan 09-01 Task 2 — LogEvent returns void and uses slog.Error on failure.

2. **For admin events, how do we get the target user's label?** RESOLVED: Log target user_id only in details_json. If label is readily available from existing service return values, include it. Avoid extra DB queries just for logging. Confirmed by Plan 09-02 Task 2 — admin events use c.Param("id") for target_user_id in details_json.

3. **Event type values in Go and frontend -- shared constants or duplicated strings?** RESOLVED: Define in Go only. Frontend dropdown values are hardcoded strings ("auth"/"generation"/"admin"/"quota") matching Go constants. This is the existing project pattern (severity, status, category strings are all duplicated rather than shared). Confirmed by Plan 09-03 Task 2 — event_type dropdown uses hardcoded values.

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | vitest ^4.1.5 (frontend) + Go testing (backend) |
| Config file | vitest is in package.json scripts; Go tests use standard `go test` |
| Quick run command | `npx vitest run` (frontend) / `go test ./backend-go/service/ -run TestLog -count=1` (backend) |
| Full suite command | `npx vitest run && cd backend-go && go test ./... -count=1` |

### Phase Requirements -> Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| LOG-01 | AuditLog model AutoMigrate creates table | integration | `go test ./backend-go/service/ -run TestAuditLog -count=1` | No -- Wave 0 |
| LOG-02 | LogEvent writes to audit_logs table | unit | `go test ./backend-go/service/ -run TestLogEvent -count=1` | No -- Wave 0 |
| LOG-03 | QueryLogs filters by event_type, severity, user, keyword, range | unit | `go test ./backend-go/service/ -run TestQueryLogs -count=1` | No -- Wave 0 |
| LOG-04 | CleanLogs deletes all or by age | unit | `go test ./backend-go/service/ -run TestCleanLogs -count=1` | No -- Wave 0 |
| LOG-05 | AdminListLogs handler returns JSON | integration | `go test ./backend-go/handler/ -run TestAdminListLogs -count=1` | No -- Wave 0 |
| LOG-06 | AdminCleanLogs handler returns deleted count | integration | `go test ./backend-go/handler/ -run TestAdminCleanLogs -count=1` | No -- Wave 0 |
| LOG-07 | adminListLogs() API client builds correct query string | unit | `npx vitest run src/admin/adminApi` | No -- Wave 0 |
| LOG-08 | Logs tab renders filter bar and table | component | `npx vitest run src/admin/AdminDashboard` | No -- Wave 0 |

### Sampling Rate

- **Per task commit:** `go test ./backend-go/service/ -run TestLog -count=1`
- **Per wave merge:** `go test ./backend-go/... -count=1 && npx vitest run`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `backend-go/service/log_test.go` -- covers LOG-01 through LOG-04 (AuditLog CRUD, LogEvent, QueryLogs with filters, CleanLogs)
- [ ] `backend-go/handler/log_test.go` -- covers LOG-05 and LOG-06 (AdminListLogs, AdminCleanLogs HTTP handlers)
- [ ] `src/admin/adminApi.test.ts` -- covers LOG-07 (API client query string building)
- [ ] `src/admin/AdminDashboard.test.tsx` -- covers LOG-08 (logs tab rendering, existing test file needs expansion with logs tab test cases)
- [ ] Framework install: none needed (vitest and Go testing already configured)

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|------------------|
| V2 Authentication | No | Audit log is read-only view of auth events, not auth itself |
| V3 Session Management | No | Not applicable to audit log |
| V4 Access Control | Yes | AdminMiddleware protects both GET and DELETE /api/admin/logs routes |
| V5 Input Validation | Yes | GORM parameterized queries prevent SQL injection; query param validation in handler |
| V6 Cryptography | No | No cryptographic operations in audit log feature |

### Known Threat Patterns for Go + GORM + SQLite

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| SQL injection via LIKE query params | Tampering | GORM `Where("user_label LIKE ?", "%"+user+"%")` uses parameterized queries -- the `?` placeholder prevents injection |
| Unauthorized log access | Information Disclosure | AdminMiddleware on all `/api/admin/logs` routes ensures only authenticated admins can view/delete logs |
| Log forging (injecting fake log entries via user-controlled fields) | Spoofing | All log fields except `details_json` are server-set (user_id from JWT, IP from gin context, message is server-composed). User input only appears in `details_json` values, not in primary fields |
| Mass log deletion by malicious admin | Tampering | Intentional -- D-05 allows admins to clear logs. The ConfirmDialog prevents accidental deletion. Audit log deletion itself should be logged (meta-audit) |
| Sensitive data in details_json | Information Disclosure | DetailsJSON should not contain passwords, API keys, or full auth tokens. Only operational metadata (task IDs, quota deltas, endpoint counts) |

## Sources

### Primary (HIGH confidence)
- Codebase: `backend-go/database/models.go` -- verified existing GORM model patterns (text ID, int64 timestamps, TableName(), column: tags)
- Codebase: `backend-go/database/database.go` -- verified AutoMigrate registration pattern (line 28)
- Codebase: `backend-go/main.go` -- verified admin route registration pattern (lines 90-117)
- Codebase: `backend-go/handler/admin.go` -- verified admin handler patterns (query params, JSON response)
- Codebase: `backend-go/handler/generate.go` -- verified generation success/failure integration points (lines 32-195)
- Codebase: `backend-go/service/analytics.go` -- verified ParseAnalyticsRange() and query building pattern
- Codebase: `backend-go/middleware/middleware.go` -- verified AdminMiddleware and GetAuthUser
- Codebase: `src/admin/AdminDashboard.tsx` -- verified Tab enum, tab switching, loaders, table rendering, ConfirmDialog usage
- Codebase: `src/admin/adminApi.ts` -- verified adminRequest<T>() pattern and API client structure
- Codebase: `src/components/ConfirmDialog.tsx` -- verified confirm dialog API
- Codebase: `src/store.ts` -- verified confirmDialog and showToast store interfaces
- Codebase: `src/components/ui/select.tsx` -- verified Radix Select component availability
- Codebase: `src/components/ui/badge.tsx` -- verified Badge component variants (success, warning, destructive)

### Secondary (MEDIUM confidence)
- CONTEXT.md Phase 09 decisions (D-01 through D-13) -- user-confirmed locked decisions

### Tertiary (LOW confidence)
- None. All claims are verified against the codebase or locked by user decisions.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- Zero new dependencies. All required libraries (GORM v1.30.0, Gin v1.10.1, React 19, Tailwind 3.4) are verified in go.mod and package.json.
- Architecture: HIGH -- The layered architecture (handler -> service -> GORM model) is directly observed in existing code. Every pattern (AutoMigrate, route registration, adminRequest, confirmDialog) has concrete examples in the codebase.
- Pitfalls: HIGH -- Six specific pitfalls identified based on actual code patterns (AutoMigrate registration, SQLite single-writer, async IP capture, JSON marshaling, GORM column naming, missing integration points).
- Event instrumentation: HIGH -- Integration points mapped by reading every handler file (auth.go, admin.go, generate.go, feedback.go, announcement.go). 24 specific LogEvent call sites identified with event types, severity levels, and details_json field specifications.

**Research date:** 2026-05-23
**Valid until:** 2026-06-23 (30 days -- this is stable internal infrastructure with no external dependencies)
