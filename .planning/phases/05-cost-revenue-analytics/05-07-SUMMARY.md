---
phase: 05-cost-revenue-analytics
plan: 07
subsystem: admin-dashboard-analytics
tags: [analytics, dashboard, charts, tables, kpi]
depends_on: ["05-06"]
requires: [adminGetBillingSummary, adminGetBillingTrend, adminGetBillingEndpointBreakdown, adminGetBillingUserBreakdown]
provides: [admin analytics tab with KPI, trend chart, endpoint/user breakdowns]
affects: [src/admin/AdminDashboard.tsx, src/admin/AdminDashboard.test.tsx]
tech-stack:
  added: []
  patterns: [TDD source-check tests, inline SVG chart, independent analytics loaders, moneyScale-aware formatting]
key-files:
  created: []
  modified:
    - src/admin/AdminDashboard.tsx
    - src/admin/AdminDashboard.test.tsx
key-decisions:
  - "Used inline SVG for lightweight trend charts — no chart library dependency"
  - "Independent loader state per analytics block (summary/trend/endpoint/user) for partial failure resilience"
  - "moneyScale read from response meta, never hardcoded — formatMoneyX10000 accepts scale parameter"
  - "Endpoint/user breakdown tables side-by-side on desktop, stacked on mobile"
decisions: []
metrics:
  start_time: "2026-05-23T02:30:00Z"
  end_time: "2026-05-23T02:43:49Z"
  duration: "~14 minutes"
  phase: 05-cost-revenue-analytics
  plan: 07
  completed_date: "2026-05-23"
---

# Phase 5 Plan 7: Analytics Tab Summary

**One-liner:** Added server-driven admin analytics tab with KPI cards, moneyScale-aware inline SVG trend chart, and endpoint/user breakdown tables

## Tasks Executed

| # | Type | Name | Commit | Status |
|---|------|------|--------|--------|
| 1 | auto (tdd) | Add analytics tab shell, KPI cards, and lightweight trend chart | `4d4c73d` (RED), `aa9aca7` (GREEN) | Complete |
| 2 | auto (tdd) | Render endpoint and user breakdown tables with independent loaders | `967e8ee` (RED), `a93f138` (GREEN) | Complete |

## TDD Gate Compliance

| Gate | Commit | Status |
|------|--------|--------|
| RED Task 1 | `4d4c73d` — test(05-07): add failing tests for analytics tab, KPI cards, and trend chart | PASS |
| GREEN Task 1 | `aa9aca7` — feat(05-07): add analytics tab shell, KPI cards, and lightweight trend chart | PASS |
| RED Task 2 | `967e8ee` — test(05-07): add failing tests for endpoint and user breakdown tables | PASS |
| GREEN Task 2 | `a93f138` — feat(05-07): add endpoint and user breakdown tables with independent loaders | PASS |

All 36 tests pass in final run.

## Deviations from Plan

None — plan executed exactly as written. One minor test assertion adjustment: the `useCallback`-wrapped `formatMoneyX10000` required a broader regex pattern in the source-check test; corrected during RED phase iteration.

## Verification

- **Tests:** 36/36 passing (`npx vitest run src/admin/AdminDashboard.test.tsx`)
- **Build:** No AdminDashboard.tsx errors. Two pre-existing errors in `src/App.tsx` (missing `ChangelogModal`) and `src/components/Header.tsx` (missing `FeedbackModal`) from Plan 05-06 — outside scope boundary.
- **Acceptance criteria:** All criteria met per plan checklist.

## Implementation Highlights

### Task 1: Analytics Tab Shell
- `Tab` union type extended with `analytics`, tab trigger placed between `config` and `announcement`
- Analytics range defaults to `'7d'` with range filter buttons (`今日`, `7天`, `30天`, `全部`)
- `loadAnalyticsSummary` and `loadAnalyticsTrend` with independent loading/error states
- KPI grid: `总收入` (blue), `总成本` (amber), `利润` (green/red), `成功图片数` (violet)
- Inline SVG trend chart with colored lines (revenue blue, cost amber, profit emerald) + success images bar chart
- `formatMoneyX10000(value, moneyScale?)` reads scale from response `meta.moneyScale`
- Header refresh supports analytics tab with `统计已刷新` toast
- Empty state: `暂无成本收益数据` with descriptive body text

### Task 2: Breakdown Tables
- `endpointRows`/`userRows` state with independent loading/error (`endpointLoading`, `endpointError`, etc.)
- `loadAnalyticsEndpointBreakdown` and `loadAnalyticsUserBreakdown` independent of summary/trend
- All four loaders wired for tab load, range change, header refresh, and all retry buttons
- Endpoint table: `端点标识, 成功图片数, 收入, 成本, 利润, 利润率`
- User table: `用户标识, 成功图片数, 收入, 成本, 利润, 利润率`
- Money formatting uses per-response `meta.moneyScale` via `endpointMeta`/`userMeta`
- `profitRateBps` formatted as percent with 2 decimals
- Error text: `统计数据加载失败，请点击"刷新统计"重试；保存价格失败时，请检查金额是否为数字且最多 4 位小数。`
- No CSV/PDF export, no row click drill-down, no export/drill-up labels

## Self-Check: PASSED

- [x] `.planning/phases/05-cost-revenue-analytics/05-07-SUMMARY.md` — created
- [x] `4d4c73d` — test commit for Task 1 RED exists
- [x] `aa9aca7` — feat commit for Task 1 GREEN exists
- [x] `967e8ee` — test commit for Task 2 RED exists
- [x] `a93f138` — feat commit for Task 2 GREEN exists
- [x] All 36 tests passing
