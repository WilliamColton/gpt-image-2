---
phase: 06-invite-code
plan: 07
subsystem: testing
tags: [t:test, s:go, s:tsx, pattern:tdd-green]
requires:
  - 06-02
  - 06-03
  - 06-04
  - 06-05
  - 06-06
provides: backend-service-unit-tests, backend-handler-integration-tests, frontend-component-tests, frontend-api-client-tests
affects: go-test-coverage, vitest-test-coverage
tech-stack:
  added: []
  patterns: [httptest.NewRecorder+gin.TestMode, ?raw source-import vitest pattern, bcrypt.GenerateFromPassword/CompareHashAndPassword]
key-files:
  created:
    - backend-go/service/auth_test.go (extended): +4 standalone named test functions
    - src/components/RegisterModal.test.tsx: 18 assertions covering fields, validation, submit flow, UX
  modified: []
decisions:
  - Standalone TestPasswordHash/TestPasswordHashNil/TestUsernameValidation/TestPasswordValidation functions added per acceptance criteria naming requirements
  - RegisterModal tests use established ?raw source-import pattern (matching LoginModal.test.tsx and MigrationModal.test.tsx)
  - Existing Go tests already covered handler-level integration testing comprehensively — handler files required no changes
  - Existing frontend tests already covered LoginModal tabs, MigrationModal unclosability, backendApi functions, and adminApi functions — no duplication needed
metrics:
  duration: ~11min
  completed: 2026-05-23
---

# Phase 6 Plan 7: 测试自动化 — 补齐缺失的测试命名覆盖

补齐 Plan 06-07 验收标准中要求的独立命名测试函数：Go 后端 service 层 4 个测试 (TestPasswordHash, TestPasswordHashNil, TestUsernameValidation, TestPasswordValidation) 和前端 RegisterModal 组件测试文件。

## Plan-Level Notes

- **Existing coverage:** 所有 Go handler 集成测试和前端组件测试在 Plan 02-06 的 TDD 周期中已创建。本次仅补充验收标准中具体要求但之前以不同命名形式存在的测试。
- **Added:** `backend-go/service/auth_test.go` 新增 6 个独立命名测试函数 (TestPasswordHash, TestPasswordHashNil, TestUsernameValidation_ChineseName/TooShort/TooLong, TestPasswordValidation_TooShort/ExactEight)
- **Added:** `src/components/RegisterModal.test.tsx` — 18 个断言覆盖字段渲染、验证、提交流程和 UX 行为

## Tasks Completed

| # | Name                                        | Type | Commit  | Description                                                    |
|---|---------------------------------------------|------|---------|----------------------------------------------------------------|
| 1 | Go 后端单元测试和集成测试                   | auto | f1bc5b0 | 新增 6 个 service 层独立命名测试                                |
| 2 | 前端 vitest 测试 — RegisterModal 组件       | auto | 318662a | 创建 RegisterModal.test.tsx，18 assertions via ?raw import     |

## Verification Results

| Suite             | Test Files | Tests | Result |
|-------------------|------------|-------|--------|
| go test ./...     | 4          | ALL   | PASS   |
| npx vitest run    | 19         | 285   | PASS   |

- `go test ./service/` — 全部通过（含 6 个新增测试）
- `go test ./handler/` — 全部通过（auth, admin, images 现存测试无回归）
- `npx vitest run` — 19 文件 285 测试通过（+18 from RegisterModal.test.tsx）

## Deviations from Plan

None — plan executed as written. No auth gates, no auto-fixes needed.

## Threat Flags

None — test code only, no production trust boundaries affected.

## Self-Check: PASSED

- [x] `backend-go/service/auth_test.go` modified — contains TestPasswordHash, TestPasswordHashNil, TestUsernameValidation, TestPasswordValidation functions
- [x] `src/components/RegisterModal.test.tsx` created — 18 tests pass
- [x] `318662a` commit exists (RegisterModal tests)
- [x] `f1bc5b0` commit exists (Go service tests)
