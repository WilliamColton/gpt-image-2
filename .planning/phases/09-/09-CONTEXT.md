# Phase 09: 操作日志 - Context

**Gathered:** 2026-05-23
**Status:** Ready for planning

<domain>
## Phase Boundary

在后端新增操作日志持久化机制并为管理后台新增"操作日志"Tab。后端在关键业务事件发生时同步写入 `audit_logs` 表，前端在管理后台提供表格查看、多维过滤和手动清理功能。

日志记录的事件范围：认证事件（登录/注册/迁移/改密）、生成事件（任务提交/成功/失败）、管理操作（配额/用户/配置变更）、配额事件（兑换/邀请奖励/已用递增）。

本阶段不包含 HTTP 请求日志入库、实时日志流/WebSocket 推送、日志导出（CSV/PDF）、自动清理策略、告警/通知。
</domain>

<decisions>
## Implementation Decisions

### 日志范围
- **D-01:** 仅记录关键业务事件，不做 HTTP 请求级别的日志入库。现有 `middleware/logger.go` 的请求日志继续输出到 stdout 即可。
- **D-02:** 四类事件全量记录——认证事件（登录/注册/迁移/改密，含成功/失败）、生成事件（任务提交/成功/失败）、管理操作（改配额/禁用用户/改端点配置/改价格等）、配额事件（兑换码使用/邀请奖励发放/配额消耗）。
- **D-03:** 每条日志公共字段采用标准方案：`event_type`、`user_id`、`user_label`、`message`、`severity`（INFO/WARN/ERROR）、IP、`details_json`（结构化详情）、`created_at`。各事件类型可按需在 `details_json` 中存储专属字段。

### 存储与保留策略
- **D-04:** 使用 SQLite 新 `audit_logs` 表存储，通过 GORM 写入，与现有数据库模型一致。不引入新的存储依赖。
- **D-05:** 手动清理策略——管理后台日志 Tab 提供两个清理操作：按天数清理（删除 N 天前的日志）和清空全部日志。不做自动清理。
- **D-06:** 管理员删除用户或任务时，关联日志不级联删除，保留完整的审计轨迹。

### 查看与过滤
- **D-07:** 管理后台新增"操作日志"Tab，通过 `Tab` 枚举扩展接入现有 AdminDashboard 的 Tab 结构。
- **D-08:** 日志以表格列表形式展示，各列对应公共字段（事件类型/消息/用户/级别/时间/IP）。与现有 admin 表格风格一致。
- **D-09:** 过滤条件全量支持：事件类型下拉筛选、严重级别下拉筛选、用户搜索（文本输入）、关键词搜索（message 内容匹配）、时间范围（今日/7天/30天/全部）。时间范围选择复用 analytics tab 现有组件。
- **D-10:** 不分页，一次性加载全部日志到前端。日志量受配额制控制，数量可控。

### 性能与级别控制
- **D-11:** INFO、WARN、ERROR 三个级别全部入库。
- **D-12:** 同步写入——业务事件触发时直接 INSERT 到 `audit_logs` 表。SQLite 写入开销可忽略，不引入 channel/goroutine 异步复杂度。
- **D-13:** 不做限流或采样。生产环境每个成功图片生成对应一条日志，日志量由配额制自然控制。

### Claude's Discretion
- `audit_logs` 表的具体字段定义和 GORM 模型
- 各事件类型在 `details_json` 中的专属字段结构
- 管理后台日志 Tab 的具体 UI 布局和样式（应与现有 glass morphism 设计系统一致）
- 后端 handler/service 分层——日志写入函数放在 `backend-go/service/`，handler 路由在 `backend-go/handler/`，API 注册在 `backend-go/main.go`
- 日志写入辅助函数的签名设计（便于各 handler/service 调用）
- 清理 API 返回的清理条数
</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### 项目级上下文
- `.planning/ROADMAP.md` — Phase 9 goal and dependencies.
- `.planning/PROJECT.md` — Project architecture, constraints, key decisions.
- `.planning/REQUIREMENTS.md` — v1 requirements baseline.

### 后端日志基础设施
- `backend-go/log/log.go` — 现有 slog logger 初始化（text/json 格式输出到 stdout）。Phase 9 在此基础上新增数据库日志层，不替代现有 slog。
- `backend-go/middleware/logger.go` — 请求日志中间件，记录 method/path/status/latency/ip/user_id。Phase 9 不修改此中间件。

### 数据库模型
- `backend-go/database/models.go` — 所有 GORM 模型定义。新增 `AuditLog` 模型遵循现有模式（文本 ID 主键、int64 时间戳）。
- `backend-go/database/database.go` — AutoMigrate 注册和数据库初始化。

### 管理后台
- `src/admin/AdminDashboard.tsx` — Tab 结构的管理后台；新增 `'logs'` tab 需要扩展 `Tab` 类型、tab 按钮、load 逻辑和渲染分支。
- `src/admin/adminApi.ts` — 管理 API 客户端（`adminRequest<T>()`）；新增日志相关 API 函数。

### 后端路由与分层
- `backend-go/main.go` — 路由集中注册。新增 `GET /api/admin/logs` 和 `DELETE /api/admin/logs` 路由。
- `backend-go/handler/admin.go` — 现有管理 API handler。可以新增日志 handler 或扩展现有 admin handler。
- `backend-go/service/` — 业务逻辑层。新增日志写入和查询 service。

### 先前阶段上下文
- `.planning/phases/04-admin-dashboard/04-CONTEXT.md` — Phase 4 管理后台 Tab 结构和配额决策。
- `.planning/phases/05-cost-revenue-analytics/05-CONTEXT.md` — Phase 5 时间范围筛选组件和分析 API 设计模式（可复用）。
- `.planning/phases/06-invite-code/06-CONTEXT.md` — Phase 6 认证和管理 API 扩展决策。
</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `src/admin/AdminDashboard.tsx` 的 Tab 切换结构——`Tab` 类型枚举、tab 按钮组、`useEffect` 按 tab 加载数据、条件渲染分支。新增 `'logs'` tab 完全遵循此模式。
- Analytic tab 的时间范围选择组件（今日/7天/30天/全部 按钮组+状态）可直接复用或参考。
- `src/components/ui/table.tsx`、`src/components/ui/input.tsx`、`src/components/ui/badge.tsx`、`src/components/ui/select.tsx`——现有 UI 原语可用于日志表格、搜索输入、级别/事件类型 badge、下拉筛选。
- `src/admin/adminApi.ts` 的 `adminRequest<T>()`——标准管理 API 请求封装，新增 API 调用直接复用。
- `backend-go/handler/admin.go` 的管理 API handler 模式和 `backend-go/service/` 的分层模式。

### Established Patterns
- 后端路由集中在 `backend-go/main.go` 注册，handler 在 `backend-go/handler/`，业务逻辑在 `backend-go/service/`，数据库模型在 `backend-go/database/models.go`。
- GORM AutoMigrate 在 `backend-go/database/database.go` 中注册新模型。
- 管理 API 通过 AdminMiddleware（`backend-go/middleware/middleware.go:60`）保护。
- Admin UI 使用 `useStore((s) => s.showToast)` 进行成功/错误反馈。
- 异步 action 后通过 `load*` 函数刷新 tab 数据。

### Integration Points
- 新增 `AuditLog` GORM 模型在 `backend-go/database/models.go`，AutoMigrate 注册在 `backend-go/database/database.go`。
- 后端各 handler（auth/generate/admin）在关键事件发生后调用日志写入 service。
- 管理后端新增 `GET /api/admin/logs`（查询+过滤）和 `DELETE /api/admin/logs`（清理）路由。
- 前端 `src/admin/adminApi.ts` 新增 `adminListLogs()` 和 `adminCleanLogs()` 函数。
- `src/admin/AdminDashboard.tsx` 扩展 `Tab` 类型为 `... | 'logs'`，新增 tab 按钮和渲染逻辑。
</code_context>

<specifics>
## Specific Ideas

- 用户搜索通过文本输入框输入用户名进行实时过滤（大小写不敏感匹配 `user_label`）。
- 关键词搜索匹配 `message` 字段内容，可与事件类型、级别筛选组合使用。
- 清理 UI 应包含确认对话（复用 `ConfirmDialog`），防止误操作，并在清理后显示清理条数。
- 时间范围默认选择"全部"，下拉选择后即时过滤。
- 日志表格按时间倒序展示（最新日志在最上面）。
- `details_json` 字段对前端透明展示——可直接以 JSON 字符串展示在详情列或 Tooltip 中，无需前端解析。
</specifics>

<deferred>
## Deferred Ideas

- HTTP 请求日志入库——当前仅 stdout，不存入 audit_logs。如需全面请求审计，可另开 phase。
- 日志导出（CSV/PDF）——当前只做查看和清理。
- 自动清理策略（定时任务/启动时自动删旧日志）——当前仅手动清理。
- 实时日志流（SSE/WebSocket 推送新日志）——当前刷新查看即可。
- 日志告警/通知——超出当前 phase 范围。
- 按用户差异化日志保留策略——当前全局统一管理。

None of these were discussed; they are noted here as natural extensions for future phases.
</deferred>

---

*Phase: 09-audit-log*
*Context gathered: 2026-05-23*
