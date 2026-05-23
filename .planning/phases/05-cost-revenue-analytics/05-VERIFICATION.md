---
phase: 05-cost-revenue-analytics
verified: 2026-05-23T10:50:00+08:00
status: human_needed
score: 28/28 must-haves verified
overrides_applied: 0
overrides: []
human_verification:
  - test: "打开 /admin 页面，导航到系统配置 tab，查看每个端点是否有成本价（元/张）输入框、全局售价（元/张）输入卡、支持 4 位小数帮助文字，输入无效值（负数、超过4位小数、空值）后检查是否显示内联错误且保存按钮被禁用，输入有效价格后点击保存价格配置按钮确认成功 toast"
    expected: "端点成本价和全局售价输入框正确渲染，校验工作正常，原子保存成功并显示 toast"
    why_human: "前端 UI 渲染、表单交互、toast 显示属于视觉用户体验，无法通过 grep 验证"
  - test: "打开 /admin 页面，导航到成本收益统计 tab，确认 tab 位置在系统配置之后公告管理之前，默认时间范围 7天/7d，KPI 卡片显示总收入、总成本、利润、成功图片数，趋势 SVG 图表绘制收入/成本/利润/成功图片数线条，端点拆分和用户拆分表格正确显示列（端点标识/用户标识、成功图片数、收入、成本、利润、利润率）"
    expected: "统计 tab 按 UI-SPEC 合同布局显示，KPI 数字格式正确，SVG 趋势图颜色区分清晰，拆分表按利润降序排列"
    why_human: "SVG 图表渲染、KPI 数值格式化、表格排序列均为前端视觉行为，无法通过静态分析验证"
  - test: "执行一次图片生成（确保端点有 costPerImageX10000 和全局 salePriceX10000 配置），查看 billing_records 表是否有对应记账行"
    expected: "成功保存的图片数量 = billing_records 行数，每行包含 task_id、user_id、user_label_snapshot、endpoint_base_url_snapshot、unit_cost_x10000、unit_sale_x10000 等快照字段"
    why_human: "需要运行后端服务并触发真实图片生成流程，无法通过单元测试或静态分析验证端到端数据流"
---

# Phase 5: 成本与收益统计 Verification Report

**Phase Goal:** 管理员可以为每个 API 端点设置成本价、配置全局售价，并查看按成功生成图片计算的成本/收入/利润图表

**Verified:** 2026-05-23T10:50:00+08:00
**Status:** human_needed
**Re-verification:** No -- initial verification

## Goal Achievement

### ROADMAP Success Criteria

| # | Success Criterion | Status | Evidence |
|---|-------------------|--------|----------|
| SC-1 | 管理后台的 API 端点配置支持为每个端点设置成本价，单位元/张 | VERIFIED | AdminDashboard.tsx line 738: `成本价（元/张）` input per endpoint; AdminUpdatePricingConfig validates CostPerImageX10000 >= 0; config.go ApiEndpoint has CostPerImageX10000 field |
| SC-2 | 管理后台支持配置全局售价，单位元/张 | VERIFIED | AdminDashboard.tsx line 754: `全局售价（元/张）` card; config.go has SalePriceX10000; adminGetPricingConfig/adminUpdatePricingConfig |
| SC-3 | 每次图片成功生成后按图片张数记录成本、收入和利润统计 | VERIFIED | handler/generate.go line 165-169: RecordBillingForSuccessfulImages called after save with saved-success slice; billing.go computes cost/revenue/profit per row |
| SC-4 | 新增管理后台图表页面展示成本、收入、利润和成功生成图片数量 | VERIFIED | AdminDashboard.tsx analytics tab: line 807 KPI cards (总收入/总成本/利润/成功图片数), line 879 SVG trend chart, lines 977-1052 endpoint/user breakdown tables |

### Observable Truths (Consolidated from All Plans)

| # | Truth | Plan | Status | Evidence |
|---|-------|------|--------|----------|
| 1 | D-12: billing_records 表独立保存成功图片记账明细 | 05-01 | VERIFIED | models.go line 100-117: BillingRecord struct, TableName="billing_records", no FK constraints; database.go line 28: AutoMigrate(&BillingRecord{}) |
| 2 | D-07: 金额以 X10000 固定点整数保存，最多 4 位小数 | 05-01 | VERIFIED | money.go line 11: MoneyScale=10000; ParseMoneyX10000 rejects >4 decimals; all arithmetic is integer-based; no float32/float64 |
| 3 | D-13/D-14: 账单快照独立于 task/user 生命周期 | 05-01 | VERIFIED | BillingRecord stores TaskID/UserID as plain text, no GORM foreign keys or OnDelete; billing_test asserts rows survive user/task deletion |
| 4 | D-04: 管理员 API 可读取端点成本价和全局售价 | 05-02 | VERIFIED | admin.go line 225-231: AdminGetPricingConfig returns endpoints + salePriceX10000 + moneyScale |
| 5 | D-05: 管理员 API 可一次性保存端点成本价和全局售价 | 05-02 | VERIFIED | admin.go line 233-284: AdminUpdatePricingConfig validates all fields, calls SetPricingConfig atomically |
| 6 | 价格配置旧 config.json 缺字段时默认 0 | 05-02 | VERIFIED | config.go line 79-99: Load() uses json.Unmarshal which decodes missing fields as 0; config_test verifies |
| 7 | D-01: 成功生成按实际保存进 outputImages 的图片数记账 | 05-03 | VERIFIED | generate.go: saveGeneratedImagesWithAttribution returns saved-success slice; billing called only for len(saved) > 0 |
| 8 | D-02: 部分保存成功时只为成功保存图片写账单 | 05-03 | VERIFIED | generate.go line 218-231: save fails skip entries without shifting later pairings |
| 9 | D-03: failover 失败端点不产生成本，成本归最终成功端点 | 05-03 | VERIFIED | openai.go line 92-103: withFailover stamps EndpointBaseURL/UnitCostX10000 only after fn succeeds; failed attempts leave no trace |
| 10 | D-06: 账单保存端点成本价和全局售价快照 | 05-03 | VERIFIED | generate.go line 236-252: buildBillingInput uses config.GetSalePriceX10000() and img.UnitCostX10000 as immutable snapshots |
| 11 | D-09/D-10: 管理员 API 可按今日/7天/30天/全部查询收入/成本/利润/成功图片数 | 05-04 | VERIFIED | analytics.go line 65-93: ParseAnalyticsRange supports all 4 values; GetBillingSummary aggregates sums |
| 12 | D-08: 管理员 API 可按同范围查询趋势、端点拆分和用户拆分 | 05-04 | VERIFIED | analytics.go: GetBillingTrend (line 142), GetBillingEndpointBreakdown (line 191), GetBillingUserBreakdown (line 240) |
| 13 | 统计直接聚合 billing_records 快照 | 05-04 | VERIFIED | analytics.go: all queries use database.BillingRecord only; no joins to User/Task tables |
| 14 | 拆分结果按利润降序、收入次序排序 | 05-04 | VERIFIED | analytics.go line 217: Order("profit_x10000 desc, revenue_x10000 desc") |
| 15 | D-11: v1 仅暴露图表/表格数据，无 CSV 导出或钻取 | 05-04 | VERIFIED | No CSV/PDF export endpoints; admin.go analytics handlers return only JSON with meta+data |
| 16 | Admin 前端有定价配置的类型化客户端 | 05-05 | VERIFIED | adminApi.ts line 146-162: PricingConfigResponse, adminGetPricingConfig, adminUpdatePricingConfig |
| 17 | Admin 前端有统计的类型化客户端 | 05-05 | VERIFIED | adminApi.ts lines 166-244: AnalyticsRange, AnalyticsMeta, 4 response types, 4 client functions |
| 18 | 金额在 API 边界保持 X10000 整数 | 05-05 | VERIFIED | adminApi.ts: all DTOs use number for revenueX10000/costX10000/profitX10000 fields; no client-side conversion |
| 19 | 系统配置页为每个端点显示成本价（元/张）输入 | 05-06 | VERIFIED | AdminDashboard.tsx line 738: `成本价（元/张）` label per endpoint |
| 20 | 系统配置页显示全局售价（元/张）输入 | 05-06 | VERIFIED | AdminDashboard.tsx line 754/758: `全局售价（元/张）` card and label |
| 21 | 保存价格配置一次性提交端点成本价和全局售价 | 05-06 | VERIFIED | AdminDashboard.tsx line 395: adminUpdatePricingConfig called once in handleSavePricingConfig |
| 22 | 价格输入无效时不能保存并显示内联错误 | 05-06 | VERIFIED | AdminDashboard.tsx line 337: `请输入非负数字，最多 4 位小数`; save button disabled on invalid inputs |
| 23 | 成本收益统计 tab 在系统配置之后、公告管理之前 | 05-07 | VERIFIED | AdminDashboard.tsx line 25: Tab type='analytics'; line 580: trigger placed after config, before announcement |
| 24 | 统计页默认时间范围 7天 | 05-07 | VERIFIED | AdminDashboard.tsx line 68: useState<AnalyticsRange>('7d') |
| 25 | 统计页展示总收入/总成本/利润/成功图片数 KPI | 05-07 | VERIFIED | AdminDashboard.tsx line 807: KPI grid with 总收入, 总成本, 利润, 成功图片数 |
| 26 | 统计页展示趋势图 | 05-07 | VERIFIED | AdminDashboard.tsx line 879: SVG trend chart with revenue/cost/profit lines + successImages bars |
| 27 | 统计页展示端点拆分和用户拆分表格 | 05-07 | VERIFIED | AdminDashboard.tsx lines 977-1052: endpoint/user breakdown tables with 端点标识/用户标识/利润率 columns |
| 28 | 时间筛选只支持今日/7天/30天/全部 | 05-07 | VERIFIED | AdminDashboard.tsx line 852: labelMap { today: '今日', '7d': '7天', '30d': '30天', all: '全部' } |

**Score: 28/28 truths verified**

### Required Artifacts

| Artifact | Plan | Status | Details |
|----------|------|--------|---------|
| backend-go/service/money.go | 05-01 | VERIFIED | MoneyScale=10000, ParseMoneyX10000, FormatMoneyX10000, all integer arithmetic, no float |
| backend-go/database/models.go BillingRecord | 05-01 | VERIFIED | 14 fields including snapshots, TableName="billing_records", no FK/OnDelete |
| backend-go/database/database.go AutoMigrate | 05-01 | VERIFIED | Line 28: AutoMigrate includes &BillingRecord{} |
| backend-go/service/billing.go | 05-01 | VERIFIED | BillingImageInput, BillingBatchInput, RecordBillingForSuccessfulImages with per-image ID generation |
| backend-go/config/config.go pricing | 05-02 | VERIFIED | CostPerImageX10000 on ApiEndpoint, SalePriceX10000 on Config, GetSalePriceX10000, SetPricingConfig, persistPricingConfig |
| backend-go/handler/admin.go pricing handlers | 05-02 | VERIFIED | AdminGetPricingConfig, AdminUpdatePricingConfig with validation |
| backend-go/main.go pricing routes | 05-02 | VERIFIED | Lines 94-95: adminAuth GET/PUT /config/pricing |
| backend-go/service/openai.go attribution | 05-03 | VERIFIED | GeneratedImage.EndpointBaseURL + UnitCostX10000, stamped in withFailover after success |
| backend-go/handler/generate.go billing integration | 05-03 | VERIFIED | savedGeneratedImage pairing, buildBillingInput, RecordBillingForSuccessfulImages after save |
| backend-go/service/analytics.go | 05-04 | VERIFIED | ParseAnalyticsRange, GetBillingSummary, GetBillingTrend, GetBillingEndpointBreakdown, GetBillingUserBreakdown, all DTOs with MoneyScale |
| backend-go/handler/admin.go analytics handlers | 05-04 | VERIFIED | AdminBillingSummary, AdminBillingTrend, AdminBillingEndpointBreakdown, AdminBillingUserBreakdown |
| backend-go/main.go analytics routes | 05-04 | VERIFIED | Lines 104-107: 4 adminAuth GET /analytics/* routes |
| src/admin/adminApi.ts | 05-05 | VERIFIED | costPerImageX10000, PricingConfigResponse, 4 analytics response wrappers, 6 client functions |
| src/admin/moneyFormat.ts | 05-06 | VERIFIED | formatMoneyInputFromX10000, parseMoneyInputToX10000, string/integer-only, no parseFloat |
| src/admin/AdminDashboard.tsx pricing UI | 05-06 | VERIFIED | 成本价/全局售价 inputs, handleSavePricingConfig, inline validation, save button |
| src/admin/AdminDashboard.tsx analytics tab | 05-07 | VERIFIED | Tab=analytics, KPI grid, SVG chart, endpoint/user tables, independent loaders, moneyScale-aware formatting |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| database.go AutoMigrate | models.go BillingRecord | AutoMigrate(&BillingRecord{}) | WIRED | database.go:28 includes &BillingRecord{} |
| billing.go | models.go BillingRecord | database.DB.Create(&records) | WIRED | billing.go:61 writes BillingRecord rows |
| AdminUpdatePricingConfig | config.SetPricingConfig | function call | WIRED | admin.go:275 calls config.SetPricingConfig |
| main.go | admin.go pricing | adminAuth /config/pricing | WIRED | main.go:94-95 GET/PUT routes |
| openai.go GeneratedImage | generate.go billing | EndpointBaseURL/UnitCostX10000 fields | WIRED | openai.go stamps attribution; generate.go buildBillingInput reads EndpointBaseURL/UnitCostX10000 |
| generate.go | billing.go | RecordBillingForSuccessfulImages | WIRED | generate.go:167 calls after save |
| admin.go analytics handlers | analytics.go service | GetBilling(Summary\|Trend\|EndpointBreakdown\|UserBreakdown) | WIRED | admin.go:288,305,324,343 call service functions |
| analytics.go queries | models.go BillingRecord | database.BillingRecord | WIRED | All analytics queries use database.BillingRecord Model() only |
| adminApi.ts pricing | /api/admin/config/pricing | adminGetPricingConfig / adminUpdatePricingConfig | WIRED | adminApi.ts:154,158 use /api/admin/config/pricing |
| adminApi.ts analytics | /api/admin/analytics/* | adminGetBilling* functions | WIRED | adminApi.ts:231,235,239,243 use /api/admin/analytics/* |
| AdminDashboard.tsx pricing | adminApi.ts | adminGetPricingConfig / adminUpdatePricingConfig | WIRED | AdminDashboard.tsx:111,395 calls pricing functions |
| AdminDashboard.tsx analytics | adminApi.ts | adminGetBilling(Summary\|Trend\|EndpointBreakdown\|UserBreakdown) | WIRED | AdminDashboard.tsx:172,186,200,214 calls analytics functions |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|-------------|--------|--------------------|--------|
| billing.go | records []BillingRecord | database.DB.Create() | Yes (real GORM insert) | FLOWING |
| analytics.go GetBillingSummary | result sumResult | SUM query on database.BillingRecord | Yes (COALESCE SUM) | FLOWING |
| analytics.go GetBillingTrend | rows []bucketRow | Group by strftime on database.BillingRecord | Yes (GROUP BY + ORDER BY) | FLOWING |
| analytics.go GetBillingEndpointBreakdown | rows []groupRow | Group by endpoint_base_url_snapshot | Yes (GROUP BY + ORDER BY) | FLOWING |
| analytics.go GetBillingUserBreakdown | rows []groupRow | Group by user_id, user_label_snapshot | Yes (GROUP BY + ORDER BY) | FLOWING |
| generate.go buildBillingInput | images []BillingImageInput | saved-success slice from save | Yes (real outputImageIDs + endpoint attribution) | FLOWING |
| AdminDashboard.tsx KPI | summary state | adminGetBillingSummary() -> fetch API | Yes (real API call via adminRequest) | FLOWING |
| AdminDashboard.tsx charts | trend state | adminGetBillingTrend() -> fetch API | Yes (real API call via adminRequest) | FLOWING |
| AdminDashboard.tsx tables | endpointRows/userRows | adminGetBillingEndpoint/UserBreakdown() -> fetch API | Yes (real API call via adminRequest) | FLOWING |
| AdminDashboard.tsx pricing | costInputDrafts/salePriceInput | adminGetPricingConfig() -> fetch API | Yes (real API call via adminRequest) | FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Backend tests pass (service) | `cd backend-go && go test ./service/...` | ok (30.686s) | PASS |
| Backend tests pass (database) | `cd backend-go && go test ./database/...` | ok (6.841s) | PASS |
| Backend tests pass (handler) | `cd backend-go && go test ./handler/...` | ok (9.625s) | PASS |
| Backend tests pass (config) | `cd backend-go && go test ./config/...` | ok (0.098s) | PASS |
| Frontend tests pass | `npx vitest run src/admin/` | 102 tests passed, 4 files | PASS |
| TypeScript build clean | `npx tsc --noEmit` | no errors | PASS |

### Probe Execution

Step 7c: SKIPPED (no probe scripts discovered for this phase — no scripts/tests/probe-*.sh exist and no PLAN/SUMMARY mention probes).

### Requirements Coverage

| Requirement | Source Plan(s) | Description | Status | Evidence |
|-------------|---------------|-------------|--------|----------|
| COST-01 | 05-02, 05-05, 05-06 | 端点成本价配置 | SATISFIED | AdminDashboard cost inputs, adminGetPricingConfig/Update, config.go CostPerImageX10000 |
| COST-02 | 05-02, 05-05, 05-06 | 全局售价配置 | SATISFIED | AdminDashboard sale price input, config.go SalePriceX10000, atomic save |
| COST-03 | 05-01, 05-03 | 生成成功记录成本/收入/利润 | SATISFIED | BillingRecord model, RecordBillingForSuccessfulImages in generate path |
| COST-04 | 05-01, 05-04, 05-05, 05-07 | 图表页面展示统计 | SATISFIED | Analytics tab with KPI/SVG chart/endpoint/user breakdown tables |

**Note:** COST-01 through COST-04 are declared in PLAN frontmatter but are NOT present in `.planning/REQUIREMENTS.md`. The REQUIREMENTS.md traceability table ends at ADMIN-05 (Phase 4). Phase 5 requirements exist only in ROADMAP.md Success Criteria. This is a documentation gap in REQUIREMENTS.md, not a codebase gap. Recommend updating REQUIREMENTS.md to include COST-01 through COST-04.

No ORPHANED requirements detected (REQUIREMENTS.md defines no Phase 5 IDs to check against).

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| N/A | N/A | No TBD/FIXME/XXX found | N/A | None in key phase files |
| N/A | N/A | No parseFloat in pricing code | N/A | D-07 compliance confirmed |
| N/A | N/A | No chart library in package.json | N/A | recharts/chart.js/echarts absent |
| N/A | N/A | No stub/placeholder/empty implementations | N/A | All functions have real logic |

### Human Verification Required

#### 1. Admin Pricing Config UI

**Test:** Open /admin page, navigate to `系统配置` tab. Verify each endpoint card has `成本价（元/张）` input with `step="0.0001"`. Verify global sale price card (`全局售价（元/张）`) with help text `支持 4 位小数`. Enter invalid values (negative, >4 decimals, empty) and check that inline error `请输入非负数字，最多 4 位小数` appears and save button is disabled. Enter valid prices and click `保存价格配置` — verify toast `价格配置已保存` appears.

**Expected:** All pricing controls render per UI-SPEC, validation works, atomic save succeeds.

**Why human:** Frontend UI rendering, form interaction, toast display are visual UX behaviors; cannot be verified via grep.

#### 2. Admin Analytics Tab Visual Display

**Test:** Open /admin page, navigate to `成本收益统计` tab. Verify tab position is after `系统配置` and before `公告管理`. Verify default time range is `7天`. Verify KPI cards show `总收入`, `总成本`, `利润`, `成功图片数` with proper numeric formatting. Verify SVG trend chart renders colored lines (blue=revenue, amber=cost, green/red=profit, violet=success images) with legend. Verify `端点拆分` table has columns: `端点标识`, `成功图片数`, `收入`, `成本`, `利润`, `利润率`. Verify `用户拆分` table has columns: `用户标识`, `成功图片数`, `收入`, `成本`, `利润`, `利润率`. Test range filter buttons `今日`, `7天`, `30天`, `全部`. Verify empty state message `暂无成本收益数据` when no billing data exists. Verify `刷新统计` button triggers reload with toast `统计已刷新`.

**Expected:** Analytics tab matches UI-SPEC layout, SVG chart draws correctly, tables sort properly, time filters work.

**Why human:** SVG rendering, KPI formatting, table alignment, interactive chart elements are visual behaviors.

#### 3. End-to-End Billing Flow

**Test:** Configure an endpoint with `costPerImageX10000` and global `salePriceX10000` via admin UI. Generate at least 1 image. After generation completes, check the `billing_records` SQLite table for a row matching the task. Verify that `success_image_count = 1`, the endpoint snapshot and cost/sale x10000 values match the config at generation time, and `profit_x10000 = salePriceX10000 - costPerImageX10000`. Delete the task and user if needed — verify billing row persists.

**Expected:** billing_records row created with correct snapshots after each successful image generation. Rows survive task/user deletion.

**Why human:** Requires running server, authenticating, generating real images, and inspecting SQLite database — cannot be done via static analysis.

### Gaps Summary

No code gaps found. All 28 must-have truths across all 7 plans are verified through codebase evidence. All 16 key artifacts exist, are substantive (not stubs), are wired to their consumers, and have data flowing through real sources (GORM DB queries, fetch API calls). All 12 key links are connected. Backend tests (4 packages) and frontend tests (102 tests, 4 files) all pass. TypeScript build is clean.

The status is `human_needed` because 3 visual/interactive behaviors require human testing: the pricing config UI form validation flow, the analytics tab chart/table rendering, and the end-to-end billing record creation flow. No automated blockers or gaps were found.

**One documentation observation:** COST-01 through COST-04 requirements are declared in PLAN frontmatter and fulfilled by the codebase, but are not present in `.planning/REQUIREMENTS.md`. The REQUIREMENTS.md traceability table was not updated for Phase 5. This does not affect the phase verification result but should be addressed for milestone completion auditability.

---

_Verified: 2026-05-23T10:50:00+08:00_
_Verifier: Claude (gsd-verifier)_
