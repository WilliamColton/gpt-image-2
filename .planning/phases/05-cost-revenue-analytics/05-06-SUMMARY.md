---
phase: 05-cost-revenue-analytics
plan: 06
subsystem: admin-pricing-ui
type: execute
tags: [pricing, admin, ui, frontend]
depends_on:
  - "05-05"
  provides:
    - "AdminDashboard pricing config UI (endpoint costs + global sale price)"
  requires:
    - "adminGetPricingConfig and adminUpdatePricingConfig from 05-05"
  affects:
    - src/admin/AdminDashboard.tsx
    - src/admin/moneyFormat.ts
tech-stack:
  added:
    - "Fixed-point money helpers (x10000 integer arithmetic)"
  patterns:
    - "TDD RED-GREEN cycle with vitest"
    - "React useState + useCallback state management"
    - "String/integer arithmetic for money (D-07 compliance)"
key-files:
  created:
    - src/admin/moneyFormat.ts
    - src/admin/moneyFormat.test.ts
    - src/admin/AdminDashboard.test.tsx
  modified:
    - src/admin/AdminDashboard.tsx
decisions:
  - "moneyFormat helpers live in their own module for testability and reuse"
  - "costInputDrafts uses Record<number, string> to keep precise string drafts until save"
  - "priceErrors is a Record<number, string | null> keyed by endpoint index for inline validation"
  - "save button disabled state computed inline in JSX via validation check closure"
metrics:
  plan_duration:
  start_time: "2026-05-23T10:01:00Z"
  end_time: "2026-05-23T10:22:00Z"
  tasks: 3
  files: 5
completed_date: "2026-05-23"
---

# Phase 05 Plan 06: Admin Pricing Config UI Summary

**One-liner:** Admin system config tab loads and saves endpoint costs and global sale price with fixed-point money helpers, inline validation, and atomic save via the Phase 5 pricing API.

## Tasks Executed

| Task | Name | Type | Commit | Files |
|------|------|------|--------|-------|
| 1 | Load pricing config and add money input state | auto (tdd) | a05d5c3, 3554ade, 2004e7f | src/admin/moneyFormat.ts, src/admin/moneyFormat.test.ts, src/admin/AdminDashboard.tsx |
| 2 | Render endpoint cost and global sale price controls | auto (tdd) | 766f76b, 9eed680 | src/admin/AdminDashboard.tsx, src/admin/AdminDashboard.test.tsx |
| 3 | Save all price fields atomically | auto (tdd) | 95a6d00 | src/admin/AdminDashboard.tsx |

## TDD Gate Compliance

All three tasks followed the RED-GREEN cycle:

| Task | RED commit | GREEN commit |
|------|------------|--------------|
| Task 1 | a05d5c3 — 28 failing tests for formatMoneyInputFromX10000 / parseMoneyInputToX10000 | 3554ade — all 28 pass + 2004e7f — integration |
| Task 2 | 766f76b — 8 failing source-check tests for pricing UI strings | 9eed680 — 7 of 8 pass (remaining 3 are Task 3 scope) |
| Task 3 | (same test file, Task 3 tests still failing after Task 2) | 95a6d00 — all 10 pass |

## Verification

- [x] `npm run build` exits 0 (TypeScript compilation + Vite build)
- [x] `npm run test` exits 0 (88 tests pass across 7 test files)
- [x] AdminDashboard.tsx contains `成本价（元/张）`
- [x] AdminDashboard.tsx contains `全局售价（元/张）`
- [x] AdminDashboard.tsx contains `step="0.0001"`
- [x] AdminDashboard.tsx contains `支持 4 位小数`
- [x] AdminDashboard.tsx contains `请输入非负数字，最多 4 位小数`
- [x] AdminDashboard.tsx contains `handleSavePricingConfig`
- [x] AdminDashboard.tsx calls `adminUpdatePricingConfig` once in save handler
- [x] AdminDashboard.tsx contains button text `保存价格配置`
- [x] AdminDashboard.tsx contains toast text `价格配置已保存`
- [x] No `parseFloat` used in moneyFormat.ts or AdminDashboard.tsx pricing code
- [x] Legacy `保存配置` button text no longer appears in config tab

## Deviations from Plan

None — plan executed exactly as written. All acceptance criteria met.

## Decisions Made

1. **moneyFormat helpers in separate module** — `formatMoneyInputFromX10000` and `parseMoneyInputToX10000` live in `src/admin/moneyFormat.ts` for testability and reuse.
2. **costInputDrafts as Record<number, string>** — Keeps precise string drafts per endpoint index until save, avoiding premature x10000 conversion.
3. **priceErrors as Record<number, string | null>** — Keyed by endpoint index for inline error display below each cost input.
4. **Save button disabled-state computed inline** — Uses an IIFE in JSX to validate all cost drafts + sale price before enabling the save button.

## Implementation Notes

- The `config` tab now loads pricing data via `adminGetPricingConfig()` instead of `adminGetEndpoints()`. This provides both endpoint configuration and pricing (costs + sale price).
- `loadPricingConfig()` converts `costPerImageX10000` and `salePriceX10000` from the API into display strings using `formatMoneyInputFromX10000`.
- `handleSavePricingConfig()` validates all inputs, builds `ApiEndpoint[]` with `costPerImageX10000` integer values, and sends a single `adminUpdatePricingConfig` request.
- On save success, the UI reloads the full pricing config and shows a success toast.
- The legacy `endpointsSaving` state was removed in favor of `pricingSaving`.
- All money parsing uses string/integer arithmetic — no `parseFloat` anywhere in pricing code (D-07 compliance).

## Self-Check: PASSED

All 5 created/modified files exist on disk.
All 6 commits (a05d5c3, 3554ade, 2004e7f, 766f76b, 9eed680, 95a6d00) are in git history.
No file deletions in any commit.
Build passes. Tests pass (88/88).
