---
phase: 05-cost-revenue-analytics
plan: 05
subsystem: admin-frontend
tags: [pricing, analytics, admin-api, typescript-dtos, tdd]
requires:
  - "05-02"  # Backend pricing config API
  - "05-04"  # Backend analytics API
provides: [adminGetPricingConfig, adminUpdatePricingConfig, adminGetBillingSummary, adminGetBillingTrend, adminGetBillingEndpointBreakdown, adminGetBillingUserBreakdown]
affects: [src/admin/adminApi.ts]
tech-stack:
  added: [vitest]
  patterns: [TDD RED-GREEN, fixed-point money (X10000), typed admin API client, analytics time range filtering]
key-files:
  created:
    - src/admin/adminApi.test.ts
  modified:
    - src/admin/adminApi.ts
decisions:
  - "All money values remain X10000 integers on API boundary; no client-side float conversion"
  - "Analytics response wrappers explicitly typed with meta + data pairs matching backend shapes"
  - "AnalyticsRange restricted to four literals: today, 7d, 30d, all"
  - "Endpoint and user breakdown rows include profitRateBps computed by backend"
  - "Existing adminGetEndpoints and adminUpdateEndpoints left unchanged; new pricing functions are additive"
metrics:
  duration_seconds: 1121
  completed_date: 2026-05-23
  tasks: 2
  files_modified: 2
---

# Phase 5 Plan 5: Admin Analytics & Pricing API Client

Typed TypeScript client functions and DTOs for admin pricing configuration and four billing analytics endpoints, preserving X10000 fixed-point money values from the Go backend.

## One-Liner

Type-safe admin client for pricing config CRUD and billing analytics (summary/trend/endpoint/user breakdown) with fixed-point money DTOs matching Go backend shapes.

## Execution Summary

Implemented two TDD cycles (RED-GREEN) on `src/admin/adminApi.ts`:

1. **Task 1 — Pricing DTOs:** Added `costPerImageX10000` to `ApiEndpoint`, defined `PricingConfigResponse` interface, and created `adminGetPricingConfig()` and `adminUpdatePricingConfig()` client functions. Both hit `/api/admin/config/pricing`.

2. **Task 2 — Analytics DTOs:** Added 10 type exports (`AnalyticsRange`, `AnalyticsMeta`, `BillingSummary`, `BillingSummaryResponse`, `BillingTrendPoint`, `BillingTrendResponse`, `BillingEndpointRow`, `BillingEndpointBreakdownResponse`, `BillingUserRow`, `BillingUserBreakdownResponse`) and 4 client functions (`adminGetBillingSummary`, `adminGetBillingTrend`, `adminGetBillingEndpointBreakdown`, `adminGetBillingUserBreakdown`). Each includes `?range={value}` query parameter.

All 19 vitest tests pass. Existing admin client functions (`adminGetEndpoints`, `adminUpdateEndpoints`) are unchanged.

## Commits

| # | Hash | Message |
|---|------|---------|
| 1 | d5338ec | test(05-05): add failing test for pricing DTOs and admin client functions |
| 2 | 757cf39 | feat(05-05): add pricing DTOs and admin client functions |
| 3 | 0efd706 | test(05-05): add failing tests for analytics DTOs and client functions |
| 4 | 63faf69 | feat(05-05): add analytics DTOs and admin client functions |

## Deviations from Plan

### Auto-fixed Issues

None — plan executed exactly as written.

## TDD Gate Compliance

```
Gate sequence per plan type: execute (individual TDD tasks)

Task 1:
  RED:   d5338ec test(05-05): add failing test for pricing DTOs...
  GREEN: 757cf39 feat(05-05): add pricing DTOs...

Task 2:
  RED:   0efd706 test(05-05): add failing tests for analytics DTOs...
  GREEN: 63faf69 feat(05-05): add analytics DTOs...
```

Each task has its own RED-GREEN pair. No REFACTOR commit was needed (no duplication or cleanup required after minimal implementation).

## Known Stubs

None — all DTOs and client functions are wired to real backend paths with proper query parameter encoding.

## Threat Flags

None — no new security surface introduced. All functions require the existing admin bearer token through the shared `adminRequest()` helper. No new network endpoints, auth paths, or schema changes at trust boundaries.

## Self-Check

- [x] `src/admin/adminApi.ts` exists and contains all required exports
- [x] `src/admin/adminApi.test.ts` exists with 19 passing tests
- [x] Commit d5338ec present: test(05-05): add failing test for pricing...
- [x] Commit 757cf39 present: feat(05-05): add pricing DTOs...
- [x] Commit 0efd706 present: test(05-05): add failing tests for analytics...
- [x] Commit 63faf69 present: feat(05-05): add analytics DTOs...
- [x] `npm run build` produces no new errors (only pre-existing missing module errors from other waves)
- [x] No modifications to STATE.md, ROADMAP.md, or other shared orchestrator artifacts
