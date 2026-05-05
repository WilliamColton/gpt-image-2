---
phase: 02-frontend-adapter
plan: 5
subsystem: ui
tags: [react, typescript, polling, backend-api, zustand]

# Dependency graph
requires:
  - phase: 01-backend-proxy
    provides: backend task management and image storage APIs
provides:
  - Backend submit functions for generate/edit/responses tasks
  - Polling-based task execution replacing direct OpenAI API calls
  - Clean api.ts with only utility re-exports
affects: [02-frontend-adapter]

# Tech tracking
tech-stack:
  added: []
  patterns: [backend-submit-with-polling, task-status-polling-loop]

key-files:
  created: []
  modified:
    - src/lib/backendApi.ts
    - src/lib/api.ts
    - src/store.ts

key-decisions:
  - "Renamed getTasks import to fetchTasks to avoid naming ambiguity with store actions"
  - "Simplified executeTask signature to just taskId since images are pre-uploaded"

patterns-established:
  - "Backend submit + polling pattern: submit task to backend, poll getTasks() every 1s until done/error/timeout"

requirements-completed: [FE-01, FE-02, FE-05]

# Metrics
duration: 4min
completed: 2026-05-05
---

# Phase 2 Plan 5: Summary

**Replaced direct OpenAI API calls with backend task submission and 1s polling loop for connection-resilient image generation**

## Performance

- **Duration:** 4 min
- **Started:** 2026-05-05T16:37:15Z
- **Completed:** 2026-05-05T16:40:36Z
- **Tasks:** 4
- **Files modified:** 3

## Accomplishments
- executeTask no longer calls OpenAI directly -- submits to backend via submitGenerateTask/submitEditTask/submitResponsesTask
- Polling loop checks task status every 1s with configurable timeout
- api.ts cleaned of all OpenAI API calling logic, now only re-exports normalizeBaseUrl
- Output images cached from backend on task completion via setCacheFromIdbOrRemote

## Task Commits

Each task was committed atomically:

1. **Task 1: Add backend submit functions** - `26bbdd9` (feat)
2. **Task 2: Rewrite executeTask with polling** - `bca76b4` (feat)
3. **Task 3: Clean up api.ts** - `4cb889c` (refactor)
4. **Task 4: Update store.ts imports** - (included in Task 2 commit)

**Plan metadata:** (pending)

## Files Created/Modified
- `src/lib/backendApi.ts` - Added submitGenerateTask, submitEditTask, submitResponsesTask functions
- `src/store.ts` - Rewrote executeTask to use backend submit + polling; updated imports
- `src/lib/api.ts` - Removed all OpenAI API calling functions; kept normalizeBaseUrl re-export

## Decisions Made
- Renamed getTasks import to fetchTasks to avoid naming ambiguity with store actions
- Simplified executeTask signature to just taskId since images are pre-uploaded before execution
- Removed inputImageDataUrls resolution block from submitTask (no longer needed for backend submission)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Frontend now submits tasks to backend endpoints (/api/generate, /api/edit, /api/responses-generate)
- Backend endpoints for these routes need to be implemented for the polling to receive results
- Ready for settings UI update (Plan 06) and E2E testing (Plan 07)

---
*Phase: 02-frontend-adapter*
*Completed: 2026-05-05*

## Self-Check: PASSED

- SUMMARY.md: FOUND
- src/lib/backendApi.ts: FOUND
- src/store.ts: FOUND
- src/lib/api.ts: FOUND
- All 4 task commits present in git log
