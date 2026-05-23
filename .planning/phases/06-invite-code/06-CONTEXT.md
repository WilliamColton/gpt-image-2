# Phase 06: 账号密码 & 邀请码 - Context

**Gathered:** 2026-05-23
**Status:** Ready for planning

<domain>
## Phase Boundary

为用户认证系统新增账号密码登录方式和邀请码注册机制，与现有兑换码系统共存。涉及后端密码认证、用户注册、老用户迁移、邀请码生成与管理、奖励规则配置、以及前端登录/注册/设置页面的改造。

本阶段不包含邮箱验证、忘记密码找回、OAuth 第三方登录、邀请码使用次数上限、邀请码链接分享。
</domain>

<decisions>
## Implementation Decisions

### 密码认证
- **D-01:** 仅密码登录（过渡方案）。新用户用邀请码注册时设密码，老用户首次登录后强制设置密码。之后只用密码登录，兑换码仅用于增加配额。
- **D-02:** 密码使用 bcrypt 哈希存储。
- **D-03:** 新增 `POST /api/auth/login-password` 独立接口，接受 `{ username, password }`。原 `POST /api/auth/login` 保留用于兑换码登录（过渡期）。
- **D-04:** 用户名规则：3-20 字符，允许中文，全局唯一。用户名用于登录和显示。
- **D-05:** 密码规则：最少 8 字符，无其他复杂度限制。
- **D-06:** 用户可在设置页修改密码，需输入旧密码、新密码、确认新密码。管理员可在管理后台为用户直接设置新密码（无需旧密码）。
- **D-07:** 不做登录失败次数锁定。暴力破解风险由 bcrypt 慢哈希缓解。
- **D-08:** 保持现有 30 天 JWT 有效期不变，无 refresh token。登出时前端清除 token 和 store 状态即可。
- **D-09:** 始终记住登录状态（localStorage 存储 token）。
- **D-10:** 密码登录成功后返回 `{ token, user, needsMigration }`，`needsMigration` 标志用于判断是否弹出迁移 Modal。

### 注册与迁移
- **D-11:** 新增 `POST /api/auth/register` 独立注册接口，接受 `{ inviteCode(可选), username, password }`。
- **D-12:** 注册即自动登录——注册接口直接返回 `{ token, user }`，前端自动进入主页。
- **D-13:** 无邀请码也可注册，获得管理员设定的默认注册配额。
- **D-14:** 老用户强制迁移：兑换码登录成功后弹出独立 Modal（不可关闭，无 X 按钮，Escape 不关闭），需填写用户名 + 密码 + 确认密码，完成后进入主页。

### 邀请码（独立于兑换码）
- **D-15:** 用户在设置页自定义邀请码（全局唯一，先到先得），可无限使用。
- **D-16:** 邀请码和兑换码完全独立——兑换码保持现有行为（纯配额充值），邀请码用于注册。
- **D-17:** 用户可随时修改自己的邀请码，旧码立即失效。
- **D-18:** 双方配额奖励：被邀请人注册时自动获得管理员设定的奖励配额，邀请人也在此时自动获得奖励配额。

### 奖励规则配置
- **D-19:** 管理后台新增"邀请码设置"tab。
- **D-20:** 管理员可分别设置邀请人奖励配额、被邀请人奖励配额（两个独立值）。
- **D-21:** 管理员可设置默认注册配额（无邀请码注册时获得的配额）。
- **D-22:** 该 tab 包含奖励配置区域和邀请码使用列表（所有用户及其邀请码、被使用次数）。

### 前端 UI
- **D-23:** LoginModal：Tab 切换登录方式（默认兑换码 Tab + 密码登录 Tab），底部有"没有邀请码？注册"链接。
- **D-24:** RegisterModal：三个字段——邀请码（可选）+ 用户名 + 密码。
- **D-25:** 迁移 Modal：三个字段——用户名 + 密码 + 确认密码。不可关闭。
- **D-26:** SettingsModal 分块展示：邀请码区（当前邀请码 + 复制 + 修改按钮）、修改密码区（旧密码 + 新密码 + 确认）、兑换码输入区（保持现有）。
- **D-27:** 登录后 header/顶部显示用户名。若未设置用户名（迁移前老用户），显示原 label。

### 管理后台扩展
- **D-28:** 用户列表每行新增"重置密码"按钮，点击弹出输入框让管理员输入新密码。
- **D-29:** 新增管理 API 路由：
  - `PUT /api/admin/users/:id/password` — 为用户设新密码
  - `GET /api/admin/invite-config` — 获取邀请奖励配置
  - `PUT /api/admin/invite-config` — 更新邀请奖励配置
  - `GET /api/admin/invites` — 查看所有用户邀请码及使用情况

### Claude's Discretion
- bcrypt 库选择（`golang.org/x/crypto/bcrypt` 为标准选择）
- 数据库迁移方案（新增字段：`password_hash`, `username`, `invite_code`, `invite_code_set_at` 等）
- Admin API 和前端请求的具体 DTO 结构
- 邀请奖励配置的存储方式（数据库新表 vs config.json 扩展）
- 具体 UI 样式（应与现有 glass morphism 设计系统一致，使用现有 UI 组件库）
- 错误提示措辞
</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### 项目级上下文
- `.planning/ROADMAP.md` — Phase 6 goal and dependencies.
- `.planning/PROJECT.md` — Project architecture, constraints, key decisions.
- `.planning/REQUIREMENTS.md` — v1 requirements baseline.

### 现有认证系统
- `backend-go/service/auth.go` — 当前 JWT 签名/验证、兑换码登录/注册、用户 CRUD、配额检查。需要扩展密码登录和邀请码逻辑。
- `backend-go/handler/auth.go` — 当前认证 handler（AuthLogin, AuthRedeem, AuthMe）。需要新增 PasswordLogin、Register 等。
- `backend-go/middleware/middleware.go` — JWT 中间件，AuthMiddleware 和 AdminMiddleware。
- `backend-go/service/models.go` — 认证相关的 service DTO（User, AuthUser, AdminUser, RedemptionCode）。

### 数据库模型
- `backend-go/database/models.go` — User 和 RedemptionCode 的 GORM 模型。User 表需新增 `password_hash`, `username`, `invite_code`, `invite_code_set_at` 等字段。

### 前端认证与设置
- `src/lib/backendApi.ts` — 前端认证 API 客户端（loginWithCode, redeemCode, getMe, token 管理）。需扩展密码登录、注册、改密、邀请码操作。
- `src/components/LoginModal.tsx` — 当前仅兑换码输入。需改造为 Tab 切换 + 密码登录 + 注册链接。
- `src/components/SettingsModal.tsx` — 当前有兑换码区域。需新增邀请码展示/管理区和改密区。

### 管理后台
- `src/admin/AdminDashboard.tsx` — Tab 结构的管理后台，需新增邀请码设置 tab。
- `src/admin/adminApi.ts` — 管理 API 客户端，需新增密码重置和邀请配置 API。
- `backend-go/handler/admin.go` — 管理 API handler，需新增密码和邀请相关路由。

### 先前的阶段上下文
- `.planning/phases/04-admin-dashboard/04-CONTEXT.md` — Phase 4 管理后台和配额决策。
- `.planning/phases/03-api-failover/03-CONTEXT.md` — Phase 3 端点配置决策。
</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `src/components/ui/` — 完整的 UI 组件库（Button, Input, Dialog, Tabs, Card 等），可复用构建 LoginModal、RegisterModal、迁移 Modal。
- `src/lib/backendApi.ts` 的 `request<T>()` 函数 — 标准的 Bearer token 请求封装，新增 API 调用直接复用。
- `src/admin/adminApi.ts` 的 `adminRequest<T>()` 函数 — 标准的管理 API 请求封装。
- `backend-go/service/auth.go` 的 `SignToken()` / `VerifyToken()` — JWT 签发和验证，新接口直接复用。

### Established Patterns
- 后端路由集中在 `backend-go/main.go` 注册，handler 在 `backend-go/handler/`，业务逻辑在 `backend-go/service/`。
- 数据库模型在 `backend-go/database/models.go`，AutoMigrate 在 `backend-go/database/database.go`。
- 管理员认证通过 `backend-go/middleware/middleware.go:60` 的 AdminMiddleware。
- 前端状态管理：登录/用户信息通过 `src/store.ts` 的 `bootstrapBackendSession()` 获取和存储。
- Admin UI 使用 Tab 切换结构，`src/admin/AdminDashboard.tsx` 中已有的 tab 枚举和切换逻辑。

### Integration Points
- 后端 auth handler 新增 `POST /api/auth/login-password` 和 `POST /api/auth/register` 路由。
- User 数据库模型新增 `password_hash`, `username`, `invite_code`, `invite_code_set_at` 字段。
- Admin handler 新增 `PUT /api/admin/users/:id/password`、`GET|PUT /api/admin/invite-config`、`GET /api/admin/invites`。
- 前端 `LoginModal.tsx` 重构为 Tab 切换结构，新建 `RegisterModal.tsx` 和迁移 Modal。
- `src/lib/backendApi.ts` 新增 `loginWithPassword()`, `register()`, `changePassword()`, `setInviteCode()`, `getInviteCode()` 等 API 函数。
- `src/admin/adminApi.ts` 新增 `adminResetPassword()`, `adminGetInviteConfig()`, `adminUpdateInviteConfig()`, `adminListInvites()`。
- 奖励配额发放逻辑集成到注册 handler 中：注册成功后自动给邀请人和被邀请人增加配额。
</code_context>

<specifics>
## Specific Ideas

- 邀请码格式为用户自定义字符串，全局唯一。用户在设置页输入任意字符合法码即可。
- 修改邀请码时，旧码立即失效，新码立即生效。
- 迁移 Modal 对老用户不可关闭——无 X 按钮、无法通过 Escape 关闭、点击 Modal 外不关闭。只有完成迁移后才消失。
- 密码修改必须在 SettingsModal 的独立区域中进行，与邀请码和兑换码区域分开。
- 管理后台邀请码 tab 使用列表形式展示每个用户及其邀请码，包含使用次数信息。
- 奖励配额和被邀请人配额是两个独立可调的数值，管理员可以分别设置。
</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.
</deferred>

---

*Phase: 06-invite-code*
*Context gathered: 2026-05-23*
