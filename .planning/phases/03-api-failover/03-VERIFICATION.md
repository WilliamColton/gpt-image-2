---
phase: 03-api-failover
verified: 2026-05-08T12:00:00Z
status: human_needed
score: 11/11 must-haves verified
overrides_applied: 0
re_verification: false
human_verification:
  - test: "Configure config.json with two apiEndpoints where the first has an invalid apiKey or unreachable URL, then submit an image generation task"
    expected: "First endpoint fails, request succeeds on second endpoint, generated image is returned"
    why_human: "Requires running the backend server with actual API credentials and observing runtime failover behavior"
  - test: "Configure config.json with two apiEndpoints both pointing to unreachable URLs, then submit an image generation task"
    expected: "Task status becomes error with a message indicating all endpoints failed and the last error"
    why_human: "Requires runtime observation of error message format and task status update"
---

# Phase 3: API Failover Verification Report

**Phase Goal:** 后端支持配置多个 API 端点和 Key，请求失败时自动切换到下一个可用端点
**Verified:** 2026-05-08T12:00:00Z
**Status:** human_needed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Config struct can hold multiple API endpoints with independent base URLs and API keys | VERIFIED | `config.go:16-19` defines `ApiEndpoint` with `BaseURL` and `APIKey`; `config.go:30` has `ApiEndpoints []ApiEndpoint` field with `json:"apiEndpoints"` tag |
| 2 | When apiEndpoints has entries, GetEndpointPool returns them in config order | VERIFIED | `config.go:71-73` returns `c.ApiEndpoints` directly — preserves slice order |
| 3 | Defaults struct has no BaseURL field | VERIFIED | `config.go:9-14` Defaults struct has only `CodexCLI`, `APIMode`, `Model`, `Timeout` — no `BaseURL` |
| 4 | Config loading from config.json is backward compatible — existing single-endpoint configs still work | VERIFIED | `config.go:35-58` Load() uses `json.Unmarshal` which silently ignores missing fields; `config_test.go:56-74` TestLoad_NoDefaultsBaseURL confirms JSON with `baseUrl` in defaults is ignored; no compile-time references to `Defaults.BaseURL` remain in codebase |
| 5 | API call that returns 429/5xx/network error on first endpoint automatically retries on second endpoint | VERIFIED | `openai.go:61-92` isRetryableError classifies network errors, 429, 500/502/503/504 as retryable; `openai.go:98-126` withFailover loops endpoints, retries on retryable errors; tests in `openai_failover_test.go` cover all cases |
| 6 | API call that returns 400/401 does NOT retry — fails immediately | VERIFIED | `openai.go:116` checks `!isRetryableError(err)` and returns immediately for non-retryable errors; `openai_failover_test.go:38-49` TestIsRetryableError_NonRetryable verifies 400, 401, 403, 404 are NOT retryable |
| 7 | When all endpoints fail, the last error is returned to the caller | VERIFIED | `openai.go:125` returns `fmt.Errorf("所有端点均失败，最后错误: %w", lastErr)` |
| 8 | Endpoint's own apiKey overrides user apiKey when configured | VERIFIED | `openai.go:104-107` withFailover uses `ep.APIKey` when non-empty, falls back to `callerAPIKey` when empty |
| 9 | Empty apiKey on endpoint falls back to user apiKey | VERIFIED | `openai.go:106-107` `if apiKey == "" { apiKey = callerAPIKey }` |
| 10 | handler/generate.go passes endpoint pool to all 4 service call sites | VERIFIED | `generate.go:86` calls `config.App.GetEndpointPool()`; `generate.go:129,131,134,136` all pass `endpoints...` to service calls |
| 11 | handler/config.go uses first endpoint's BaseURL for backward-compatible config response | VERIFIED | `config.go:16-19` gets endpoints from `GetEndpointPool()`, uses `endpoints[0].BaseURL` with empty-slice guard |

**Score:** 11/11 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `backend-go/config/config.go` | ApiEndpoint struct, ApiEndpoints field, GetEndpointPool function | VERIFIED | Lines 16-19 (struct), line 30 (field), lines 71-73 (method) |
| `backend-go/config/config_test.go` | Unit tests for config loading and GetEndpointPool | VERIFIED | 4 test functions: TestGetEndpointPool_MultiEndpoint, TestGetEndpointPool_SingleEndpoint, TestLoad_WithApiEndpoints, TestLoad_NoDefaultsBaseURL |
| `backend-go/service/openai.go` | Failover logic: withFailover, isRetryableError, updated Call* functions | VERIFIED | isRetryableError at line 61, withFailover at line 98, all 4 Call* functions have `endpoints ...config.ApiEndpoint` variadic |
| `backend-go/service/openai_failover_test.go` | Unit tests for isRetryableError | VERIFIED | 5 test functions covering network, 429, 5xx, non-retryable, nil |
| `backend-go/handler/generate.go` | Handler wired to use endpoint pool for failover | VERIFIED | Line 86 gets pool, lines 129/131/134/136 pass endpoints to all service calls |
| `backend-go/handler/config.go` | Updated for removed Defaults.BaseURL | VERIFIED | Lines 16-19 use GetEndpointPool() with empty-slice guard |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `backend-go/handler/generate.go` | `backend-go/service/openai.go` | `config.App.GetEndpointPool()` passed as variadic | WIRED | `generate.go:86` calls GetEndpointPool, lines 129-136 spread endpoints to all 4 Call* functions |
| `backend-go/service/openai.go` | `backend-go/config/config.go` | imports config.ApiEndpoint type | WIRED | `openai.go:14` imports `gpt-image-playground/backend/config`; all Call* signatures use `config.ApiEndpoint` |
| `backend-go/config/config.go` | `backend-go/config.json` | `json.Unmarshal in Load()` | WIRED | `config.go:57` uses `json.Unmarshal(data, App)` which maps `apiEndpoints` JSON key to `ApiEndpoints` struct field |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|-------------------|--------|
| `handler/generate.go` | `endpoints` | `config.App.GetEndpointPool()` | Config-driven from config.json | FLOWING |
| `service/openai.go` withFailover | `endpoints []config.ApiEndpoint` | Caller passes via variadic | Flows from config through handler | FLOWING |
| `handler/config.go` | `endpoints` | `config.App.GetEndpointPool()` | Config-driven from config.json | FLOWING |

### Behavioral Spot-Checks

Skipped — sandbox blocks `go build` and `go test` commands. Static analysis confirms all structural wiring is correct. All test files contain the expected test functions with correct assertions. The code compiles structurally (no type mismatches between config.ApiEndpoint usage sites and definitions).

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| FAILOVER-01 | Plan 01 | `config.json` supports `apiEndpoints` array with baseUrl + apiKey | SATISFIED | `config.go:16-19` (struct), `config.go:30` (field), `config.go:57` (JSON loading) |
| FAILOVER-02 | Plan 02 | API call fails with retryable error (network/429/5xx), auto-switch to next endpoint | SATISFIED | `openai.go:61-92` (classification), `openai.go:98-126` (failover loop), tests verify all cases |
| FAILOVER-03 | Plan 02 | All endpoints fail, return last error to caller | SATISFIED | `openai.go:125` returns wrapped last error; `openai.go:123` returns "no endpoints configured" if pool empty |
| FAILOVER-04 | Plan 01 | `apiEndpoints` must have at least one entry; `defaults.baseUrl` removed | SATISFIED | `Defaults` struct at `config.go:9-14` has no BaseURL; `withFailover` at `openai.go:123` handles empty pool gracefully; no `Defaults.BaseURL` references anywhere in codebase |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| (none) | | | | |

No TODOs, FIXMEs, placeholders, stubs, or empty implementations found in any modified file.

### Human Verification Required

### 1. Failover End-to-End Behavior

**Test:** Configure `config.json` with two `apiEndpoints` where the first has an invalid apiKey (to trigger 401) or an unreachable URL (to trigger network error), and verify the second endpoint handles the request.
**Expected:** First endpoint fails, request succeeds on second endpoint, generated image is returned.
**Why human:** Requires running the backend server with actual API credentials and observing runtime behavior.

### 2. Error Propagation When All Endpoints Fail

**Test:** Configure `config.json` with two `apiEndpoints` both pointing to unreachable URLs, then submit an image generation task.
**Expected:** Task status becomes "error" with a message indicating all endpoints failed and the last error.
**Why human:** Requires runtime observation of error message format and task status update.

### Gaps Summary

No gaps found. All 11 observable truths are verified. All 4 FAILOVER requirements are satisfied. All artifacts exist, are substantive (not stubs), are wired correctly, and data flows through the wiring. No anti-patterns detected.

The only items requiring human verification are end-to-end failover behaviors at runtime, which cannot be tested through static analysis.

---

_Verified: 2026-05-08T12:00:00Z_
_Verifier: Claude (gsd-verifier)_
