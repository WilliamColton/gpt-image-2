# Phase 4: 管理后台 - Context

**Gathered:** 2026-05-08
**Status:** Ready for planning

<domain>
## Phase Boundary

为后端和前端添加管理后台功能：管理员可以通过 `/admin` 页面管理用户的图片生成配额、查看使用统计、禁用/启用用户。

仅管理员可访问（通过 adminApikey 认证）。配额为总量限制（不按月重置）。

</domain>

<decisions>
## Implementation Decisions

### 管理后台功能范围
- 用户列表 + 配额管理：查看所有用户，设置每个用户的图片生成配额
- 使用统计：查看每个用户已生成的图片数量
- 禁用/启用用户：临时禁止某个用户使用生成服务
- 全局配额设置不需要（用户要求中未选择）

### 配额机制
- 总量限制：每个用户总共只能生成 N 张，不按月重置
- 配额耗尽时，提交任务返回"配额已用完"错误
- 需要在用户表中添加 quota（配额总量）和 used_count（已使用数量）字段
- 管理员可以手动增减用户配额（如 +5 张、-3 张），即调整 quota 字段
- 管理员可以重置用户配额（将 used_count 清零，恢复满额可用）
- UI 上提供输入框 + "增加"/"减少"按钮，以及"重置"按钮

### 认证方式
- 复用 config.json 中已有的 adminApikey
- 前端 /admin 页面通过 adminApikey 登录
- 后端管理 API 通过 adminApikey 验证身份

### Claude's Discretion
- 前端管理页面的 UI 设计风格（应与现有应用一致）
- 配额字段的数据类型和默认值
- 是否需要分页（用户量大时）
- 管理 API 的具体路由设计

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### 后端配置
- `backend-go/config/config.go` — Config 结构体（含 AdminApikey）
- `backend-go/config.json` — 实际配置文件

### 用户管理
- `backend-go/service/user.go` — 用户模型和数据库操作（FindUserByID 等）
- `backend-go/handler/generate.go` — 任务执行入口（需要检查配额和禁用状态）

### 现有路由
- `backend-go/router/router.go` — 路由注册

### 前端
- `frontend/src/` — 现有前端代码结构

</canonical_refs>

<specifics>
## Specific Ideas

当前用户表（service/user.go）已有字段：
- ID, Username, ApikeyCipher, CreatedAt

需要添加：
- Quota int（配额总量，0 = 无限制）
- UsedCount int（已使用数量）
- Disabled bool（是否禁用）

当前 adminApikey 仅用于 config.json 中的管理密钥，需要新增管理 API 路由。

</specifics>

<deferred>
## Deferred Ideas

- 全局默认配额设置（新用户注册后的默认额度）
- 配额重置功能
- 用户删除
- 操作日志

</deferred>

---

*Phase: 04-admin-dashboard*
*Context gathered: 2026-05-08*
