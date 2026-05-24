<!-- refreshed: 2026-05-24 -->
# Architecture

**Analysis Date:** 2026-05-24

## System Overview

```text
┌──────────────────────────────────────────────────────────────────────────┐
│                        Frontend (React SPA)                              │
│                          Vite + TypeScript                               │
├──────────────────────┬──────────────────┬────────────────────────────────┤
│   Main App (src/)    │  Admin Panel     │   Shared Components (ui/)      │
│  `src/App.tsx`       │  `src/admin/`    │   `src/components/ui/`         │
│  Zustand store       │  AdminDashboard  │   Radix UI + Tailwind (Zinc)   │
│  IndexedDB cache     │  adminApi.ts     │   shadcn/ui conventions        │
└──────────┬───────────┴────────┬─────────┴───────────────┬────────────────┘
           │                    │                          │
           ▼                    ▼                          ▼
┌──────────────────────────────────────────────────────────────────────────┐
│              Backend API Server (Go + Gin)                                │
│                         `backend-go/`                                     │
│                                                                           │
│  ┌──────────┐  ┌──────────┐  ┌──────────────┐  ┌──────────────────────┐ │
│  │ handlers │  │ services │  │  middleware   │  │  config (hot-reload) │ │
│  │          │  │          │  │              │  │                      │ │
│  │ auth.go  │  │ auth.go  │  │ AuthMidware  │  │ ApiEndpoint pool     │ │
│  │ generate │  │ openai.go│  │ AdminMidware │  │ JWT + pricing config │ │
│  │ tasks.go │  │ queue.go │  │ RequestLog   │  │ Persisted JSON       │ │
│  │ images.go│  │ billing  │  │              │  │                      │ │
│  │ admin.go │  │ image.go │  │              │  │                      │ │
│  └──────────┘  └──────────┘  └──────────────┘  └──────────────────────┘ │
└──────────┬───────────────────────────────────────────────────────────────┘
           │
           ▼
┌──────────────────────────────────────┐    ┌──────────────────────────────┐
│  SQLite (GORM)                       │    │  File System                 │
│  `backend-go/data/app.sqlite`        │    │  `backend-go/upload/`        │
│  WAL mode, single connection         │    │  Per-user subdirectories     │
│  Tables: users, tasks, images,       │    │  Images stored as files      │
│  billing_records, redemption_codes,  │    │  SHA-256 deduplication       │
│  announcements, feedbacks,           │    │                              │
│  changelog_entries                   │    │                              │
└──────────────────────────────────────┘    └──────────────────────────────┘
           │
           ▼
┌──────────────────────────────────────┐
│  External API: OpenAI Images API     │
│  Multi-endpoint pool with failover   │
│  `/v1/images/generations`            │
│  `/v1/images/edits`                  │
│  Concurrency slot management         │
│  Per-endpoint API key isolation      │
└──────────────────────────────────────┘
```

## Architecture Pattern

**Overall:** Single-Page Application (SPA) with a monolithic REST API backend. The frontend is a Vite-built React SPA with code-split admin route. The backend is a Go/Gin HTTP API server backed by SQLite. No microservices, no message queue -- all state lives in SQLite and the filesystem.

**Key Characteristics:**
- Client-rendered SPA with lazy-loaded admin route (no SSR)
- Monolithic Go backend serving both user-facing and admin APIs
- Polling + SSE dual-strategy for real-time task status updates
- Multi-endpoint API failover pool with per-endpoint concurrency slots
- Image storage on local filesystem with IndexedDB client-side cache
- JWT-based authentication with bearer tokens, persisted in localStorage
- SPA and admin panel share the same Zustand store but render separate React roots

## Component Responsibilities

### Frontend

| Component | Responsibility | File |
|-----------|----------------|------|
| App | Root user-facing app; mounts global UI shell, modals, and providers | `src/App.tsx` |
| AdminPage | Admin route root; conditionally renders login or dashboard | `src/admin/AdminPage.tsx` |
| AdminDashboard | Full admin panel: user management, codes, API endpoints, analytics, announcements, changelog, invites | `src/admin/AdminDashboard.tsx` |
| main.tsx | Entry point; routes / and /admin to separate React roots, registers service worker | `src/main.tsx` |
| store.ts | Zustand global state; auth, settings, tasks, image cache, polling, SSE, all business actions | `src/store.ts` |
| backendApi.ts | User-facing API client; fetch wrappers, auth header injection, SSE streaming | `src/lib/backendApi.ts` |
| db.ts | IndexedDB wrapper for client-side image storage and SHA-256 hashing | `src/lib/db.ts` |
| Header | Navigation bar with settings, help, appearance, feedback buttons | `src/components/Header.tsx` |
| InputBar | Floating bottom bar for prompt input, image uploads, parameter controls | `src/components/InputBar.tsx` |
| TaskGrid | Task list with drag-select, filtering, search | `src/components/TaskGrid.tsx` |
| TaskCard | Individual task display: thumbnail, status, actions | `src/components/TaskCard.tsx` |
| SettingsModal | API mode, model, timeout, Codex CLI, theme settings | `src/components/SettingsModal.tsx` |
| LoginModal | Code-based and password-based login UI | `src/components/LoginModal.tsx` |

### Backend

| Component | Responsibility | File |
|-----------|----------------|------|
| main.go | Server bootstrap: config load, DB init, route registration, middleware chain, server startup | `backend-go/main.go` |
| handler/generate.go | Image generation/editing endpoint; validates request, queues task, spawns async goroutine for execution | `backend-go/handler/generate.go` |
| handler/tasks.go | Task CRUD endpoints + SSE streaming endpoint that polls DB every 1s | `backend-go/handler/tasks.go` |
| handler/auth.go | Authentication: login (code/password), register, redeem, migrate, password change, invite code management | `backend-go/handler/auth.go` |
| handler/admin.go | Admin endpoints: user management, code generation, endpoint config, analytics, announcement, changelog, invites | `backend-go/handler/admin.go` |
| handler/images.go | Image upload/download/delete; handles multipart form uploads and file serving | `backend-go/handler/images.go` |
| handler/config.go | Public config endpoint exposing model, apiMode, timeout, Codex CLI, inviteEnabled | `backend-go/handler/config.go` |
| service/openai.go | OpenAI API client; generations/edits calls, failover orchestration, concurrent multi-image requests, data URL conversion | `backend-go/service/openai.go` |
| service/queue.go | Concurrency slot manager; per-endpoint semaphore pool, blocking acquire with notification | `backend-go/service/queue.go` |
| service/auth.go | JWT sign/verify, user CRUD, redemption codes, password auth (bcrypt), registration, invitation system | `backend-go/service/auth.go` |
| service/billing.go | Billing record creation; immutable cost/revenue/profit snapshots per image | `backend-go/service/billing.go` |
| service/image.go | Image storage: save buffer to filesystem with SHA-256 dedup, read data URL from file, parse/format data URLs | `backend-go/service/image.go` |
| service/task.go | Task CRUD with JSON field marshaling; pending image counting for quota checks | `backend-go/service/task.go` |
| service/analytics.go | Aggregation queries for billing analytics: summary, trend, endpoint/user breakdown | `backend-go/service/analytics.go` |
| middleware/middleware.go | AuthMiddleware (JWT + user status check) and AdminMiddleware (JWT + role check) | `backend-go/middleware/middleware.go` |
| middleware/logger.go | Request logging middleware with method, path, status, latency, IP, user_id | `backend-go/middleware/logger.go` |
| config/config.go | Runtime config: JSON file loading, endpoint pool (hot-swappable via admin), pricing, invite config persistence | `backend-go/config/config.go` |
| database/database.go | GORM initialization; SQLite WAL mode, AutoMigrate, admin/announcement seeding | `backend-go/database/database.go` |
| database/models.go | GORM model definitions for all 8 tables | `backend-go/database/models.go` |

## Layers

### Frontend Layer: React SPA
- **Purpose:** User-facing image generation playground and admin dashboard
- **Location:** `src/`
- **Contains:** Components, Zustand store, API client, utility libraries
- **Depends on:** Backend API (via fetch), OpenAI API indirectly through backend proxy
- **Used by:** End users in browser

### Frontend: Store Layer (Zustand)
- **Purpose:** Centralized state management with localStorage persistence
- **Location:** `src/store.ts`
- **Contains:** Auth state, settings, tasks, UI state, image cache, business logic functions
- **Depends on:** `src/lib/backendApi.ts`, `src/lib/db.ts`
- **Used by:** All React components via `useStore` hook

### Frontend: UI Component Library
- **Purpose:** Reusable UI primitives following shadcn/ui patterns
- **Location:** `src/components/ui/`
- **Contains:** Button, Card, Dialog, Input, Select, Tabs, Switch, Badge, etc. (Radix-based)
- **Depends on:** @radix-ui packages, class-variance-authority, tailwindcss
- **Used by:** Application components and admin dashboard

### Backend: Handler Layer
- **Purpose:** HTTP request parsing, validation, response formatting; delegates to services
- **Location:** `backend-go/handler/`
- **Contains:** Gin handler functions for each route group
- **Depends on:** `backend-go/service/`, `backend-go/middleware/`, `backend-go/config/`
- **Used by:** Gin router in `main.go`

### Backend: Service Layer
- **Purpose:** Business logic; database operations, external API calls, authentication
- **Location:** `backend-go/service/`
- **Contains:** All business logic modules (auth, openai, queue, billing, image, task, analytics, announcement, changelog, feedback)
- **Depends on:** `backend-go/database/`, `backend-go/config/`, `backend-go/util/`
- **Used by:** Handler layer

### Backend: Persistence Layer
- **Purpose:** Database ORM and file storage
- **Location:** `backend-go/database/` and `backend-go/upload/`
- **Contains:** GORM models + SQLite, filesystem image storage per-user
- **Depends on:** SQLite driver, filesystem
- **Used by:** Service layer

## Data Flow

### Primary Request Path: Image Generation

1. **User submits prompt** in `InputBar` component (`src/components/InputBar.tsx`) -- triggers `submitTask()` in store
2. **Frontend creates optimistic task** with status `queued`, assigns a client-side ID, prepends to task list
3. **Image uploads execute** -- if input images exist or a mask draft is present, `uploadImage()` POSTs to `/api/images` per image
4. **Task submitted to backend** via `submitGenerateTask()` or `submitEditTask()` depending on mask presence -- POSTs to `/api/generate` or `/api/edit` in `src/lib/backendApi.ts`
5. **Backend handler validates** request, checks quota (`service.CheckQuota`), creates task record (`service.UpsertTask`), spawns goroutine for async execution (`backend-go/handler/generate.go:76`)
6. **Async goroutine loads images** from filesystem, acquires concurrency slot (`service.AcquireSlotFrom` in `backend-go/service/queue.go`), calls OpenAI API with failover (`service.CallImagesGenerations` or `service.CallImagesEdits` in `backend-go/service/openai.go`)
7. **OpenAI API fails over** across endpoints: tries each in priority order; on failure, advances to next endpoint; stamps attribution on generated images (`backend-go/service/openai.go:62-115`)
8. **Output images saved** to filesystem with SHA-256 dedup (`service.SaveDataURLImage` in `backend-go/service/image.go`)
9. **Billing records created** as immutable snapshots of cost/revenue/profit per image (`service.RecordBillingForSuccessfulImages` in `backend-go/service/billing.go`)
10. **Task marked done** with output image IDs, actual params, elapsed time; persisted to DB
11. **Frontend receives updates** via SSE stream (`/api/tasks/:id/stream`) -- store processes each event, updates task in state, triggers toast on completion (`src/store.ts:696-741`)
12. **SSE fallback to polling**: if SSE connection fails, a 5-second interval poller checks all running/queued tasks (`src/store.ts:79-115`)

### Secondary Flow: Image Retrieval / Caching

1. **TaskCard mounts** and needs a thumbnail: calls `ensureImageCached(id)` (`src/store.ts:132-150`)
2. **Check memory cache** first (Map<string, string>); if hit and is data URL, return immediately
3. **Check IndexedDB** via `getImage(id)` (`src/lib/db.ts:38-40`); if hit, store in memory cache
4. **Fetch from backend** via `fetchAndCacheImage(id)`: GET `/api/images/:id?token=...`, blob response, convert to base64 data URL (`src/store.ts:184-202`)
5. **Store in both caches**: memory Map + IndexedDB via `putImage()` for offline persistence
6. **Image source URL construction**: `getRemoteImageDataUrl(id)` builds `{VITE_BACKEND_URL}/api/images/{id}?token={JWT}` (`src/store.ts:483-485`)

### Tertiary Flow: Settings Sync

1. **On app init**: `initStore()` fetches public announcement, latest changelog, public config in parallel (`src/store.ts:518-541`)
2. **If backend token exists**: `bootstrapBackendSession()` calls `/api/auth/me`, `/api/tasks`, `/api/config/public` to hydrate auth, tasks, settings
3. **Public config** (`/api/config/public`) returns `AppConfig` (model, apiMode, timeout, codexCli, inviteEnabled) -- no auth required
4. **Settings persisted** to localStorage via Zustand `persist` middleware with `partialize` (saves: settings, authUser, params, dismissed flags)

## Frontend-Backend Communication Pattern

**Protocol:** REST (JSON) + SSE (Server-Sent Events) for task streaming

**Transport:**
- All API calls go through fetch-based clients (`src/lib/backendApi.ts` for user, `src/admin/adminApi.ts` for admin)
- JWT bearer token in `Authorization` header for authenticated routes
- Query param `?token=` for image GET requests (allows `<img src>` to work without header injection)
- SSE endpoint at `/api/tasks/:id/stream` uses `text/event-stream` with 1-second DB polling, 15-second heartbeat, 10-minute timeout

**Auth Headers:**
- User: `Authorization: Bearer <JWT>` (from localStorage key `gpt-image-playground-token`)
- Admin: `Authorization: Bearer <admin-JWT>` (from localStorage key `gpt-image-playground-admin-token`)

**Error Handling (Frontend):**
- The `request()` function in `src/lib/backendApi.ts:33-53` catches HTTP errors, parses JSON error/message, throws as Error
- Store actions wrap async calls in try/catch; on failure update task status to `error` with the error message
- Public endpoints (announcement, changelog) silently return null on fetch failure

**Error Handling (Backend):**
- Gin handlers return `gin.H{"error": "message"}` with appropriate HTTP status codes
- Async goroutine failures call `failTask()` which updates the task record with error message and marks it as error
- Service-level errors are logged with `slog.Error()` and returned as formatted errors

## Authentication / Authorization Flow

1. **Login via code**: POST `/api/auth/login` with `{code}` -- validates redemption code in DB, creates user if new or returns existing user; returns JWT
2. **Login via password**: POST `/api/auth/login-password` with `{username, password}` -- bcrypt verification, returns JWT
3. **Register**: POST `/api/auth/register` with `{inviteCode, username, password}` -- creates user, awards inviter/invitee quota, returns JWT
4. **Admin login**: POST `/api/admin/login` with `{apikey}` -- simple string comparison against `config.App.AdminApikey`; issues a JWT with `role=admin`
5. **JWT structure**: HS256 signed with `config.App.JWTSecret`; claims: `sub` (userID), `role` (admin/user), `exp` (30 days)
6. **AuthMiddleware**: Extracts Bearer token from header or `?token=` query param, verifies JWT, looks up user in DB, checks `status != "disabled"`, sets `user` and `userID` in Gin context (`backend-go/middleware/middleware.go:14-48`)
7. **AdminMiddleware**: Verifies JWT and checks `role == "admin"`; returns 403 if not (`backend-go/middleware/middleware.go:60-78`)

## Caching Strategy

**Frontend Image Cache (3-tier):**
1. **Memory cache** (`imageCache` Map<string, string>): keyed by image ID, stores data URLs; cleared on logout/init
2. **IndexedDB** (`gpt-image-playground` / `images` store): persistent browser storage; keyed by SHA-256 hash of image content
3. **Backend fetch** as last resort: GET `/api/images/:id?token=...` with `cache: 'no-store'`

**Backend Image Dedup:**
- SHA-256 hash of image buffer before saving; if same hash already exists for user, skip write and return existing ID (`backend-go/service/image.go:46-49`)

**Backend DB:**
- SQLite with WAL journal mode for concurrent reads
- Single connection pool (`SetMaxOpenConns(1)`) -- serialized writes, safe for concurrent reads via WAL

## Build and Deployment Architecture

**Frontend Build:**
- **Tool:** Vite 6 with `@vitejs/plugin-react`
- **Config:** `vite.config.ts` -- supports dev proxy (`dev-proxy.config.json`), `base: './'` for relative paths
- **Output:** `dist/` directory, served as static files
- **Type checking:** `tsc -b` before `vite build` in production

**Backend Build:**
- **Tool:** Go 1.25 with module support (`go.mod`)
- **Entry:** `backend-go/main.go`
- **Dependencies:** Gin, GORM, SQLite, golang-jwt, bcrypt, OpenAI SDK v3
- **Config:** `backend-go/config.json` (JSON file, mutable at runtime via admin API)

**Dev Proxy:**
- Vite dev server proxies `/api` requests to backend (configurable via `dev-proxy.config.json`)
- Default backend URL from `VITE_BACKEND_URL` env var or `http://localhost:3001` in production builds

**Service Worker:**
- `public/sw.js` registered only in production builds
- Provides PWA offline support
- Unregistered in dev mode

## Key Abstractions

**Endpoint Pool + Failover:**
- Purpose: Multi-provider API key rotation with priority-based ordering and per-endpoint concurrency limits
- Location: `backend-go/config/config.go` (config), `backend-go/service/queue.go` (slots), `backend-go/service/openai.go` (failover)
- Pattern: Each endpoint has `{baseUrl, apiKey, maxConcurrency, priority, costPerImageX10000}`. Failover iterates endpoints in priority order, acquires a concurrency semaphore, calls API. On failure, releases and tries next. On success, stamps endpoint attribution on each generated image.

**Task Lifecycle:**
- States: `queued` -> `running` -> `done` | `error`
- Queued: created by handler, persisted to DB
- Running: set by goroutine after acquiring concurrency slot
- Done: set after all output images saved and billing recorded
- Error: set by `failTask()` on any failure

**Concurrent vs Single Generation:**
- Normal mode (`n=1`): single API call
- Codex CLI + `n>1`: N concurrent API calls (each `n=1`) via goroutines, merged results
- Image edits with `n>1`: N concurrent `/v1/images/edits` calls, merged results

**Codex CLI Compatibility:**
- When `codexCli=true`, prompts are prefixed with `"Use the following text as the complete prompt. Do not rewrite it:\n"` to prevent the API from rewriting user prompts
- When generating multiple images (`n>1`) in Codex CLI mode, N separate concurrent `/v1/images/generations` calls are made instead of one call with `n=N`

**Image ID System:**
- Backend: 21-character random ID using `[a-zA-Z0-9_-]` charset (from `util.GenerateID()`)
- Frontend: 3-segment ID combining timestamp + counter + random (from `genId()` in store)
- Images are deduplicated server-side by SHA-256 hash; same image content returns the same ID

## Architectural Constraints

- **Threading:** Go backend uses goroutines for async task execution. Concurrent API calls use `sync.WaitGroup` to collect results. Concurrency slots use `sync.Map` + channel-based semaphores per endpoint. **Single SQLite connection** (`SetMaxOpenConns(1)`) means all DB writes are serialized.
- **Global state:** Config singleton (`config.App`) is shared across all handlers/services. Endpoint pool is protected by `sync.RWMutex`. Frontend Zustand store is a global singleton accessed via `useStore()` hook.
- **Circular imports:** Not detected.
- **File storage:** Images stored on local filesystem under `backend-go/upload/{userID}/`. No cloud storage integration. No cleanup mechanism for orphaned files when tasks are deleted.

## Error Handling

**Backend Strategy:**
- **Pattern:** Use `slog.Error()` for server-side errors, return structured JSON errors (`gin.H{"error": "message"}`) with appropriate HTTP codes
- **Async tasks:** `failTask()` in `handler/generate.go:197-211` persists error to task record, which surfaces to client via SSE/polling
- **Failover errors:** Each endpoint failure logs a warning; only after all endpoints fail is an error returned
- **Auth errors:** JWT failure returns 401; admin role check returns 403

**Frontend Strategy:**
- **Pattern:** `try/catch` in all async actions; errors are displayed via `showToast()` and/or stored in task records
- **API client:** `request()` in `backendApi.ts` parses error responses and throws as Error with message
- **Safe fetch functions:** Public data fetchers (`getPublicAnnouncement`, `getLatestPublicChangelog`, `getPublicChangelogEntries`) silently return null/empty on error
- **Image cache failures:** Non-fatal; images fall back to backend URL if no data URL available

## Cross-Cutting Concerns

**Logging:** Structured logging via Go `log/slog` (text format for dev, JSON for production via `glog.Init(false)`). Request-level logging via middleware captures method, path, status, latency, IP, userID.

**Validation:** Handler-level input validation (required fields, numeric ranges). Service-level validation (string length, category/status enums, bcrypt verification).

**Authentication:** JWT with 30-day expiry. Stored in localStorage on client. Token sent as Bearer header for all authenticated requests, or as `?token=` query param for image GET requests (served as `<img src>` which cannot inject headers).

---

*Architecture analysis: 2026-05-24*
