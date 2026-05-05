---
plan: 7
phase: 2
status: complete
completed: 2026-05-06
---

# Summary: 端到端测试验证

## What Changed

Updated test suite to verify the new backend submit + polling flow introduced in Plan 05.

### store.test.ts
- Added module mocks for `backendApi` (submitGenerateTask, fetchTasks, putRemoteTask, uploadImage) and `db`
- New `submitTask backend submission flow` test suite with 6 tests:
  - Creates task with running status immediately
  - Calls submitGenerateTask to submit to backend
  - Polls for completion and updates task to done
  - Updates task to error when backend returns error status
  - Does not submit when prompt is empty
  - Does not submit when not authenticated

### api.test.ts
- Removed obsolete `callImageApi` tests (function removed in Plan 05)
- Added minimal `normalizeBaseUrl` tests (the only remaining export)

## Test Results

All 66 tests pass across 10 test files.

## Manual Verification Notes

Tasks 3-6 from the plan (manual verification of Images API, Responses API, Codex CLI, Edit modes) require a running backend with OPENAI_API_KEY configured. These should be verified by the user before marking Phase 2 complete.

## Key Files
- src/store.test.ts
- src/lib/api.test.ts
