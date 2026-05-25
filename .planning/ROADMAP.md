# Roadmap: GPT Image Playground Stability Program

## Overview

This MVP brownfield stability roadmap fixes the selected known bugs from the codebase map without expanding into v2 production hardening, broad security redesign, architecture migration, CI/CD, deployment, or new product features. The phases are vertical bug-fix slices: first restore the local test command needed for confidence, then harden access correctness, task deletion/finalization, invite quota accounting, and finally close the milestone with full backend and regression verification.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions, if needed later

- [ ] **Phase 1: Vitest Harness Baseline** - `npm test` starts reliably and dependency setup is explicit.
- [ ] **Phase 2: Authentication and Admin Access Integrity** - Redemption-code races, admin authorization freshness, and auth contract drift are fixed together.
- [ ] **Phase 3: Task Deletion and Generation Finalization Integrity** - Deleted queued/running tasks cannot reappear or finalize side effects after deletion.
- [ ] **Phase 4: Invite Reward Atomicity** - Invited-user creation and inviter quota rewards succeed or fail as one unit.
- [ ] **Phase 5: Stability Regression Closure** - Fixed known bugs are covered by targeted regressions and the backend suite passes.

## Phase Details

### Phase 1: Vitest Harness Baseline
**Goal:** Developers can run the JavaScript test command for this checkout without Rollup optional-package startup failure.
**Mode:** mvp
**Depends on:** Nothing (first phase)
**Requirements:** TEST-01, TEST-02
**Success Criteria** (what must be TRUE):
  1. Running `npm test` reaches Vitest execution instead of failing during Rollup startup because an optional native package is missing.
  2. A fresh or repaired checkout has an explicit dependency repair path that restores the optional native package state needed by Rollup/Vitest.
  3. The test setup failure mode is no longer a guessing exercise: the required install state is documented or enforced in project scripts/manifests.
**Plans:** 2 plans

Plans:
- [ ] 01-01: Restore `npm test` startup by correcting the local dependency/test-runner path.
- [ ] 01-02: Document or enforce the Rollup/Vitest optional dependency recovery path.

### Phase 2: Authentication and Admin Access Integrity
**Goal:** Users and admins only receive access that is currently authorized by persisted account state and the documented auth API contract.
**Mode:** mvp
**Depends on:** Phase 1
**Requirements:** AUTH-01, AUTH-02, AUTH-03, AUTH-04, AUTH-05
**Success Criteria** (what must be TRUE):
  1. Concurrent first-use login attempts with the same unused redemption code leave exactly one successful user/code claim.
  2. A losing concurrent redemption-code request receives no usable token, no usable extra account, and no quota grant.
  3. Protected admin API requests reject a validly signed admin-role JWT when the persisted admin/account state no longer authorizes that subject.
  4. Protected admin API requests still allow a currently valid admin identity after checking persisted state.
  5. Auth responses expose `needsMigration` according to one explicit backend contract that is proven by service tests.
**Plans:** 3 plans

Plans:
- [ ] 02-01: Make redemption-code first-use claims atomic and cover concurrent success/failure outcomes.
- [ ] 02-02: Enforce persisted admin/account authorization in admin middleware and cover stale-token denial.
- [ ] 02-03: Align the `needsMigration` JSON contract with backend service tests.

### Phase 3: Task Deletion and Generation Finalization Integrity
**Goal:** Deleted queued or running generation tasks stay deleted and cannot later finalize user, quota, billing, or output state as if they still existed.
**Mode:** mvp
**Depends on:** Phase 2
**Requirements:** TASK-01, TASK-02, TASK-03, TASK-04
**Success Criteria** (what must be TRUE):
  1. A user can delete a queued or running task and it does not reappear after backend generation work completes.
  2. Completion of a deleted queued/running task does not create successful task output state, consume quota, create billing rows, or attach generated output to that deleted task.
  3. Backend finalization checks task existence or cancellation state before persisting completion data.
  4. Regression coverage reproduces and prevents queued/running deletion completion races.
**Plans:** 3 plans

Plans:
- [ ] 03-01: Add task cancellation or tombstone semantics for queued/running deletion.
- [ ] 03-02: Guard generation finalization and side effects when the task no longer exists or is cancelled.
- [ ] 03-03: Add queued/running deletion race regression coverage.

### Phase 4: Invite Reward Atomicity
**Goal:** Invite registration and inviter quota reward accounting cannot diverge silently.
**Mode:** mvp
**Depends on:** Phase 3
**Requirements:** INVT-01, INVT-02, INVT-03
**Success Criteria** (what must be TRUE):
  1. Invited-user registration succeeds only when the invited account creation and inviter reward update both persist.
  2. If inviter quota reward persistence fails, invited-user creation fails and the invited user cannot proceed with a silently missing reward.
  3. Regression coverage proves the invite reward success path and failure rollback path.
**Plans:** 2 plans

Plans:
- [ ] 04-01: Make invited-user creation and inviter reward updates transactional.
- [ ] 04-02: Add invite reward success and failure regression coverage.

### Phase 5: Stability Regression Closure
**Goal:** The fixed known bugs are proven by targeted regression coverage and the backend test suite passes from the backend module.
**Mode:** mvp
**Depends on:** Phase 4
**Requirements:** TEST-03, TEST-04
**Success Criteria** (what must be TRUE):
  1. Running `go test ./...` from `backend-go/` passes after the known bug fixes are in place.
  2. Each fixed known bug in this milestone has targeted regression coverage at the service, handler, or JavaScript test layer where the defect occurred.
  3. Final verification demonstrates the v1 stability bugs are closed without manual data repair or one-off recovery steps.
**Plans:** 2 plans

Plans:
- [ ] 05-01: Run and repair backend-wide test failures related to the fixed known bugs.
- [ ] 05-02: Audit milestone-wide regression coverage and capture final release evidence.

## Requirement Coverage

| Requirement | Phase |
|-------------|-------|
| AUTH-01 | Phase 2 |
| AUTH-02 | Phase 2 |
| AUTH-03 | Phase 2 |
| AUTH-04 | Phase 2 |
| AUTH-05 | Phase 2 |
| TASK-01 | Phase 3 |
| TASK-02 | Phase 3 |
| TASK-03 | Phase 3 |
| TASK-04 | Phase 3 |
| INVT-01 | Phase 4 |
| INVT-02 | Phase 4 |
| INVT-03 | Phase 4 |
| TEST-01 | Phase 1 |
| TEST-02 | Phase 1 |
| TEST-03 | Phase 5 |
| TEST-04 | Phase 5 |

Coverage: 16/16 v1 requirements mapped. No orphaned requirements. No duplicated mappings.

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> 3 -> 4 -> 5

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Vitest Harness Baseline | 0/2 | Not started | - |
| 2. Authentication and Admin Access Integrity | 0/3 | Not started | - |
| 3. Task Deletion and Generation Finalization Integrity | 0/3 | Not started | - |
| 4. Invite Reward Atomicity | 0/2 | Not started | - |
| 5. Stability Regression Closure | 0/2 | Not started | - |
