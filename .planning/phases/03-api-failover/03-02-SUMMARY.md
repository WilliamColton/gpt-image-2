---
phase: 03-api-failover
plan: 02
subsystem: api
tags: [failover, openai, error-handling, retry]

requires:
  - phase: 03-api-failover/01
    provides: "ApiEndpoint struct, GetEndpointPool(), config layer"
provides:
  - "Failover engine: withFailover + isRetryableError"
  - "All Call* functions support endpoint pool via variadic parameter"
  - "Handler wired to pass endpoint pool to all service calls"
affects: [service, handler]

tech-stack:
  added: []
  patterns: ["failover-with-error-classification", "variadic-endpoint-pool"]

key-files:
  created:
    - backend-go/service/openai_failover_test.go
  modified:
    - backend-go/service/openai.go
    - backend-go/handler/generate.go
    - backend-go/handler/config.go

key-decisions:
  - "Endpoint apiKey overrides caller apiKey when non-empty; empty falls back to caller"
  - "Immediate switching: no delay between endpoint attempts"
  - "Non-retryable errors (4xx except 429) fail immediately without trying other endpoints"

patterns-established:
  - "Failover pattern: withFailover loop with isRetryableError classification"
  - "Variadic endpoint parameter: endpoints ...config.ApiEndpoint on all public Call* functions"

requirements-completed: [FAILOVER-02, FAILOVER-03]

duration: 12min
completed: 2026-05-08
---

# Phase 03: API Failover Summary

**Failover engine with isRetryableError classification — network/429/5xx errors auto-retry on next endpoint, 4xx fails fast**

## Performance

- **Duration:** ~12 min
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Implemented `isRetryableError` — classifies network errors, HTTP 429, HTTP 5xx as retryable; 400/401/403/404 as non-retryable
- Implemented `withFailover` execution loop — iterates endpoint pool in order, uses endpoint's own apiKey or falls back to caller's key
- Refactored all 4 `Call*` functions (`CallImagesGenerations`, `CallImagesEdits`, `CallImagesGenerationsConcurrent`, `CallImagesEditsConcurrent`) to accept endpoint pool via variadic parameter
- Fixed `handler/config.go` to use first endpoint's BaseURL for backward-compatible config response
- Wired `GetEndpointPool()` into `handler/generate.go` — all service calls now use failover

## Task Commits

Each task was committed atomically:

1. **Task 1: Add failover engine and error classification** - `a6b016a` (feat)
2. **Task 2: Wire endpoint pool into handler** - `bee6360` (feat)

## Files Created/Modified
- `backend-go/service/openai.go` — Failover engine (isRetryableError, withFailover), refactored Call* functions with variadic endpoints
- `backend-go/service/openai_failover_test.go` — 5 test functions for error classification (network, 429, 5xx, non-retryable, nil)
- `backend-go/handler/generate.go` — Wired GetEndpointPool() to all 4 service call sites
- `backend-go/handler/config.go` — Fixed for removed Defaults.BaseURL, uses first endpoint's URL

## Decisions Made
- Endpoint apiKey overrides caller apiKey when non-empty; empty string falls back to caller's key (D-03)
- Immediate switching strategy — no delay between endpoint attempts (D-02)
- Non-retryable errors (4xx except 429) fail immediately without trying other endpoints
- `handler/config.go` uses first endpoint's BaseURL for backward-compatible public config response

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] handler/config.go referenced removed Defaults.BaseURL**
- **Found during:** Task 1 implementation
- **Issue:** `handler/config.go:14` referenced `config.App.Defaults.BaseURL` which was removed in Plan 01
- **Fix:** Updated to use `config.App.GetEndpointPool()[0].BaseURL` with empty-slice guard
- **Files modified:** backend-go/handler/config.go
- **Verification:** Code review confirms correct endpoint pool usage

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Fix necessary for compilation after Plan 01's Defaults.BaseURL removal. No scope creep.

## Issues Encountered
- Sandbox blocked git operations — orchestrator committed on behalf of executor

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 3 complete — failover mechanism active for all image generation/editing paths
- Users need to add `apiEndpoints` array to their `config.json` with at least one endpoint

---
*Phase: 03-api-failover*
*Completed: 2026-05-08*
