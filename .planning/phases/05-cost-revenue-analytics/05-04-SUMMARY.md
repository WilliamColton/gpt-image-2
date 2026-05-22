---
phase: 05-cost-revenue-analytics
plan: 04
subsystem: backend
tags:
  - analytics
  - billing
  - admin-api
  - fixed-point
  - tdd
requires:
  - "05-01" # BillingRecord model + billing write service
  - "05-02" # adminAuth middleware + admin handler patterns
provides:
  - "05-05" # TypeScript client interfaces consume these endpoint shapes
affects: []
tech-stack:
  added: []
  patterns:
    - Analytics service returns (data, meta, error) triple with explicit MoneyScale
    - Admin analytics handlers use ?range= query parameter with server-side parsing
    - SQLite strftime('%Y-%m-%d', created_at/1000, 'unixepoch') for date bucketing
    - profitRateBps = profitX10000 * 10000 / revenueX10000 (0 when revenue is 0)
key-files:
  created:
    - backend-go/service/analytics.go
    - backend-go/service/analytics_test.go
    - backend-go/handler/admin_analytics_test.go
  modified:
    - backend-go/handler/admin.go
    - backend-go/main.go
decisions:
  - EndpointLabel defaults to endpoint base URL in endpoint breakdown rows
  - All four analytics endpoints use independent service calls for partial failure isolation
  - Invalid range values return HTTP 400 with JSON error instead of silent fallback
  - Empty range query defaults to 7d, matching research recommendation
metrics:
  duration: "~15min"
  completed_date: "2026-05-23T01:55:00+08:00"
---

# Phase 05 Plan 04: Backend Analytics API Summary

Billing analytics aggregation service and four authenticated admin API endpoints (summary, trend, endpoint breakdown, user breakdown) with explicit meta wrappers and fixed-point money fields.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed incorrect cost assertion in endpoint breakdown test**
- **Found during:** Task 1 GREEN
- **Issue:** Test seeded api1 rows with CostX10000 summing to 20000 but asserted 30000
- **Fix:** Corrected assertion value from 30000 to 20000 in `TestGetBillingEndpointBreakdown`
- **Files modified:** backend-go/service/analytics_test.go
- **Commit:** b890f74

**2. [Rule 1 - Bug] Fixed timezone boundary issue in handler data test**
- **Found during:** Task 2 GREEN
- **Issue:** `TestAdminBillingSummary_ContainsData` used UTC noon timestamp with `range=today`, but `time.Now()` in the handler runs in local timezone (UTC+8), causing the record to fall outside today's range
- **Fix:** Changed to use `now.Add(-1*time.Hour)` and `range=7d` for timezone-agnostic test data
- **Files modified:** backend-go/handler/admin_analytics_test.go
- **Commit:** a68f6dc

## TDD Gate Compliance

| Gate | Commit | Status |
|------|--------|--------|
| RED (Task 1 - Service) | 175c12f | PASSED |
| GREEN (Task 1 - Service) | b890f74 | PASSED |
| RED (Task 2 - Handlers) | 5875809 | PASSED |
| GREEN (Task 2 - Handlers) | a68f6dc | PASSED |

## Verification

All tests pass:
- `go test ./service/...` -- 16 analytics tests + 11 pre-existing tests (PASS)
- `go test ./handler/...` -- 10 analytics handler tests + 5 pre-existing pricing tests (PASS)
- `go test ./config/...` -- 13 pre-existing tests (PASS, no regressions)
- `go test ./database/...` -- 2 pre-existing tests (PASS, no regressions)

## Self-Check: PASSED

- [x] backend-go/service/analytics.go exists
- [x] backend-go/service/analytics_test.go exists
- [x] backend-go/handler/admin.go updated with 4 analytics handlers
- [x] backend-go/handler/admin_analytics_test.go exists
- [x] backend-go/main.go updated with 4 analytics routes
- [x] All 4 commits verified in git log
- [x] No stubs or placeholders found
