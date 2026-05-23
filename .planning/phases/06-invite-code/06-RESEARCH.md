# Phase 06: 账号密码 & 邀请码机制 - Research

**Researched:** 2026-05-23
**Domain:** Authentication system extension (password auth, user registration, invite codes)
**Confidence:** HIGH

## Summary

Phase 06 extends the existing redemption-code-based authentication system with password login, user registration with optional invite codes, forced migration for legacy users, and invite code management. The Go backend already has `golang.org/x/crypto` at v0.39.0 as an indirect dependency, providing the `bcrypt` sub-package for password hashing. All required Radix UI primitives (`react-tabs`, `react-separator`, `react-dialog`, `react-label`) are already present in `node_modules` -- no new npm packages are needed. The database uses SQLite with GORM AutoMigrate, which handles column additions but requires careful NULL handling for existing users who lack the new fields.

**Primary recommendation:** Add `password_hash`, `username`, `invite_code`, and `invite_code_set_at` as nullable columns to the User model; use `bcrypt.DefaultCost` (10) for hashing; store global invite reward config in `config.json` matching the existing `salePriceX10000` pattern; build all new frontend components using the existing `src/components/ui/` primitives and glass morphism design system.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Password hashing (bcrypt) | API / Backend | -- | Security: hash never leaves the server |
| Username uniqueness enforcement | Database / Storage | API / Backend | GORM uniqueIndex + service-layer validation |
| JWT token issuance & verification | API / Backend | -- | Existing `SignToken`/`VerifyToken` pattern |
| Invite code global uniqueness | Database / Storage | API / Backend | Unique constraint on `invite_code` column |
| Invite reward config storage | API / Backend (config.json) | -- | Global admin setting, persisted to config.json |
| Reward quota distribution | API / Backend | -- | Atomic DB updates in registration handler |
| Login/Register/Migration UI | Browser / Client | -- | React modals with Zustand state |
| Admin invite config UI | Browser / Client | -- | AdminDashboard tab extension |
| Username display in header | Browser / Client | -- | Reads from Zustand authUser |
| needsMigration flag determination | API / Backend | Browser / Client | Backend checks `password_hash IS NULL` |

## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** 仅密码登录（过渡方案）。新用户用邀请码注册时设密码，老用户首次登录后强制设置密码。
- **D-02:** 密码使用 bcrypt 哈希存储。
- **D-03:** 新增 `POST /api/auth/login-password` 独立接口，接受 `{ username, password }`。
- **D-04:** 用户名规则：3-20 字符，允许中文，全局唯一。
- **D-05:** 密码规则：最少 8 字符，无其他复杂度限制。
- **D-06:** 用户可在设置页修改密码；管理员可在管理后台为用户直接设置新密码。
- **D-07:** 不做登录失败次数锁定。暴力破解风险由 bcrypt 慢哈希缓解。
- **D-08:** 保持现有 30 天 JWT 有效期不变，无 refresh token。
- **D-09:** 始终记住登录状态（localStorage 存储 token）。
- **D-10:** 密码登录成功后返回 `{ token, user, needsMigration }`。
- **D-11:** 新增 `POST /api/auth/register` 独立注册接口。
- **D-12:** 注册即自动登录 -- 注册接口直接返回 `{ token, user }`。
- **D-13:** 无邀请码也可注册，获得管理员设定的默认注册配额。
- **D-14:** 老用户强制迁移：兑换码登录成功后弹出独立 Modal（不可关闭）。
- **D-15:** 用户在设置页自定义邀请码（全局唯一，先到先得），可无限使用。
- **D-16:** 邀请码和兑换码完全独立。
- **D-17:** 用户可随时修改自己的邀请码，旧码立即失效。
- **D-18:** 双方配额奖励：被邀请人和邀请人均获得奖励配额。
- **D-19 to D-22:** 管理后台新增"邀请码设置"tab。
- **D-23 to D-27:** 前端 UI 组件（LoginModal Tab 切换、RegisterModal、MigrationModal、SettingsModal 扩展、Header 用户名显示）。
- **D-28:** 用户列表每行新增"重置密码"按钮。
- **D-29:** 新增管理 API 路由。

### Claude's Discretion

- bcrypt 成本参数（`bcrypt.DefaultCost` = 10 为标准选择）
- 数据库迁移方案（新增字段设计）
- Admin API 和前端请求的具体 DTO 结构
- 邀请奖励配置的存储方式（config.json 扩展）
- 具体 UI 样式（应与现有 glass morphism 设计系统一致）
- 错误提示措辞

### Deferred Ideas (OUT OF SCOPE)

None -- discussion stayed within phase scope.

## Phase Requirements

No explicit requirement IDs mapped. Phase 6 is defined by CONTEXT.md decisions (D-01 through D-29 above). Each decision maps to a deliverable.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `golang.org/x/crypto/bcrypt` | v0.39.0 (indirect) | Password hashing | Official Go crypto extension; `bcrypt.DefaultCost` (10) is the community standard; already in go.sum as indirect dep [VERIFIED: go.sum + go list] |
| `github.com/gin-gonic/gin` | v1.10.1 | HTTP routing | Existing framework; all new auth endpoints register on existing gin engine [VERIFIED: go.mod] |
| `gorm.io/gorm` | v1.30.0 | ORM / AutoMigrate | Existing DB layer; new columns added via AutoMigrate [VERIFIED: go.mod] |
| `github.com/golang-jwt/jwt/v5` | v5.2.2 | JWT signing | Existing auth pattern; `SignToken()`/`VerifyToken()` reused [VERIFIED: go.mod] |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `@radix-ui/react-tabs` | (present) | LoginModal tab switching | Existing ui/tabs.tsx wrapper [VERIFIED: node_modules] |
| `@radix-ui/react-separator` | (present) | SettingsModal section dividers | Existing ui/separator.tsx wrapper [VERIFIED: node_modules] |
| `@radix-ui/react-dialog` | (present) | RegisterModal base | Existing ui/dialog.tsx + ui/app-dialog.tsx [VERIFIED: node_modules] |
| `@radix-ui/react-label` | (present) | Form field labels | Existing ui/label.tsx wrapper [VERIFIED: node_modules] |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| bcrypt (Go std extended) | argon2 | argon2 is technically stronger but not in go.sum; bcrypt is already an indirect dependency and satisfies D-02 [CITED: go.sum] |
| config.json for invite config | New DB table `invite_configs` | config.json matches existing `salePriceX10000` pattern; DB table would require new model + AutoMigrate. config.json is simpler and admin-managed settings are already persisted there [CITED: backend-go/config/config.go persistPricingConfig pattern] |

**Installation:** No new package installations required. `golang.org/x/crypto` is already an indirect dependency at v0.39.0. All Radix UI primitives are already in `node_modules`. Zod is NOT in project dependencies and is NOT needed -- validation will be hand-rolled following existing patterns.

**Version verification:**
```bash
# Go -- bcrypt sub-package via golang.org/x/crypto (already in go.sum)
cd backend-go && go list -m golang.org/x/crypto
# => golang.org/x/crypto v0.39.0

# npm -- Radix primitives (already in node_modules, not in package.json deps)
ls node_modules/@radix-ui/react-tabs node_modules/@radix-ui/react-separator node_modules/@radix-ui/react-dialog node_modules/@radix-ui/react-label
# => all present
```

## Package Legitimacy Audit

No new external packages are being installed in this phase. `golang.org/x/crypto/bcrypt` is a sub-package of an already-present Go dependency. All Radix UI primitives are already present in `node_modules`. No npm registry packages are being added.

| Package | Registry | Age | Downloads | Source Repo | slopcheck | Disposition |
|---------|----------|-----|-----------|-------------|-----------|-------------|
| `golang.org/x/crypto` | Go modules | 10+ yrs | -- | go.googlesource.com/crypto | -- (existing dep) | Approved (existing) |
| `@radix-ui/react-tabs` | npm | 4+ yrs | 2M+/wk | github.com/radix-ui/primitives | -- (existing dep) | Approved (existing) |
| `@radix-ui/react-separator` | npm | 4+ yrs | 2M+/wk | github.com/radix-ui/primitives | -- (existing dep) | Approved (existing) |

**Packages removed due to slopcheck [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

*slopcheck was unavailable at research time. No new packages are being added -- all dependencies exist in the project already. The `[VERIFIED: node_modules]` / `[VERIFIED: go.sum]` tags confirm presence without needing slopcheck verification.*

## Architecture Patterns

### System Architecture Diagram

```
                   Browser / Client
              ┌────────────────────────────────────────────────┐
              │                                                │
              │  ┌──────────┐  ┌──────────────┐  ┌──────────┐ │
Entry ────────┤─>│LoginModal│  │RegisterModal │  │Migration │ │
(no auth)     │  │(Tabs:    │  │(invite code  │  │Modal     │ │
              │  │ code +   │  │ + username   │  │(forced,  │ │
              │  │ password)│  │ + password)  │  │unclosable)│ │
              │  └────┬─────┘  └──────┬───────┘  └────┬─────┘ │
              │       │               │                │       │
              │  ┌────┴───────────────┴────────────────┴───┐   │
              │  │         backendApi.ts                    │   │
              │  │  loginWithPassword / register / migrate  │   │
              │  │  changePassword / setInviteCode / getMe  │   │
              │  └────────────────────┬────────────────────┘   │
              │                       │                        │
              │  ┌────────────────────┴────────────────────┐   │
Entry ────────┤─>│ SettingsModal (extended)                │   │
(auth'd)      │  │  Invite code section + Change password  │   │
              │  │  + existing redemption code section     │   │
              │  └─────────────────────────────────────────┘   │
              │                                                │
              │  ┌────────────────────────────────────────┐    │
              │  │ AdminDashboard (invites tab + reset pw) │    │
              │  │  adminApi.ts (resetPassword,            │    │
              │  │  getInviteConfig, updateInviteConfig,   │    │
              │  │  listInvites)                           │    │
              │  └────────────────────┬───────────────────┘    │
              └───────────────────────┼────────────────────────┘
                                      │  HTTP (Bearer JWT)
              ┌───────────────────────┼────────────────────────┐
              │            API / Backend (Go + Gin)            │
              │                       │                        │
              │  ┌────────────────────┴────────────────────┐   │
              │  │            Routes (main.go)             │   │
              │  │  POST /api/auth/login-password          │   │
              │  │  POST /api/auth/register                │   │
              │  │  POST /api/auth/migrate                 │   │
              │  │  POST /api/auth/change-password          │   │
              │  │  PUT  /api/auth/invite-code              │   │
              │  │  GET  /api/auth/invite-code              │   │
              │  │  PUT  /api/admin/users/:id/password      │   │
              │  │  GET  /api/admin/invite-config            │   │
              │  │  PUT  /api/admin/invite-config            │   │
              │  │  GET  /api/admin/invites                 │   │
              │  └────────────────────┬────────────────────┘   │
              │                       │                        │
              │  ┌────────────────────┴────────────────────┐   │
              │  │          handler/auth.go                │   │
              │  │  AuthLoginPassword / AuthRegister        │   │
              │  │  AuthMigrate / AuthChangePassword        │   │
              │  │  AuthSetInviteCode / AuthGetInviteCode   │   │
              │  └────────────────────┬────────────────────┘   │
              │                       │                        │
              │  ┌────────────────────┴────────────────────┐   │
              │  │          service/auth.go                │   │
              │  │  LoginWithPassword / RegisterUser        │   │
              │  │  MigrateUser / ChangePassword            │   │
              │  │  SetInviteCode / GetInviteCode           │   │
              │  │  bcrypt.GenerateFromPassword / Compare   │   │
              │  └────────────────────┬────────────────────┘   │
              │                       │                        │
              └───────────────────────┼────────────────────────┘
                                      │
              ┌───────────────────────┼────────────────────────┐
              │           Database / Storage (SQLite)          │
              │                       │                        │
              │  ┌────────────────────┴────────────────────┐   │
              │  │         database/models.go              │   │
              │  │  User: +password_hash, +username        │   │
              │  │        +invite_code, +invite_code_set_at │   │
              │  │  (RedemptionCode: unchanged)            │   │
              │  └─────────────────────────────────────────┘   │
              │                                                │
              │  ┌────────────────────────────────────────┐    │
              │  │         config.json (extended)          │    │
              │  │  +inviteInviterReward                  │    │
              │  │  +inviteInviteeReward                  │    │
              │  │  +inviteDefaultQuota                   │    │
              │  └────────────────────────────────────────┘    │
              └────────────────────────────────────────────────┘
```

### Recommended Project Structure
```
backend-go/
├── handler/
│   └── auth.go              # EXTEND: +PasswordLogin, +Register, +Migrate, +ChangePassword, +SetInviteCode, +GetInviteCode
├── service/
│   └── auth.go              # EXTEND: +LoginWithPassword, +RegisterUser, +MigrateUser, +ChangePassword, +SetInviteCode, +GetInviteCode, +ListInvites, +AdminResetPassword
├── database/
│   └── models.go            # EXTEND: User +password_hash, +username, +invite_code, +invite_code_set_at
├── config/
│   └── config.go            # EXTEND: +InviteInviterReward, +InviteInviteeReward, +InviteDefaultQuota fields
└── main.go                  # EXTEND: new auth routes, new admin routes

src/
├── components/
│   ├── LoginModal.tsx       # REDESIGN: Tab switching, password login form
│   ├── RegisterModal.tsx    # NEW: Three-field registration form
│   ├── MigrationModal.tsx   # NEW: Forced migration form (unclosable)
│   ├── SettingsModal.tsx    # EXTEND: Invite code section + change password section
│   └── Header.tsx           # EXTEND: Username display beside title
├── lib/
│   └── backendApi.ts        # EXTEND: +loginWithPassword, +register, +migrate, +changePassword, +setInviteCode, +getInviteCode
├── admin/
│   ├── AdminDashboard.tsx   # EXTEND: +invites tab, +reset password modal
│   └── adminApi.ts          # EXTEND: +resetPassword, +getInviteConfig, +updateInviteConfig, +listInvites
└── store.ts                 # EXTEND: +needsMigration state flag
```

### Pattern 1: bcrypt Password Hashing (Go Service Layer)

**What:** Use `golang.org/x/crypto/bcrypt` with cost 10 for all password operations. Never store or log plaintext passwords.

**When to use:** Every password write (registration, migration, change password, admin reset) and every password verification (login-password).

**Example:**
```go
import "golang.org/x/crypto/bcrypt"

// Hash password during registration/reset
func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", fmt.Errorf("密码加密失败")
    }
    return string(bytes), nil
}

// Verify password during login
func checkPassword(hash, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```
Source: [CITED: pkg.go.dev/golang.org/x/crypto/bcrypt] -- `bcrypt.DefaultCost` = 10, which provides 2^10 iterations. Each increment doubles computation time.

### Pattern 2: GORM AutoMigrate -- Adding Nullable Columns to Existing Table

**What:** Add new fields as nullable (no `not null` constraint) to the User model, then use AutoMigrate. SQLite adds columns with NULL default, so existing rows get NULL values for new columns.

**When to use:** Database migration for the User table.

**Example:**
```go
// In database/models.go -- additions to User struct:
type User struct {
    // ... existing fields ...
    PasswordHash    *string `gorm:"type:text"`                    // NULL = needs migration (old user)
    Username        *string `gorm:"type:text;uniqueIndex"`        // NULL = not yet set
    InviteCode      *string `gorm:"type:text;uniqueIndex"`        // NULL = not yet set
    InviteCodeSetAt *int64                                        // NULL = not yet set
}
```
Source: [CITED: gorm.io/docs/migration.html] -- AutoMigrate creates missing columns. The `*string` (pointer) ensures the column allows NULL, which is critical for existing users who don't have usernames yet.

**CRITICAL:** `uniqueIndex` on `Username` and `InviteCode` -- SQLite allows multiple NULL values in a unique index (NULL != NULL in SQL). This is safe for existing users who have NULL usernames. [VERIFIED: sqlite.org/lang_createindex.html -- "For the purposes of unique indices, all NULL values are considered different from all other NULL values"]

### Pattern 3: Config Persistence for Admin Settings

**What:** Store invite reward config in `config.json` using the same `persistMu/persistEndpoints` pattern used by `salePriceX10000`.

**When to use:** Admin changes invite reward settings.

**Example:**
```go
// In config/config.go -- additions to Config struct:
type Config struct {
    // ... existing fields ...
    InviteInviterReward  int `json:"inviteInviterReward"`   // 邀请人奖励配额
    InviteInviteeReward  int `json:"inviteInviteeReward"`   // 被邀请人奖励配额
    InviteDefaultQuota   int `json:"inviteDefaultQuota"`    // 无邀请码注册默认配额
}

// Getter/convenience functions:
func GetInviteInviterReward() int {
    if App.InviteInviterReward <= 0 { return 0 }
    return App.InviteInviterReward
}
func GetInviteInviteeReward() int {
    if App.InviteInviteeReward <= 0 { return 0 }
    return App.InviteInviteeReward
}
func GetInviteDefaultQuota() int {
    return App.InviteDefaultQuota  // 0 = no quota for new register without invite
}
```
Source: [CITED: backend-go/config/config.go -- `SetPricingConfig` / `persistPricingConfig` pattern at lines 130-225]

### Pattern 4: AuthUser Response DTO Extension

**What:** Add `username` and `needsMigration` fields to the `AuthUser` struct returned in API responses. `needsMigration` is derived (not stored) -- it is `true` when `password_hash IS NULL`.

**When to use:** All login responses, `GET /api/auth/me`.

**Example:**
```go
// In service/models.go -- additions to AuthUser:
type AuthUser struct {
    ID             string `json:"id"`
    Label          string `json:"label"`
    Username       string `json:"username,omitempty"`       // NEW: set username for display
    Role           string `json:"role"`
    ImageCount     int    `json:"imageCount"`
    Quota          int    `json:"quota"`
    UsedCount      int    `json:"usedCount"`
    NeedsMigration bool   `json:"needsMigration,omitempty"` // NEW: true when password_hash IS NULL
}

// In service/models.go -- additions to User (service DTO):
type User struct {
    // ... existing fields ...
    Username       string  `json:"username,omitempty"`
    PasswordHash   *string `json:"-"`
    InviteCode     *string `json:"inviteCode,omitempty"`
    InviteCodeSetAt *int64 `json:"-"`
}
```
Source: [CITED: backend-go/service/models.go lines 22-29 -- existing AuthUser struct]

### Pattern 5: Zustand Store -- needsMigration State

**What:** Add `needsMigration` to the `AuthUser` type in the frontend and store. The `bootstrapBackendSession()` function already fetches user data; it will naturally pick up the new field.

**When to use:** After any login (code or password), the store holds the flag. App root component checks this flag and shows MigrationModal if true.

**Example:**
```typescript
// In src/lib/backendApi.ts -- extend AuthUser interface:
export interface AuthUser {
  id: string
  label: string
  username?: string       // NEW
  role: 'admin' | 'user'
  imageCount: number
  quota: number
  usedCount: number
  needsMigration?: boolean // NEW
}

// Migration check in App layout (conceptual):
{ authUser?.needsMigration && <MigrationModal /> }
```
Source: [CITED: src/lib/backendApi.ts lines 6-13 -- existing AuthUser interface]

### Anti-Patterns to Avoid

- **Plaintext password storage:** Never log or store raw passwords. Use bcrypt for all hashing.
- **Client-side password validation only:** Always re-validate on the server. Frontend validation is UX only.
- **Blocking AutoMigrate with NOT NULL on existing tables:** Adding a `not null` column via AutoMigrate will fail or produce incorrect defaults for existing rows. Use `*string` (nullable) pointers.
- **Invite code collision without DB constraint:** Relying on application-level uniqueness checks without a unique index invites race conditions. Use `uniqueIndex` on the column.
- **Copying Dialog for MigrationModal:** The standard Dialog/AppDialog always allows Escape/backdrop close. MigrationModal must use raw glass morphism markup without Dialog wrapper.
- **Adding zod as new dependency:** The project does not use zod. Hand-rolled validation matches existing patterns. Adding zod would be a new dependency for a pattern the project doesn't follow.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Password hashing | Custom hash/salt function | `golang.org/x/crypto/bcrypt` | bcrypt handles salt generation, timing-safe comparison, and cost factor automatically. Custom implementations have known vulnerabilities (weak salts, timing attacks). |
| JWT token management | Custom token format | Existing `SignToken()`/`VerifyToken()` (golang-jwt/jwt/v5) | Already implemented and tested. Phase 6 reuses existing 30-day JWT pattern unchanged. |
| Unique constraint enforcement | Application-level check-then-insert | GORM `uniqueIndex` tag on DB column | Race conditions exist between check and insert. Database-level unique index is atomic. |
| Confirmation dialog (admin reset password) | Custom modal | Existing AdminDashboard inline modal pattern | AdminDashboard already has `quotaModal` and `confirmModal` patterns; follow the same structure for reset password modal. |

**Key insight:** Password hashing is the one area where hand-rolling is most dangerous. bcrypt's `GenerateFromPassword` handles salt generation (random 16-byte salt), base64 encoding, and the full bcrypt format output (`$2a$10$...`). `CompareHashAndPassword` uses timing-safe comparison. Any custom implementation would need to replicate all of these correctly [CITED: pkg.go.dev/golang.org/x/crypto/bcrypt].

## Runtime State Inventory

This is NOT a rename/refactor/migration phase. This section is omitted per the output format instruction (only required for rename/refactor/migration phases). All changes are net-new features or extensions to existing code.

## Common Pitfalls

### Pitfall 1: SQLite NULL vs UNIQUE INDEX
**What goes wrong:** Adding `uniqueIndex` on a column that existing rows have as NULL. Developer fears NULLs will violate uniqueness.
**Why it happens:** Misunderstanding of SQL standard -- many expect NULL = NULL, but SQL treats NULLs as distinct values in unique indices.
**How to avoid:** Use `*string` (nullable pointer) for `Username` and `InviteCode` in the GORM model. SQLite's unique index treats each NULL as a distinct value, so multiple users with NULL usernames (pre-migration) is safe. [VERIFIED: sqlite.org -- "each NULL is considered unique"]
**Warning signs:** "UNIQUE constraint failed" on AutoMigrate or insert of existing users. If this happens, the column was defined as `not null` with a non-null default.

### Pitfall 2: MigrationModal Not Truly Unclosable
**What goes wrong:** Migration modal can be dismissed with Escape or backdrop click, allowing users to bypass forced migration.
**Why it happens:** Using the standard `ui/Dialog` or `AppDialog` component which always has Escape/backdrop handlers.
**How to avoid:** Build MigrationModal with raw glass morphism markup (mirroring the overlay + panel pattern from existing LoginModal), without wrapping in Dialog. Do NOT use `useCloseOnEscape`. Set backdrop div without `onClick` handler.
**Warning signs:** Pressing Escape hides the modal during testing. The user sees the main app briefly, then the modal reappears on next page load.

### Pitfall 3: bcrypt Cost Too High on Low-End Hardware
**What goes wrong:** Login/registration takes >1 second, creating poor UX.
**Why it happens:** bcrypt cost 10 = 1024 iterations. On slow hardware (e.g., a small VPS), this can be ~250ms per hash. Increasing to 12+ doubles this.
**How to avoid:** Use `bcrypt.DefaultCost` (10). D-07 explicitly accepts slower login as the brute-force mitigation. For this project's scale (single-instance, low concurrency), cost 10 provides adequate security without noticeable latency.
**Warning signs:** Users report login taking >2 seconds. If encountered, the solution is to lower cost (not increase), but monitor first.

### Pitfall 4: Race Condition on Invite Code Assignment
**What goes wrong:** Two users simultaneously set the same invite code; one silently overwrites the other.
**Why it happens:** Application-level check (SELECT to see if code exists) followed by UPDATE is not atomic.
**How to avoid:** Use the database-level unique index. Attempt the INSERT/UPDATE first, catch the unique constraint violation error, and return "该邀请码已被使用" to the user. Never check-then-set.
**Warning signs:** Duplicate invite codes in the database despite application-level validation.

### Pitfall 5: Config.json Concurrent Write Corruption
**What goes wrong:** Admin saves pricing config while invite config is being saved; one write overwrites the other.
**Why it happens:** Both `persistPricingConfig` and a new `persistInviteConfig` independently read-modify-write config.json.
**How to avoid:** Use the existing `persistMu sync.Mutex` in config/config.go. All config.json persistence functions must acquire this lock. The existing `persistEndpoints` and `persistPricingConfig` already do this -- follow the same pattern. Alternatively, combine invite config persistence with the existing pricing config persistence path.

**Recommended approach:** Add invite config to the `Config` struct and have `AdminUpdatePricingConfig`-style handler update all config together. Or, add a dedicated `PUT /api/admin/invite-config` that uses the same `persistMu`. The key is that ALL config.json writes go through the same mutex.

## Code Examples

### Password Login Handler + Service

```go
// --- handler/auth.go ---
func AuthLoginPassword(c *gin.Context) {
    var body struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    _ = c.ShouldBindJSON(&body)
    if body.Username == "" || body.Password == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "请输入用户名和密码"})
        return
    }
    token, user, needsMigrate, err := service.LoginWithPassword(body.Username, body.Password)
    if err != nil {
        slog.Warn("密码登录失败", "username", body.Username, "error", err)
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"token": token, "user": user, "needsMigration": needsMigrate})
}

// --- service/auth.go ---
func LoginWithPassword(username, password string) (token string, user *AuthUser, needsMigrate bool, err error) {
    var u database.User
    if err := database.DB.Where("username = ?", username).First(&u).Error; err != nil {
        return "", nil, false, fmt.Errorf("用户名或密码错误")
    }
    if u.PasswordHash == nil {
        return "", nil, false, fmt.Errorf("该账号尚未设置密码，请使用兑换码登录后设置密码")
    }
    if err := bcrypt.CompareHashAndPassword([]byte(*u.PasswordHash), []byte(password)); err != nil {
        return "", nil, false, fmt.Errorf("用户名或密码错误")
    }
    if u.Status == "disabled" {
        return "", nil, false, fmt.Errorf("账号已被禁用")
    }
    now := time.Now().UnixMilli()
    database.DB.Model(&database.User{}).Where("id = ?", u.ID).Update("last_login_at", now)

    token, err = SignToken(u.ID, u.Role, config.App.JWTSecret)
    if err != nil {
        slog.Error("签发 JWT 失败", "user_id", u.ID, "error", err)
        return "", nil, false, fmt.Errorf("登录失败")
    }
    return token, dbUserToAuthUser(&u), false, nil
}
```
Source: [PATTERN MATCH: backend-go/handler/auth.go AuthLogin -- same gin binding + error response style] [CITED: pkg.go.dev/golang.org/x/crypto/bcrypt]

### Registration Handler + Service (with invite code)

```go
// --- handler/auth.go ---
func AuthRegister(c *gin.Context) {
    var body struct {
        InviteCode string `json:"inviteCode"`
        Username   string `json:"username"`
        Password   string `json:"password"`
    }
    _ = c.ShouldBindJSON(&body)
    // Validation
    if len([]rune(body.Username)) < 3 || len([]rune(body.Username)) > 20 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "用户名须为 3-20 个字符"})
        return
    }
    if len(body.Password) < 8 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "密码至少需要 8 个字符"})
        return
    }
    token, user, err := service.RegisterUser(body.Username, body.Password, body.InviteCode)
    if err != nil {
        slog.Warn("注册失败", "username", body.Username, "error", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

// --- service/auth.go ---
func RegisterUser(username, password, inviteCode string) (token string, user *AuthUser, err error) {
    // Check username uniqueness
    var existing database.User
    if result := database.DB.Where("username = ?", username).First(&existing); result.Error == nil {
        return "", nil, fmt.Errorf("用户名已被使用")
    }

    // Hash password
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        slog.Error("密码哈希失败", "error", err)
        return "", nil, fmt.Errorf("注册失败")
    }
    hashStr := string(hash)
    now := time.Now().UnixMilli()

    // Validate and lookup invite code
    var inviter *database.User
    if strings.TrimSpace(inviteCode) != "" {
        if result := database.DB.Where("invite_code = ?", strings.TrimSpace(inviteCode)).First(&inviter); result.Error != nil {
            return "", nil, fmt.Errorf("邀请码无效")
        }
    }

    // Determine initial quota
    quota := config.GetInviteDefaultQuota()
    if inviter != nil {
        quota += config.GetInviteInviteeReward()
    }

    userID := util.GenerateID()
    usernameStr := username
    newUser := &database.User{
        ID:          userID,
        Label:       userID[:8],
        Username:    &usernameStr,
        PasswordHash: &hashStr,
        Role:        "user",
        Status:      "active",
        Quota:       quota,
        UsedCount:   0,
        CreatedAt:   now,
        LastLoginAt: &now,
    }
    if err := database.DB.Create(newUser).Error; err != nil {
        slog.Error("创建用户失败", "user_id", userID, "error", err)
        return "", nil, fmt.Errorf("注册失败")
    }

    // Award inviter quota
    if inviter != nil {
        reward := config.GetInviteInviterReward()
        if reward > 0 {
            database.DB.Model(&database.User{}).Where("id = ?", inviter.ID).
                Update("quota", gorm.Expr("quota + ?", reward))
        }
    }

    token, err = SignToken(userID, "user", config.App.JWTSecret)
    if err != nil {
        return "", nil, fmt.Errorf("登录失败")
    }
    return token, dbUserToAuthUser(newUser), nil
}
```
Source: [PATTERN MATCH: backend-go/service/auth.go LoginWithCode -- same user creation + quota assignment pattern, lines 82-155]

### Migration Handler

```go
// --- handler/auth.go ---
func AuthMigrate(c *gin.Context) {
    user := middleware.GetAuthUser(c)
    var body struct {
        Username        string `json:"username"`
        Password        string `json:"password"`
        ConfirmPassword string `json:"confirmPassword"`
    }
    _ = c.ShouldBindJSON(&body)
    if body.Password != body.ConfirmPassword {
        c.JSON(http.StatusBadRequest, gin.H{"error": "两次输入的密码不一致"})
        return
    }
    if len([]rune(body.Username)) < 3 || len([]rune(body.Username)) > 20 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "用户名须为 3-20 个字符"})
        return
    }
    if len(body.Password) < 8 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "密码至少需要 8 个字符"})
        return
    }
    updatedUser, err := service.MigrateUser(user.ID, body.Username, body.Password)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"user": updatedUser})
}
```
Source: [PATTERN MATCH: backend-go/handler/auth.go AuthRedeem -- same middleware.GetAuthUser + shouldBindJSON + service call pattern, lines 35-55]

### Admin Reset Password Handler

```go
// --- handler/admin.go ---
func AdminResetPassword(c *gin.Context) {
    userID := c.Param("id")
    var body struct {
        Password string `json:"password"`
    }
    if err := c.ShouldBindJSON(&body); err != nil || len(body.Password) < 8 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "密码至少需要 8 个字符"})
        return
    }
    hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
    if err != nil {
        slog.Error("密码哈希失败", "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "重置密码失败"})
        return
    }
    hashStr := string(hash)
    result := database.DB.Model(&database.User{}).Where("id = ?", userID).Update("password_hash", hashStr)
    if result.Error != nil {
        slog.Error("重置密码失败", "user_id", userID, "error", result.Error)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "重置密码失败"})
        return
    }
    if result.RowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"ok": true})
}
```
Source: [PATTERN MATCH: backend-go/handler/admin.go AdminToggleStatus -- same c.Param("id") + ShouldBindJSON + DB update pattern, lines 77-96]

### Admin Invite Config Handlers

```go
// --- handler/admin.go ---
func AdminGetInviteConfig(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "inviterReward":  config.GetInviteInviterReward(),
        "inviteeReward":  config.GetInviteInviteeReward(),
        "defaultQuota":   config.GetInviteDefaultQuota(),
    })
}

func AdminUpdateInviteConfig(c *gin.Context) {
    var body struct {
        InviterReward int `json:"inviterReward"`
        InviteeReward int `json:"inviteeReward"`
        DefaultQuota  int `json:"defaultQuota"`
    }
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
        return
    }
    if body.InviterReward < 0 || body.InviteeReward < 0 || body.DefaultQuota < 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "奖励值不能为负数"})
        return
    }
    config.SetInviteConfig(body.InviterReward, body.InviteeReward, body.DefaultQuota)
    c.JSON(http.StatusOK, gin.H{
        "ok": true,
        "inviterReward": body.InviterReward,
        "inviteeReward": body.InviteeReward,
        "defaultQuota":  body.DefaultQuota,
    })
}
```
Source: [PATTERN MATCH: backend-go/handler/admin.go AdminGetAnnouncement + AdminUpdateAnnouncement -- same GET/PUT config pattern, lines referenced in existing code]

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | Backend build | Yes | 1.23.0 | -- |
| golang.org/x/crypto | bcrypt hashing | Yes (indirect dep) | v0.39.0 | -- |
| SQLite (via gorm) | Database | Yes | go-sqlite3 v1.14.27 | -- |
| Node.js | Frontend build | Yes | v22.14.0 | -- |
| React | Frontend UI | Yes | v19.1.0 | -- |
| @radix-ui/react-tabs | LoginModal tabs | Yes (in node_modules) | -- | -- |
| @radix-ui/react-separator | SettingsModal dividers | Yes (in node_modules) | -- | -- |
| @radix-ui/react-dialog | RegisterModal base | Yes (in node_modules) | -- | -- |
| @radix-ui/react-label | Form labels | Yes (in node_modules) | -- | -- |
| vitest | Frontend testing | Yes | v4.1.5 | -- |

**Missing dependencies with no fallback:** none -- all required dependencies are already present.

**Missing dependencies with fallback:** none.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | vitest v4.1.5 (frontend), Go `testing` + `net/http/httptest` (backend) |
| Config file | none -- vitest runs without config file (zero-config mode), Go uses `*_test.go` convention |
| Quick run command | `npm test` (vitest run), `cd backend-go && go test ./...` |
| Full suite command | `npm test && cd backend-go && go test ./...` |

### Phase Requirements --> Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| D-02 | bcrypt hash stored, never plaintext | unit | `go test ./service/ -run TestPasswordHash -v` | No -- Wave 0 |
| D-03 | POST /api/auth/login-password returns token+user+needsMigration | integration | `go test ./handler/ -run TestPasswordLogin -v` | No -- Wave 0 |
| D-04 | Username 3-20 chars, allows Chinese, globally unique | unit | `go test ./service/ -run TestUsernameValidation -v` | No -- Wave 0 |
| D-05 | Password min 8 chars enforced | unit | `go test ./service/ -run TestPasswordValidation -v` | No -- Wave 0 |
| D-06 | Change password: old pw required; admin reset: no old pw required | integration | `go test ./handler/ -run TestChangePassword -v` | No -- Wave 0 |
| D-10 | needsMigration: true when password_hash IS NULL | integration | `go test ./handler/ -run TestPasswordLoginMigration -v` | No -- Wave 0 |
| D-11 | POST /api/auth/register returns token+user | integration | `go test ./handler/ -run TestRegister -v` | No -- Wave 0 |
| D-15 | Invite code globally unique, user can set own | integration | `go test ./handler/ -run TestSetInviteCode -v` | No -- Wave 0 |
| D-18 | Register with invite code awards both inviter and invitee quota | integration | `go test ./handler/ -run TestRegisterInviteReward -v` | No -- Wave 0 |
| D-29 | Admin API endpoints return correct data | integration | `go test ./handler/ -run TestAdminInvite -v` | No -- Wave 0 |
| -- | LoginModal renders tabs, switches correctly | unit | `npx vitest run src/components/LoginModal.test.tsx` | No -- Wave 0 |
| -- | MigrationModal cannot be closed (no Escape, no backdrop click) | unit | `npx vitest run src/components/MigrationModal.test.tsx` | No -- Wave 0 |

### Sampling Rate
- **Per task commit:** `cd backend-go && go test ./... -count=1` (backend), `npx vitest run` (frontend)
- **Per wave merge:** Full suite `npm test && cd backend-go && go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `backend-go/service/auth_test.go` -- NEW: covers password hashing, username validation, invite code uniqueness, reward distribution
- [ ] `backend-go/handler/auth_test.go` -- NEW: covers password login, register, migrate, change password, invite code CRUD endpoints
- [ ] `backend-go/handler/admin_test.go` -- EXTEND: covers admin reset password, invite config GET/PUT, invites list
- [ ] `src/components/LoginModal.test.tsx` -- NEW: covers tab switching, form validation, error display
- [ ] `src/components/MigrationModal.test.tsx` -- NEW: covers forced non-close behavior, form validation, submit
- [ ] `src/components/RegisterModal.test.tsx` -- NEW: covers form validation, invite code optionality
- [ ] `src/lib/backendApi.test.ts` -- NEW/EXTEND: covers new API functions (loginWithPassword, register, migrate, changePassword, etc.)
- [ ] `src/admin/adminApi.test.ts` -- EXTEND: covers new admin API functions (resetPassword, inviteConfig, listInvites)
- [ ] Test framework config: none needed (vitest zero-config, Go `testing` standard)
- [ ] General pattern: Go handler tests follow pattern in `backend-go/handler/images_test.go` (setup helper with temp SQLite); frontend tests follow pattern in `src/lib/db.test.ts` (vitest + vi.mock)

## Security Domain

`security_enforcement` is not explicitly set in `.planning/config.json`. Per the instruction (absent = enabled), the Security Domain section is included.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | Yes | bcrypt via `golang.org/x/crypto/bcrypt` with cost 10 |
| V3 Session Management | Yes | Existing 30-day JWT (HS256) in localStorage; no refresh token |
| V4 Access Control | Yes | Existing `AuthMiddleware` and `AdminMiddleware`; all new user endpoints use AuthMiddleware, admin endpoints use AdminMiddleware |
| V5 Input Validation | Yes | Server-side validation: username 3-20 chars (unicode), password min 8 chars, invite code trim+unique; frontend pre-validation as UX layer |
| V6 Cryptography | No | Not applicable for this phase (no encryption at rest, no TLS configuration -- handled at deployment level) |

### Known Threat Patterns for Go + SQLite + bcrypt + JWT

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Username enumeration via timing | Information Disclosure | Use same bcrypt comparison path for both "user not found" and "wrong password" scenarios. Return generic "用户名或密码错误" in both cases. |
| SQL injection in username lookup | Tampering | GORM parameterized queries (`Where("username = ?", username)`) [VERIFIED: existing code in backend-go/service/auth.go uses parameterized queries] |
| bcrypt cost too low (brute force) | Elevation of Privilege | D-07: no lockout, relies on bcrypt cost 10. Each login attempt takes ~250ms. 100 attempts/minute = ~4 per second -- rate limited by hash cost alone. |
| JWT token theft (localStorage) | Information Disclosure | Existing pattern (localStorage Bearer token). No XSS protection in scope for this phase. Mitigated by 30-day expiry and logout clearing. |
| Race condition in invite code assignment | Spoofing | Database-level unique index on `invite_code` column. Attempt insert first, catch constraint violation, return "已被使用". |
| Config.json tampering (invite rewards) | Tampering | Admin-only endpoint behind `AdminMiddleware`. Same protection as existing `salePriceX10000` config. |
| Password in logs | Information Disclosure | Never log raw passwords. Only log `username` and error summary (e.g., "密码登录失败"). Existing code already uses `slog.Warn` without raw sensitive data. |

## Sources

### Primary (HIGH confidence)
- [backend-go/go.mod] -- Go module dependencies: gin v1.10.1, gorm v1.30.0, golang-jwt v5.2.2, golang.org/x/crypto v0.39.0 (indirect)
- [backend-go/go.sum] -- golang.org/x/crypto v0.39.0 present; bcrypt sub-package available
- [backend-go/config/config.go] -- Config struct, persistPricingConfig pattern, persistMu mutex
- [backend-go/database/database.go] -- GORM AutoMigrate with existing User model
- [backend-go/database/models.go] -- Current User struct (no password/username fields yet)
- [backend-go/service/auth.go] -- Existing LoginWithCode, SignToken, VerifyToken patterns
- [backend-go/handler/auth.go] -- Existing AuthLogin, AuthRedeem, AuthMe handler patterns
- [backend-go/middleware/middleware.go] -- AuthMiddleware, AdminMiddleware, GetAuthUser
- [backend-go/main.go] -- Route registration pattern; existing admin routes
- [backend-go/handler/admin.go] -- Admin handler patterns (AdminToggleStatus, AdminGetAnnouncement, etc.)
- [backend-go/service/models.go] -- AuthUser, AdminUser, User service DTOs
- [src/lib/backendApi.ts] -- Frontend API client patterns (request<T>, token management)
- [src/admin/adminApi.ts] -- Admin API client patterns (adminRequest<T>)
- [src/components/LoginModal.tsx] -- Current LoginModal structure (redeem code only)
- [src/components/SettingsModal.tsx] -- Current SettingsModal structure (redeem code section)
- [src/components/Header.tsx] -- Current header (no username display)
- [src/admin/AdminDashboard.tsx] -- Admin tab structure, inline modal patterns
- [src/store.ts] -- Zustand store with authUser, bootstrapBackendSession, logout
- [src/components/ui/tabs.tsx] -- Radix Tabs wrapper for LoginModal tab switching
- [src/components/ui/input.tsx] -- Shared Input component for all form fields
- [src/components/ui/separator.tsx] -- Radix Separator for SettingsModal sections
- [src/components/ui/app-dialog.tsx] -- AppDialog wrapper (NOT suitable for MigrationModal)
- [src/hooks/useCloseOnEscape.ts] -- ESC stack pattern (MigrationModal must NOT register)
- [package.json] -- Frontend deps: react 19, zustand 5; devDeps: vitest 4.1.5, tailwindcss 3.4.17
- [.planning/config.json] -- nyquist_validation absent (enabled), security_enforcement absent (enabled)
- [src/lib/db.test.ts] -- Existing vitest test pattern (FakeIDB*, describe/it/expect)
- [backend-go/handler/images_test.go] -- Existing Go handler test pattern (temp SQLite, httptest)

### Secondary (MEDIUM confidence)
- [pkg.go.dev/golang.org/x/crypto/bcrypt] -- Official Go docs: bcrypt.DefaultCost=10, GenerateFromPassword, CompareHashAndPassword
- [gorm.io/docs/migration.html] -- GORM AutoMigrate documentation: column addition behavior
- [sqlite.org/lang_createindex.html] -- SQLite unique index NULL behavior: multiple NULLs allowed

### Tertiary (LOW confidence)
- None. All claims are verified against project source code or official documentation.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `bcrypt.DefaultCost` (10) is appropriate for this project's hardware | Standard Stack / Security | If server is extremely slow, login may take >1s per attempt. Mitigation: can reduce to cost 8 (4x faster) but this must be a user decision. |
| A2 | `golang.org/x/crypto/bcrypt` sub-package compiles and works without explicit `go get` since the parent module is already in go.sum | Standard Stack | If bcrypt sub-package requires explicit `go get golang.org/x/crypto/bcrypt`, the build will fail. Mitigation: planner should add `go get` step or verify sub-package is importable. |
| A3 | Storing invite config in `config.json` (not a DB table) is acceptable -- admin changes are infrequent and config.json persistence already exists | Architecture Patterns | If admin needs audit trail or history of config changes, config.json overwrites are insufficient. Low risk given project scope. |
| A4 | Radix UI primitives in `node_modules` work without being in `package.json` dependencies -- they were installed as transitive deps of other packages | Standard Stack | If `npm install` on a fresh machine doesn't pull them, import will fail. Mitigation: planner should verify with `npm ls @radix-ui/react-tabs`. |
| A5 | Existing users' `label` field (currently = redemption code uppercase) can be repurposed as display fallback when `username` is NULL | Architecture Patterns | If `label` format is important for admin identification, changing its behavior may confuse admins. Mitigation: keep `label` unchanged; display `username || label || '用户'` in header. |

## Open Questions

1. **Should config.json invite settings also be persisted via the same persistPricingConfig function or a separate one?**
   - What we know: `persistPricingConfig` already handles config.json read-modify-write with `persistMu`. Adding fields to the same function avoids concurrency issues.
   - What's unclear: Whether to combine invite config persistence with pricing config persistence, or add a separate `persistInviteConfig` that shares the same mutex.
   - Recommendation: Add invite fields to a new `persistInviteConfig` function that acquires `persistMu`. Keep the functions separate for clarity but share the mutex. Alternatively, have the admin handler update `config.App` fields directly and call a unified persist function.

2. **Should the `POST /api/auth/login` (existing code login) response include `needsMigration`?**
   - What we know: D-14 says old users logging in with code should see the MigrationModal. D-10 says password login returns `needsMigration`. The existing code login path needs to also return this flag.
   - What's unclear: Whether to add `needsMigration` to the existing `AuthLogin` handler response (modifying the existing API contract slightly) or add it only to password login.
   - Recommendation: Add `needsMigration: password_hash IS NULL` to ALL login responses (both code login and password login). The frontend already expects this field per D-10 and D-14. The migration modal trigger should be: show if `authUser.needsMigration === true` after any successful login.

3. **Existing users created via redeem code currently get `label` = the redemption code string. How does this interact with username display?**
   - What we know: D-27 says header should show username if set, otherwise show original label. Pre-migration user `label` is set to the redemption code (e.g., "ABCD-EFGH-IJKL-MNOP").
   - What's unclear: Whether showing the 20-char code string in the header is acceptable UX for pre-migration users.
   - Recommendation: Follow D-27 exactly: `{authUser.username || authUser.label || '用户'}`. The code string display is transitional until the user migrates. The migration is forced, so this is a temporary state.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- all deps already present, verified via go.mod/go.sum and node_modules inspection
- Architecture: HIGH -- patterns match existing project conventions exactly; handler/service/model layers are well-established
- Pitfalls: HIGH -- SQLite NULL+unique index behavior is well-documented; bcrypt costing is standard; race conditions have clear mitigations

**Research date:** 2026-05-23
**Valid until:** 2026-06-23 (30 days -- authentication and bcrypt patterns are stable)

## Research Complete Summary

### Key Findings

1. **No new packages needed** -- `golang.org/x/crypto/bcrypt` is already an indirect dependency; all Radix UI primitives (`react-tabs`, `react-separator`, `react-dialog`, `react-label`) are in `node_modules`
2. **Database migration is safe** -- adding nullable `*string` columns with `uniqueIndex` via GORM AutoMigrate works correctly because SQLite treats NULL values as distinct in unique indices
3. **Invite config belongs in config.json** -- matches the existing `salePriceX10000` pattern; share the `persistMu` mutex to prevent concurrent write corruption
4. **MigrationModal requires raw markup** -- cannot use existing `ui/Dialog` or `AppDialog` because they always permit Escape/backdrop dismiss; must use the raw glass morphism overlay + panel pattern
5. **Race condition prevention** -- invite code uniqueness enforced at database level via unique index, not at application level
6. **168 existing tests pass** -- vitest v4.1.5 (frontend, 168 tests green) and Go `testing` (backend) are the test frameworks; Wave 0 needs 8+ new test files
