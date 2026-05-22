---
phase: 05-cost-revenue-analytics
plan: 02
subsystem: backend-config-pricing
tags: [backend, config, pricing, admin-api, tdd]
requires: ["05-01"]
provides: pricing-config-api
affects:
  - backend-go/config/config.go
  - backend-go/handler/admin.go
  - backend-go/main.go
tech-stack:
  added: []
  patterns:
    - "Fixed-point X10000 integer pricing (cost & salePrice)"
    - "Atomic config.json persistence for endpoints + salePrice"
    - "Gin test-mode handler testing with temp config directories"
    - "Config DI via exported GetRootDir/SetRootDir for test isolation"
key-files:
  created:
    - backend-go/handler/admin_pricing_test.go
  modified:
    - backend-go/config/config.go
    - backend-go/config/config_test.go
    - backend-go/handler/admin.go
    - backend-go/main.go
decisions:
  - "Converted getRootDir from function to variable for test dependency injection"
  - "Added GetRootDir/SetRootDir exported functions for handler tests to isolate config filesystem"
  - "persistPricingConfig writes both apiEndpoints and salePriceX10000 in one atomic write to config.json"
metrics:
  duration: TBD
  completed_date: 2026-05-23
---

# Phase 05 Plan 02: Backend Pricing Configuration Summary

Endpoint cost price and global sale price backend configuration, including config.json persistence, admin read/write API, route registration, and validation tests.

## Tasks Completed

| # | Task | Type | Commits | Key Files |
|---|------|------|---------|-----------|
| 1 | Extend config with fixed-point pricing fields | auto (tdd) | faf864b, f8e6ab9 | config/config.go, config/config_test.go |
| 2 | Add admin pricing config handlers | auto (tdd) | 67231b7, 90e744b | handler/admin.go, handler/admin_pricing_test.go |
| 3 | Register pricing routes | auto | e752d5b | main.go |

## Implementation Details

### Task 1: Config Pricing Fields
- Added `CostPerImageX10000 int64` to `ApiEndpoint` with JSON key `costPerImageX10000`
- Added `SalePriceX10000 int64` to `Config` with JSON key `salePriceX10000`
- Added `GetSalePriceX10000() int64` for runtime access
- Added `SetPricingConfig(eps []ApiEndpoint, salePriceX10000 int64)` for atomic sorted persistence
- Added `persistPricingConfig` writes both `apiEndpoints` and `salePriceX10000` in one file write
- Missing JSON fields decode as `0` for backward compatibility with old `config.json`
- Converted `getRootDir` from function to variable; exported `GetRootDir`/`SetRootDir` for test DI
- 6 new tests: missing fields default, JSON round-trip, getter, priority sort, file persistence, cost preserved by SetEndpoints

### Task 2: Admin Pricing Handlers
- `AdminGetPricingConfig(c *gin.Context)` returns `endpoints`, `salePriceX10000`, `moneyScale`
- `AdminUpdatePricingConfig(c *gin.Context)` validates all fields:
  - Non-negative `salePriceX10000`, non-negative `costPerImageX10000`
  - Valid `baseUrl`, non-negative `maxConcurrency` and `priority`
  - Calls `config.SetPricingConfig` + `service.RefreshLimiters` on success
  - Returns `ok`, `endpoints`, `salePriceX10000`, `moneyScale`
- Updated `AdminUpdateEndpoints` to reject `CostPerImageX10000 < 0`
- 5 new tests: GET returns moneyScale, PUT success, negative sale price rejection, negative cost rejection, existing endpoint route cost validation

### Task 3: Route Registration
- `adminAuth.GET("/config/pricing", handler.AdminGetPricingConfig)`
- `adminAuth.PUT("/config/pricing", handler.AdminUpdatePricingConfig)`
- Only exposed under authenticated admin middleware (`adminAuth` group)
- Existing `/config/endpoints` routes preserved unchanged

## Deviations from Plan

None - plan executed exactly as written.

## TDD Gate Compliance

Both TDD tasks followed the RED (failing tests) -> GREEN (implementation) pattern:

| Task | RED Commit | GREEN Commit |
|------|------------|--------------|
| 1 | faf864b `test(05-02): add failing test for pricing config fields` | f8e6ab9 `feat(05-02): implement pricing config fields and atomic persistence` |
| 2 | 67231b7 `test(05-02): add failing test for admin pricing config handlers` | 90e744b `feat(05-02): implement admin pricing config handlers` |

All tests pass: `cd backend-go && go test ./...` exits 0 for all packages.

## Deferred Items

None.

## Known Stubs

None.

## Threat Flags

None - all threat model mitigations from PLAN.md implemented:
- Input validation (T-05-02-01): non-negative cost/sale price, URL validation, concurrency/priority checks
- No public pricing endpoints (T-05-02-02): routes only under adminAuth middleware
- Audit logging (T-05-02-03): slog.Info on successful pricing config updates

## Self-Check: PASSED

- All 6 key files exist on disk
- All 5 commit hashes verified in git log
- `go test ./...` passes across all packages
