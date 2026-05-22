# Phase 05: 成本与收益统计 - Context

**Gathered:** 2026-05-22
**Status:** Ready for planning

<domain>
## Phase Boundary

为管理后台新增成本与收益统计能力：管理员可以为每个 API 端点设置成本价（元/张），配置全局售价（元/张），系统在图片成功生成后按实际保存成功的图片张数记录成本、收入和利润，并在管理后台新增图表页面展示总览、趋势、端点拆分和用户拆分。

本阶段不包含支付、充值、用户端价格展示、按用户差异化售价、CSV 导出或点击钻取明细。

</domain>

<decisions>
## Implementation Decisions

### 记账口径
- **D-01:** 一条任务成功生成多张图片时，只按最终成功保存到本地并进入 `outputImages` 的图片张数入账。
- **D-02:** 部分成功时，成功保存几张就记几张；失败或未保存成功的图片不产生收入和成本。
- **D-03:** failover 过程中，成本只归属最终成功生成图片的端点；失败端点不计成本。

### 定价配置
- **D-04:** 每个 API 端点配置当前成本价，单位元/张。
- **D-05:** 管理后台系统配置中新增全局售价，单位元/张。
- **D-06:** 每条成功记账记录必须保存当时的成本价和售价快照；后续修改端点成本价或全局售价不得改变历史统计。
- **D-07:** 成本价和售价输入最多支持 4 位小数，规划实现时应避免浮点金额误差（例如用整数最小单位或等价方式持久化）。

### 图表范围
- **D-08:** 新增管理后台图表页面，第一版同时包含核心总览、趋势、端点拆分和用户拆分。
- **D-09:** 核心指标至少包括总收入、总成本、利润、成功图片数。
- **D-10:** 时间筛选支持今日、7 天、30 天、全部。
- **D-11:** 第一版只展示图表和表格，不做 CSV 导出，不做点击钻取账单明细。

### 数据保留
- **D-12:** 使用单独账单/用量统计表记录成本收益明细，而不是只扩展任务表或只存汇总表。
- **D-13:** 每条账单记录至少需要保留 taskId、userId、用户可读标识快照、成功端点标识/URL、成功图片张数、成本价快照、售价快照、成本、收入、利润、创建时间。
- **D-14:** 管理员删除任务或用户时，历史成本收益统计保留，不随任务/用户删除而丢失。

### Claude's Discretion
- 图表库或是否用现有组件手写简单图表由 planner/researcher 决定，但必须符合现有 React + Tailwind + admin UI 风格。
- 管理 API 的具体路由命名、统计查询 DTO、数据库索引设计由 planner 决定。

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Planning context
- `.planning/ROADMAP.md` — Phase 5 goal and success criteria.
- `.planning/PROJECT.md` — Project architecture goal and constraints.
- `.planning/REQUIREMENTS.md` — Existing v1 requirements and out-of-scope boundaries.
- `.planning/phases/03-api-failover/03-CONTEXT.md` — Existing endpoint pool/failover decisions.
- `.planning/phases/04-admin-dashboard/04-CONTEXT.md` — Existing admin dashboard and quota decisions.

### Backend generation and accounting integration
- `backend-go/handler/generate.go` — Current successful generation path saves output images and increments `used_count`; new accounting should hook into the success path after output IDs are known.
- `backend-go/service/openai.go` — Current failover abstraction returns image results but does not expose which endpoint succeeded; planning must address success endpoint attribution.
- `backend-go/service/task.go` — Task persistence and task DTO conversion.
- `backend-go/service/models.go` — Service DTOs exposed to frontend/admin clients.
- `backend-go/database/models.go` — GORM models; add billing/usage model here or nearby following existing pattern.

### Endpoint configuration
- `backend-go/config/config.go` — `ApiEndpoint` and `Config` structs, endpoint pool sorting/persistence, config.json mutation.
- `backend-go/handler/admin.go` — Admin endpoint config handlers and validation.
- `backend-go/main.go` — Admin route registration.

### Admin frontend
- `src/admin/AdminDashboard.tsx` — Current tabbed admin dashboard, endpoint config UI, table/card primitives usage.
- `src/admin/adminApi.ts` — Admin API client and `ApiEndpoint` TypeScript type.
- `src/components/ui/` — Existing UI primitives for admin cards, tables, tabs, inputs, badges, dialogs.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `src/admin/AdminDashboard.tsx` already uses tab-based admin sections; Phase 5 can add a statistics/analytics tab using the same layout.
- `src/components/ui/card.tsx`, `src/components/ui/table.tsx`, `src/components/ui/tabs.tsx`, `src/components/ui/input.tsx`, `src/components/ui/badge.tsx` provide the visual primitives needed for summary cards, filters, and tables.
- `src/admin/adminApi.ts` centralizes authenticated admin requests and should receive new pricing/statistics client functions.

### Established Patterns
- Backend routes are registered centrally in `backend-go/main.go`, handlers live in `backend-go/handler/`, business logic in `backend-go/service/`, and persistent models in `backend-go/database/models.go`.
- Endpoint configuration is mutable through admin APIs and persisted to `config.json` through `backend-go/config/config.go`.
- Successful generation currently saves images first, then increments user `used_count`; cost/revenue accounting should use the same successful output count rather than requested `n`.
- Admin UI async actions use `useStore((s) => s.showToast)` for success/error feedback.

### Integration Points
- Add `costPerImage` or equivalent field to the endpoint configuration path, including Go `config.ApiEndpoint`, TypeScript `ApiEndpoint`, admin endpoint form, and admin endpoint validation.
- Add global sale price configuration to backend config/admin API and surface it in the admin system configuration UI.
- Add success endpoint attribution to the OpenAI/failover result so `executeImageGeneration` can write accurate billing records.
- Add admin statistics endpoint(s) that aggregate billing records by time range, endpoint, and user.

</code_context>

<specifics>
## Specific Ideas

- Unit for both cost and sale price is 元/张.
- Price precision should support up to 4 decimal places.
- The analytics page should show all first-version dimensions requested by the user: total overview, trend, endpoint breakdown, and user breakdown.
- Preserve billing history even if source task or user is deleted.

</specifics>

<deferred>
## Deferred Ideas

- CSV/PDF export for finance reconciliation.
- Click-through billing detail page from charts or tables.
- Per-user or per-plan sale price.
- Payment, recharge, invoice, or user-visible pricing flows.
- Endpoint failure-attempt cost accounting.

</deferred>

---

*Phase: 05-成本与收益统计*
*Context gathered: 2026-05-22*
