# GPT Image Playground Stability Program

## What This Is

GPT Image Playground is an existing React/PWA + Go/Gin backend application for OpenAI-compatible image generation and editing. It already supports user authentication, task submission, image upload/storage, Server-Sent Events task updates, admin management, endpoint failover, quotas, invite flows, billing records, announcements, changelog, and feedback.

This project initialization defines the next milestone: fix known bugs and improve reliability where defects reduce customer experience, increase operational cost, or weaken operational trust.

## Core Value

Users and admins can rely on image generation, authentication, task lifecycle, quota, and billing workflows to behave correctly without manual recovery or avoidable support cost.

## Requirements

### Validated

- ✓ Users can authenticate with redemption codes, username/password registration, migration flows, and JWT-backed sessions — existing
- ✓ Users can submit text-to-image and image-edit tasks through the React workspace — existing
- ✓ Users can upload, retrieve, reuse, and delete image assets backed by local filesystem storage and SQLite metadata — existing
- ✓ Users can receive generation progress and completion updates through SSE with polling-style backend task reads — existing
- ✓ Admins can manage users, redemption codes, API endpoints, pricing, analytics, announcements, feedback, changelog, and invite settings — existing
- ✓ Backend can call OpenAI-compatible image APIs through prioritized endpoints with failover, concurrency limits, and per-image billing attribution — existing
- ✓ Frontend and backend have targeted unit, service, and handler tests for API clients, store behavior, auth, billing, images, config, and handlers — existing

### Active

- [ ] Fix concurrent first-use redemption-code login so only one user can claim a single unused code and only successful claimants receive access/quota.
- [ ] Fix queued/running task deletion so deleted work cannot silently continue, recreate the task, consume quota, create billing rows, or orphan generated outputs.
- [ ] Fix the backend `needsMigration` JSON contract so service tests and API behavior agree on whether false values are omitted or explicitly returned.
- [ ] Restore the frontend test command so `npm test` can run reliably in a fresh or repaired checkout instead of failing before Vitest starts.
- [ ] Fix invite registration so invited-user creation and inviter reward updates succeed or fail atomically, with no silent reward loss after user creation.
- [ ] Strengthen admin middleware so admin access checks validate current persisted admin/account state rather than trusting any still-valid admin-role JWT alone.

### Out of Scope

- Broad productionization work such as Docker, CI/CD pipelines, hosted deployment manifests, external monitoring, or infrastructure automation — valuable, but outside the known-bug milestone.
- Large architectural refactors such as splitting `src/store.ts`, extracting every admin dashboard tab, or replacing GORM AutoMigrate — defer unless directly required to fix an active bug.
- New product features, new image provider capabilities, UI redesign, payment integrations, or social/community features — this milestone is reliability-focused.
- Full migration from SQLite/local filesystem to network database/object storage — not required to fix the current known bug list.
- Comprehensive security hardening such as rate limiting, HttpOnly cookie auth, signed image URLs, or CORS policy redesign — defer unless needed for the admin middleware fix.

## Context

This is a brownfield stability milestone based on the existing codebase map in `.planning/codebase/`. The codebase currently has a React 19 + Vite + Zustand frontend and a Go 1.25 + Gin + GORM/SQLite backend. Runtime images are stored on the local filesystem, task metadata and billing records are stored in SQLite, and external generation calls go through OpenAI-compatible endpoint configuration.

The highest-priority known bugs from the codebase map are concentrated around concurrency, task lifecycle, auth/session trust, invite/quota accounting, test reliability, and API contract drift. These issues matter because they can directly harm customer experience, cause quota or billing divergence, waste provider/API resources, create support load, or reduce confidence in releases.

Planning should favor small, verifiable bug fixes with targeted tests over broad rewrites. Each fix should include regression coverage at the layer where the bug occurs: backend service/handler tests for transactional and auth behavior, frontend tests where user-visible behavior or test runner reliability is affected, and explicit manual verification steps for task lifecycle behavior when needed.

## Constraints

- **Scope**: Use the known bug list from the codebase map as the v1 source of truth — the user chose not to broaden this roadmap into general production hardening.
- **Prioritization**: Balance customer experience, operational cost, and safety/reliability rather than optimizing only one category.
- **Workflow**: Interactive GSD mode with standard phase granularity, parallel execution where independent plans allow it, and research/plan-check/verifier enabled.
- **Tech stack**: Preserve the current React/Vite/Zustand frontend and Go/Gin/GORM/SQLite backend unless a local change is necessary for a bug fix.
- **Persistence**: Do not read or commit `backend-go/config.json`, `dev-proxy.config.json`, runtime SQLite files, uploaded images, `dist/`, `node_modules/`, or `.claude/worktrees/`.
- **Testing**: Prefer co-located Vitest tests for frontend logic and Go `testing` with temporary SQLite/Gin `httptest` for backend service/handler behavior.
- **Security**: Treat auth, admin access, token handling, upload paths, and endpoint API keys as sensitive boundaries; validate at HTTP/API boundaries and avoid exposing secrets.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Treat this as a brownfield stability/bug-fix milestone | The application already has major product capabilities; current value comes from making them reliable | — Pending |
| Use the codebase map known bugs as v1 scope | The user wants to focus on known defects rather than broad hardening or new features | — Pending |
| Balance customer experience, operational cost, and safety/reliability | The user explicitly chose balanced prioritization across bug classes | — Pending |
| Use standard phase granularity with interactive checkpoints | Stability fixes need reviewable slices without over-fragmenting every bug | — Pending |
| Keep research, plan checking, and verifier enabled | Bug-fix phases need implementation-specific context and regression validation | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-05-25 after initialization*
