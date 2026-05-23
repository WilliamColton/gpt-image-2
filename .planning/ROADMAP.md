# Roadmap: GPT Image Playground — 后端接管 API 调用

**Created:** 2026-05-05
**Granularity:** Coarse

## 架构变更说明

**旧架构:** 浏览器 → 直接调用 OpenAI API → 页面刷新 = 连接丢失

**新架构:** 浏览器 → 提交任务到后端 → 后端独立调用 OpenAI → 前端轮询获取结果

后端完全接管 API 调用，前端只负责提交任务和展示结果。页面刷新不影响正在进行的任务。

## Phases

### Phase 1: 后端任务执行层

**Goal:** Go 后端能接收生成任务、独立调用 OpenAI API、存储结果

**Requirements:** TASK-01, TASK-02, TASK-03, TASK-04, TASK-05, TASK-06, CFG-01, CFG-02

**Success Criteria:**
1. 后端 `POST /api/generate` 接收任务参数，异步调用 OpenAI Images API，结果存入数据库
2. 后端 `POST /api/edit` 接收任务参数和参考图，异步调用 OpenAI Edits API，结果存入数据库
3. 后端 `POST /api/responses-generate` 接收任务参数，异步调用 Responses API，结果存入数据库
4. `GET /api/tasks/:id` 返回任务状态 (pending/processing/completed/failed) 和结果
5. API Key 从后端环境变量读取，不从前端传递
6. 任务结果包含 base64 图片、实际参数、改写提示词

**Plans:**
- 01: 新增任务执行 service (OpenAI API 调用逻辑)
- 02: 新增 generate/edit/responses handler 和路由
- 03: 任务状态管理和图片存储
- 04: 环境变量配置 (API Key, Base URL, Model)

**Depends on:** —

---

### Phase 2: 前端适配

**Goal:** 前端改为提交任务 + 轮询结果，移除直接调用 OpenAI 的逻辑

**Requirements:** FE-01, FE-02, FE-03, FE-04, FE-05, CFG-03

**Success Criteria:**
1. 前端 `callImageApi` 改为提交任务到后端，轮询等待结果
2. 页面刷新后任务卡片显示正确状态，完成后可查看结果
3. Images API / Responses API 模式切换正常工作
4. Codex CLI 兼容模式行为不变
5. 前端设置中 API Key 变为可选 (后端有 key 时无需填写)

**Plans:**
- 05: 改造 api.ts — 提交任务 + 轮询
- 06: 更新前端设置和配置逻辑
- 07: 端到端测试验证

**Depends on:** Phase 1

---

### Phase 3: API 降级机制

**Goal:** 后端支持配置多个 API 端点和 Key，请求失败时自动切换到下一个可用端点

**Requirements:** FAILOVER-01, FAILOVER-02, FAILOVER-03, FAILOVER-04

**Success Criteria:**
1. `config.json` 的 `apiEndpoints` 数组支持配置多个 base URL 和对应的 API Key（至少一项）
2. API 调用失败时（网络错误、429、5xx）自动尝试下一个端点
3. 所有端点均失败时返回最后遇到的错误
4. `defaults.baseUrl` 字段移除，端点配置统一通过 `apiEndpoints`

**Plans:**
- [x] 03-01-PLAN.md — 端点池配置结构和加载逻辑
- [x] 03-02-PLAN.md — 降级调用逻辑集成到 OpenAI service

**Depends on:** Phase 1

---

### Phase 4: 管理后台 ✓ 2026-05-08

**Goal:** 管理员可以通过 /admin 页面管理用户配额、查看使用统计、禁用/启用用户

**Requirements:** ADMIN-01, ADMIN-02, ADMIN-03, ADMIN-04, ADMIN-05

**Success Criteria:**
1. `/admin` 页面需要通过 adminApikey 认证才能访问
2. 管理员可以查看所有用户的列表，包含注册时间、已用配额、剩余配额
3. 管理员可以为每个用户设置图片生成总量配额，支持手动增减额度（如 +5 张、-3 张），并支持重置（已用张数清零）
4. 管理员可以查看每个用户已生成的图片数量（使用统计）
5. 管理员可以禁用/启用用户，被禁用用户无法提交生成任务
6. 配额耗尽的用户无法提交生成任务，返回明确的错误提示

**Plans:**
- [x] 04-01-PLAN.md — 后端管理 API: 数据库迁移、管理员认证、用户管理接口、配额执行
- [x] 04-02-PLAN.md — 前端管理页面: admin 登录、用户列表、配额管理 UI

**Depends on:** Phase 1

---

## Summary

| Phase | Name | Requirements | Plans |
|-------|------|--------------|-------|
| 1 | 后端任务执行层 | 8 | 4 |
| 2 | 前端适配 | 6 | 3 |
| 3 | API 降级机制 | 4 | 2 |
| 4 | 管理后台 | 5 | 2 |
| 5 | 成本与收益统计 | 4 | 7 |
| 6 | 5/7 | In Progress|  |

**Total:** phases, 27 requirements, 18 plans

### Phase 5: 成本与收益统计

**Goal:** 管理员可以为每个 API 端点设置成本价、配置全局售价，并查看按成功生成图片计算的成本/收入/利润图表

**Requirements:** COST-01, COST-02, COST-03, COST-04

**Success Criteria:**
1. 管理后台的 API 端点配置支持为每个端点设置成本价，单位元/张
2. 管理后台支持配置全局售价，单位元/张
3. 每次图片成功生成后按图片张数记录成本、收入和利润统计
4. 新增管理后台图表页面展示成本、收入、利润和成功生成图片数量

**Plans:**
- [x] 05-01-PLAN.md — 成本收益账单表、固定点金额 helper 和账单写入基础
- [x] 05-02-PLAN.md — 端点成本价与全局售价配置 API
- [x] 05-03-PLAN.md — 生成成功路径记账与端点归属传递
- [x] 05-04-PLAN.md — 成本收益聚合查询与管理统计 API
- [x] 05-05-PLAN.md — adminApi 定价与统计客户端类型
- [x] 05-06-PLAN.md — 系统配置页价格输入与一次性保存
- [x] 05-07-PLAN.md — 成本收益统计 tab 的 KPI、趋势与拆分表

**Depends on:** Phase 4

### Phase 6: 新增账号密码机制和邀请码机制，合并兑换码机制

**Goal:** [To be planned]
**Requirements**: TBD
**Depends on:** Phase 5
**Plans:** 5/7 plans executed

Plans:
- [x] 06-01-PLAN.md — invite-code data layer foundation
- [x] 06-02-PLAN.md — password auth and invite code service functions
- [x] 06-03-PLAN.md — auth and admin handler endpoints + routing
- [x] 06-04-PLAN.md — frontend API client extension
- [ ] 06-05-PLAN.md — LoginModal/RegisterModal redesign
- [x] 06-06-PLAN.md — admin invites tab + reset password modal
- [ ] 06-07-PLAN.md — TBD

---
*Roadmap created: 2026-05-05*