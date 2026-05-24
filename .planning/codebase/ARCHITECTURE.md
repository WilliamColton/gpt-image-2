<!-- refreshed: 2026-05-24 -->
# Architecture

**Analysis Date:** 2026-05-24

## System Overview

```text
┌─────────────────────────────────────────────────────────────┐
│                      Browser / PWA Shell                    │
│  `index.html` → `src/main.tsx` → home or admin React tree    │
├──────────────────────┬──────────────────────────────────────┤
│ Home workspace        │ Admin workspace                      │
│ `src/App.tsx`         │ `src/admin/AdminPage.tsx`            │
│ `src/components/*`    │ `src/admin/AdminDashboard.tsx`       │
└──────────┬───────────┴──────────────────┬───────────────────┘
           │                              │
           ▼                              ▼
┌─────────────────────────────────────────────────────────────┐
│              Client state, API wrappers, local caches        │
│ `src/store.ts`, `src/lib/backendApi.ts`,                    │
│ `src/admin/adminApi.ts`, `src/lib/db.ts`                    │
└────────────────────────────┬────────────────────────────────┘
                             │ REST + multipart + SSE
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                    Go Gin HTTP API                          │
│ `backend-go/main.go`, `backend-go/middleware/*`,            │
│ `backend-go/handler/*`                                      │
└────────────────────────────┬────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                    Backend service layer                    │
│ `backend-go/service/auth.go`, `task.go`, `image.go`,        │
│ `openai.go`, `queue.go`, `billing.go`, `analytics.go`       │
└───────────────┬────────────────────────────┬────────────────┘
                │                            │
                ▼                            ▼
┌───────────────────────────────┐  ┌──────────────────────────┐
│ SQLite + filesystem runtime   │  │ OpenAI-compatible APIs   │
│ `backend-go/database/*`       │  │ `config.ApiEndpoint` pool│
│ `backend-go/data/app.sqlite`  │  │ `backend-go/config/*`    │
│ `backend-go/upload/<user>/`   │  │ `backend-go/service/*`   │
└───────────────────────────────┘  └──────────────────────────┘
```

## Component Responsibilities

| Component | Responsibility | File |
|-----------|----------------|------|
| Vite entry | Mounts React, installs mobile viewport guards, registers/unregisters service worker, chooses home vs admin route by `window.location.pathname`. | `src/main.tsx` |
| Home app shell | Initializes store, applies theme, opens changelog, renders header/search/grid/input/modals. | `src/App.tsx` |
| Admin shell | Gates `/admin` behind stored admin token and lazy-loads the dashboard. | `src/admin/AdminPage.tsx` |
| Admin dashboard | Owns admin tabs for users, codes, API endpoints, pricing, analytics, announcements, feedback, changelog, invites. | `src/admin/AdminDashboard.tsx` |
| Home components | Render task cards, grid selection, input bar, image detail/lightbox, settings, masks, auth, announcements, changelog, feedback. | `src/components/*.tsx` |
| UI primitives | Wrap Radix/shadcn-style primitives and shared visual components. | `src/components/ui/*.tsx` |
| Client state/actions | Centralizes Zustand state, persistent settings/session, image memory cache, task submission, SSE/polling, local/remote task mutations. | `src/store.ts` |
| Public API client | Encapsulates authenticated user API calls, token storage, task/image endpoints, public config, SSE stream parsing. | `src/lib/backendApi.ts` |
| Admin API client | Encapsulates admin token storage and `/api/admin/*` calls. | `src/admin/adminApi.ts` |
| Browser image store | Stores image data URLs in IndexedDB and hashes data URLs for client-side deduplication. | `src/lib/db.ts` |
| Backend bootstrap/router | Loads config, initializes logger/runtime dirs/database, wires middleware and all routes, starts Gin server. | `backend-go/main.go` |
| Auth/admin middleware | Verifies JWTs, loads active users, places `AuthUser` and `userID` in Gin context, protects admin endpoints. | `backend-go/middleware/middleware.go` |
| Request logger | Logs method/path/status/latency/IP/user ID after each Gin request. | `backend-go/middleware/logger.go` |
| HTTP handlers | Bind/validate request payloads, read path/query params, call services, return JSON/files/SSE. | `backend-go/handler/*.go` |
| Domain services | Own auth, quota, task persistence, image filesystem operations, OpenAI calls, endpoint limiting, billing, analytics, announcements, changelog, feedback. | `backend-go/service/*.go` |
| Database layer | Owns GORM connection, AutoMigrate, bootstrap admin/announcement rows, and persisted model definitions. | `backend-go/database/database.go`, `backend-go/database/models.go` |
| Runtime config | Loads `config.json`, exposes public config values, protects endpoint pool with mutexes, persists endpoint/pricing/invite config changes. | `backend-go/config/config.go` |
| Utilities | Generate IDs, hash/encrypt API keys, and constrain upload paths to the upload root. | `backend-go/util/*.go` |
| PWA shell cache | Caches static shell assets and bypasses `/api/*` requests. | `public/sw.js` |

## Pattern Overview

**Overall:** Single-page React application plus layered Go API monolith.

**Key Characteristics:**
- The frontend uses a composition shell: `src/main.tsx` selects `src/App.tsx` or `src/admin/AdminPage.tsx`; feature screens compose components from `src/components/`, `src/admin/`, and `src/components/ui/`.
- The frontend uses a centralized action store: business actions such as submission, upload, reuse, deletion, polling, and SSE updates live in `src/store.ts`; components call store actions instead of owning workflow state.
- The backend uses explicit layers: `backend-go/main.go` declares routes, `backend-go/handler/*.go` handles HTTP concerns, `backend-go/service/*.go` handles domain logic, and `backend-go/database/*.go` handles GORM models/connection.
- Long-running image generation is asynchronous: `backend-go/handler/generate.go` creates a queued task, starts a goroutine, updates persisted task state, and `backend-go/handler/tasks.go` streams task status by polling SQLite.
- External image generation uses an endpoint pool with priority, failover, per-endpoint concurrency limits, and cost attribution in `backend-go/service/openai.go` and `backend-go/service/queue.go`.

## Layers

**Browser shell and routing:**
- Purpose: Mount the React app, choose home or admin UI, manage service worker lifecycle, and apply global CSS.
- Location: `index.html`, `src/main.tsx`, `src/index.css`, `public/sw.js`, `public/manifest.webmanifest`
- Contains: HTML root, React root creation, service worker registration, PWA metadata, Tailwind base/theme CSS.
- Depends on: `react`, `react-dom`, `src/App.tsx`, `src/admin/AdminPage.tsx`, `src/lib/viewport.ts`.
- Used by: Browser navigation to `/` and `/admin`.

**Home UI components:**
- Purpose: Render the image generation workspace, task list, image viewers, mask editor, settings/auth dialogs, announcements, changelog, and feedback UI.
- Location: `src/App.tsx`, `src/components/*.tsx`
- Contains: Presentational and interaction-heavy React components such as `src/components/InputBar.tsx`, `src/components/TaskGrid.tsx`, `src/components/TaskCard.tsx`, `src/components/DetailModal.tsx`, `src/components/MaskEditorModal.tsx`.
- Depends on: `src/store.ts`, `src/types.ts`, `src/lib/*.ts`, `src/components/ui/*.tsx`.
- Used by: `src/App.tsx`.

**Admin UI components:**
- Purpose: Render management screens for users, redemption codes, endpoints, pricing, billing analytics, announcement, feedback, changelog, password reset, and invite settings.
- Location: `src/admin/AdminPage.tsx`, `src/admin/AdminLogin.tsx`, `src/admin/AdminDashboard.tsx`, `src/admin/adminApi.ts`, `src/admin/moneyFormat.ts`
- Contains: Admin route gate, login form, dashboard tabs, admin-specific API wrapper and money formatting helpers.
- Depends on: `src/store.ts` for theme/toasts, `src/admin/adminApi.ts` for HTTP, `src/components/ui/*.tsx` for controls.
- Used by: `src/main.tsx` when path is `/admin` or `/admin/*`.

**Shared client state and workflows:**
- Purpose: Own application state, persisted user settings/session, image cache, task lifecycle, SSE/polling fallback, and high-level operations.
- Location: `src/store.ts`
- Contains: Zustand state shape, `imageCache`, `activeStreams`, polling timer, `initStore`, `bootstrapBackendSession`, `submitTask`, `executeTask`, `reuseConfig`, `editOutputs`, `removeTask`, `addImageFromFile`.
- Depends on: `src/lib/backendApi.ts`, `src/lib/db.ts`, `src/lib/canvasImage.ts`, `src/lib/mask.ts`, `src/lib/size.ts`, `src/types.ts`.
- Used by: `src/App.tsx`, `src/components/*.tsx`, `src/admin/*.tsx`.

**Client API/data utilities:**
- Purpose: Keep browser storage, network calls, image manipulation, masking, sizing, clipboard, viewport, and parameter display reusable.
- Location: `src/lib/*.ts`, `src/lib/*.tsx`, `src/admin/adminApi.ts`
- Contains: `fetch` wrappers, auth/admin token helpers, IndexedDB helpers, image canvas helpers, mask preprocessing, size normalization, viewport transforms, clipboard helpers.
- Depends on: Browser APIs (`fetch`, `localStorage`, `indexedDB`, canvas, clipboard), `src/types.ts`.
- Used by: `src/store.ts`, `src/components/*.tsx`, `src/admin/*.tsx`, `vite.config.ts`.

**HTTP router and middleware:**
- Purpose: Register all backend routes, apply CORS/request logging/auth/admin middleware, configure multipart limits.
- Location: `backend-go/main.go`, `backend-go/middleware/*.go`
- Contains: Gin router groups for `/api/auth`, `/api/config`, `/api/images`, `/api/tasks`, `/api/generate`, `/api/edit`, `/api/feedback`, `/api/admin/*`.
- Depends on: `backend-go/config`, `backend-go/database`, `backend-go/handler`, `backend-go/middleware`, `github.com/gin-gonic/gin`, `github.com/gin-contrib/cors`.
- Used by: Backend process started from `backend-go/main.go`.

**Handler layer:**
- Purpose: Translate HTTP requests/responses into service calls; handlers should not own persistence conversions.
- Location: `backend-go/handler/*.go`
- Contains: Auth handlers, image upload/download/delete, task list/update/delete/SSE, generation dispatch, config, admin, announcement, changelog, feedback.
- Depends on: `backend-go/service`, `backend-go/middleware`, `backend-go/config`, `github.com/gin-gonic/gin`.
- Used by: Route declarations in `backend-go/main.go`.

**Service layer:**
- Purpose: Own domain rules, persistence workflows, external API calls, async task transitions, quota logic, billing, analytics, endpoint limiting, and filesystem image storage.
- Location: `backend-go/service/*.go`
- Contains: `service.TaskRecord` conversions, quota transaction, OpenAI-compatible client calls, endpoint failover, image save/read/delete, JWT/password auth, invitation logic, billing rows, analytics aggregation, announcement/changelog/feedback rules.
- Depends on: `backend-go/database`, `backend-go/config`, `backend-go/util`, OpenAI SDK, GORM.
- Used by: `backend-go/handler/*.go`, `backend-go/middleware/middleware.go`.

**Persistence and runtime files:**
- Purpose: Store durable application records and image files.
- Location: `backend-go/database/*.go`, `backend-go/data/app.sqlite`, `backend-go/upload/<userID>/`
- Contains: GORM models for users, redemption codes, images, tasks, announcements, feedback, changelog entries, billing records; SQLite database; uploaded/generated image files.
- Depends on: SQLite/GORM and filesystem access.
- Used by: `backend-go/service/*.go`.

**External provider integration:**
- Purpose: Call OpenAI-compatible image generation/edit APIs through configured endpoint pool.
- Location: `backend-go/service/openai.go`, `backend-go/service/queue.go`, `backend-go/config/config.go`
- Contains: `config.ApiEndpoint`, OpenAI client creation, failover loop, endpoint limiters, Codex CLI compatibility prompt, generated image attribution.
- Depends on: `github.com/openai/openai-go/v3`, `backend-go/config`.
- Used by: `backend-go/handler/generate.go` through `service.CallImagesGenerations*` and `service.CallImagesEdits*`.

## Data Flow

### Primary Request Path

1. Browser loads `index.html`; `src/main.tsx:23` detects `/admin` routes, otherwise mounts `src/App.tsx:35`.
2. `src/App.tsx:27` calls `initStore()` from `src/store.ts:518` to load public announcement/changelog/config and bootstrap an existing backend session.
3. User submits from `src/components/InputBar.tsx:199` via `submitTask()` in `src/store.ts:554`.
4. `src/store.ts:571` validates mask ordering/coverage with `src/lib/mask.ts` and `src/lib/canvasImage.ts`, then `src/store.ts:597` normalizes task params with `src/lib/size.ts`.
5. `src/store.ts:620` creates an optimistic queued `TaskRecord` in Zustand so the UI updates before uploads complete.
6. `src/store.ts:640` uploads mask/input images through `uploadImage()` in `src/lib/backendApi.ts:189`; the backend route `backend-go/main.go:71` calls `handler.ImagesUpload` in `backend-go/handler/images.go:15`, which saves files through `service.SaveImageBuffer` in `backend-go/service/image.go:42`.
7. `src/store.ts:696` calls `submitGenerateTask()` or `submitEditTask()` in `src/lib/backendApi.ts:226`; routes in `backend-go/main.go:83` dispatch to `handler.GenerateImage` in `backend-go/handler/generate.go:32`.
8. `backend-go/handler/generate.go:67` calls `service.CheckQuotaAndCreateTask` in `backend-go/service/task.go:153`, which transactionally checks pending quota and upserts the queued task.
9. `backend-go/handler/generate.go:79` starts `executeImageGeneration` as a goroutine; `backend-go/handler/generate.go:87` loads configured endpoints from `config.GetEndpointPool` in `backend-go/config/config.go:28`.
10. `backend-go/handler/generate.go:93` reads input/mask images through `service.ReadImageDataURLForUser` in `backend-go/service/image.go:101`, then calls generation/edit service functions in `backend-go/service/openai.go:152`, `backend-go/service/openai.go:216`, or concurrent variants in `backend-go/service/openai.go:223` and `backend-go/service/openai.go:324`.
11. `backend-go/service/openai.go:62` runs endpoint failover and `backend-go/service/queue.go:77` acquires per-endpoint concurrency slots; `backend-go/handler/generate.go:130` marks the task running after a slot is acquired.
12. `backend-go/handler/generate.go:157` saves generated images through `service.SaveDataURLImage` in `backend-go/service/image.go:83`; `backend-go/handler/generate.go:169` records per-image billing rows through `service.RecordBillingForSuccessfulImages` in `backend-go/service/billing.go:32`.
13. `backend-go/handler/generate.go:185` sets `OutputImages`, metadata, status, finish time, and elapsed time, then persists through `service.UpsertTask` in `backend-go/service/task.go:132`.
14. `src/store.ts:711` opens an SSE status stream with `streamTaskStatus()` in `src/lib/backendApi.ts:247`; the backend route `backend-go/main.go:78` calls `handler.TaskStream` in `backend-go/handler/tasks.go:121`.
15. `backend-go/handler/tasks.go:161` polls SQLite every second and emits `task-update` events until `done` or `error`; `src/lib/backendApi.ts:277` parses `data:` lines and `src/store.ts:675` merges completed tasks into Zustand.

### Admin Configuration and Analytics Path

1. `src/main.tsx:23` routes `/admin` to lazy-loaded `src/admin/AdminPage.tsx:26`.
2. `src/admin/AdminPage.tsx:8` checks admin token presence through `isAdminLoggedIn()` in `src/admin/adminApi.ts:107`.
3. `src/admin/AdminLogin.tsx:25` submits admin credentials through `adminLogin()` in `src/admin/adminApi.ts:73`; `backend-go/main.go:89` routes `/api/admin/login` to `handler.AdminLogin` in `backend-go/handler/admin.go:17`.
4. `src/admin/AdminDashboard.tsx:122` and related loaders call typed admin wrappers in `src/admin/adminApi.ts` for users, codes, pricing, endpoints, analytics, announcements, feedback, changelog, invites, and password reset.
5. `backend-go/main.go:91` protects admin routes with `middleware.AdminMiddleware` from `backend-go/middleware/middleware.go:60`.
6. Endpoint/pricing/invite updates flow through `backend-go/handler/admin.go:198`, `backend-go/handler/admin.go:242`, and `backend-go/handler/admin.go:399`, then mutate/persist runtime config through `backend-go/config/config.go:35`, `backend-go/config/config.go:134`, and `backend-go/config/config.go:316`.
7. Billing analytics routes in `backend-go/main.go:113` call handler functions in `backend-go/handler/admin.go:303`, which aggregate `billing_records` through `backend-go/service/analytics.go:105`, `backend-go/service/analytics.go:140`, `backend-go/service/analytics.go:189`, and `backend-go/service/analytics.go:238`.

### Authentication and Session Path

1. Public user auth calls are defined in `src/lib/backendApi.ts:56`, `src/lib/backendApi.ts:65`, `src/lib/backendApi.ts:81`, and `src/lib/backendApi.ts:90`.
2. `backend-go/main.go:55` registers `/api/auth/*` routes; `backend-go/handler/auth.go:18` handles code login, `backend-go/handler/auth.go:76` handles password login, and `backend-go/handler/auth.go:97` handles registration.
3. `backend-go/service/auth.go:20` signs JWTs; `backend-go/service/auth.go:30` verifies them; `backend-go/middleware/middleware.go:14` loads active users and injects `AuthUser` into Gin context.
4. `src/lib/backendApi.ts:18` stores the user token in `localStorage`; `src/store.ts:443` validates a stored token with `getMe()` and reloads backend tasks.
5. Admin auth uses a separate `localStorage` key in `src/admin/adminApi.ts:4` and the admin-only JWT path in `backend-go/handler/admin.go:17`.

### Image Storage and Cache Path

1. File input/paste/drop flows call `addImageFromFile()` in `src/store.ts:900`, which hashes the data URL with `hashDataUrl()` in `src/lib/db.ts:48` and places it in the in-memory `imageCache` at `src/store.ts:44`.
2. Before generation, `src/store.ts:657` uploads local data URLs and replaces local IDs with backend image IDs returned by `src/lib/backendApi.ts:189`.
3. The backend deduplicates by user and SHA-256 in `service.SaveImageBuffer()` at `backend-go/service/image.go:42`, stores files under `backend-go/upload/<userID>/`, and stores metadata in the `images` table from `backend-go/database/models.go:33`.
4. Render paths call `ensureImageCached()` in `src/store.ts:132`; it prefers in-memory data URLs, then IndexedDB via `src/lib/db.ts:38`, then `/api/images/:id` from `src/lib/backendApi.ts:201`.
5. The image download route `backend-go/main.go:71` uses `middleware.AuthMiddleware` and `handler.ImagesGet` in `backend-go/handler/images.go:52`, which resolves paths safely through `backend-go/util/paths.go:29`.

**State Management:**
- Use Zustand in `src/store.ts` for all shared frontend state: auth user, settings, prompt, input images, mask draft, task list, search/filter, selection, modal IDs, announcements, changelog, toasts, and confirm dialogs.
- Persist only durable client preferences/session fragments through `persist()` in `src/store.ts:291`: settings, auth user, params, dismissed Codex prompts, seen announcement timestamp, and dismissed changelog keys.
- Keep image bytes out of persisted Zustand; use `imageCache` in `src/store.ts:44`, IndexedDB in `src/lib/db.ts`, backend files in `backend-go/upload/`, and backend metadata in `backend-go/database/models.go`.
- Treat backend SQLite as authoritative for authenticated task history; `src/store.ts:443` reloads tasks from `getTasks()` after session bootstrap.
- Backend global runtime state lives in `config.App` and `config.ApiEndpoints` in `backend-go/config/config.go`, `database.DB` in `backend-go/database/database.go`, endpoint limiters in `backend-go/service/queue.go`, and the default slog logger in `backend-go/log/log.go`.

## Key Abstractions

**TaskRecord:**
- Purpose: Represents a generation/edit job across UI, API, service, and database layers.
- Examples: `src/types.ts:64`, `backend-go/service/models.go:103`, `backend-go/database/models.go:46`, `backend-go/service/task.go:14`
- Pattern: Frontend uses typed `TaskRecord`; backend service converts JSON-friendly `TaskRecord` to GORM `database.Task` with JSON-encoded params/image arrays and metadata.

**TaskParams:**
- Purpose: Carries image generation parameters (`size`, `quality`, `output_format`, `output_compression`, `moderation`, `n`).
- Examples: `src/types.ts:27`, `backend-go/service/models.go:94`, `backend-go/service/openai.go:126`, `src/store.ts:597`
- Pattern: UI normalizes unsupported values before submission; backend passes params to OpenAI-compatible SDK calls and captures actual params returned by the provider.

**InputImage / StoredImage / Image:**
- Purpose: Separate transient browser input previews, IndexedDB records, and backend-persisted image metadata/files.
- Examples: `src/types.ts:47`, `src/types.ts:92`, `backend-go/service/models.go:83`, `backend-go/database/models.go:33`, `backend-go/service/image.go:42`
- Pattern: Data URLs are local preview/cache payloads; backend image IDs identify uploaded/generated/mask files and are stored in task image ID arrays.

**AuthUser / User / AdminUser:**
- Purpose: Represent authenticated user state, persisted user records, and admin list rows with quota/status fields.
- Examples: `src/lib/backendApi.ts:6`, `backend-go/service/models.go:35`, `backend-go/database/models.go:3`, `backend-go/service/auth.go:47`, `src/admin/adminApi.ts:6`
- Pattern: JWT `sub` and `role` authorize requests; middleware reloads the persisted user and rejects disabled accounts.

**ApiEndpoint:**
- Purpose: Represents an OpenAI-compatible backend endpoint with API key, priority, max concurrency, and per-image cost.
- Examples: `backend-go/config/config.go:12`, `src/admin/adminApi.ts:27`, `backend-go/service/openai.go:62`, `backend-go/service/queue.go:11`
- Pattern: Admin config updates replace a sorted endpoint pool; generation code acquires a limiter slot, tries endpoints in priority order, and stamps endpoint/cost attribution on generated images.

**BillingRecord:**
- Purpose: Immutable per-successful-image accounting snapshot.
- Examples: `backend-go/database/models.go:106`, `backend-go/service/billing.go:10`, `backend-go/service/analytics.go:105`, `backend-go/handler/admin.go:303`
- Pattern: Generation writes one row per saved output image with endpoint, sale, cost, revenue, profit, and user label snapshots; admin analytics aggregate these rows by range.

**Announcement / Changelog / Feedback:**
- Purpose: Support public announcement/changelog display and user bug/feature feedback with admin management.
- Examples: `src/types.ts:101`, `src/types.ts:110`, `src/types.ts:128`, `backend-go/database/models.go:70`, `backend-go/database/models.go:79`, `backend-go/database/models.go:93`, `backend-go/service/announcement.go`, `backend-go/service/changelog.go`, `backend-go/service/feedback.go`
- Pattern: Public read endpoints hydrate home UI at startup; admin endpoints create/update/status-change records through services.

**Modal and escape stack:**
- Purpose: Coordinate overlay visibility and close only the topmost modal on Escape.
- Examples: `src/store.ts:253`, `src/hooks/useCloseOnEscape.ts:7`, `src/components/ConfirmDialog.tsx`, `src/components/MaskEditorModal.tsx`
- Pattern: Global modal IDs/booleans live in Zustand when shared; local modal state is acceptable for component-owned overlays; `useCloseOnEscape` handles stacked Escape behavior.

## Entry Points

**Browser app:**
- Location: `index.html`
- Triggers: Browser loads the Vite-built SPA.
- Responsibilities: Provides `#root`, PWA metadata links, and module script to `/src/main.tsx`.

**React runtime:**
- Location: `src/main.tsx`
- Triggers: Vite module load from `index.html`.
- Responsibilities: Installs mobile viewport guards, handles service worker lifecycle, selects home vs admin route, and mounts React.

**Home workspace:**
- Location: `src/App.tsx`
- Triggers: `src/main.tsx` when path is not `/admin`.
- Responsibilities: Initializes store, applies theme, listens for image drag prevention, displays the main workspace and global modals.

**Admin workspace:**
- Location: `src/admin/AdminPage.tsx`
- Triggers: `src/main.tsx` when path is `/admin` or `/admin/*`.
- Responsibilities: Checks admin token, renders admin login or dashboard, applies theme.

**Backend service:**
- Location: `backend-go/main.go`
- Triggers: Go process start in the `backend-go` module.
- Responsibilities: Loads config, initializes logging/directories/database, registers middleware/routes, and runs Gin on `config.App.Port`.

**Service worker:**
- Location: `public/sw.js`
- Triggers: `src/main.tsx:9` registers it in production.
- Responsibilities: Cache app shell/static GET responses, serve cached `index.html` for navigations, bypass `/api/*`.

**Vite config:**
- Location: `vite.config.ts`
- Triggers: `npm run dev`, `npm run build`, `npm run preview`.
- Responsibilities: Configure React plugin, relative base, dev proxy normalization from `dev-proxy.config.json`, and compile-time `__DEV_PROXY_CONFIG__` define.

## Architectural Constraints

- **Threading:** Frontend code in `src/*.tsx` and `src/store.ts` runs on the browser event loop; backend HTTP handlers in `backend-go/handler/*.go` run concurrently per request; image generation runs in goroutines started at `backend-go/handler/generate.go:79`; concurrent multi-image provider calls use goroutines in `backend-go/service/openai.go:223` and `backend-go/service/openai.go:324`.
- **SQLite concurrency:** `backend-go/database/database.go:20` enables WAL and foreign keys; `backend-go/database/database.go:25` sets `SetMaxOpenConns(1)`, so backend DB access is serialized through one open SQLite connection.
- **Global state:** `config.App` in `backend-go/config/config.go:81`, `database.DB` in `backend-go/database/database.go:15`, `config.ApiEndpoints` in `backend-go/config/config.go:20`, endpoint limiters in `backend-go/service/queue.go:15`, `imageCache`/`activeStreams` in `src/store.ts:44` and `src/store.ts:70`, and `escStack` in `src/hooks/useCloseOnEscape.ts:7` are module-level state.
- **Config secrets:** `backend-go/config.json` exists as ignored runtime config and contains backend configuration/secrets; never read or commit its contents. Use `backend-go/config/config.go` to understand schema/defaults.
- **Upload path safety:** All backend file reads should resolve relative upload paths through `backend-go/util/paths.go:29`; do not join untrusted paths directly under `backend-go/upload/`.
- **Authentication tokens:** User tokens are stored with key `gpt-image-playground-token` in `src/lib/backendApi.ts:4`; admin tokens are stored with key `gpt-image-playground-admin-token` in `src/admin/adminApi.ts:4`; do not mix user and admin API clients.
- **Circular imports:** No circular dependency chain is detected in the scanned source tree; preserve the one-way flow `components/admin → store/API/lib → backendApi/adminApi` on the frontend and `main → handler/middleware → service → database/config/util` on the backend.
- **Route ownership:** Backend routes are centralized in `backend-go/main.go`; adding a handler without adding a route in `backend-go/main.go` leaves it unreachable.
- **PWA API bypass:** `public/sw.js:27` bypasses paths beginning with `/api/`; new API routes should stay under `/api/` to avoid static shell caching.
- **Agent worktrees:** `.claude/worktrees/` contains agent scratch worktrees, not product source; edit root `src/` and `backend-go/` paths instead.

## Anti-Patterns

### Handler-to-database shortcut

**What happens:** Persistence and GORM model conversion are concentrated in `backend-go/service/*.go` and `backend-go/database/*.go`, while handlers in `backend-go/handler/*.go` bind HTTP input and call services.
**Why it's wrong:** Writing new `database.DB` queries inside handlers couples HTTP payload shape to database schema and bypasses shared conversions such as `toTaskModel()` in `backend-go/service/task.go:14` and auth helpers in `backend-go/service/auth.go`.
**Do this instead:** Add a service function in the relevant file under `backend-go/service/` and call it from a small handler in `backend-go/handler/*.go`; register the route in `backend-go/main.go`.

### Component-level HTTP scattering

**What happens:** HTTP details are centralized in `src/lib/backendApi.ts` for user/public APIs and `src/admin/adminApi.ts` for admin APIs; components such as `src/components/InputBar.tsx` and `src/admin/AdminDashboard.tsx` call store/API functions.
**Why it's wrong:** Adding raw `fetch` calls in feature components duplicates token handling, error extraction, URL construction, and cache behavior from `src/lib/backendApi.ts:34` and `src/admin/adminApi.ts:51`.
**Do this instead:** Add a typed wrapper to `src/lib/backendApi.ts` or `src/admin/adminApi.ts`, then invoke it from `src/store.ts` for shared workflows or from the feature component for local admin/UI actions.

### Local-only task mutation

**What happens:** `updateTaskLocal()` in `src/store.ts:755` changes only Zustand state, while exported `updateTaskInStore()` in `src/store.ts:759` also persists the task through `putRemoteTask()` in `src/lib/backendApi.ts:210`.
**Why it's wrong:** User-visible edits such as favorites or metadata changes disappear after session bootstrap if they only update local state.
**Do this instead:** Use `updateTaskInStore()` from components such as `src/components/TaskCard.tsx`; reserve `updateTaskLocal()` for internal transient state changes inside `src/store.ts`.

### Editing generated/runtime directories

**What happens:** Build output, runtime data, uploaded images, local config, and agent worktrees exist under `dist/`, `backend-go/data/`, `backend-go/upload/`, `backend-go/config.json`, and `.claude/worktrees/`.
**Why it's wrong:** These paths are generated or local-only and are ignored by `.gitignore`; code changes there do not update the application source of truth.
**Do this instead:** Edit source/config templates under `src/`, `backend-go/`, `public/`, `vite.config.ts`, `tailwind.config.js`, `tsconfig.json`, and package/module manifests.

## Error Handling

**Strategy:** Validate at the edge, return JSON errors for HTTP failures, persist async task failures, and surface frontend failures through thrown `Error` objects and toasts.

**Patterns:**
- Frontend API wrappers in `src/lib/backendApi.ts:34` and `src/admin/adminApi.ts:51` parse `{ error }` or `{ message }` response bodies and throw `Error` instances.
- Store workflows in `src/store.ts` catch backend/session/upload/SSE errors and call `showToast()` or mark tasks as `error`.
- Backend handlers validate request bodies with `c.ShouldBindJSON()` or explicit checks in files such as `backend-go/handler/auth.go`, `backend-go/handler/generate.go`, `backend-go/handler/admin.go`, and `backend-go/handler/feedback.go`.
- Backend services return Go `error` values with user-facing Chinese messages for validation and generic messages for persistence/provider failures.
- Async generation failures call `failTask()` in `backend-go/handler/generate.go:200`, which logs the failure and persists `status = "error"`, `error`, `finishedAt`, and `elapsed`.
- Request/operational logs use `log/slog` initialized in `backend-go/log/log.go`; request logging is applied in `backend-go/main.go:39`.

## Cross-Cutting Concerns

**Logging:** Backend uses `slog` via `backend-go/log/log.go` and middleware in `backend-go/middleware/logger.go`; frontend logs service worker registration failures to `console.error` in `src/main.tsx:13` and otherwise favors toasts via `src/store.ts:421`.

**Validation:** Frontend normalizes image params in `src/store.ts:597`, validates masks through `src/lib/mask.ts` and `src/lib/canvasImage.ts`, and limits input images in `src/components/InputBar.tsx:12`; backend validates auth, quota, route bodies, admin endpoint configs, changelog lengths, feedback categories/statuses, usernames/passwords, and path resolution in `backend-go/handler/*.go`, `backend-go/service/*.go`, and `backend-go/util/paths.go`.

**Authentication:** User APIs use JWT bearer tokens or `token` query fallback for image/SSE access in `backend-go/middleware/middleware.go:14`; admin APIs require admin JWTs through `backend-go/middleware/middleware.go:60`; frontend clients keep user/admin token storage separate in `src/lib/backendApi.ts` and `src/admin/adminApi.ts`.

**Authorization/data ownership:** Image and task routes always filter by `user.ID` via handlers and services such as `backend-go/handler/images.go:52`, `backend-go/handler/tasks.go:16`, `backend-go/service/image.go:91`, and `backend-go/service/task.go:118`; admin routes use admin middleware and service methods for global records.

**Configuration:** Public frontend-visible config is served by `backend-go/handler/config.go:12`; backend runtime config loads/persists through `backend-go/config/config.go`; Vite dev proxy config is normalized by `src/lib/devProxy.ts` and consumed by `vite.config.ts`.

**Persistence:** GORM models live in `backend-go/database/models.go`; migrations are automatic through `database.DB.AutoMigrate()` in `backend-go/database/database.go:28`; frontend uses IndexedDB for local image data URLs in `src/lib/db.ts` and localStorage through Zustand persist in `src/store.ts`.

**External integrations:** OpenAI-compatible provider calls are isolated to `backend-go/service/openai.go`; endpoint selection/limiters are isolated to `backend-go/service/queue.go`; API keys live in runtime config handled by `backend-go/config/config.go` and must not be exposed through frontend files.

---

*Architecture analysis: 2026-05-24*
